package prompt

// Prompt is the normalized input passed to the LLM layer.
type Prompt struct {
	System   string
	Question string
}

// Type identifies a prompt template family.
type Type string

const (
	TypeQA            Type = "qa"
	TypeRAGQA         Type = "rag_qa"
	TypeLogAnalysis   Type = "log_analysis"
	TypeSummary       Type = "summary"
	TypeTaskBreakdown Type = "task_breakdown"
	TypeReport        Type = "report"
)

// Request describes the inputs required to build a prompt.
type Request struct {
	Type    Type
	Input   string
	Context string
}

// Builder converts domain prompt requests into LLM prompts.
type Builder interface {
	Build(req Request) (Prompt, error)
}

type template interface {
	Build(req Request) Prompt
}
