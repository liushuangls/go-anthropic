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
	Model    Model       `json:"model"`
	Messages []Message   `json:"messages"`
	System   interface{} `json:"system,omitempty"`

	Tools        []ToolDefinition     `json:"tools,omitempty"`
	ToolChoice   *ToolChoice          `json:"tool_choice,omitempty"`
	Thinking     *Thinking            `json:"thinking,omitempty"`
	CacheControl *MessageCacheControl `json:"cache_control,omitempty"`
	OutputConfig *OutputConfig        `json:"output_config,omitempty"`
}

// newCountTokensRequest builds the restricted count_tokens body from a
// MessagesRequest, preserving the system / multi-system handling that
// MessagesRequest.MarshalJSON performs.
func newCountTokensRequest(request MessagesRequest) countTokensRequest {
	body := countTokensRequest{
		Model:        request.Model,
		Messages:     request.Messages,
		Tools:        request.Tools,
		ToolChoice:   request.ToolChoice,
		Thinking:     request.Thinking,
		CacheControl: request.CacheControl,
		OutputConfig: request.OutputConfig,
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
