package anthropic

import (
	"context"
	"net/http"
)

type CompleteRequest struct {
	Model             string `json:"model"`
	Prompt            string `json:"prompt"`
	MaxTokensToSample int    `json:"max_tokens_to_sample"`

	StopSequences []string       `json:"stop_sequences,omitempty"`
	Temperature   *float64       `json:"temperature,omitempty"`
	TopP          *float64       `json:"top_p,omitempty"`
	TopK          int            `json:"top_k,omitempty"`
	MetaData      map[string]any `json:"meta_data,omitempty"`
	Stream        bool           `json:"stream,omitempty"`
}

type CompleteResponse struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Completion string `json:"completion"`
	// possible values are: stop_sequence、max_tokens、null
	StopReason string `json:"stop_reason"`
	Model      string `json:"model"`
}

func (c *Client) CreateComplete(ctx context.Context, request CompleteRequest) (response CompleteResponse, err error) {
	urlSuffix := "/complete"
	req, err := c.newRequest(ctx, http.MethodPost, urlSuffix, request)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
