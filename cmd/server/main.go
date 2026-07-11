package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"agent-demo/internal/agent"
	"agent-demo/internal/config"
	"agent-demo/internal/document"
	"agent-demo/internal/handler"
	"agent-demo/internal/knowledge"
	"agent-demo/internal/llm"
	"agent-demo/internal/retriever"
	"agent-demo/internal/tool"
	"agent-demo/internal/workflow"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	llmClient, mode, err := newLLMClientFromConfig(cfg)
	if err != nil {
		log.Fatalf("create llm client: %v", err)
	}
	log.Printf("LLM mode: %s", mode)

	unifiedRetriever, err := newRetrieverFromDefaultKnowledge(cfg.Knowledge.RootDir)
	if err != nil {
		log.Fatalf("load default knowledge: %v", err)
	}
	toolRegistry := newToolRegistry(cfg)
	workflowRegistry, err := newWorkflowRegistry(cfg, toolRegistry)
	if err != nil {
		log.Fatalf("create workflow registry: %v", err)
	}
	agentCore := agent.NewAgentWithOptions(llmClient, unifiedRetriever, agent.Options{
		TopK:               cfg.RAG.TopK,
		SessionMaxMessages: cfg.Session.MaxMessages,
		MaxHistoryMessage:  cfg.Session.RecentLimit,
		ToolRegistry:       toolRegistry,
		ToolsEnabled:       cfg.ToolEnabled(),
		WorkflowRegistry:   workflowRegistry,
	})

	chatHandler := handler.NewChatHandler(agentCore)
	fileHandler := handler.NewFileHandler(cfg.Upload.Dir, cfg.UploadMaxBytes(), unifiedRetriever)
	knowledgeHandler := handler.NewKnowledgeHandlerWithTopK(unifiedRetriever, cfg.RAG.TopK)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/chat", chatHandler)
	mux.HandleFunc("/api/v1/files/upload", fileHandler.Upload)
	mux.HandleFunc("/api/v1/knowledge", knowledgeHandler.List)
	mux.HandleFunc("/api/v1/knowledge/retrieve", knowledgeHandler.Recall)

	log.Printf("agent-demo server started, addr=%s", cfg.Server.Addr)

	if err := http.ListenAndServe(cfg.Server.Addr, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func loadConfig() (*config.Config, error) {
	path := strings.TrimSpace(os.Getenv("CONFIG_PATH"))
	if path == "" {
		path = config.DefaultConfigPath
	}
	return config.Load(path)
}

func newRetrieverFromDefaultKnowledge(dir string) (*retriever.UnifiedRetriever, error) {
	docs, err := document.LoadFromDir(dir)
	if err != nil {
		return nil, fmt.Errorf("load docs: %w", err)
	}

	unifiedRetriever := retriever.NewUnifiedRetriever()
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID:     "default",
		Chunks: document.SplitByParagraph(docs),
	})
	return unifiedRetriever, nil
}

func newToolRegistry(cfg *config.Config) *tool.Registry {
	if cfg == nil || !cfg.ToolEnabled() {
		return nil
	}

	registry := tool.NewRegistry()
	registry.Register(tool.FileReaderTool{RootDir: cfg.ToolRootDir()})
	registry.Register(tool.LogAnalyzerTool{})
	if len(cfg.Tool.HTTPAllowedHosts) > 0 {
		httpClient := tool.NewHTTPClient(cfg.Tool.HTTPAllowedHosts, cfg.HTTPToolTimeout())
		registry.Register(tool.HTTPGetTool{Client: httpClient})
		registry.Register(tool.HTTPPostTool{Client: httpClient})
	}
	return registry
}

func newWorkflowRegistry(cfg *config.Config, tools *tool.Registry) (*workflow.Registry, error) {
	registry := workflow.NewRegistry()
	if cfg == nil || len(cfg.Workflow.Definitions) == 0 {
		return registry, nil
	}
	if tools == nil {
		return nil, fmt.Errorf("workflow definitions require tools to be enabled")
	}

	for _, definition := range cfg.Workflow.Definitions {
		nodes := make([]workflow.Node, 0, len(definition.Nodes))
		for _, nodeConfig := range definition.Nodes {
			selectedTool, ok := tools.Get(nodeConfig.Tool)
			if !ok {
				return nil, fmt.Errorf("workflow %q node %q: tool not found: %s", definition.ID, nodeConfig.Name, nodeConfig.Tool)
			}
			nodes = append(nodes, workflow.ToolNode{
				NameValue:      nodeConfig.Name,
				Tool:           selectedTool,
				InputTemplate:  nodeConfig.Input,
				OutputKeyValue: nodeConfig.OutputKey,
			})
		}
		if err := registry.Register(&workflow.Workflow{ID: definition.ID, Nodes: nodes}); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func newLLMClientFromConfig(cfg *config.Config) (llm.Client, string, error) {
	if cfg == nil {
		cfg = &config.Config{}
		cfg.ApplyDefaults()
	}

	mode := strings.ToLower(strings.TrimSpace(cfg.LLM.Mode))

	switch mode {
	case "", "openai":
		apiKey := firstNonEmpty(cfg.LLM.APIKey, os.Getenv("OPENAI_API_KEY"))
		model := firstNonEmpty(cfg.LLM.Model, os.Getenv("LLM_MODEL"))
		client, err := llm.NewOpenAIClientWithConfig(apiKey, model, cfg.LLM.BaseURL, cfg.LLMTimeout())
		if err != nil {
			return nil, "openai", err
		}
		return client, "openai", nil
	case "mock":
		return llm.NewMockClient(), "mock", nil
	case "gemini":
		apiKey := firstNonEmpty(cfg.LLM.APIKey, os.Getenv("GEMINI_API_KEY"))
		model := firstNonEmpty(cfg.LLM.Model, os.Getenv("LLM_MODEL"))
		client, err := llm.NewGeminiClientWithConfig(apiKey, model, cfg.LLM.BaseURL, cfg.LLMTimeout())
		if err != nil {
			return nil, "gemini", err
		}
		return client, "gemini", nil
	default:
		return nil, "", fmt.Errorf("unsupported LLM_MODE: %s", mode)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
