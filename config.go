package anthropic

import (
	"net/http"
)

const (
	anthropicAPIURLv1              = "https://api.anthropic.com/v1"
	defaultEmptyMessagesLimit uint = 300
)

const (
	APIVersion20230601 = "2023-06-01"
)

const (
	BetaTools20240404         = "tools-2024-04-04"
	BetaTools20240516         = "tools-2024-05-16"
	BetaPromptCaching20240731 = "prompt-caching-2024-07-31"

	BetaMaxTokens35Sonnet20240715 = "max-tokens-3-5-sonnet-2024-07-15"
)

// ClientConfig is a configuration of a client.
type ClientConfig struct {
	apiKey string

	BaseURL     string
	APIVersion  string
	BetaVersion string
	HTTPClient  *http.Client

	EmptyMessagesLimit uint
}

type ClientOption func(c *ClientConfig)

func newConfig(apiKey string, opts ...ClientOption) ClientConfig {
	c := ClientConfig{
		apiKey: apiKey,

		BaseURL:    anthropicAPIURLv1,
		APIVersion: APIVersion20230601,
		HTTPClient: &http.Client{},

		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

func WithBaseURL(baseUrl string) ClientOption {
	return func(c *ClientConfig) {
		c.BaseURL = baseUrl
	}
}

func WithAPIVersion(apiVersion string) ClientOption {
	return func(c *ClientConfig) {
		c.APIVersion = apiVersion
	}
}

func WithHTTPClient(cli *http.Client) ClientOption {
	return func(c *ClientConfig) {
		c.HTTPClient = cli
	}
}

func WithEmptyMessagesLimit(limit uint) ClientOption {
	return func(c *ClientConfig) {
		c.EmptyMessagesLimit = limit
	}
}

func WithBetaVersion(betaVersion string) ClientOption {
	return func(c *ClientConfig) {
		c.BetaVersion = betaVersion
	}
}
