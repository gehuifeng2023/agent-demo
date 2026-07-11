package workflow

import (
	"context"
	"fmt"
	"strings"
)

type Node interface {
	Name() string
	OutputKey() string
	Execute(ctx context.Context, wf *Context) error
}

type Workflow struct {
	ID    string
	Nodes []Node
}

type Registry struct {
	workflows map[string]*Workflow
}

func NewRegistry() *Registry {
	return &Registry{workflows: make(map[string]*Workflow)}
}

func (r *Registry) Register(wf *Workflow) error {
	if r == nil {
		return fmt.Errorf("workflow registry is nil")
	}
	if err := Validate(wf); err != nil {
		return err
	}
	if _, ok := r.workflows[wf.ID]; ok {
		return fmt.Errorf("workflow %q already registered", wf.ID)
	}
	r.workflows[wf.ID] = wf
	return nil
}

func (r *Registry) Get(id string) (*Workflow, bool) {
	if r == nil {
		return nil, false
	}
	wf, ok := r.workflows[id]
	return wf, ok
}

func Validate(wf *Workflow) error {
	if wf == nil {
		return fmt.Errorf("workflow is nil")
	}
	wf.ID = strings.TrimSpace(wf.ID)
	if wf.ID == "" {
		return fmt.Errorf("workflow ID is empty")
	}
	if len(wf.Nodes) == 0 {
		return fmt.Errorf("workflow %q has no nodes", wf.ID)
	}

	names := make(map[string]struct{}, len(wf.Nodes))
	outputs := make(map[string]struct{}, len(wf.Nodes))
	for _, node := range wf.Nodes {
		if node == nil {
			return fmt.Errorf("workflow %q has nil node", wf.ID)
		}
		name := strings.TrimSpace(node.Name())
		if name == "" {
			return fmt.Errorf("workflow %q has node with empty name", wf.ID)
		}
		if _, ok := names[name]; ok {
			return fmt.Errorf("workflow %q has duplicate node name %q", wf.ID, name)
		}
		names[name] = struct{}{}

		outputKey := strings.TrimSpace(node.OutputKey())
		if outputKey == "" {
			return fmt.Errorf("workflow %q node %q has empty output key", wf.ID, name)
		}
		if _, ok := outputs[outputKey]; ok {
			return fmt.Errorf("workflow %q has duplicate output key %q", wf.ID, outputKey)
		}
		if err := validateNode(node, outputs); err != nil {
			return fmt.Errorf("workflow %q node %q: %w", wf.ID, name, err)
		}
		outputs[outputKey] = struct{}{}
	}
	return nil
}

func validateNode(node Node, availableResults map[string]struct{}) error {
	switch n := node.(type) {
	case ToolNode:
		return n.validate(availableResults)
	case *ToolNode:
		if n == nil {
			return fmt.Errorf("tool node is nil")
		}
		return n.validate(availableResults)
	default:
		return fmt.Errorf("unsupported node type %T", node)
	}
}
