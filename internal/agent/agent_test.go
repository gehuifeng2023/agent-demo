package agent

import (
	"context"
	"testing"

	"agent-demo/internal/document"
	"agent-demo/internal/intent"
	"agent-demo/internal/knowledge"
	"agent-demo/internal/llm"
	"agent-demo/internal/model"
	"agent-demo/internal/prompt"
	"agent-demo/internal/retriever"
	"agent-demo/internal/session"
)

func TestChatReturnsSourcesForRAGQuestion(t *testing.T) {
	agent := &Agent{
		llmClient:         llm.NewMockClient(),
		promptFactory:     prompt.NewFactory(),
		classifier:        intent.NewClassifier(),
		retriever:         testUnifiedRetriever(),
		sessionStore:      session.NewMemoryStore(30),
		maxHistoryMessage: 8,
	}

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
	agent := &Agent{
		llmClient:         llm.NewMockClient(),
		promptFactory:     prompt.NewFactory(),
		classifier:        intent.NewClassifier(),
		retriever:         testUnifiedRetriever(),
		sessionStore:      session.NewMemoryStore(30),
		maxHistoryMessage: 8,
	}

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
	agent := &Agent{
		llmClient:         llm.NewMockClient(),
		promptFactory:     prompt.NewFactory(),
		classifier:        intent.NewClassifier(),
		retriever:         testUnifiedRetriever(),
		sessionStore:      session.NewMemoryStore(30),
		maxHistoryMessage: 8,
	}

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
