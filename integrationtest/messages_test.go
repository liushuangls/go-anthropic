package integrationtest

import (
	"context"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
)

func TestIntegrationMessages(t *testing.T) {
	testAPIKey(t)
	client := anthropic.NewClient(APIKey)
	ctx := context.Background()

	request := anthropic.MessagesRequest{
		Model:       anthropic.ModelClaude3Haiku20240307,
		MultiSystem: anthropic.NewMultiSystemMessages("you are an assistant", "you are snarky"),
		MaxTokens:   10,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage("What is your name?"),
			anthropic.NewAssistantTextMessage("My name is Claude."),
			anthropic.NewUserTextMessage("What is your favorite color?"),
		},
	}

	t.Run("CreateMessages on real API", func(t *testing.T) {
		betaOpts := anthropic.WithBetaVersion(
			anthropic.BetaTools20240404,
			anthropic.BetaMaxTokens35Sonnet20240715,
		)
		newClient := anthropic.NewClient(APIKey, betaOpts)

		resp, err := newClient.CreateMessages(ctx, request)
		if err != nil {
			t.Fatalf("CreateMessages error: %s", err)
		}
		t.Logf("CreateMessages resp: %+v", resp)

		t.Run("RateLimitHeaders are present", func(t *testing.T) {
			rateLimHeader, err := resp.GetRateLimitHeaders()
			if err != nil {
				t.Fatalf("GetRateLimitHeaders error: %s", err)
			}
			t.Logf("RateLimitHeaders: %+v", rateLimHeader)
		})
	})

	t.Run("CreateMessagesStream on real API", func(t *testing.T) {
		streamRequest := anthropic.MessagesStreamRequest{
			MessagesRequest: request,
		}
		resp, err := client.CreateMessagesStream(ctx, streamRequest)
		if err != nil {
			t.Fatalf("CreateMessagesStream error: %s", err)
		}
		t.Logf("CreateMessagesStream resp: %+v", resp)

		t.Run("RateLimitHeaders are present", func(t *testing.T) {
			rateLimHeader, err := resp.GetRateLimitHeaders()
			if err != nil {
				t.Fatalf("GetRateLimitHeaders error: %s", err)
			}
			t.Logf("RateLimitHeaders: %+v", rateLimHeader)
		})
	})

	t.Run("CreateMessages on real API with claude-3-5-haiku", func(t *testing.T) {
		newClient := anthropic.NewClient(APIKey)
		req := request
		req.Model = anthropic.ModelClaude3Dot5Haiku20241022

		resp, err := newClient.CreateMessages(ctx, req)
		if err != nil {
			t.Fatalf("CreateMessages error: %s", err)
		}
		t.Logf("CreateMessages resp: %+v", resp)
		t.Logf("CreteMessages resp content: %s", resp.GetFirstContentText())

		t.Run("RateLimitHeaders are present", func(t *testing.T) {
			rateLimHeader, err := resp.GetRateLimitHeaders()
			if err != nil {
				t.Fatalf("GetRateLimitHeaders error: %s", err)
			}
			t.Logf("RateLimitHeaders: %+v", rateLimHeader)
		})
	})
}
