package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	modelName  string
	httpClient *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model             string    `json:"model"`
	Messages          []Message `json:"messages"`
	MaxTokens         int       `json:"max_tokens,omitempty"`
	Temperature       float64   `json:"temperature,omitempty"`
	RepetitionPenalty float64   `json:"repetition_penalty,omitempty"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

type chatChoice struct {
	Message Message `json:"message"`
}

func NewClient(baseURL, modelName string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		modelName:  modelName,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("llama health status %d", resp.StatusCode)
	}
	return nil
}

// GenerateOpts holds per-request inference parameters.
type GenerateOpts struct {
	Messages          []Message
	MaxTokens         int
	Temperature       float64
	RepetitionPenalty float64
}

func (c *Client) Generate(ctx context.Context, opts GenerateOpts) (string, error) {
	if len(opts.Messages) == 0 {
		return "", fmt.Errorf("no messages provided")
	}

	payload, err := json.Marshal(chatRequest{
		Model:             c.modelName,
		Messages:          opts.Messages,
		MaxTokens:         opts.MaxTokens,
		Temperature:       opts.Temperature,
		RepetitionPenalty: opts.RepetitionPenalty,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("llama completion status %d", resp.StatusCode)
	}

	var completion chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return "", err
	}
	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("llama completion returned no choices")
	}
	text := strings.TrimSpace(completion.Choices[0].Message.Content)
	if text == "" {
		return "", fmt.Errorf("llama completion returned empty content")
	}
	return text, nil
}
