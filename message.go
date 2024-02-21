package anthropic

import (
	"context"
	"net/http"
)

type (
	MessagesResponseType string
)

const (
	MessagesResponseMsg MessagesResponseType = "message"
	MessagesResponseErr MessagesResponseType = "error"
)

type MessagesRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`

	System        string         `json:"system,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	StopSequences []string       `json:"stop_sequences,omitempty"`
	Stream        bool           `json:"stream,omitempty"`
	Temperature   *float32       `json:"temperature,omitempty"`
	TopK          *float32       `json:"top_k,omitempty"`
	TopP          *int           `json:"top_p,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MessagesResponse struct {
	ID           string               `json:"id"`
	Type         MessagesResponseType `json:"type"`
	Role         string               `json:"role"`
	Content      []MessagesContent    `json:"content"`
	Model        string               `json:"model"`
	StopReason   string               `json:"stop_reason"`
	StopSequence string               `json:"stop_sequence"`
	Usage        MessagesUsage        `json:"usage"`
}

type MessagesContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type MessagesUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (c *Client) CreateMessages(ctx context.Context, request MessagesRequest) (response MessagesResponse, err error) {
	request.Stream = false

	urlSuffix := "/messages"
	req, err := c.newRequest(ctx, http.MethodPost, urlSuffix, request)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
