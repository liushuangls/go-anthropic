package anthropic

import (
	"net/http"
	"testing"

	"github.com/liushuangls/go-anthropic/v2/internal/test"
)

func TestWithBetaVersion(t *testing.T) {
	is := test.NewRequire(t)
	t.Run("single beta version", func(t *testing.T) {
		opt := withBetaVersion("fake-version")
		request, err := http.NewRequest("GET", "http://example.com", nil)
		is.NoError(err)

		opt(request)

		is.Equal("fake-version", request.Header.Get("anthropic-beta"))
	})

	t.Run("multiple beta versions", func(t *testing.T) {
		opt := withBetaVersion("fake-version1", "fake-version2")
		request, err := http.NewRequest("GET", "http://example.com", nil)
		is.NoError(err)

		opt(request)

		is.Equal("fake-version1,fake-version2", request.Header.Get("anthropic-beta"))
	})
}
