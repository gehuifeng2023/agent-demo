package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPGetToolExecutesAllowlistedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("unexpected authorization header %q", got)
		}
		_, _ = fmt.Fprint(w, "hello")
	}))
	defer server.Close()

	input := mustMarshalHTTPInput(t, HTTPRequestInput{
		URL:     server.URL + "/health",
		Headers: map[string]string{"Authorization": "Bearer test-token"},
	})
	got, err := (HTTPGetTool{Client: newTestHTTPClient(t, server.URL)}).Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("execute GET: %v", err)
	}
	if got != "status=200\nbody=hello" {
		t.Fatalf("unexpected result %q", got)
	}
}

func TestHTTPPostToolSendsJSONBodyAndHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("X-Request-ID"); got != "request-1" {
			t.Fatalf("unexpected request header %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("unexpected content type %q", got)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["name"] != "agent" {
			t.Fatalf("unexpected body %#v", body)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprint(w, "created")
	}))
	defer server.Close()

	input := mustMarshalHTTPInput(t, HTTPRequestInput{
		URL:     server.URL,
		Headers: map[string]string{"X-Request-ID": "request-1"},
		Body:    json.RawMessage(`{"name":"agent"}`),
	})
	got, err := (HTTPPostTool{Client: newTestHTTPClient(t, server.URL)}).Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("execute POST: %v", err)
	}
	if got != "status=201\nbody=created" {
		t.Fatalf("unexpected result %q", got)
	}
}

func TestHTTPToolReturnsNonSuccessResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	input := mustMarshalHTTPInput(t, HTTPRequestInput{URL: server.URL})
	got, err := (HTTPGetTool{Client: newTestHTTPClient(t, server.URL)}).Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("execute GET: %v", err)
	}
	if !strings.Contains(got, "status=503") || !strings.Contains(got, "unavailable") {
		t.Fatalf("unexpected result %q", got)
	}
}

func TestHTTPToolRejectsInvalidOrUnallowlistedURL(t *testing.T) {
	client := NewHTTPClient([]string{"api.example.com"}, time.Second)
	tool := HTTPGetTool{Client: client}

	for _, input := range []string{
		`not JSON`,
		`{}`,
		`{"url":"ftp://api.example.com/data"}`,
		`{"url":"https://other.example.com/data"}`,
		`{"url":"/relative"}`,
	} {
		if _, err := tool.Execute(context.Background(), input); err == nil {
			t.Fatalf("expected error for %s", input)
		}
	}
}

func TestHTTPToolRejectsRedirectToUnallowlistedHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://untrusted.example.com/secret", http.StatusFound)
	}))
	defer server.Close()

	input := mustMarshalHTTPInput(t, HTTPRequestInput{URL: server.URL})
	_, err := (HTTPGetTool{Client: newTestHTTPClient(t, server.URL)}).Execute(context.Background(), input)
	if err == nil || !strings.Contains(err.Error(), "host not allowed") {
		t.Fatalf("expected redirect allowlist error, got %v", err)
	}
}

func TestHTTPToolHonorsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	input := mustMarshalHTTPInput(t, HTTPRequestInput{URL: server.URL})
	_, err := (HTTPGetTool{Client: NewHTTPClient([]string{hostFromURL(t, server.URL)}, 10*time.Millisecond)}).Execute(context.Background(), input)
	if err == nil || !strings.Contains(err.Error(), "send HTTP request") {
		t.Fatalf("expected timeout request error, got %v", err)
	}
}

func TestHTTPToolTruncatesLargeResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, strings.Repeat("x", int(maxHTTPResponseBytes+1)))
	}))
	defer server.Close()

	input := mustMarshalHTTPInput(t, HTTPRequestInput{URL: server.URL})
	got, err := (HTTPGetTool{Client: newTestHTTPClient(t, server.URL)}).Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("execute GET: %v", err)
	}
	if !strings.HasSuffix(got, fmt.Sprintf("\n[body truncated at %d bytes]", maxHTTPResponseBytes)) {
		t.Fatalf("expected truncation marker, got suffix %q", got[len(got)-80:])
	}
}

func mustMarshalHTTPInput(t *testing.T, input HTTPRequestInput) string {
	t.Helper()
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	return string(data)
}

func newTestHTTPClient(t *testing.T, rawURL string) *HTTPClient {
	t.Helper()
	return NewHTTPClient([]string{hostFromURL(t, rawURL)}, time.Second)
}

func hostFromURL(t *testing.T, rawURL string) string {
	t.Helper()
	u, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	return u.URL.Hostname()
}
