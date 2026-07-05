package agent

import (
	"agent-demo/internal/document"
	"agent-demo/internal/retriever"
	"context"
	"fmt"

	"agent-demo/internal/intent"
	"agent-demo/internal/llm"
	"agent-demo/internal/prompt"
)

type Agent struct {
	llmClient     llm.Client
	promptFactory *prompt.Factory
	classifier    *intent.Classifier
	retriever     *retriever.KeywordRetriever
}

func NewAgent(llmClient llm.Client) (*Agent, error) {
	docs, err := document.LoadFromDir("docs")
	if err != nil {
		return nil, fmt.Errorf("load docs: %w", err)
	}

	chunks := document.SplitByParagraph(docs)

	return &Agent{
		llmClient:     llmClient,
		promptFactory: prompt.NewFactory(),
		classifier:    intent.NewClassifier(),
		retriever:     retriever.NewKeywordRetriever(chunks),
	}, nil
}

func (a *Agent) Chat(ctx context.Context, question string, requestType string) (string, string, error) {
	if question == "" {
		return "", "", fmt.Errorf("question is empty")
	}

	intentType, err := a.resolveIntent(ctx, question, requestType)
	if err != nil {
		return "", "", fmt.Errorf("resolve intent: %w", err)
	}

	promptText, err := a.buildPrompt(intentType, question)
	if err != nil {
		return "", "", fmt.Errorf("build prompt: %w", err)
	}

	answer, err := a.llmClient.Generate(ctx, promptText)
	if err != nil {
		return "", "", fmt.Errorf("generate answer: %w", err)
	}

	return answer, string(intentType), nil
}

func (a *Agent) buildPrompt(intentType prompt.Type, question string) (string, error) {
	if intentType == prompt.TypeChat {
		chunks := a.retriever.Retrieve(question, 3)
		if len(chunks) > 0 {
			return prompt.BuildRAGPrompt(question, chunks), nil
		}
	}

	promptType := toPromptType(intentType)
	return a.promptFactory.Build(promptType, question)
}

func (a *Agent) resolveIntent(ctx context.Context, question string, requestType string) (prompt.Type, error) {
	if requestType != "" {
		return prompt.Type(requestType), nil
	}

	return a.classifier.Classify(ctx, question)
}

func toPromptType(intentType prompt.Type) prompt.Type {
	switch intentType {
	case prompt.TypeLogAnalysis:
		return prompt.TypeLogAnalysis
	case prompt.TypeSummarize:
		return prompt.TypeSummarize
	case prompt.TypeTaskBreakdown:
		return prompt.TypeTaskBreakdown
	default:
		return prompt.TypeChat
	}
}
