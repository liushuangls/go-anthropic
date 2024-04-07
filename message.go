package anthropic

import (
	"context"
	"net/http"
)

type MessagesResponseType string

const (
	MessagesResponseTypeMessage MessagesResponseType = "message"
	MessagesResponseTypeError   MessagesResponseType = "error"
)

type MessagesContentType string

const (
	MessagesContentTypeText       MessagesContentType = "text"
	MessagesContentTypeImage      MessagesContentType = "image"
	MessagesContentTypeToolResult MessagesContentType = "tool_result"
	MessagesContentTypeToolUse    MessagesContentType = "tool_use"
)

type MessagesStopReason string

const (
	MessagesStopReasonEndTurn      MessagesStopReason = "end_turn"
	MessagesStopReasonMaxTokens    MessagesStopReason = "max_tokens"
	MessagesStopReasonStopSequence MessagesStopReason = "stop_sequence"
	MessagesStopReasonToolUse      MessagesStopReason = "tool_use"
)

type MessagesRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`

	System        string           `json:"system,omitempty"`
	Metadata      map[string]any   `json:"metadata,omitempty"`
	StopSequences []string         `json:"stop_sequences,omitempty"`
	Stream        bool             `json:"stream,omitempty"`
	Temperature   *float32         `json:"temperature,omitempty"`
	TopP          *float32         `json:"top_p,omitempty"`
	TopK          *int             `json:"top_k,omitempty"`
	Tools         []ToolDefinition `json:"tools,omitempty"`
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
		Role:    RoleUser,
		Content: []MessageContent{NewTextMessageContent(text)},
	}
}

func NewAssistantTextMessage(text string) Message {
	return Message{
		Role:    RoleAssistant,
		Content: []MessageContent{NewTextMessageContent(text)},
	}
}

func NewToolResultsMessage(toolUseID, content string, isError bool) Message {
	return Message{
		Role:    RoleUser,
		Content: []MessageContent{NewToolResultMessageContent(toolUseID, content, isError)},
	}
}

func (m Message) GetFirstContent() MessageContent {
	if len(m.Content) == 0 {
		return MessageContent{}
	}
	return m.Content[0]
}

type MessageContent struct {
	Type MessagesContentType `json:"type"`

	Text *string `json:"text,omitempty"`

	Source *MessageContentImageSource `json:"source,omitempty"`

	*MessageContentToolResult

	*MessageContentToolUse
}

func NewTextMessageContent(text string) MessageContent {
	return MessageContent{
		Type: MessagesContentTypeText,
		Text: &text,
	}
}

func NewImageMessageContent(source MessageContentImageSource) MessageContent {
	return MessageContent{
		Type:   MessagesContentTypeImage,
		Source: &source,
	}
}

func NewToolResultMessageContent(toolUseID, content string, isError bool) MessageContent {
	return MessageContent{
		Type:                     MessagesContentTypeToolResult,
		MessageContentToolResult: NewMessageContentToolResult(toolUseID, content, isError),
	}
}

func NewToolUseMessageContent(toolUseID, name string, input map[string]any) MessageContent {
	return MessageContent{
		Type: MessagesContentTypeToolUse,
		MessageContentToolUse: &MessageContentToolUse{
			ID:    toolUseID,
			Name:  name,
			Input: input,
		},
	}
}

func (m *MessageContent) GetText() string {
	if m.Text != nil {
		return *m.Text
	}
	return ""
}

func (m *MessageContent) ConcatText(s string) {
	if m.Text == nil {
		m.Text = &s
	} else {
		*m.Text += s
	}
}

type MessageContentToolResult struct {
	ToolUseID *string          `json:"tool_use_id,omitempty"`
	Content   []MessageContent `json:"content,omitempty"`
	IsError   *bool            `json:"is_error,omitempty"`
}

func NewMessageContentToolResult(toolUseID, content string, isError bool) *MessageContentToolResult {
	return &MessageContentToolResult{
		ToolUseID: &toolUseID,
		Content: []MessageContent{
			{
				Type: MessagesContentTypeText,
				Text: &content,
			},
		},
		IsError: &isError,
	}
}

type MessageContentImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      any    `json:"data"`
}

type MessageContentToolUse struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Input any    `json:"input,omitempty"`
}

type MessagesResponse struct {
	ID           string               `json:"id"`
	Type         MessagesResponseType `json:"type"`
	Role         string               `json:"role"`
	Content      []MessageContent     `json:"content"`
	Model        string               `json:"model"`
	StopReason   MessagesStopReason   `json:"stop_reason"`
	StopSequence string               `json:"stop_sequence"`
	Usage        MessagesUsage        `json:"usage"`
}

// GetFirstContentText get Content[0].Text avoid panic
func (m MessagesResponse) GetFirstContentText() string {
	if len(m.Content) == 0 {
		return ""
	}
	return m.Content[0].GetText()
}

type MessagesUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// InputSchema is an object describing the tool.
	// You can pass json.RawMessage to describe the schema,
	// or you can pass in a struct which serializes to the proper JSON schema.
	// The jsonschema package is provided for convenience, but you should
	// consider another specialized library if you require more complex schemas.
	InputSchema any `json:"input_schema"`
}

func (c *Client) CreateMessages(ctx context.Context, request MessagesRequest) (response MessagesResponse, err error) {
	request.Stream = false

	var setters []requestSetter
	if len(request.Tools) > 0 {
		setters = append(setters, withBetaVersion(c.config.BetaVersion))
	}

	urlSuffix := "/messages"
	req, err := c.newRequest(ctx, http.MethodPost, urlSuffix, request, setters...)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
