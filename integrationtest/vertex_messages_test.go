package integrationtest

import (
	"context"
	"fmt"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
	"golang.org/x/oauth2/google"
)

func TestVertexIntegrationMessages(t *testing.T) {
	testVertexAPIKey(t)

	ts, err := google.JWTAccessTokenSourceWithScope(
		[]byte(VertexAPIKey),
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/cloud-platform.read-only",
	)
	if err != nil {
		fmt.Println("Error creating token source")
		return
	}

	// use JWTAccessTokenSourceWithScope
	token, err := ts.Token()
	if err != nil {
		fmt.Println("Error getting token")
		return
	}

	client := anthropic.NewClient(
		token.AccessToken,
		anthropic.WithVertexAI(VertexAPIProject, VertexAPILocation),
	)
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
		// betaOpts := anthropic.WithBetaVersion(
		// 	anthropic.BetaTools20240404,
		// 	anthropic.BetaMaxTokens35Sonnet20240715,
		// )
		newClient := anthropic.NewClient(
			token.AccessToken,
			anthropic.WithVertexAI(VertexAPIProject, VertexAPILocation),
		)

		resp, err := newClient.CreateMessages(ctx, request)
		if err != nil {
			t.Fatalf("CreateMessages error: %s", err)
		}
		t.Logf("CreateMessages resp: %+v", resp)
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
	})
}
