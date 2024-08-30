package anthropic

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type RateLimitHeaders struct {
	// The maximum number of requests allowed within the rate limit window.
	RequestsLimit int `json:"anthropic-ratelimit-requests-limit"`
	// The number of requests remaining within the current rate limit window.
	RequestsRemaining int `json:"anthropic-ratelimit-requests-remaining"`
	// The time when the request rate limit window will reset, provided in RFC 3339 format.
	RequestsReset time.Time `json:"anthropic-ratelimit-requests-reset"`
	// The maximum number of tokens allowed within the rate limit window.
	TokensLimit int `json:"anthropic-ratelimit-tokens-limit"`
	// The number of tokens remaining, rounded to the nearest thousand, within the current rate limit window.
	TokensRemaining int `json:"anthropic-ratelimit-tokens-remaining"`
	// The time when the token rate limit window will reset, provided in RFC 3339 format.
	TokensReset time.Time `json:"anthropic-ratelimit-tokens-reset"`
	// The number of seconds until the rate limit window resets.
	RetryAfter int `json:"retry-after"`
}

func newRateLimitHeaders(h http.Header) (RateLimitHeaders, error) {
	var errs []error

	parseHeader := func(key string) int {
		value, err := strconv.Atoi(h.Get(key))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to parse %s: %w", key, err))
			return 0
		}
		return value
	}

	parseTimeHeader := func(key string) time.Time {
		value, err := time.Parse(time.RFC3339, h.Get(key))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to parse %s: %w", key, err))
			return time.Time{}
		}
		return value
	}

	headers := RateLimitHeaders{}
	headers.RequestsLimit = parseHeader("anthropic-ratelimit-requests-limit")
	headers.RequestsRemaining = parseHeader("anthropic-ratelimit-requests-remaining")
	headers.RequestsReset = parseTimeHeader("anthropic-ratelimit-requests-reset")

	headers.TokensLimit = parseHeader("anthropic-ratelimit-tokens-limit")
	headers.TokensRemaining = parseHeader("anthropic-ratelimit-tokens-remaining")
	headers.TokensReset = parseTimeHeader("anthropic-ratelimit-tokens-reset")

	headers.RetryAfter = parseHeader("retry-after")
	if headers.RetryAfter == 0 && len(errs) > 0 {
		headers.RetryAfter = -1
	}

	if len(errs) > 0 {
		return headers, fmt.Errorf("multiple errors occurred: %v", errs)
	}

	return headers, nil
}
