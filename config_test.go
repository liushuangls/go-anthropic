package anthropic_test

import (
	"net/http"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/stretchr/testify/require"
)

func TestWithBaseURL(t *testing.T) {
	fakeBaseURL := "fake-url!"
	opt := anthropic.WithBaseURL(fakeBaseURL)

	c := anthropic.ClientConfig{}
	opt(&c)

	if c.BaseURL != fakeBaseURL {
		t.Errorf("unexpected BaseURL: %s", c.BaseURL)
	}
}

func TestWithAPIVersion(t *testing.T) {
	is := require.New(t)
	t.Run("single generic version", func(t *testing.T) {
		fakeAPIVersion := anthropic.APIVersion("fake-version")
		opt := anthropic.WithAPIVersion(fakeAPIVersion)

		c := anthropic.ClientConfig{}
		opt(&c)

		is.Equal(fakeAPIVersion, c.APIVersion)
	})
}

func TestWithHTTPClient(t *testing.T) {
	is := require.New(t)
	fakeHTTPClient := http.Client{}
	fakeHTTPClient.Timeout = 1234

	opt := anthropic.WithHTTPClient(&fakeHTTPClient)

	c := anthropic.ClientConfig{}
	opt(&c)

	is.Equal(&fakeHTTPClient, c.HTTPClient)
}

func TestWithEmptyMessagesLimit(t *testing.T) {
	is := require.New(t)
	fakeLimit := uint(1234)
	opt := anthropic.WithEmptyMessagesLimit(fakeLimit)

	c := anthropic.ClientConfig{}
	opt(&c)

	is.Equal(fakeLimit, c.EmptyMessagesLimit)
}

func TestWithBetaVersion(t *testing.T) {
	is := require.New(t)
	t.Run("single generic version", func(t *testing.T) {
		fakeBetaVersion := anthropic.BetaVersion("fake-version")
		opt := anthropic.WithBetaVersion(fakeBetaVersion)
		c := anthropic.ClientConfig{}
		opt(&c)

		is.Equal(fakeBetaVersion, c.BetaVersion[0])
		is.Equal(1, len(c.BetaVersion))
	})

	t.Run("multiple versions", func(t *testing.T) {
		opt := anthropic.WithBetaVersion("foo", "bar")
		c := anthropic.ClientConfig{}
		opt(&c)

		is.Equal(2, len(c.BetaVersion))
		is.Equal(anthropic.BetaVersion("foo"), c.BetaVersion[0])
		is.Equal(anthropic.BetaVersion("bar"), c.BetaVersion[1])
	})
}
