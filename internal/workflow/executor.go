package workflow

import (
	"context"
	"fmt"
)

type Executor struct{}

func (Executor) Run(ctx context.Context, wf *Workflow, wfCtx *Context) error {
	if err := Validate(wf); err != nil {
		return err
	}
	if wfCtx == nil {
		return fmt.Errorf("workflow context is nil")
	}
	if wfCtx.Results == nil {
		wfCtx.Results = make(map[string]string)
	}

	for _, node := range wf.Nodes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := node.Execute(ctx, wfCtx); err != nil {
			return fmt.Errorf("node %s: %w", node.Name(), err)
		}
	}
	return nil
}
