package anthropic_test

import (
	"net/http"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
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
	t.Run("single generic version", func(t *testing.T) {
		fakeAPIVersion := anthropic.APIVersion("fake-version")
		opt := anthropic.WithAPIVersion(fakeAPIVersion)

		c := anthropic.ClientConfig{}
		opt(&c)

		if c.APIVersion != fakeAPIVersion {
			t.Errorf("unexpected APIVersion: %s", c.APIVersion)
		}
	})
}

func TestWithHTTPClient(t *testing.T) {
	fakeHTTPClient := http.Client{}
	fakeHTTPClient.Timeout = 1234

	opt := anthropic.WithHTTPClient(&fakeHTTPClient)

	c := anthropic.ClientConfig{}
	opt(&c)

	if c.HTTPClient != &fakeHTTPClient {
		t.Errorf("unexpected HTTPClient: %v", c.HTTPClient)
	}
}

func TestWithEmptyMessagesLimit(t *testing.T) {
	fakeLimit := uint(1234)
	opt := anthropic.WithEmptyMessagesLimit(fakeLimit)

	c := anthropic.ClientConfig{}
	opt(&c)

	if c.EmptyMessagesLimit != fakeLimit {
		t.Errorf("unexpected EmptyMessagesLimit: %d", c.EmptyMessagesLimit)
	}
}

func TestWithBetaVersion(t *testing.T) {
	t.Run("single generic version", func(t *testing.T) {
		fakeBetaVersion := anthropic.BetaVersion("fake-version")
		opt := anthropic.WithBetaVersion(fakeBetaVersion)
		c := anthropic.ClientConfig{}
		opt(&c)

		if c.BetaVersion[0] != fakeBetaVersion {
			t.Errorf("unexpected BetaVersion: %s", c.BetaVersion)
		}
		if len(c.BetaVersion) != 1 {
			t.Errorf("unexpected BetaVersion length: %d", len(c.BetaVersion))
		}
	})

	t.Run("multiple versions", func(t *testing.T) {
		opt := anthropic.WithBetaVersion("foo", "bar")
		c := anthropic.ClientConfig{}
		opt(&c)

		if len(c.BetaVersion) != 2 {
			t.Errorf("unexpected BetaVersion length: %d", len(c.BetaVersion))
		}
	})
}
