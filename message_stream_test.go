package anthropic_test

import (
	"context"
	"net/http"
	"testing"

	"anthropic"
	"anthropic/internal/test"
)

func TestMessagesStream(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handlerMessagesStream)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
	)
	resp, err := client.CreateMessagesSteam(context.Background(), anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				{Role: anthropic.RoleUser, Content: "What is your name?"},
			},
			MaxTokens: 1000,
		},
		OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			t.Logf("CreateMessagesSteam delta resp: %+v", data)
		},
	})
	if err != nil {
		t.Fatalf("CreateMessagesSteam error: %s", err)
	}

	t.Logf("CreateMessagesSteam resp: %+v", resp)
}

func handlerMessagesStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")

	var dataBytes []byte

	dataBytes = append(dataBytes, []byte("event: message_start\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_start","message":{"id":"1","type":"message","role":"assistant","content":[],"model":"claude-instant-1.2","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":14,"output_tokens":1}}}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: content_block_start\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: ping\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type": "ping"}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"My"}}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" name"}}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" is"}}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" Claude"}}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"."}}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: content_block_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_stop","index":0}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: message_delta\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":9}}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: message_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_stop"}`+"\n\n")...)

	_, _ = w.Write(dataBytes)
}
