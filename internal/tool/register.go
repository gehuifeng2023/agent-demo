package tool

import (
	"fmt"
)

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: map[string]Tool{}}
}

func (r *Registry) Register(t Tool) {
	if t == nil {
		return
	}
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) MustGet(name string) Tool {
	t, ok := r.tools[name]
	if !ok {
		panic(fmt.Sprintf("tool not found: %s", name))
	}
	return t
}
