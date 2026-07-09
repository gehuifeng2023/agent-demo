package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"agent-demo/internal/agent"
	"agent-demo/internal/document"
	"agent-demo/internal/handler"
	"agent-demo/internal/knowledge"
	"agent-demo/internal/llm"
	"agent-demo/internal/retriever"
)

func main() {
	llmClient, mode, err := newLLMClientFromEnv()
	if err != nil {
		log.Fatalf("create llm client: %v", err)
	}
	log.Printf("LLM mode: %s", mode)

	unifiedRetriever, err := newRetrieverFromDefaultKnowledge("knowledge_attachment/default/")
	if err != nil {
		log.Fatalf("load default knowledge: %v", err)
	}
	agentCore := agent.NewAgent(llmClient, unifiedRetriever)

	chatHandler := handler.NewChatHandler(agentCore)
	fileHandler := handler.NewFileHandler("knowledge_attachment/days", 20<<20, unifiedRetriever)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/chat", chatHandler)
	mux.HandleFunc("/api/v1/files/upload", fileHandler.Upload)

	addr := ":8080"

	log.Printf("agent-demo server started, addr=%s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
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

func newLLMClientFromEnv() (llm.Client, string, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_MODE")))

	switch mode {
	case "", "openai":
		client, err := llm.NewOpenAIClient()
		if err != nil {
			return nil, "openai", err
		}
		return client, "openai", nil
	case "mock":
		return llm.NewMockClient(), "mock", nil
	case "gemini":
		client, err := llm.NewGeminiClient()
		if err != nil {
			return nil, "gemini", err
		}
		return client, "gemini", nil
	default:
		return nil, "", fmt.Errorf("unsupported LLM_MODE: %s", mode)
	}
}
