package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"inferflow/internal/llm"
	"inferflow/internal/proxy"
)

type fakeGenerator struct {
	healthErr error
	output    string
	genErr    error
	lastOpts  llm.GenerateOpts
}

func (f *fakeGenerator) HealthCheck(context.Context) error {
	return f.healthErr
}

func (f *fakeGenerator) Generate(_ context.Context, opts llm.GenerateOpts) (string, error) {
	f.lastOpts = opts
	if f.genErr != nil {
		return "", f.genErr
	}
	return f.output, nil
}

func TestHealthz(t *testing.T) {
	handler := NewHandler(&fakeGenerator{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestInferPassesMessages(t *testing.T) {
	client := &fakeGenerator{output: "hello from llama"}
	handler := NewHandler(client)

	body := proxy.BackendRequest{
		Model: "Qwen/Qwen2.5-0.5B-Instruct",
		Messages: []proxy.ChatMessage{
			{Role: "system", Content: "You are helpful."},
			{Role: "user", Content: "Say hello"},
		},
		MaxTokens:   100,
		Temperature: 0.7,
	}
	req := httptest.NewRequest(http.MethodPost, "/infer", mustJSON(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if len(client.lastOpts.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(client.lastOpts.Messages))
	}
	if client.lastOpts.MaxTokens != 100 {
		t.Fatalf("expected max_tokens=100, got %d", client.lastOpts.MaxTokens)
	}

	var resp proxy.BackendResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.OutputText != "hello from llama" {
		t.Fatalf("unexpected output: %q", resp.OutputText)
	}
}

func TestInferHandlesGeneratorFailure(t *testing.T) {
	handler := NewHandler(&fakeGenerator{genErr: errors.New("boom")})
	req := httptest.NewRequest(http.MethodPost, "/infer", mustJSON(proxy.BackendRequest{
		Model: "Qwen/Qwen2.5-0.5B-Instruct",
		Messages: []proxy.ChatMessage{
			{Role: "user", Content: "hello"},
		},
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
}

func mustJSON(v any) *bytes.Buffer {
	data, _ := json.Marshal(v)
	return bytes.NewBuffer(data)
}
