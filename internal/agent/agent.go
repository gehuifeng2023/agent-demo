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
	"agent-demo/internal/model"
	"agent-demo/internal/prompt"
	"agent-demo/internal/retriever"
	"agent-demo/internal/session"
	"agent-demo/internal/tool"
	"agent-demo/internal/workflow"
)

type Agent struct {
	llmClient         llm.Client
	promptFactory     *prompt.Factory
	classifier        *intent.Classifier
	retriever         *retriever.UnifiedRetriever
	sessionStore      session.Store
	maxHistoryMessage int
	topK              int
	toolRegistry      *tool.Registry
	toolsEnabled      bool
	workflowRegistry  *workflow.Registry
}

type Options struct {
	TopK               int
	SessionMaxMessages int
	MaxHistoryMessage  int
	ToolRegistry       *tool.Registry
	ToolsEnabled       bool
	WorkflowRegistry   *workflow.Registry
}

func NewAgent(llmClient llm.Client, unifiedRetriever *retriever.UnifiedRetriever) *Agent {
	return NewAgentWithOptions(llmClient, unifiedRetriever, Options{})
}

func NewAgentWithOptions(llmClient llm.Client, unifiedRetriever *retriever.UnifiedRetriever, options Options) *Agent {
	if unifiedRetriever == nil {
		unifiedRetriever = retriever.NewUnifiedRetriever()
	}
	if options.TopK <= 0 {
		options.TopK = 3
	}
	if options.SessionMaxMessages <= 0 {
		options.SessionMaxMessages = 30
	}
	if options.MaxHistoryMessage <= 0 {
		options.MaxHistoryMessage = 8
	}

	return &Agent{
		llmClient:         llmClient,
		promptFactory:     prompt.NewFactory(),
		classifier:        intent.NewClassifier(),
		retriever:         unifiedRetriever,
		sessionStore:      session.NewMemoryStore(options.SessionMaxMessages),
		maxHistoryMessage: options.MaxHistoryMessage,
		topK:              options.TopK,
		toolRegistry:      options.ToolRegistry,
		toolsEnabled:      options.ToolsEnabled,
		workflowRegistry:  options.WorkflowRegistry,
	}
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

	chunks := a.retriever.Retrieve(question, compactStrings(req.KnowledgeBaseIDs), compactStrings(req.FileIDs), a.topK)
	toolContext, err := a.executeWorkflowOrTool(ctx, sessionID, question, req.WorkflowID)
	if err != nil {
		return "", "", sessionID, buildSources(chunks), err
	}

	promptText, err := a.buildPrompt(intentType, question, history, chunks, toolContext)
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

func (a *Agent) buildPrompt(intentType prompt.Type, question string, history []session.Message, chunks []document.Chunk, toolContext string) (string, error) {
	questionWithTools := withToolContext(question, toolContext)

	if intentType == prompt.TypeChat {
		if len(chunks) > 0 {
			return prompt.BuildRAGPrompt(questionWithTools, chunks, history), nil
		}
		if looksLikeDocumentQuestion(question) {
			return prompt.BuildRAGPrompt(questionWithTools, nil, history), nil
		}

		input := prompt.WithHistory(questionWithTools, history)
		promptText, err := a.promptFactory.Build(prompt.TypeChat, input)
		return promptText, err
	}

	promptText, err := a.promptFactory.Build(intentType, questionWithTools)
	return promptText, err
}

func (a *Agent) executeToolIfNeeded(ctx context.Context, question string) (string, error) {
	if !a.toolsEnabled || a.toolRegistry == nil {
		return "", nil
	}

	toolName := tool.RouteTool(question)
	if toolName == "" {
		return "", nil
	}

	selectedTool, ok := a.toolRegistry.Get(toolName)
	if !ok {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}

	input := tool.ExtractToolInput(toolName, question)
	if input == "" {
		return "", fmt.Errorf("execute tool %s: input is empty", toolName)
	}

	output, err := selectedTool.Execute(ctx, input)
	if err != nil {
		return "", fmt.Errorf("execute tool %s: %w", toolName, err)
	}

	return fmt.Sprintf("工具：%s\n输入：%s\n输出：\n%s", selectedTool.Name(), input, strings.TrimSpace(output)), nil
}

func (a *Agent) executeWorkflowOrTool(ctx context.Context, sessionID, question, workflowID string) (string, error) {
	workflowID = strings.TrimSpace(workflowID)
	if workflowID == "" {
		return a.executeToolIfNeeded(ctx, question)
	}
	if a.workflowRegistry == nil {
		return "", fmt.Errorf("workflow not configured: %s", workflowID)
	}
	wf, ok := a.workflowRegistry.Get(workflowID)
	if !ok {
		return "", fmt.Errorf("workflow not found: %s", workflowID)
	}
	wfCtx := workflow.NewContext(sessionID, question)
	if err := (workflow.Executor{}).Run(ctx, wf, wfCtx); err != nil {
		return "", fmt.Errorf("execute workflow %s: %w", workflowID, err)
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "工作流：%s", wf.ID)
	for _, node := range wf.Nodes {
		fmt.Fprintf(&builder, "\n节点：%s\n输出：\n%s", node.Name(), strings.TrimSpace(wfCtx.Results[node.OutputKey()]))
	}
	return builder.String(), nil
}

func withToolContext(question string, toolContext string) string {
	toolContext = strings.TrimSpace(toolContext)
	if toolContext == "" {
		return question
	}

	return fmt.Sprintf(`%s

【工具上下文】
%s

请优先依据工具上下文回答。`, question, toolContext)
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
