package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Message is a single chat turn in the OpenAI-compatible format.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolCall represents a model-requested function invocation.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolCallFunc `json:"function"`
}

type ToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolSpec describes a tool advertised to the model.
type ToolSpec struct {
	Type     string       `json:"type"`
	Function FunctionSpec `json:"function"`
}

type FunctionSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ChatRequest is the body we send to /chat/completions.
type ChatRequest struct {
	Model    string     `json:"model"`
	Messages []Message  `json:"messages"`
	Tools    []ToolSpec `json:"tools,omitempty"`
}

// ChatResponse is the subset of the response we care about.
type ChatResponse struct {
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    any    `json:"code"`
	} `json:"error,omitempty"`
}

// LLMClient is the minimal interface the agent needs.
type LLMClient interface {
	Chat(req ChatRequest) (*ChatResponse, error)
}

// OpenRouterClient is a small stdlib-only HTTP client for OpenRouter.
type OpenRouterClient struct {
	APIKey  string
	Model   string
	BaseURL string
	HTTP    *http.Client
}

func NewOpenRouterClient(apiKey, model string) *OpenRouterClient {
	return &OpenRouterClient{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: "https://openrouter.ai/api/v1",
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *OpenRouterClient) Chat(req ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = c.Model
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("HTTP-Referer", "https://learnwithparam.com")
	httpReq.Header.Set("X-Title", "coding-agents-from-scratch-go")

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("openrouter HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var out ChatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w (body=%s)", err, string(raw))
	}
	if out.Error != nil {
		return nil, fmt.Errorf("openrouter error: %s", out.Error.Message)
	}
	return &out, nil
}
