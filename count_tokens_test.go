package anthropic_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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

	t.Run("count tokens success with MultiSystem", func(t *testing.T) {
		resp, err := client.CountTokens(context.Background(), request)
		if err != nil {
			t.Fatalf("CountTokens error: %v", err)
		}

		t.Logf("CountTokens resp: %+v", resp)
	})

	t.Run("count tokens success with System", func(t *testing.T) {
		resp, err := client.CountTokens(context.Background(), anthropic.MessagesRequest{
			Model:  anthropic.ModelClaude3Dot5HaikuLatest,
			System: "you are an assistant and you are snarky",
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
				anthropic.NewAssistantTextMessage("My name is Claude."),
				anthropic.NewUserTextMessage("What is your favorite color?"),
			},
		})
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

func TestVertexCountTokens(t *testing.T) {
	project := "project"
	location := "location"

	baseEndpoint := fmt.Sprintf(
		"/v1/projects/%s/locations/%s/publishers/anthropic/models",
		project,
		location,
	)
	expectedPath := baseEndpoint + "/count-tokens:rawPredict"
	expectedModel := "claude-haiku-4-5@20251001"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != expectedPath {
			http.Error(w, "unexpected path: "+r.URL.Path, http.StatusNotFound)
			return
		}

		if got := r.Header.Get("Authorization"); got != "Bearer "+test.GetTestToken() {
			http.Error(w, "missing or invalid auth header", http.StatusUnauthorized)
			return
		}

		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "could not read request", http.StatusInternalServerError)
			return
		}

		var raw map[string]json.RawMessage
		if err := json.Unmarshal(rawBody, &raw); err != nil {
			http.Error(w, "could not parse request", http.StatusBadRequest)
			return
		}

		var model string
		if err := json.Unmarshal(raw["model"], &model); err != nil {
			http.Error(w, "missing or invalid model", http.StatusBadRequest)
			return
		}

		if model != expectedModel {
			http.Error(
				w,
				"unexpected model: "+model,
				http.StatusBadRequest,
			)
			return
		}

		if model == string(anthropic.ModelClaudeHaiku4Dot5V20251001) {
			http.Error(
				w,
				"snapshot model must be normalized for Vertex count-tokens",
				http.StatusBadRequest,
			)
			return
		}

		var version string
		if err := json.Unmarshal(raw["anthropic_version"], &version); err != nil {
			http.Error(w, "missing anthropic_version", http.StatusBadRequest)
			return
		}
		if version != string(anthropic.APIVersionVertex20231016) {
			http.Error(w, "unexpected anthropic_version: "+version, http.StatusBadRequest)
			return
		}

		if _, ok := raw["messages"]; !ok {
			http.Error(w, "messages: Field required", http.StatusBadRequest)
			return
		}

		if _, ok := raw["thinking"]; ok {
			http.Error(w, "thinking: Field must not be present", http.StatusBadRequest)
			return
		}

		res := anthropic.CountTokensResponse{InputTokens: 321}
		resBytes, _ := json.Marshal(res)
		_, _ = w.Write(resBytes)
	}))
	defer ts.Close()

	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithVertexAI(project, location),
		anthropic.WithBaseURL(ts.URL+baseEndpoint),
		anthropic.WithHTTPClient(http.DefaultClient),
	)

	resp, err := client.CountTokens(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaudeHaiku4Dot5V20251001,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage("Count these tokens."),
		},
		Thinking: &anthropic.Thinking{
			Type: anthropic.ThinkingTypeAdaptive,
		},
	})
	if err != nil {
		t.Fatalf("CountTokens error: %v", err)
	}

	if resp.InputTokens != 321 {
		t.Fatalf("InputTokens mismatch. got %d, want 321", resp.InputTokens)
	}
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
