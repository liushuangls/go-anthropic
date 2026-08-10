package anthropic

import (
	"context"
	"net/http"
)

type CountTokensResponse struct {
	httpHeader

	InputTokens int `json:"input_tokens"`
}

// countTokensRequest is the restricted body sent to the count_tokens
// endpoint. The endpoint only accepts a subset of the message creation
// fields and rejects any extra inputs (e.g. max_tokens) with a 400. We
// therefore serialize only the permitted fields here rather than reusing
// MessagesRequest, whose max_tokens field has no omitempty and would always
// be present.
type countTokensRequest struct {
	Model            Model       `json:"model"`
	AnthropicVersion string      `json:"anthropic_version,omitempty"`
	Messages         []Message   `json:"messages"`
	System           interface{} `json:"system,omitempty"`

	Tools      []ToolDefinition `json:"tools,omitempty"`
	ToolChoice *ToolChoice      `json:"tool_choice,omitempty"`
}

var _ VertexAISupport = (*countTokensRequest)(nil)

func (m countTokensRequest) GetModel() Model {
	return m.Model
}

func (m *countTokensRequest) SetAnthropicVersion(version APIVersion) {
	m.AnthropicVersion = string(version)
}

func (m countTokensRequest) IsStreaming() bool {
	return false
}

// newCountTokensRequest builds the restricted count_tokens body from a
// MessagesRequest, preserving the system / multi-system handling that
// MessagesRequest.MarshalJSON performs.
func newCountTokensRequest(request MessagesRequest) countTokensRequest {
	body := countTokensRequest{
		Model:            request.Model,
		AnthropicVersion: request.AnthropicVersion,
		Messages:         request.Messages,
		Tools:            request.Tools,
		ToolChoice:       request.ToolChoice,
	}

	if len(request.MultiSystem) > 0 {
		body.System = request.MultiSystem
	} else if len(request.System) > 0 {
		body.System = request.System
	}

	return body
}

func (c *Client) CountTokens(
	ctx context.Context,
	request MessagesRequest,
) (response CountTokensResponse, err error) {
	var setters []requestSetter
	if len(c.config.BetaVersion) > 0 {
		setters = append(setters, withBetaVersion(c.config.BetaVersion...))
	}

	body := newCountTokensRequest(request)

	urlSuffix := "/messages/count_tokens"
	req, err := c.newRequest(ctx, http.MethodPost, urlSuffix, &body, setters...)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
