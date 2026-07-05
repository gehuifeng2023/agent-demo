package prompt

import "fmt"

type Factory struct {
	builders map[Type]Builder
}

func NewFactory() *Factory {
	return &Factory{
		builders: map[Type]Builder{
			TypeChat:          NewChatBuilder(),
			TypeLogAnalysis:   NewLogAnalysisBuilder(),
			TypeSummarize:     NewSummarizeBuilder(),
			TypeTaskBreakdown: NewTaskBreakdownBuilder(),
		},
	}
}

func (f *Factory) Build(promptType Type, input string) (string, error) {
	builder, ok := f.builders[promptType]
	if !ok {
		return "", fmt.Errorf("unsupported prompt type: %s", promptType)
	}

	return builder.Build(input), nil
}
