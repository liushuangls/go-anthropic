package anthropic

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithBetaVersion(t *testing.T) {
	is := require.New(t)
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
