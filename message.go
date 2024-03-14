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
	TopP          *float32       `json:"top_p,omitempty"`
	TopK          *int           `json:"top_k,omitempty"`
}

func (m *MessagesRequest) SetTemperature(t float32) {
	m.Temperature = &t
}

func (m *MessagesRequest) SetTopP(p float32) {
	m.TopP = &p
}

func (m *MessagesRequest) SetTopK(k int) {
	m.TopK = &k
}

type Message struct {
	Role    string           `json:"role"`
	Content []MessageContent `json:"content"`
}

func NewUserTextMessage(text string) Message {
	return Message{
		Role:    "user",
		Content: []MessageContent{NewTextMessageContent(text)},
	}
}

func NewAssistantTextMessage(text string) Message {
	return Message{
		Role:    "assistant",
		Content: []MessageContent{NewTextMessageContent(text)},
	}
}

func (m Message) GetFirstContent() MessageContent {
	if len(m.Content) == 0 {
		return MessageContent{}
	}
	return m.Content[0]
}

type MessageContent struct {
	Type   string                     `json:"type"`
	Text   *string                    `json:"text,omitempty"`
	Source *MessageContentImageSource `json:"source,omitempty"`
}

func NewTextMessageContent(text string) MessageContent {
	return MessageContent{
		Type: "text",
		Text: &text,
	}
}

func NewImageMessageContent(source MessageContentImageSource) MessageContent {
	return MessageContent{
		Type:   "image",
		Source: &source,
	}
}

func (m MessageContent) IsTextContent() bool {
	return m.Type == "text"
}

func (m MessageContent) IsImageContent() bool {
	return m.Type == "image"
}

func (m MessageContent) GetText() string {
	if m.IsTextContent() && m.Text != nil {
		return *m.Text
	}
	return ""
}

type MessageContentImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      any    `json:"data"`
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

// GetFirstContentText get Content[0].Text avoid panic
func (m MessagesResponse) GetFirstContentText() string {
	if len(m.Content) == 0 {
		return ""
	}
	return m.Content[0].Text
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
