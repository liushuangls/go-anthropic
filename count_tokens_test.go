package anthropic_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/liushuangls/go-anthropic/v2/internal/test"
)

func TestCountTokens(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages/count_tokens", handleCountTokens)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
		anthropic.WithBetaVersion(anthropic.BetaTokenCounting20241101),
	)

	request := anthropic.MessagesRequest{
		Model:       anthropic.ModelClaude3Dot5HaikuLatest,
		MultiSystem: anthropic.NewMultiSystemMessages("you are an assistant", "you are snarky"),
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage("What is your name?"),
			anthropic.NewAssistantTextMessage("My name is Claude."),
			anthropic.NewUserTextMessage("What is your favorite color?"),
		},
	}

	t.Run("count tokens success", func(t *testing.T) {
		resp, err := client.CountTokens(context.Background(), request)
		if err != nil {
			t.Fatalf("CountTokens error: %v", err)
		}

		t.Logf("CountTokens resp: %+v", resp)
	})

	t.Run("count tokens failure", func(t *testing.T) {
		request.MaxTokens = 10
		_, err := client.CountTokens(context.Background(), request)
		if err == nil {
			t.Fatalf("CountTokens expected error, got nil")
		}

		t.Logf("CountTokens error: %v", err)
	})
}

func handleCountTokens(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	var req anthropic.MessagesRequest
	if req, err = getRequest[anthropic.MessagesRequest](r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}
	if req.MaxTokens > 0 {
		http.Error(w, "max_tokens: Extra inputs are not permitted", http.StatusBadRequest)
		return
	}

	betaHeaders := r.Header.Get("Anthropic-Beta")
	if !strings.Contains(betaHeaders, string(anthropic.BetaTokenCounting20241101)) {
		http.Error(w, "missing beta version header", http.StatusBadRequest)
		return
	}

	res := anthropic.CountTokensResponse{
		InputTokens: 100,
	}
	resBytes, _ = json.Marshal(res)
	_, _ = w.Write(resBytes)
}
