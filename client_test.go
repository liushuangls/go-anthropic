package anthropic

import (
	"net/http"
	"testing"
)

func TestWithBetaVersion(t *testing.T) {
	t.Run("single beta version", func(t *testing.T) {
		opt := withBetaVersion("fake-version")
		request, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("http.NewRequest error: %s", err)
		}
		opt(request)

		if req := request.Header.Get("anthropic-beta"); req != "fake-version" {
			t.Errorf("unexpected BetaVersion: %s", req)
		}
	})

	t.Run("multiple beta versions", func(t *testing.T) {
		opt := withBetaVersion("fake-version1", "fake-version2")
		request, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("http.NewRequest error: %s", err)
		}
		opt(request)

		if req := request.Header.Get("anthropic-beta"); req != "fake-version1,fake-version2" {
			t.Errorf("unexpected BetaVersion: %s", req)
		}
	})
}
