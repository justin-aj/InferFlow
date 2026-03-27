package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"inferflow/internal/proxy"
)

func TestEndToEndChatCompletion(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz":
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		case "/infer":
			writeJSON(w, http.StatusOK, map[string]string{
				"model":       "mock-llm",
				"output_text": "Mock response to: integration prompt",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer backend.Close()

	srv := newTestServer(t, []string{backend.URL})
	defer func() { _ = srv.Shutdown() }()

	app := httptest.NewServer(srv.httpSrv.Handler)
	defer app.Close()

	reqBody := proxy.ChatCompletionRequest{
		Model: "mock-llm",
		Messages: []proxy.ChatMessage{{
			Role:    "user",
			Content: "integration prompt",
		}},
	}
	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(app.URL+"/v1/chat/completions", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}
