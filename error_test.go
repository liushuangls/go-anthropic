package anthropic_test

import (
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/stretchr/testify/require"
)

func TestIsXError(t *testing.T) {
	is := require.New(t)
	countBool := func(bools []bool) int {
		count := 0
		for _, b := range bools {
			if b {
				count++
			}
		}
		return count
	}

	errTypes := []anthropic.ErrType{
		anthropic.ErrTypeInvalidRequest,
		anthropic.ErrTypeAuthentication,
		anthropic.ErrTypePermission,
		anthropic.ErrTypeNotFound,
		anthropic.ErrTypeTooLarge,
		anthropic.ErrTypeRateLimit,
		anthropic.ErrTypeApi,
		anthropic.ErrTypeOverloaded,
	}
	isErrFuncs := func(e anthropic.APIError) []bool {
		return []bool{
			e.IsInvalidRequestErr(),
			e.IsAuthenticationErr(),
			e.IsPermissionErr(),
			e.IsNotFoundErr(),
			e.IsTooLargeErr(),
			e.IsRateLimitErr(),
			e.IsApiErr(),
			e.IsOverloadedErr(),
		}
	}

	apiErrors := []anthropic.APIError{}
	for _, errType := range errTypes {
		apiErrors = append(apiErrors, anthropic.APIError{
			Type:    errType,
			Message: "fake message",
		})
	}

	for i, e := range apiErrors {
		isErrorType := isErrFuncs(e)

		// Expect only one error type to be true for each error
		numErrorType := countBool(isErrorType)
		is.Equal(1, numErrorType, "Expected 1 error type to be true, got %d, for error %T", numErrorType, i)

		// Expect the error type to be true for the correct error
		is.True(isErrorType[i], "Expected error type %T to be true, got false", e)
	}
}
