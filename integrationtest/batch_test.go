package integrationtest

import (
	"context"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
)

func TestIntegrationBatch(t *testing.T) {
	testAPIKey(t)
	client := anthropic.NewClient(APIKey)
	ctx := context.Background()

	request := anthropic.BatchRequest{
		Requests: []anthropic.InnerRequests{
			{
				CustomId: "fake-id-1",
				Params: anthropic.MessagesRequest{
					Model:       anthropic.ModelClaude3Haiku20240307,
					MultiSystem: anthropic.NewMultiSystemMessages("you are an assistant", "you are snarky"),
					MaxTokens:   10,
					Messages: []anthropic.Message{
						anthropic.NewUserTextMessage("What is your name?"),
						anthropic.NewAssistantTextMessage("My name is Claude."),
						anthropic.NewUserTextMessage("What is your favorite color?"),
					},
				},
			},
		},
	}

	t.Run("CreateMessages on real API", func(t *testing.T) {
		betaOpts := anthropic.WithBetaVersion(
			anthropic.BetaTools20240404,
			anthropic.BetaMaxTokens35Sonnet20240715,
			anthropic.BetaMessageBatches20240924,
		)
		newClient := anthropic.NewClient(APIKey, betaOpts)

		resp, err := newClient.CreateBatch(ctx, request)
		if err != nil {
			t.Fatalf("CreateBatch error: %s", err)
		}
		t.Logf("CreateBatch resp: %+v", resp)

		t.Run("RateLimitHeaders are present", func(t *testing.T) {
			rateLimHeader, err := resp.GetRateLimitHeaders()
			if err != nil {
				t.Fatalf("GetRateLimitHeaders error: %s", err)
			}
			t.Logf("RateLimitHeaders: %+v", rateLimHeader)
		})
	})

	t.Run("RetrieveBatch on real API", func(t *testing.T) {
		resp, err := client.RetrieveBatch(ctx, "fake-id-1")
		if err != nil {
			t.Fatalf("RetrieveBatch error: %s", err)
		}
		t.Logf("RetrieveBatch resp: %+v", resp)

		t.Run("RateLimitHeaders are present", func(t *testing.T) {
			rateLimHeader, err := resp.GetRateLimitHeaders()
			if err != nil {
				t.Fatalf("GetRateLimitHeaders error: %s", err)
			}
			t.Logf("RateLimitHeaders: %+v", rateLimHeader)
		})
	})

	t.Run("RetrieveBatchResults on real API", func(t *testing.T) {
		resp, err := client.RetrieveBatchResults(ctx, "fake-id-1")
		if err != nil {
			t.Fatalf("RetrieveBatchResults error: %s", err)
		}
		t.Logf("RetrieveBatchResults resp: %+v", resp)

		t.Run("RateLimitHeaders are present", func(t *testing.T) {
			rateLimHeader, err := resp.GetRateLimitHeaders()
			if err != nil {
				t.Fatalf("GetRateLimitHeaders error: %s", err)
			}
			t.Logf("RateLimitHeaders: %+v", rateLimHeader)
		})
	})

	t.Run("ListBatches on real API", func(t *testing.T) {
		lbr := anthropic.ListBatchRequest{} // todo test args
		resp, err := client.ListBatches(ctx, lbr)
		if err != nil {
			t.Fatalf("ListBatches error: %s", err)
		}
		t.Logf("ListBatches resp: %+v", resp)

		t.Run("RateLimitHeaders are present", func(t *testing.T) {
			rateLimHeader, err := resp.GetRateLimitHeaders()
			if err != nil {
				t.Fatalf("GetRateLimitHeaders error: %s", err)
			}
			t.Logf("RateLimitHeaders: %+v", rateLimHeader)
		})
	})

	t.Run("CancelBatch on real API", func(t *testing.T) {
		resp, err := client.CancelBatch(ctx, "fake-id-1")
		if err != nil {
			t.Fatalf("CancelBatch error: %s", err)
		}
		t.Logf("CancelBatch resp: %+v", resp)

		t.Run("RateLimitHeaders are present", func(t *testing.T) {
			rateLimHeader, err := resp.GetRateLimitHeaders()
			if err != nil {
				t.Fatalf("GetRateLimitHeaders error: %s", err)
			}
			t.Logf("RateLimitHeaders: %+v", rateLimHeader)
		})
	})
}
