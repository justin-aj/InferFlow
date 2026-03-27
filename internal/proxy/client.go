package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ajinfrank/inferflow/internal/router"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type BackendRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type BackendResponse struct {
	Model      string `json:"model"`
	OutputText string `json:"output_text"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) HealthCheck(ctx context.Context, backend *router.Backend) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, backend.BaseURL+"/healthz", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("health check status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) SendChatCompletion(ctx context.Context, backend *router.Backend, reqBody ChatCompletionRequest) (ChatCompletionResponse, error) {
	payload, err := json.Marshal(BackendRequest{
		Model:    reqBody.Model,
		Messages: reqBody.Messages,
		Stream:   reqBody.Stream,
	})
	if err != nil {
		return ChatCompletionResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, backend.BaseURL+"/infer", bytes.NewReader(payload))
	if err != nil {
		return ChatCompletionResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ChatCompletionResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return ChatCompletionResponse{}, fmt.Errorf("backend status %d", resp.StatusCode)
	}

	var backendResp BackendResponse
	if err := json.NewDecoder(resp.Body).Decode(&backendResp); err != nil {
		return ChatCompletionResponse{}, err
	}

	promptTokens := estimateTokensFromMessages(reqBody.Messages)
	completionTokens := estimateTokens(backendResp.OutputText)
	now := time.Now().Unix()

	return ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", now),
		Object:  "chat.completion",
		Created: now,
		Model:   chooseModel(backendResp.Model, reqBody.Model),
		Choices: []Choice{{
			Index: 0,
			Message: ChatMessage{
				Role:    "assistant",
				Content: backendResp.OutputText,
			},
			FinishReason: "stop",
		}},
		Usage: Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}, nil
}

func estimateTokensFromMessages(messages []ChatMessage) int {
	var builder strings.Builder
	for _, msg := range messages {
		builder.WriteString(msg.Content)
		builder.WriteByte(' ')
	}
	return estimateTokens(builder.String())
}

func estimateTokens(text string) int {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return 0
	}
	runes := len([]rune(trimmed))
	tokens := runes / 4
	if runes%4 != 0 {
		tokens++
	}
	if tokens == 0 {
		return 1
	}
	return tokens
}

func chooseModel(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	if fallback != "" {
		return fallback
	}
	return "mock-llm"
}
