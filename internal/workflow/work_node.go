package workflow

import (
	"context"
	"fmt"
	"strings"

	"agent-demo/internal/tool"
)

type ToolNode struct {
	NameValue      string
	Tool           tool.Tool
	InputTemplate  string
	OutputKeyValue string
}

func (n ToolNode) Name() string {
	return n.NameValue
}
func (n ToolNode) OutputKey() string {
	return n.OutputKeyValue
}

func (n ToolNode) Execute(ctx context.Context, wf *Context) error {
	if n.Tool == nil {
		return fmt.Errorf("tool is nil")
	}
	if wf == nil {
		return fmt.Errorf("workflow context is nil")
	}
	input, err := ResolveTemplate(n.InputTemplate, wf)
	if err != nil {
		return err
	}
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("resolved input is empty")
	}

	output, err := n.Tool.Execute(ctx, input)
	if err != nil {
		return err
	}
	if wf.Results == nil {
		wf.Results = make(map[string]string)
	}
	wf.Results[n.OutputKey()] = output
	return nil
}

func (n ToolNode) validate(availableResults map[string]struct{}) error {
	if n.Tool == nil {
		return fmt.Errorf("tool is nil")
	}
	return ValidateTemplate(n.InputTemplate, availableResults)
}
