package integrationtest

import (
	"context"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
)

func TestIntegrationComplete(t *testing.T) {
	testAPIKey(t)

	t.Run("CreateComplete on real API", func(t *testing.T) {
		client := anthropic.NewClient(APIKey)
		ctx := context.Background()

		request := anthropic.CompleteRequest{
			Model:             anthropic.ModelClaudeInstant1Dot2,
			Prompt:            "\n\nHuman: What is your name?\n\nAssistant:",
			MaxTokensToSample: 1000,
		}

		resp, err := client.CreateComplete(ctx, request)
		if err != nil {
			t.Fatalf("CreateComplete error: %s", err)
		}
		t.Logf("CreateComplete resp: %+v", resp)

		// RateLimitHeaders are not present on the Completions endpoint
	})
}
