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

	// The number of tokens written to the cache when creating a new entry.
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	// The number of tokens retrieved from the cache for associated request.
	CacheReadInputTokens int `json:"cache_read_input_tokens"`
}

func newRateLimitHeaders(h http.Header) (RateLimitHeaders, error) {
	errs := []error{}

	requestsLimit, err := strconv.Atoi(h.Get("anthropic-ratelimit-requests-limit"))
	errs = append(errs, err)
	requestsRemaining, err := strconv.Atoi(h.Get("anthropic-ratelimit-requests-remaining"))
	errs = append(errs, err)
	requestsReset, err := time.Parse(time.RFC3339, h.Get("anthropic-ratelimit-requests-reset"))
	errs = append(errs, err)

	tokensLimit, err := strconv.Atoi(h.Get("anthropic-ratelimit-tokens-limit"))
	errs = append(errs, err)
	tokensRemaining, err := strconv.Atoi(h.Get("anthropic-ratelimit-tokens-remaining"))
	errs = append(errs, err)
	tokensReset, err := time.Parse(time.RFC3339, h.Get("anthropic-ratelimit-tokens-reset"))
	errs = append(errs, err)

	retryAfter, err := strconv.Atoi(h.Get("retry-after"))
	errs = append(errs, err)

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
