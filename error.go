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
	return e.Message
}

func (e *RequestError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("anthropic request error: status code %d", e.StatusCode)
}

func (e *RequestError) Unwrap() error {
	return e.Err
}
