package anthropic_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"anthropic"
	"anthropic/internal/test"
)

var (
	testCompletionStreamContent = []string{"My", " name", " is", " Claude", "."}
)

func TestCompleteStream(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/complete", handlerCompleteStream)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
	)
	var receivedContent string
	resp, err := client.CreateStreamComplete(context.Background(), anthropic.CompleteStreamRequest{
		CompleteRequest: anthropic.CompleteRequest{
			Model:             anthropic.ModelClaudeInstant1Dot2,
			Prompt:            "\n\nHuman: What is your name?\n\nAssistant:",
			MaxTokensToSample: 1000,
		},
		OnCompletion: func(data anthropic.CompleteResponse) {
			receivedContent += data.Completion
			//t.Logf("CreateStreamComplete OnCompletion data: %+v", data)
		},
	})
	if err != nil {
		t.Fatalf("CreateStreamComplete error: %s", err)
	}

	expected := strings.Join(testCompletionStreamContent, "")
	if receivedContent != expected {
		t.Fatalf("CreateStreamComplete content not match expected: %s, got: %s", expected, receivedContent)
	}
	if resp.Completion != expected {
		t.Fatalf("CreateStreamComplete content not match expected: %s, got: %s", expected, resp.Completion)
	}
	t.Logf("CreateStreamComplete resp: %+v", resp)
}

func handlerCompleteStream(w http.ResponseWriter, r *http.Request) {
	request, err := getCompleteRequest(r)
	if err != nil {
		http.Error(w, "request error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")

	var dataBytes []byte

	dataBytes = append(dataBytes, []byte("event: ping\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type": "ping"}`+"\n\n")...)

	if request.Temperature != nil && *request.Temperature > 1 {
		dataBytes = append(dataBytes, []byte("event: error\n")...)
		dataBytes = append(dataBytes, []byte(`data: {"type": "error", "error": {"type": "overloaded_error", "message": "Overloaded"}}`+"\n\n")...)
	}

	for _, t := range testCompletionStreamContent {
		dataBytes = append(dataBytes, []byte("event: completion\n")...)
		dataBytes = append(dataBytes, []byte(fmt.Sprintf(`data: {"type":"completion","id":"compl_01GatBXF5t5K51mYzbVgRJfZ","completion":"%s","stop_reason":null,"model":"claude-instant-1.2","stop":null,"log_id":"compl_01GatBXF5t5K51mYzbVgRJfZ"}`, t)+"\n\n")...)
	}

	dataBytes = append(dataBytes, []byte("event: completion\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"completion","id":"compl_01GatBXF5t5K51mYzbVgRJfZ","completion":"","stop_reason":"stop_sequence","model":"claude-instant-1.2","stop":null,"log_id":"compl_01GatBXF5t5K51mYzbVgRJfZ"}`+"\n\n")...)

	_, _ = w.Write(dataBytes)
}
