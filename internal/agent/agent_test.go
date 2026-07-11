package agent

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agent-demo/internal/document"
	"agent-demo/internal/intent"
	"agent-demo/internal/knowledge"
	"agent-demo/internal/llm"
	"agent-demo/internal/model"
	"agent-demo/internal/prompt"
	"agent-demo/internal/retriever"
	"agent-demo/internal/session"
	"agent-demo/internal/tool"
	"agent-demo/internal/workflow"
)

func TestChatReturnsSourcesForRAGQuestion(t *testing.T) {
	agent := testAgent(testUnifiedRetriever())

	_, answerType, sessionID, sources, err := agent.Chat(context.Background(), model.ChatRequest{Question: "什么是 RAG？"})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if answerType != string(prompt.TypeChat) {
		t.Fatalf("expected chat type, got %q", answerType)
	}
	if sessionID == "" {
		t.Fatal("expected session id to be generated")
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	if sources[0].File != "docs/faq.md" {
		t.Fatalf("expected source file docs/faq.md, got %q", sources[0].File)
	}
	if sources[0].ChunkID != "docs/faq.md-0" {
		t.Fatalf("expected chunk id docs/faq.md-0, got %q", sources[0].ChunkID)
	}
}

func TestChatUsesSelectedKnowledgeBase(t *testing.T) {
	agent := testAgent(testUnifiedRetriever())

	_, _, _, sources, err := agent.Chat(context.Background(), model.ChatRequest{
		Question:         "什么是 RAG？",
		KnowledgeBaseIDs: []string{"default"},
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	if sources[0].File != "docs/faq.md" {
		t.Fatalf("expected docs/faq.md, got %q", sources[0].File)
	}
}

func TestChatIgnoresUnknownKnowledgeBase(t *testing.T) {
	agent := testAgent(testUnifiedRetriever())

	_, _, _, sources, err := agent.Chat(context.Background(), model.ChatRequest{
		Question:         "什么是 RAG？",
		KnowledgeBaseIDs: []string{"missing"},
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if len(sources) != 0 {
		t.Fatalf("expected no sources, got %d", len(sources))
	}
}

func TestChatUsesKnowledgeAddedAfterAgentCreation(t *testing.T) {
	unifiedRetriever := retriever.NewUnifiedRetriever()
	agent := testAgent(unifiedRetriever)

	unifiedRetriever.StoreFileChunks("file-1", []document.Chunk{
		{
			ID:       "uploads/file-1.txt-1",
			Source:   "uploads/file-1.txt",
			Content:  "GammaProject 支持动态扩充知识库。",
			Position: 1,
		},
	})

	_, _, _, sources, err := agent.Chat(context.Background(), model.ChatRequest{
		Question: "GammaProject 支持什么？",
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	if sources[0].File != "uploads/file-1.txt" {
		t.Fatalf("expected uploaded source, got %q", sources[0].File)
	}
}

func TestChatUsesConfiguredTopK(t *testing.T) {
	unifiedRetriever := retriever.NewUnifiedRetriever()
	unifiedRetriever.StoreFileChunks("file-1", []document.Chunk{
		{ID: "uploads/file-1.txt-1", Source: "uploads/file-1.txt", Content: "AlphaProject first match", Position: 1},
		{ID: "uploads/file-1.txt-2", Source: "uploads/file-1.txt", Content: "AlphaProject second match", Position: 2},
	})
	agent := NewAgentWithOptions(llm.NewMockClient(), unifiedRetriever, Options{TopK: 1})

	_, _, _, sources, err := agent.Chat(context.Background(), model.ChatRequest{
		Question: "AlphaProject",
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source from configured topK, got %d", len(sources))
	}
}

func TestChatInjectsToolContext(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "faq.md"), []byte("FileReaderTool content"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	registry := tool.NewRegistry()
	registry.Register(tool.FileReaderTool{RootDir: root})
	agent := NewAgentWithOptions(llm.NewMockClient(), retriever.NewUnifiedRetriever(), Options{
		ToolRegistry: registry,
		ToolsEnabled: true,
	})

	answer, _, _, sources, err := agent.Chat(context.Background(), model.ChatRequest{
		Question: "请读取 faq.md 并总结",
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if !strings.Contains(answer, "FileReaderTool content") {
		t.Fatalf("expected tool content in answer prompt, got %q", answer)
	}
	if len(sources) != 0 {
		t.Fatalf("expected no RAG sources, got %d", len(sources))
	}
}

func TestChatInjectsLogAnalyzerToolContext(t *testing.T) {
	registry := tool.NewRegistry()
	registry.Register(tool.LogAnalyzerTool{})
	agent := NewAgentWithOptions(llm.NewMockClient(), retriever.NewUnifiedRetriever(), Options{
		ToolRegistry: registry,
		ToolsEnabled: true,
	})

	answer, answerType, _, sources, err := agent.Chat(context.Background(), model.ChatRequest{
		Question: "帮我分析日志 request_id=abc trace_id=t1 status=502 upstream timeout",
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if answerType != string(prompt.TypeLogAnalysis) {
		t.Fatalf("expected log_analysis type, got %q", answerType)
	}
	for _, want := range []string{
		"工具：log_analyzer",
		`"error_type": "gateway_502"`,
		`"request_id": "abc"`,
		"网关或上游服务异常",
	} {
		if !strings.Contains(answer, want) {
			t.Fatalf("expected answer to contain %q, got %q", want, answer)
		}
	}
	if len(sources) != 0 {
		t.Fatalf("expected no RAG sources, got %d", len(sources))
	}
}

func TestChatInjectsHTTPToolContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "HTTP tool content")
	}))
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("parse test server URL: %v", err)
	}
	registry := tool.NewRegistry()
	registry.Register(tool.HTTPGetTool{Client: tool.NewHTTPClient([]string{request.URL.Hostname()}, 0)})
	agent := NewAgentWithOptions(llm.NewMockClient(), retriever.NewUnifiedRetriever(), Options{
		ToolRegistry: registry,
		ToolsEnabled: true,
	})

	answer, _, _, _, err := agent.Chat(context.Background(), model.ChatRequest{
		Question: fmt.Sprintf(`GET {"url":%q}`, server.URL),
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if !strings.Contains(answer, "工具：http_get") || !strings.Contains(answer, "HTTP tool content") {
		t.Fatalf("expected HTTP tool context in answer, got %q", answer)
	}
}

func TestChatInjectsExplicitWorkflowContextWithoutAutomaticToolRouting(t *testing.T) {
	workflowRegistry := workflow.NewRegistry()
	if err := workflowRegistry.Register(&workflow.Workflow{ID: "manual", Nodes: []workflow.Node{
		workflow.ToolNode{NameValue: "step", Tool: workflowTestTool{output: "workflow result"}, InputTemplate: "{{question}}", OutputKeyValue: "result"},
	}}); err != nil {
		t.Fatalf("register workflow: %v", err)
	}
	tools := tool.NewRegistry()
	tools.Register(failingTool{})
	agent := NewAgentWithOptions(llm.NewMockClient(), retriever.NewUnifiedRetriever(), Options{
		ToolRegistry:     tools,
		ToolsEnabled:     true,
		WorkflowRegistry: workflowRegistry,
	})

	answer, _, _, _, err := agent.Chat(context.Background(), model.ChatRequest{
		Question:   "请读取 faq.md",
		WorkflowID: "manual",
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if !strings.Contains(answer, "工作流：manual") || !strings.Contains(answer, "workflow result") {
		t.Fatalf("expected workflow context in answer, got %q", answer)
	}
}

func TestChatReturnsUnknownWorkflowError(t *testing.T) {
	agent := NewAgentWithOptions(llm.NewMockClient(), retriever.NewUnifiedRetriever(), Options{
		WorkflowRegistry: workflow.NewRegistry(),
	})
	_, _, _, _, err := agent.Chat(context.Background(), model.ChatRequest{
		Question:   "test",
		WorkflowID: "missing",
	})
	if err == nil || !strings.Contains(err.Error(), "workflow not found: missing") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestChatDoesNotTriggerToolForNormalQuestion(t *testing.T) {
	registry := tool.NewRegistry()
	registry.Register(failingTool{})
	agent := NewAgentWithOptions(llm.NewMockClient(), retriever.NewUnifiedRetriever(), Options{
		ToolRegistry: registry,
		ToolsEnabled: true,
	})

	_, _, _, _, err := agent.Chat(context.Background(), model.ChatRequest{
		Question: "什么是 RAG？",
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
}

func TestChatReturnsToolError(t *testing.T) {
	registry := tool.NewRegistry()
	registry.Register(tool.FileReaderTool{RootDir: t.TempDir()})
	agent := NewAgentWithOptions(llm.NewMockClient(), retriever.NewUnifiedRetriever(), Options{
		ToolRegistry: registry,
		ToolsEnabled: true,
	})

	_, _, _, _, err := agent.Chat(context.Background(), model.ChatRequest{
		Question: "请读取 missing.md",
	})
	if err == nil {
		t.Fatal("expected tool error")
	}
	if !strings.Contains(err.Error(), "execute tool file_reader") {
		t.Fatalf("expected tool error context, got %v", err)
	}
}

func testAgent(unifiedRetriever *retriever.UnifiedRetriever) *Agent {
	return &Agent{
		llmClient:         llm.NewMockClient(),
		promptFactory:     prompt.NewFactory(),
		classifier:        intent.NewClassifier(),
		retriever:         unifiedRetriever,
		sessionStore:      session.NewMemoryStore(30),
		maxHistoryMessage: 8,
		topK:              3,
	}
}

func testUnifiedRetriever() *retriever.UnifiedRetriever {
	unifiedRetriever := retriever.NewUnifiedRetriever()
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID: "default",
		Chunks: []document.Chunk{
			{
				ID:       "docs/faq.md-0",
				Source:   "docs/faq.md",
				Content:  "RAG 是 Retrieval-Augmented Generation 的缩写。它的核心思想是：先从知识库中检索相关内容，再让大模型基于这些内容回答问题。",
				Position: 0,
			},
		},
	})
	return unifiedRetriever
}

type failingTool struct{}

func (failingTool) Name() string { return "file_reader" }

func (failingTool) Description() string { return "failing tool" }

func (failingTool) Execute(context.Context, string) (string, error) {
	return "", os.ErrInvalid
}

type workflowTestTool struct {
	output string
}

func (t workflowTestTool) Name() string        { return "workflow_test" }
func (t workflowTestTool) Description() string { return "workflow test tool" }
func (t workflowTestTool) Execute(context.Context, string) (string, error) {
	return t.output, nil
}
