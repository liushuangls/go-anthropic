package anthropic_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
)

// MessageContent embeds multiple pointer structs with overlapping JSON tags
// (tool_use_id/content on tool_result + web_search_tool_result; id/name/input on
// tool_use + server_tool_use). Without a custom MarshalJSON, Go's encoding/json
// drops those ambiguous promoted fields, so tool blocks serialize empty and the
// API never receives them. These tests lock in correct serialization.

func TestToolResultMarshalIncludesFields(t *testing.T) {
	b, err := json.Marshal(anthropic.NewToolResultMessageContent("toolu_123", "the result", false))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"type":"tool_result"`, `"tool_use_id":"toolu_123"`, `"content"`, `the result`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("tool_result marshal missing %q: %s", want, b)
		}
	}
}

func TestToolUseMarshalIncludesFields(t *testing.T) {
	b, err := json.Marshal(anthropic.NewToolUseMessageContent("toolu_123", "get_weather", json.RawMessage(`{"city":"NYC"}`)))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"type":"tool_use"`, `"id":"toolu_123"`, `"name":"get_weather"`, `"input"`, `"city":"NYC"`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("tool_use marshal missing %q: %s", want, b)
		}
	}
}

func TestServerToolUseMarshalIncludesFields(t *testing.T) {
	b, err := json.Marshal(anthropic.NewServerToolUseContent("srvtoolu_1", "web_search", json.RawMessage(`{"query":"golang"}`)))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"id":"srvtoolu_1"`, `"name":"web_search"`, `"query":"golang"`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("server_tool_use marshal missing %q: %s", want, b)
		}
	}
}

// Non-tool blocks must keep their original encoding.
func TestTextBlockMarshalUnchanged(t *testing.T) {
	b, err := json.Marshal(anthropic.NewTextMessageContent("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"type":"text"`) || !strings.Contains(string(b), `"text":"hello"`) {
		t.Errorf("text block marshal changed unexpectedly: %s", b)
	}
}
