package anthropic_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/liushuangls/go-anthropic/v2/internal/test"
	"github.com/stretchr/testify/require"
)

var (
	testCompletionStreamContent = []string{"My", " name", " is", " Claude", "."}
)

func TestCompleteStream(t *testing.T) {
	is := require.New(t)
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
	resp, err := client.CreateCompleteStream(context.Background(), anthropic.CompleteStreamRequest{
		CompleteRequest: anthropic.CompleteRequest{
			Model:             anthropic.ModelClaudeInstant1Dot2,
			Prompt:            "\n\nHuman: What is your name?\n\nAssistant:",
			MaxTokensToSample: 1000,
		},
		OnCompletion: func(data anthropic.CompleteResponse) {
			receivedContent += data.Completion
		},
		OnPing:  func(data anthropic.CompleteStreamPingData) {},
		OnError: func(response anthropic.ErrorResponse) {},
	})
	is.NoError(err)

	expected := strings.Join(testCompletionStreamContent, "")
	is.Equal(expected, receivedContent)
	is.Equal(expected, resp.Completion)

	t.Logf("CreateCompleteStream resp: %+v", resp)
}

func TestCompleteStreamError(t *testing.T) {
	is := require.New(t)
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
	param := anthropic.CompleteStreamRequest{
		CompleteRequest: anthropic.CompleteRequest{
			Model:             anthropic.ModelClaudeInstant1Dot2,
			Prompt:            "\n\nHuman: What is your name?\n\nAssistant:",
			MaxTokensToSample: 1000,
		},
		OnCompletion: func(data anthropic.CompleteResponse) {
			receivedContent += data.Completion
		},
		OnPing:  func(data anthropic.CompleteStreamPingData) {},
		OnError: func(response anthropic.ErrorResponse) {},
	}
	param.SetTemperature(2)
	_, err := client.CreateCompleteStream(context.Background(), param)
	is.Error(err)
	is.Contains(err.Error(), "Overloaded")

	var e *anthropic.APIError
	is.True(errors.As(err, &e), "should api error")

	t.Logf("CreateCompleteStream error: %+v", err)
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
