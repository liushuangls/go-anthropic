package integrationtest

import (
	"context"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/liushuangls/go-anthropic/v2/internal/test"
)

func TestIntegrationComplete(t *testing.T) {
	testAPIKey(t)
	is := test.NewRequire(t)

	t.Run("CreateComplete on real API", func(t *testing.T) {
		client := anthropic.NewClient(APIKey)
		ctx := context.Background()

		request := anthropic.CompleteRequest{
			Model:             anthropic.ModelClaudeInstant1Dot2,
			Prompt:            "\n\nHuman: What is your name?\n\nAssistant:",
			MaxTokensToSample: 10,
		}

		resp, err := client.CreateComplete(ctx, request)
		is.NoError(err)

		t.Logf("CreateComplete resp: %+v", resp)

		// RateLimitHeaders are not present on the Completions endpoint
	})
}
