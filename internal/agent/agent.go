package agent

import (
	"agent-demo/internal/model"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"strings"
	"time"

	"agent-demo/internal/document"
	"agent-demo/internal/intent"
	"agent-demo/internal/knowledge"
	"agent-demo/internal/llm"
	"agent-demo/internal/prompt"
	"agent-demo/internal/retriever"
	"agent-demo/internal/session"
)

type Agent struct {
	llmClient         llm.Client
	promptFactory     *prompt.Factory
	classifier        *intent.Classifier
	retriever         *retriever.UnifiedRetriever
	sessionStore      session.Store
	maxHistoryMessage int
}

func NewAgent(llmClient llm.Client) (*Agent, error) {
	docs, err := document.LoadFromDir("docs")
	if err != nil {
		return nil, fmt.Errorf("load docs: %w", err)
	}

	chunks := document.SplitByParagraph(docs)
	unifiedRetriever := retriever.NewUnifiedRetriever()
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID:     "default",
		Chunks: chunks,
	})

	return &Agent{
		llmClient:         llmClient,
		promptFactory:     prompt.NewFactory(),
		classifier:        intent.NewClassifier(),
		retriever:         unifiedRetriever,
		sessionStore:      session.NewMemoryStore(30),
		maxHistoryMessage: 8,
	}, nil
}

func (a *Agent) StoreFileChunks(fileID string, chunks []document.Chunk) {
	if a.retriever == nil {
		return
	}

	a.retriever.StoreFileChunks(fileID, chunks)
}

func (a *Agent) Chat(ctx context.Context, req model.ChatRequest) (string, string, string, []model.Source, error) {
	sessionID := req.SessionID
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		sessionID = newSessionID()
	}

	question := req.Question
	question = strings.TrimSpace(question)
	if question == "" {
		return "", "", sessionID, nil, fmt.Errorf("question is empty")
	}

	history, err := a.sessionStore.Recent(ctx, sessionID, a.maxHistoryMessage)
	if err != nil {
		return "", "", sessionID, nil, fmt.Errorf("load session history: %w", err)
	}
	log.Printf("history: %#v", history)

	intentType, err := a.resolveIntent(ctx, question, req.Type)
	if err != nil {
		return "", "", sessionID, nil, fmt.Errorf("resolve intent: %w", err)
	}

	chunks := a.retriever.Retrieve(question, compactStrings(req.KnowledgeBaseIDs), compactStrings(req.FileIDs), 3)

	promptText, err := a.buildPrompt(intentType, question, history, chunks)
	if err != nil {
		return "", "", sessionID, buildSources(chunks), fmt.Errorf("build prompt: %w", err)
	}
	answer, err := a.llmClient.Generate(ctx, promptText)
	if err != nil {
		return "", "", sessionID, buildSources(chunks), fmt.Errorf("generate answer: %w", err)
	}

	now := time.Now()

	if err := a.sessionStore.Append(
		ctx,
		sessionID,
		session.Message{
			Role:      session.RoleUser,
			Content:   question,
			CreatedAt: now,
		},
		session.Message{
			Role:      session.RoleAssistant,
			Content:   answer,
			CreatedAt: now,
		},
	); err != nil {
		return "", "", sessionID, buildSources(chunks), fmt.Errorf("save session history: %w", err)
	}

	return answer, string(intentType), sessionID, buildSources(chunks), nil
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func (a *Agent) buildPrompt(intentType prompt.Type, question string, history []session.Message, chunks []document.Chunk) (string, error) {
	if intentType == prompt.TypeChat {
		if len(chunks) > 0 {
			return prompt.BuildRAGPrompt(question, chunks, history), nil
		}
		if looksLikeDocumentQuestion(question) {
			return prompt.BuildRAGPrompt(question, nil, history), nil
		}

		input := prompt.WithHistory(question, history)
		promptText, err := a.promptFactory.Build(prompt.TypeChat, input)
		return promptText, err
	}

	promptText, err := a.promptFactory.Build(intentType, question)
	return promptText, err
}

func buildSources(chunks []document.Chunk) []model.Source {
	result := make([]model.Source, 0, len(chunks))

	for _, chunk := range chunks {
		result = append(result, model.Source{
			File:     chunk.Source,
			ChunkID:  chunk.ID,
			Content:  chunk.Content,
			Position: chunk.Position,
		})
	}

	return result
}

func looksLikeDocumentQuestion(question string) bool {
	text := strings.ToLower(strings.TrimSpace(question))

	keywords := []string{
		"文档",
		"知识库",
		"资料",
		"根据",
		"这份",
		"这个项目",
		"说明书",
		"设计文档",
		"接口文档",
		"部署文档",
		"有没有提到",
		"是否支持",
	}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}

func (a *Agent) resolveIntent(ctx context.Context, question string, requestType string) (prompt.Type, error) {
	requestType = strings.TrimSpace(requestType)
	if requestType != "" {
		return prompt.Type(requestType), nil
	}

	return a.classifier.Classify(ctx, question)
}

func newSessionID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("s-%d", time.Now().UnixNano())
	}

	return fmt.Sprintf("s-%x", b)
}
