package llm

import (
	"context"
	"time"
)

type StreamClient interface {
	Stream(ctx context.Context, prompt string) (<-chan string, <-chan error)
}

type MockStreamClient struct{}

func (c MockStreamClient) Stream(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	out := make(chan string)
	errs := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errs)
		for _, part := range []string{"正在分析...", "\n结论：", "这是流式输出示例。"} {
			select {
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			case out <- part:
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()
	return out, errs
}

func (c *MockClient) Stream(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	return MockStreamClient{}.Stream(ctx, prompt)
}
