package integrationtest

import (
	"net/http"

	"github.com/henvic/httpretty"
)

func NewHttpPrettyClient(responseBody bool) *http.Client {
	logger := &httpretty.Logger{
		Time:           true,
		TLS:            true,
		RequestHeader:  true,
		RequestBody:    true,
		ResponseHeader: true,
		ResponseBody:   responseBody,
		Colors:         true,
		Formatters:     []httpretty.Formatter{&httpretty.JSONFormatter{}},
	}

	return &http.Client{
		Transport: logger.RoundTripper(http.DefaultTransport),
	}
}
