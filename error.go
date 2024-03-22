package anthropic

import "fmt"

const (
	// ErrInvalidRequest There was an issue with the format or content of your request.
	ErrInvalidRequest = "invalid_request_error"
	// ErrAuthentication There's an issue with your API key.
	ErrAuthentication = "authentication_error"
	// ErrPermission Your API key does not have permission to use the specified resource.
	ErrPermission = "permission_error"
	// ErrNotFound The requested resource was not found.
	ErrNotFound = "not_found_error"
	// ErrRateLimit Your account has hit a rate limit.
	ErrRateLimit = "rate_limit_error"
	// ErrApi An unexpected error has occurred internal to Anthropic's systems.
	ErrApi = "api_error"
	// ErrOverloaded Anthropic's API is temporarily overloaded.
	ErrOverloaded = "overloaded_error"
)

// APIError provides error information returned by the Anthropic API.
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e *APIError) IsInvalidRequestErr() bool {
	return e.Type == ErrInvalidRequest
}

func (e *APIError) IsAuthenticationErr() bool {
	return e.Type == ErrAuthentication
}

func (e *APIError) IsPermissionErr() bool {
	return e.Type == ErrPermission
}

func (e *APIError) IsNotFoundErr() bool {
	return e.Type == ErrNotFound
}

func (e *APIError) IsRateLimitErr() bool {
	return e.Type == ErrRateLimit
}

func (e *APIError) IsApiErr() bool {
	return e.Type == ErrApi
}

func (e *APIError) IsOverloadedErr() bool {
	return e.Type == ErrOverloaded
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
