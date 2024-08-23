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
	errs := []error{}

	// Requests
	requestsLimit, err := strconv.Atoi(h.Get("anthropic-ratelimit-requests-limit"))
	if err != nil {
		err = fmt.Errorf("failed to parse anthropic-ratelimit-requests-limit: %w", err)
		errs = append(errs, err)
	}
	requestsRemaining, err := strconv.Atoi(h.Get("anthropic-ratelimit-requests-remaining"))
	if err != nil {
		err = fmt.Errorf("failed to parse anthropic-ratelimit-requests-remaining: %w", err)
		errs = append(errs, err)
	}
	requestsReset, err := time.Parse(time.RFC3339, h.Get("anthropic-ratelimit-requests-reset"))
	if err != nil {
		err = fmt.Errorf("failed to parse anthropic-ratelimit-requests-reset: %w", err)
		errs = append(errs, err)
	}

	// Tokens
	tokensLimit, err := strconv.Atoi(h.Get("anthropic-ratelimit-tokens-limit"))
	if err != nil {
		err = fmt.Errorf("failed to parse anthropropic-ratelimit-tokens-limit: %w", err)
		errs = append(errs, err)
	}
	tokensRemaining, err := strconv.Atoi(h.Get("anthropic-ratelimit-tokens-remaining"))
	if err != nil {
		err = fmt.Errorf("failed to parse anthropropic-ratelimit-tokens-remaining: %w", err)
		errs = append(errs, err)
	}
	tokensReset, err := time.Parse(time.RFC3339, h.Get("anthropic-ratelimit-tokens-reset"))
	errs = append(errs, err)

	// RetryAfter
	retryAfter, err := strconv.Atoi(h.Get("retry-after"))
	if err != nil {
		retryAfter = -1
	}

	headers := RateLimitHeaders{
		RequestsLimit:     requestsLimit,
		RequestsRemaining: requestsRemaining,
		RequestsReset:     requestsReset,
		TokensLimit:       tokensLimit,
		TokensRemaining:   tokensRemaining,
		TokensReset:       tokensReset,
		RetryAfter:        retryAfter,
	}

	for _, e := range errs {
		if e != nil {
			return headers, e
		}
	}
	return headers, nil
}
