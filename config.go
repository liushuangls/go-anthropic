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
	ApiKey     string
	BaseURL    string
	APIVersion string
	HTTPClient *http.Client

	EmptyMessagesLimit uint
}

func DefaultConfig(apikey string) ClientConfig {
	return ClientConfig{
		ApiKey:     apikey,
		BaseURL:    anthropicAPIURLv1,
		APIVersion: APIVersion20230601,
		HTTPClient: &http.Client{},

		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}
}
