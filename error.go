package anthropic

import "fmt"

const (
	// InvalidRequestErr There was an issue with the format or content of your request.
	InvalidRequestErr = "invalid_request_error"
	// AuthenticationErr There's an issue with your API key.
	AuthenticationErr = "authentication_error"
	// PermissionErr Your API key does not have permission to use the specified resource.
	PermissionErr = "permission_error"
	// NotFoundErr The requested resource was not found.
	NotFoundErr = "not_found_error"
	// RateLimitErr Your account has hit a rate limit.
	RateLimitErr = "rate_limit_error"
	// ApiErr An unexpected error has occurred internal to Anthropic's systems.
	ApiErr = "api_error"
	// OverloadedErr Anthropic's API is temporarily overloaded.
	OverloadedErr = "overloaded_error"
)

// APIError provides error information returned by the Anthropic API.
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e *APIError) IsInvalidRequestErr() bool {
	return e.Type == InvalidRequestErr
}

func (e *APIError) AuthenticationErr() bool {
	return e.Type == AuthenticationErr
}

func (e *APIError) PermissionErr() bool {
	return e.Type == PermissionErr
}

func (e *APIError) NotFoundErr() bool {
	return e.Type == NotFoundErr
}

func (e *APIError) RateLimitErr() bool {
	return e.Type == RateLimitErr
}

func (e *APIError) ApiErr() bool {
	return e.Type == ApiErr
}

func (e *APIError) OverloadedErr() bool {
	return e.Type == OverloadedErr
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
