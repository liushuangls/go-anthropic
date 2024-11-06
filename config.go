package anthropic

import (
	"net/http"
)

const (
	anthropicAPIURLv1              = "https://api.anthropic.com/v1"
	defaultEmptyMessagesLimit uint = 300
)

type APIVersion string

const (
	APIVersion20230601 APIVersion = "2023-06-01"
)

type BetaVersion string

const (
	BetaTools20240404             BetaVersion = "tools-2024-04-04"
	BetaTools20240516             BetaVersion = "tools-2024-05-16"
	BetaPromptCaching20240731     BetaVersion = "prompt-caching-2024-07-31"
	BetaMessageBatches20240924    BetaVersion = "message-batches-2024-09-24"
	BetaTokenCounting20241101     BetaVersion = "token-counting-2024-11-01"
	BetaMaxTokens35Sonnet20240715 BetaVersion = "max-tokens-3-5-sonnet-2024-07-15"
)

// ClientConfig is a configuration of a client.
type ClientConfig struct {
	apiKey string

	BaseURL     string
	APIVersion  APIVersion
	BetaVersion []BetaVersion
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

func WithAPIVersion(apiVersion APIVersion) ClientOption {
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

func WithBetaVersion(betaVersion ...BetaVersion) ClientOption {
	return func(c *ClientConfig) {
		c.BetaVersion = betaVersion
	}
}
