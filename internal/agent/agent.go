package agent

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"strings"
	"time"

	"agent-demo/internal/document"
	"agent-demo/internal/intent"
	"agent-demo/internal/llm"
	"agent-demo/internal/prompt"
	"agent-demo/internal/retriever"
	"agent-demo/internal/session"
)

type Agent struct {
	llmClient         llm.Client
	promptFactory     *prompt.Factory
	classifier        *intent.Classifier
	retriever         *retriever.KeywordRetriever
	sessionStore      session.Store
	maxHistoryMessage int
}

func NewAgent(llmClient llm.Client) (*Agent, error) {
	docs, err := document.LoadFromDir("docs")
	if err != nil {
		return nil, fmt.Errorf("load docs: %w", err)
	}

	chunks := document.SplitByParagraph(docs)

	return &Agent{
		llmClient:         llmClient,
		promptFactory:     prompt.NewFactory(),
		classifier:        intent.NewClassifier(),
		retriever:         retriever.NewKeywordRetriever(chunks),
		sessionStore:      session.NewMemoryStore(30),
		maxHistoryMessage: 8,
	}, nil
}

func (a *Agent) Chat(ctx context.Context, sessionID string, question string, requestType string) (string, string, string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		sessionID = newSessionID()
	}

	question = strings.TrimSpace(question)
	if question == "" {
		return "", "", sessionID, fmt.Errorf("question is empty")
	}

	history, err := a.sessionStore.Recent(ctx, sessionID, a.maxHistoryMessage)
	if err != nil {
		return "", "", sessionID, fmt.Errorf("load session history: %w", err)
	}
	log.Printf("history: %#v", history)

	intentType, err := a.resolveIntent(ctx, question, requestType)
	if err != nil {
		return "", "", sessionID, fmt.Errorf("resolve intent: %w", err)
	}

	promptText, err := a.buildPrompt(intentType, question, history)
	if err != nil {
		return "", "", sessionID, fmt.Errorf("build prompt: %w", err)
	}

	answer, err := a.llmClient.Generate(ctx, promptText)
	if err != nil {
		return "", "", sessionID, fmt.Errorf("generate answer: %w", err)
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
		return "", "", sessionID, fmt.Errorf("save session history: %w", err)
	}

	return answer, string(intentType), sessionID, nil
}

func (a *Agent) buildPrompt(intentType prompt.Type, question string, history []session.Message) (string, error) {
	if intentType == prompt.TypeChat {
		chunks := a.retriever.Retrieve(question, 3)
		if len(chunks) > 0 {
			return prompt.BuildRAGPrompt(question, chunks, history), nil
		}

		input := prompt.WithHistory(question, history)
		return a.promptFactory.Build(prompt.TypeChat, input)
	}

	return a.promptFactory.Build(intentType, question)
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
