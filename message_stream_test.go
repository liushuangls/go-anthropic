package anthropic_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/liushuangls/go-anthropic/v2/internal/test"
)

var (
	testMessagesStreamContent = []string{"My", " name", " is", " Claude", "."}
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
	var received string
	resp, err := client.CreateMessagesStream(context.Background(), anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
		},
		OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			received += data.Delta.GetText()
			//t.Logf("CreateMessagesStream delta resp: %+v", data)
		},
		OnError:             func(response anthropic.ErrorResponse) {},
		OnPing:              func(data anthropic.MessagesEventPingData) {},
		OnMessageStart:      func(data anthropic.MessagesEventMessageStartData) {},
		OnContentBlockStart: func(data anthropic.MessagesEventContentBlockStartData) {},
		OnContentBlockStop:  func(data anthropic.MessagesEventContentBlockStopData) {},
		OnMessageDelta:      func(data anthropic.MessagesEventMessageDeltaData) {},
		OnMessageStop:       func(data anthropic.MessagesEventMessageStopData) {},
	})
	if err != nil {
		t.Fatalf("CreateMessagesStream error: %s", err)
	}

	expectedContent := strings.Join(testMessagesStreamContent, "")
	if received != expectedContent {
		t.Fatalf("CreateMessagesStream content not match expected: %s, got: %s", expectedContent, received)
	}
	if resp.GetFirstContentText() != expectedContent {
		t.Fatalf("CreateMessagesStream content not match expected: %s, got: %s", expectedContent, resp.GetFirstContentText())
	}

	t.Logf("CreateMessagesStream resp: %+v", resp)
}

func TestMessagesStreamError(t *testing.T) {
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
	param := anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
		},
		OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			t.Logf("CreateMessagesStream delta resp: %+v", data)
		},
		OnError: func(response anthropic.ErrorResponse) {},
	}
	param.SetTemperature(2)
	param.SetTopP(2)
	param.SetTopK(1)
	_, err := client.CreateMessagesStream(context.Background(), param)
	if err == nil {
		t.Fatalf("CreateMessagesStream expect error, but not")
	}

	t.Logf("CreateMessagesStream error: %s", err)
}

func handlerMessagesStream(w http.ResponseWriter, r *http.Request) {
	request, err := getMessagesRequest(r)
	if err != nil {
		http.Error(w, "request error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")

	var dataBytes []byte

	if request.Temperature != nil && *request.Temperature > 1 {
		dataBytes = append(dataBytes, []byte("event: error\n")...)
		dataBytes = append(dataBytes, []byte(`data: {"type": "error", "error": {"type": "overloaded_error", "message": "Overloaded"}}`+"\n\n")...)
	}

	dataBytes = append(dataBytes, []byte("event: message_start\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_start","message":{"id":"1","type":"message","role":"assistant","content":[],"model":"claude-instant-1.2","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":14,"output_tokens":1}}}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: content_block_start\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: ping\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type": "ping"}`+"\n\n")...)

	for _, t := range testMessagesStreamContent {
		dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
		dataBytes = append(dataBytes, []byte(fmt.Sprintf(`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"%s"}}`, t)+"\n\n")...)
	}

	dataBytes = append(dataBytes, []byte("event: content_block_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_stop","index":0}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: message_delta\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":9}}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: message_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_stop"}`+"\n\n")...)

	_, _ = w.Write(dataBytes)
}
