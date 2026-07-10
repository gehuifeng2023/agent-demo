package tool

import (
	"context"
	"testing"
)

type testTool struct{}

func (testTool) Name() string { return "test" }

func (testTool) Description() string { return "test tool" }

func (testTool) Execute(context.Context, string) (string, error) { return "ok", nil }

func TestRegistryGet(t *testing.T) {
	registry := NewRegistry()
	registry.Register(testTool{})

	if _, ok := registry.Get("test"); !ok {
		t.Fatal("expected registered tool")
	}
	if _, ok := registry.Get("missing"); ok {
		t.Fatal("expected missing tool")
	}
}
