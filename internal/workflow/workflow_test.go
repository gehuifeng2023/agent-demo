package workflow

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"agent-demo/internal/tool"
)

type testTool struct {
	name string
	run  func(context.Context, string) (string, error)
}

func (t testTool) Name() string        { return t.name }
func (t testTool) Description() string { return t.name }
func (t testTool) Execute(ctx context.Context, input string) (string, error) {
	return t.run(ctx, input)
}

func TestExecutorRunsToolNodesInOrder(t *testing.T) {
	var inputs []string
	first := testTool{name: "first", run: func(_ context.Context, input string) (string, error) {
		inputs = append(inputs, input)
		return "first-result", nil
	}}
	second := testTool{name: "second", run: func(_ context.Context, input string) (string, error) {
		inputs = append(inputs, input)
		return "second-result", nil
	}}
	wf := &Workflow{ID: "ordered", Nodes: []Node{
		ToolNode{NameValue: "first", Tool: first, InputTemplate: "{{question}}", OutputKeyValue: "first"},
		ToolNode{NameValue: "second", Tool: second, InputTemplate: "prefix {{results.first}}", OutputKeyValue: "second"},
	}}
	wfCtx := NewContext("session-1", "question")

	if err := (Executor{}).Run(context.Background(), wf, wfCtx); err != nil {
		t.Fatalf("run workflow: %v", err)
	}
	if got := strings.Join(inputs, ","); got != "question,prefix first-result" {
		t.Fatalf("unexpected inputs %q", got)
	}
	if wfCtx.Results["second"] != "second-result" {
		t.Fatalf("unexpected results %#v", wfCtx.Results)
	}
}

func TestExecutorReadsAndAnalyzesLogFile(t *testing.T) {
	wf := &Workflow{ID: "analyze_test_file", Nodes: []Node{
		ToolNode{
			NameValue:      "read_log",
			Tool:           tool.FileReaderTool{RootDir: filepath.Join("..", "..", "knowledge_attachment")},
			InputTemplate:  "test_file/test_file.log",
			OutputKeyValue: "log_content",
		},
		ToolNode{
			NameValue:      "analyze_log",
			Tool:           tool.LogAnalyzerTool{},
			InputTemplate:  "{{results.log_content}}",
			OutputKeyValue: "analysis",
		},
	}}
	wfCtx := NewContext("session-1", "analyze test log")

	if err := (Executor{}).Run(context.Background(), wf, wfCtx); err != nil {
		t.Fatalf("run workflow: %v", err)
	}
	if !strings.Contains(wfCtx.Results["log_content"], "502") {
		t.Fatalf("expected read log content to contain 502, got %q", wfCtx.Results["log_content"])
	}
	if !strings.Contains(wfCtx.Results["analysis"], `"error_type": "gateway_502"`) {
		t.Fatalf("expected gateway analysis, got %q", wfCtx.Results["analysis"])
	}
}

func TestExecutorStopsOnToolError(t *testing.T) {
	called := false
	wf := &Workflow{ID: "fail-fast", Nodes: []Node{
		ToolNode{NameValue: "failed", Tool: testTool{name: "failed", run: func(context.Context, string) (string, error) { return "", errors.New("boom") }}, InputTemplate: "{{question}}", OutputKeyValue: "failed"},
		ToolNode{NameValue: "after", Tool: testTool{name: "after", run: func(context.Context, string) (string, error) { called = true; return "", nil }}, InputTemplate: "{{question}}", OutputKeyValue: "after"},
	}}

	err := (Executor{}).Run(context.Background(), wf, NewContext("", "question"))
	if err == nil || !strings.Contains(err.Error(), "node failed: boom") {
		t.Fatalf("unexpected error %v", err)
	}
	if called {
		t.Fatal("expected execution to stop after failed node")
	}
}

func TestResolveTemplateRejectsUnknownResult(t *testing.T) {
	_, err := ResolveTemplate("{{results.missing}}", NewContext("", "question"))
	if err == nil || !strings.Contains(err.Error(), "not available") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestExecutorRejectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	wf := &Workflow{ID: "cancelled", Nodes: []Node{
		ToolNode{NameValue: "node", Tool: testTool{name: "node", run: func(context.Context, string) (string, error) { return "ok", nil }}, InputTemplate: "{{question}}", OutputKeyValue: "result"},
	}}
	if err := (Executor{}).Run(ctx, wf, NewContext("", "question")); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestRegistryRejectsInvalidWorkflow(t *testing.T) {
	registry := NewRegistry()
	err := registry.Register(&Workflow{ID: "empty"})
	if err == nil || !strings.Contains(err.Error(), "has no nodes") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestRegistryRejectsForwardResultReference(t *testing.T) {
	tool := testTool{name: "tool", run: func(context.Context, string) (string, error) { return "ok", nil }}
	err := NewRegistry().Register(&Workflow{ID: "forward", Nodes: []Node{
		ToolNode{NameValue: "first", Tool: tool, InputTemplate: "{{results.second}}", OutputKeyValue: "first"},
		ToolNode{NameValue: "second", Tool: tool, InputTemplate: "{{question}}", OutputKeyValue: "second"},
	}})
	if err == nil || !strings.Contains(err.Error(), "not available before this node") {
		t.Fatalf("unexpected error %v", err)
	}
}
