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

// ClientConfig is a configuration of a client.
type ClientConfig struct {
	apikey string

	BaseURL    string
	APIVersion string
	HTTPClient *http.Client

	EmptyMessagesLimit uint
}

type ClientOption func(c *ClientConfig)

func NewConfig(apikey string, opts ...ClientOption) ClientConfig {
	c := ClientConfig{
		apikey: apikey,

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
