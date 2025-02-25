package integrationtest

import (
	"context"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
)

func printMessagesContent(t *testing.T, cs []anthropic.MessageContent) {
	for i, content := range cs {
		printMessageContent(t, i, content)
	}
}

func printMessageContent(t *testing.T, i int, content anthropic.MessageContent) {
	switch content.Type {
	case anthropic.MessagesContentTypeText, anthropic.MessagesContentTypeTextDelta:
		t.Logf("CreateMessages resp text content[%d]: %s", i, *content.Text)
	case anthropic.MessagesContentTypeThinking, anthropic.MessagesContentTypeThinkingDelta,
		anthropic.MessagesContentTypeSignatureDelta:
		t.Logf(
			"CreateMessages resp thking content[%d]: %+v",
			i,
			content.MessageContentThinking,
		)
	default:
		t.Logf("CreateMessages resp %s content[%d]: %+v", content.Type, i, content)
	}
}

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

	t.Run("CreateMessages on real API with claude-3-7-sonnet", func(t *testing.T) {
		newClient := anthropic.NewClient(APIKey)
		req := request
		req.Model = anthropic.ModelClaude3Dot7Sonnet20250219
		req.MaxTokens = 2048
		req.Thinking = &anthropic.Thinking{
			Type:         anthropic.ThinkingTypeEnabled,
			BudgetTokens: 1024,
		}

		resp, err := newClient.CreateMessages(ctx, req)
		if err != nil {
			t.Fatalf("CreateMessages error: %s", err)
		}
		t.Logf("CreateMessages resp: %+v", resp)
		printMessagesContent(t, resp.Content)

		t.Run("RateLimitHeaders are present", func(t *testing.T) {
			rateLimHeader, err := resp.GetRateLimitHeaders()
			if err != nil {
				t.Fatalf("GetRateLimitHeaders error: %s", err)
			}
			t.Logf("RateLimitHeaders: %+v", rateLimHeader)
		})
	})

	t.Run("CreateMessagesStream on real API with claude-3-7-sonnet", func(t *testing.T) {
		newClient := anthropic.NewClient(APIKey)
		var req anthropic.MessagesStreamRequest
		req.MessagesRequest = request
		req.Model = anthropic.ModelClaude3Dot7Sonnet20250219
		req.MaxTokens = 2048
		req.Stream = true
		req.Thinking = &anthropic.Thinking{
			Type:         anthropic.ThinkingTypeEnabled,
			BudgetTokens: 1024,
		}
		req.OnContentBlockStart = func(data anthropic.MessagesEventContentBlockStartData) {
			t.Logf("OnContentBlockStart: ")
			printMessageContent(t, 0, data.ContentBlock)
		}
		req.OnContentBlockDelta = func(data anthropic.MessagesEventContentBlockDeltaData) {
			t.Logf("OnContentBlockDelta: ")
			printMessageContent(t, 0, data.Delta)
		}

		resp, err := newClient.CreateMessagesStream(ctx, req)
		if err != nil {
			t.Fatalf("CreateMessagesStream error: %s", err)
		}
		printMessagesContent(t, resp.Content)

		t.Run("RateLimitHeaders are present", func(t *testing.T) {
			rateLimHeader, err := resp.GetRateLimitHeaders()
			if err != nil {
				t.Fatalf("GetRateLimitHeaders error: %s", err)
			}
			t.Logf("RateLimitHeaders: %+v", rateLimHeader)
		})
	})
}

func TestComputerUse(t *testing.T) {
	testAPIKey(t)
	client := anthropic.NewClient(
		APIKey,
		anthropic.WithBetaVersion(anthropic.BetaComputerUse20241022),
	)
	ctx := context.Background()
	var temperature float32 = 0.0

	req := anthropic.MessagesRequest{
		Model:     anthropic.ModelClaude3Dot5SonnetLatest,
		MaxTokens: 4096,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage("use chrome to google today's weather"),
		},
		Tools: []anthropic.ToolDefinition{
			anthropic.NewComputerUseToolDefinition("computer", 1024, 768, nil),
		},
		Temperature: &temperature,
	}

	t.Run("CreateMessages on real API with computer use", func(t *testing.T) {
		resp, err := client.CreateMessages(ctx, req)
		if err != nil {
			t.Fatalf("CreateMessages error: %s", err)
		}
		t.Logf("CreateMessages resp: %+v", resp)
		t.Logf("CreateMessages resp content: %s", resp.GetFirstContentText())
	})
}
