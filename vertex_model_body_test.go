package anthropic

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestVertexRequestOmitsModelFromBody is a regression test for the Vertex AI
// (and Bedrock) request path. On those platforms the model is carried in the
// URL and MUST NOT appear in the request body; the adapters strip it by setting
// Model to "" (Vertex does this inside SetAnthropicVersion). If the "model"
// field is not omitempty, the blanked model serializes as `"model":""` and
// Vertex rejects the request with:
//
//	model: Extra inputs are not permitted
//
// See PR #117, which removed omitempty and regressed this behavior.
func TestVertexRequestOmitsModelFromBody(t *testing.T) {
	req := MessagesRequest{
		Model:     ModelClaudeOpus4Dot6,
		MaxTokens: 1024,
		Messages: []Message{
			NewUserTextMessage("hello"),
		},
	}

	// Exactly what VertexAdapter.PrepareRequest does before the body is marshalled.
	req.SetAnthropicVersion(APIVersionVertex20231016)

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if strings.Contains(string(body), `"model"`) {
		t.Fatalf("Vertex body must not contain a \"model\" field, got: %s", body)
	}

	// max_tokens is a valid Vertex body field and must still be present.
	if !strings.Contains(string(body), `"max_tokens"`) {
		t.Fatalf("body must retain \"max_tokens\", got: %s", body)
	}

	// anthropic_version must be injected for the Vertex path.
	if !strings.Contains(string(body), `"anthropic_version":"vertex-2023-10-16"`) {
		t.Fatalf("body must carry anthropic_version, got: %s", body)
	}
}

// TestDirectRequestKeepsModelInBody guards the other direction: on the direct
// Anthropic API the model is required in the body and must always be sent.
func TestDirectRequestKeepsModelInBody(t *testing.T) {
	req := MessagesRequest{
		Model:     ModelClaudeOpus4Dot6,
		MaxTokens: 1024,
		Messages: []Message{
			NewUserTextMessage("hello"),
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !strings.Contains(string(body), `"model":"claude-opus-4-6"`) {
		t.Fatalf("direct API body must contain the model, got: %s", body)
	}
}
