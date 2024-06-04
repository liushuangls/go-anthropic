package anthropic

import (
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

func newRateLimitHeaders(h http.Header) RateLimitHeaders {
	requestsLimit, _ := strconv.Atoi(h.Get("anthropic-ratelimit-requests-limit"))
	requestsRemaining, _ := strconv.Atoi(h.Get("anthropic-ratelimit-requests-remaining"))
	requestsReset, _ := time.Parse(time.RFC3339, h.Get("anthropic-ratelimit-requests-reset"))

	tokensLimit, _ := strconv.Atoi(h.Get("anthropic-ratelimit-tokens-limit"))
	tokensRemaining, _ := strconv.Atoi(h.Get("anthropic-ratelimit-tokens-remaining"))
	tokensReset, _ := time.Parse(time.RFC3339, h.Get("anthropic-ratelimit-tokens-reset"))

	retryAfter, _ := strconv.Atoi(h.Get("anthropic-ratelimit-retry-after"))

	return RateLimitHeaders{
		RequestsLimit:     requestsLimit,
		RequestsRemaining: requestsRemaining,
		RequestsReset:     requestsReset,
		TokensLimit:       tokensLimit,
		TokensRemaining:   tokensRemaining,
		TokensReset:       tokensReset,
		RetryAfter:        retryAfter,
	}
}
