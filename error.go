package anthropic

import "fmt"

// APIError provides error information returned by the Anthropic API.
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// RequestError provides information about generic request errors.
type RequestError struct {
	StatusCode int
	Err        error
}

type ErrorResponse struct {
	Type  string    `json:"type"`
	Error *APIError `json:"error,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("anthropic api error type: %s, message: %s", e.Type, e.Message)
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("anthropic request error status code: %d, err: %s", e.StatusCode, e.Err)
}

func (e *RequestError) Unwrap() error {
	return e.Err
}
