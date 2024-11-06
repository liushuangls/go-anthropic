package integrationtest

import (
	"context"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
)

func TestCountTokens(t *testing.T) {
	testAPIKey(t)
	client := anthropic.NewClient(
		APIKey,
		anthropic.WithBetaVersion(anthropic.BetaTokenCounting20241101),
	)
	ctx := context.Background()

	request := anthropic.MessagesRequest{
		Model:       anthropic.ModelClaude3Dot5HaikuLatest,
		MultiSystem: anthropic.NewMultiSystemMessages("you are an assistant", "you are snarky"),
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage("What is your name?"),
			anthropic.NewAssistantTextMessage("My name is Claude."),
			anthropic.NewUserTextMessage("What is your favorite color?"),
		},
	}

	t.Run("CountTokens on real API", func(t *testing.T) {
		resp, err := client.CountTokens(ctx, request)
		if err != nil {
			t.Fatalf("CountTokens error: %s", err)
		}
		t.Logf("CountTokens resp: %+v", resp)
	})
}
