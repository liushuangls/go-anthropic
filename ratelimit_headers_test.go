package anthropic

import (
	"net/http"
	"testing"
)

func TestNewRateLimitHeadersTokensOptional(t *testing.T) {
	// A valid response (e.g. count_tokens / batch / split-token) that omits the
	// combined anthropic-ratelimit-tokens-* headers must not produce an error.
	h := http.Header{}
	h.Set("anthropic-ratelimit-requests-limit", "1000")
	h.Set("anthropic-ratelimit-requests-remaining", "999")
	h.Set("anthropic-ratelimit-requests-reset", "2022-01-01T00:00:00Z")
	h.Set("anthropic-ratelimit-input-tokens-limit", "2000")
	h.Set("anthropic-ratelimit-output-tokens-limit", "3000")

	headers, err := newRateLimitHeaders(h)
	if err != nil {
		t.Fatalf("expected no error when tokens-* headers are absent, got: %s", err)
	}
	if headers.RequestsLimit != 1000 {
		t.Fatalf("expected RequestsLimit 1000, got %d", headers.RequestsLimit)
	}
}
