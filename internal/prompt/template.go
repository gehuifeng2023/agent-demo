package prompt

type Type string

const (
	TypeChat          Type = "chat"
	TypeLogAnalysis   Type = "log_analysis"
	TypeSummarize     Type = "summarize"
	TypeTaskBreakdown Type = "task_breakdown"
)

type Builder interface {
	Build(input string) string
}
