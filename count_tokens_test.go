package anthropic_test

import (
	"context"
	"encoding/json"
	"io"
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

	t.Run("count tokens omits max_tokens even when set", func(t *testing.T) {
		// The count_tokens endpoint rejects max_tokens. Even if the caller
		// populates MaxTokens on the MessagesRequest, the SDK must not send
		// it. The mock handler 400s if the raw body carries max_tokens.
		req := request
		req.MaxTokens = 10
		resp, err := client.CountTokens(context.Background(), req)
		if err != nil {
			t.Fatalf("CountTokens error: %v", err)
		}

		t.Logf("CountTokens resp: %+v", resp)
	})
}

func handleCountTokens(w http.ResponseWriter, r *http.Request) {
	var resBytes []byte

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}

	// The count_tokens endpoint rejects extra inputs such as max_tokens.
	// Inspect the raw bytes so an always-present "max_tokens":0 cannot slip
	// through. The endpoint returns 400 "Extra inputs are not permitted".
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &raw); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}
	if _, ok := raw["max_tokens"]; ok {
		http.Error(w, "max_tokens: Extra inputs are not permitted", http.StatusBadRequest)
		return
	}
	if _, ok := raw["model"]; !ok {
		http.Error(w, "model: Field required", http.StatusBadRequest)
		return
	}
	if _, ok := raw["messages"]; !ok {
		http.Error(w, "messages: Field required", http.StatusBadRequest)
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
