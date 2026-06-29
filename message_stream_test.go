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
	"github.com/liushuangls/go-anthropic/v2/jsonschema"
)

var (
	testMessagesStreamContent    = []string{"My", " name", " is", " Claude", "."}
	testMessagesJsonDeltaContent = []string{`{\"location\":`, `\"San Francisco, CA\"}`}
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
			Model: anthropic.ModelClaude3Haiku20240307,
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
		OnContentBlockStop:  func(data anthropic.MessagesEventContentBlockStopData, content anthropic.MessageContent) {},
		OnMessageDelta:      func(data anthropic.MessagesEventMessageDeltaData) {},
		OnMessageStop:       func(data anthropic.MessagesEventMessageStopData) {},
	})
	if err != nil {
		t.Fatalf("CreateMessagesStream error: %s", err)
	}

	expectedContent := strings.Join(testMessagesStreamContent, "")
	if received != expectedContent {
		t.Fatalf(
			"CreateMessagesStream content not match expected: %s, got: %s",
			expectedContent,
			received,
		)
	}
	if resp.GetFirstContentText() != expectedContent {
		t.Fatalf(
			"CreateMessagesStream content not match expected: %s, got: %s",
			expectedContent,
			resp.GetFirstContentText(),
		)
	}

	headers, err := resp.GetRateLimitHeaders()
	if err != nil {
		t.Fatalf("CreateMessagesStream GetRateLimitHeaders error: %s", err)
	}
	t.Logf("CreateMessagesStream rate limit headers: %+v", headers)

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
			Model: anthropic.ModelClaude3Haiku20240307,
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

func TestCreateMessagesStream(t *testing.T) {
	t.Run("Does not error for empty unknown messages below limit", func(t *testing.T) {
		emptyMessagesLimit := 100
		server := test.NewTestServer()
		server.RegisterHandler("/v1/messages",
			handlerMessagesStreamEmptyMessages(emptyMessagesLimit-1, "fake: {}"),
		)

		ts := server.AnthropicTestServer()
		ts.Start()
		defer ts.Close()
		baseUrl := ts.URL + "/v1"

		client := anthropic.NewClient(
			test.GetTestToken(),
			anthropic.WithBaseURL(baseUrl),
			anthropic.WithEmptyMessagesLimit(uint(emptyMessagesLimit)),
		)
		_, err := client.CreateMessagesStream(context.Background(), anthropic.MessagesStreamRequest{
			MessagesRequest: anthropic.MessagesRequest{
				Model:     anthropic.ModelClaude3Haiku20240307,
				Messages:  []anthropic.Message{},
				MaxTokens: 1000,
			},
		})
		if err != nil {
			t.Fatalf("CreateMessagesStream error: %s", err)
		}
	})

	t.Run("Error for empty unknown messages above limit", func(t *testing.T) {
		emptyMessagesLimit := 100
		server := test.NewTestServer()
		server.RegisterHandler(
			"/v1/messages",
			handlerMessagesStreamEmptyMessages(emptyMessagesLimit, "fake: {}"),
		)

		ts := server.AnthropicTestServer()
		ts.Start()
		defer ts.Close()
		baseUrl := ts.URL + "/v1"

		client := anthropic.NewClient(
			test.GetTestToken(),
			anthropic.WithBaseURL(baseUrl),
			anthropic.WithEmptyMessagesLimit(uint(emptyMessagesLimit-1)),
		)
		_, err := client.CreateMessagesStream(context.Background(), anthropic.MessagesStreamRequest{
			MessagesRequest: anthropic.MessagesRequest{
				Model: anthropic.ModelClaude3Haiku20240307,
				Messages: []anthropic.Message{
					anthropic.NewUserTextMessage("What's the weather like?"),
				},
				MaxTokens: 1000,
			},
		})
		if err == nil {
			t.Fatalf("Expected error for empty messages above limit, got nil")
		}

		if !errors.Is(err, anthropic.ErrTooManyEmptyStreamMessages) {
			t.Fatalf("Expected error to be ErrTooManyEmptyStreamMessages, got: %v", err)
		}
	})
}

func TestCreateMessagesStreamHandlesSparseContentBlockIndexes(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handlerMessagesStreamSparseContentBlockIndex)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
	)

	resp, err := client.CreateMessagesStream(context.Background(), anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
		},
	})
	if err != nil {
		t.Fatalf("CreateMessagesStream error: %s", err)
	}

	if len(resp.Content) != 3 {
		t.Fatalf("Content length mismatch. got %d, want %d", len(resp.Content), 3)
	}
	if resp.Content[2].GetText() != "sparse block" {
		t.Fatalf(
			"Content text mismatch. got %s, want %s",
			resp.Content[2].GetText(),
			"sparse block",
		)
	}
}

func TestCreateMessagesStreamHandlesSparseContentBlockDeltaIndexes(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handlerMessagesStreamSparseContentBlockDeltaIndex)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
	)

	resp, err := client.CreateMessagesStream(context.Background(), anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
		},
	})
	if err != nil {
		t.Fatalf("CreateMessagesStream error: %s", err)
	}

	if len(resp.Content) != 3 {
		t.Fatalf("Content length mismatch. got %d, want %d", len(resp.Content), 3)
	}
	if resp.Content[2].GetText() != "delta block" {
		t.Fatalf("Content text mismatch. got %s, want %s", resp.Content[2].GetText(), "delta block")
	}
}

func TestCreateMessagesStreamReturnsErrorForNegativeContentBlockIndexes(t *testing.T) {
	tests := []struct {
		name    string
		handler test.Handler
	}{
		{
			name:    "content_block_start",
			handler: handlerMessagesStreamNegativeContentBlockStartIndex,
		},
		{
			name:    "content_block_delta",
			handler: handlerMessagesStreamNegativeContentBlockDeltaIndex,
		},
		{
			name:    "content_block_stop",
			handler: handlerMessagesStreamNegativeContentBlockStopIndex,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := test.NewTestServer()
			server.RegisterHandler("/v1/messages", tt.handler)

			ts := server.AnthropicTestServer()
			ts.Start()
			defer ts.Close()

			baseUrl := ts.URL + "/v1"
			client := anthropic.NewClient(
				test.GetTestToken(),
				anthropic.WithBaseURL(baseUrl),
			)

			_, err := client.CreateMessagesStream(
				context.Background(),
				anthropic.MessagesStreamRequest{
					MessagesRequest: anthropic.MessagesRequest{
						Model: anthropic.ModelClaude3Haiku20240307,
						Messages: []anthropic.Message{
							anthropic.NewUserTextMessage("What is your name?"),
						},
						MaxTokens: 1000,
					},
				},
			)
			if err == nil {
				t.Fatalf("CreateMessagesStream expect error, but not")
			}
		})
	}
}

func TestMessagesStreamToolUse(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handlerMessagesStreamToolUse)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	cli := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
	)

	request := anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Opus20240229,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is the weather like in San Francisco?"),
			},
			MaxTokens: 1000,
			Tools: []anthropic.ToolDefinition{
				{
					Name:        "get_weather",
					Description: "Get the current weather in a given location",
					InputSchema: jsonschema.Definition{
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"location": {
								Type:        jsonschema.String,
								Description: "The city and state, e.g. San Francisco, CA",
							},
							"unit": {
								Type:        jsonschema.String,
								Enum:        []string{"celsius", "fahrenheit"},
								Description: "The unit of temperature, either 'celsius' or 'fahrenheit'",
							},
						},
						Required: []string{"location"},
					},
				},
			},
		},
		OnContentBlockStop: func(data anthropic.MessagesEventContentBlockStopData, content anthropic.MessageContent) {
			t.Logf("content block stop, index: %d", data.Index)
			switch content.Type {
			case anthropic.MessagesContentTypeText:
				t.Logf("content block stop, text: %s", content.GetText())
			case anthropic.MessagesContentTypeToolUse:
				t.Logf("content blog stop, tool_use: %+v, input: %s",
					*content.MessageContentToolUse,
					content.MessageContentToolUse.Input,
				)
			}
		},
	}

	resp, err := cli.CreateMessagesStream(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	request.Messages = append(request.Messages, anthropic.Message{
		Role:    anthropic.RoleAssistant,
		Content: resp.Content,
	})

	var toolUse *anthropic.MessageContentToolUse

	for _, m := range resp.Content {
		if m.Type == anthropic.MessagesContentTypeToolUse {
			toolUse = m.MessageContentToolUse
		}
	}

	if toolUse == nil {
		t.Fatalf("tool use not found")
	}

	request.Messages = append(
		request.Messages,
		anthropic.NewToolResultsMessage(toolUse.ID, "65 degrees", false),
	)

	resp, err = cli.CreateMessagesStream(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	var hasDegrees bool
	for _, m := range resp.Content {
		if m.Type == anthropic.MessagesContentTypeText {
			if strings.Contains(m.GetText(), "65 degrees") {
				hasDegrees = true
				break
			}
		}
	}
	if !hasDegrees {
		t.Fatalf("Expected response to contain '65 degrees', got: %+v", resp.Content)
	}
}

func TestMessagesStreamToolUseWithoutParameters(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handlerMessagesStreamToolUseWithoutParameters)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	cli := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
	)

	request := anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Opus20240229,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is the weather like in San Francisco?"),
			},
			MaxTokens: 1000,
			Tools: []anthropic.ToolDefinition{
				{
					Name:        "get_weather",
					Description: "Get the current weather in a given location",
					InputSchema: jsonschema.Definition{
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"location": {
								Type:        jsonschema.String,
								Description: "The city and state, e.g. San Francisco, CA",
							},
						},
						Required: []string{"location"},
					},
				},
			},
		},
		OnContentBlockStop: func(data anthropic.MessagesEventContentBlockStopData, content anthropic.MessageContent) {
			t.Logf("content block stop, index: %d", data.Index)
			switch content.Type {
			case anthropic.MessagesContentTypeText:
				t.Logf("content block stop, text: %s", content.GetText())
			case anthropic.MessagesContentTypeToolUse:
				t.Logf("content blog stop, tool_use: %+v, input: %s",
					*content.MessageContentToolUse,
					content.MessageContentToolUse.Input,
				)
			}
		},
	}

	resp, err := cli.CreateMessagesStream(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	request.Messages = append(request.Messages, anthropic.Message{
		Role:    anthropic.RoleAssistant,
		Content: resp.Content,
	})

	var toolUse *anthropic.MessageContentToolUse

	for _, m := range resp.Content {
		if m.Type == anthropic.MessagesContentTypeToolUse {
			toolUse = m.MessageContentToolUse
		}
	}

	if toolUse == nil {
		t.Fatalf("tool use not found")
	}

	request.Messages = append(
		request.Messages,
		anthropic.NewToolResultsMessage(toolUse.ID, "65 degrees", false),
	)

	resp, err = cli.CreateMessagesStream(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	var hasDegrees bool
	for _, m := range resp.Content {
		if m.Type == anthropic.MessagesContentTypeText {
			if strings.Contains(m.GetText(), "65 degrees") {
				hasDegrees = true
				break
			}
		}
	}
	if !hasDegrees {
		t.Fatalf("Expected response to contain '65 degrees', got: %+v", resp.Content)
	}
}

func handlerMessagesStream(w http.ResponseWriter, r *http.Request) {
	request, err := getRequest[anthropic.MessagesRequest](r)
	if err != nil {
		http.Error(w, "request error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")

	w.Header().Set("anthropic-ratelimit-requests-limit", "1000")
	w.Header().Set("anthropic-ratelimit-requests-remaining", "999")
	w.Header().Set("anthropic-ratelimit-requests-reset", "2022-01-01T00:00:00Z")
	w.Header().Set("anthropic-ratelimit-tokens-limit", "1000")
	w.Header().Set("anthropic-ratelimit-tokens-remaining", "999")
	w.Header().Set("anthropic-ratelimit-tokens-reset", "2022-01-01T00:00:00Z")
	w.Header().Set("retry-after", "0")

	var dataBytes []byte

	if request.Temperature != nil && *request.Temperature > 1 {
		dataBytes = append(dataBytes, []byte("event: error\n")...)
		dataBytes = append(
			dataBytes,
			[]byte(
				`data: {"type": "error", "error": {"type": "overloaded_error", "message": "Overloaded"}}`+"\n\n",
			)...)
	}

	dataBytes = append(dataBytes, []byte("event: message_start\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"message_start","message":{"id":"1","type":"message","role":"assistant","content":[],"model":"claude-3-haiku-20240307","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":14,"output_tokens":1}}}`+"\n\n",
		)...)

	dataBytes = append(dataBytes, []byte("event: content_block_start\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`+"\n\n",
		)...)

	dataBytes = append(dataBytes, []byte("event: ping\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type": "ping"}`+"\n\n")...)

	for _, t := range testMessagesStreamContent {
		dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
		dataBytes = append(
			dataBytes,
			[]byte(
				fmt.Sprintf(
					`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"%s"}}`,
					t,
				)+"\n\n",
			)...)
	}

	dataBytes = append(dataBytes, []byte("event: content_block_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_stop","index":0}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: message_delta\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":9}}`+"\n\n",
		)...)

	dataBytes = append(dataBytes, []byte("event: message_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_stop"}`+"\n\n")...)

	_, _ = w.Write(dataBytes)
}

func handlerMessagesStreamSparseContentBlockIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")

	var dataBytes []byte

	dataBytes = append(dataBytes, []byte("event: message_start\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"message_start","message":{"id":"1","type":"message","role":"assistant","content":[],"model":"claude-3-haiku-20240307","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":14,"output_tokens":1}}}`+"\n\n",
		)...)

	dataBytes = append(dataBytes, []byte("event: content_block_start\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`+"\n\n",
		)...)
	dataBytes = append(dataBytes, []byte("event: content_block_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_stop","index":0}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: content_block_start\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"content_block_start","index":2,"content_block":{"type":"text","text":""}}`+"\n\n",
		)...)
	dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"content_block_delta","index":2,"delta":{"type":"text_delta","text":"sparse block"}}`+"\n\n",
		)...)
	dataBytes = append(dataBytes, []byte("event: content_block_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_stop","index":2}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: message_delta\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":9}}`+"\n\n",
		)...)

	dataBytes = append(dataBytes, []byte("event: message_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_stop"}`+"\n\n")...)

	_, _ = w.Write(dataBytes)
}

func handlerMessagesStreamSparseContentBlockDeltaIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")

	var dataBytes []byte

	dataBytes = append(dataBytes, []byte("event: message_start\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"message_start","message":{"id":"1","type":"message","role":"assistant","content":[],"model":"claude-3-haiku-20240307","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":14,"output_tokens":1}}}`+"\n\n",
		)...)

	dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"content_block_delta","index":2,"delta":{"type":"text_delta","text":"delta block"}}`+"\n\n",
		)...)

	dataBytes = append(dataBytes, []byte("event: content_block_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_stop","index":2}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: message_delta\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":9}}`+"\n\n",
		)...)

	dataBytes = append(dataBytes, []byte("event: message_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_stop"}`+"\n\n")...)

	_, _ = w.Write(dataBytes)
}

func handlerMessagesStreamNegativeContentBlockStartIndex(w http.ResponseWriter, r *http.Request) {
	writeMessagesStreamWithNegativeContentBlockIndex(
		w,
		"content_block_start",
		`{"type":"content_block_start","index":-1,"content_block":{"type":"text","text":""}}`,
	)
}

func handlerMessagesStreamNegativeContentBlockDeltaIndex(w http.ResponseWriter, r *http.Request) {
	writeMessagesStreamWithNegativeContentBlockIndex(
		w,
		"content_block_delta",
		`{"type":"content_block_delta","index":-1,"delta":{"type":"text_delta","text":"negative index"}}`,
	)
}

func handlerMessagesStreamNegativeContentBlockStopIndex(w http.ResponseWriter, r *http.Request) {
	writeMessagesStreamWithNegativeContentBlockIndex(
		w,
		"content_block_stop",
		`{"type":"content_block_stop","index":-1}`,
	)
}

func writeMessagesStreamWithNegativeContentBlockIndex(w http.ResponseWriter, event, data string) {
	w.Header().Set("Content-Type", "text/event-stream")

	var dataBytes []byte

	dataBytes = append(dataBytes, []byte("event: message_start\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"message_start","message":{"id":"1","type":"message","role":"assistant","content":[],"model":"claude-3-haiku-20240307","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":14,"output_tokens":1}}}`+"\n\n",
		)...)

	dataBytes = append(dataBytes, []byte("event: "+event+"\n")...)
	dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

	_, _ = w.Write(dataBytes)
}

func handlerMessagesStreamToolUse(w http.ResponseWriter, r *http.Request) {
	messagesReq, err := getRequest[anthropic.MessagesRequest](r)
	if err != nil {
		http.Error(w, "request error", http.StatusBadRequest)
		return
	}

	var hasToolResult bool

	for _, m := range messagesReq.Messages {
		for _, c := range m.Content {
			if c.Type == anthropic.MessagesContentTypeToolResult {
				hasToolResult = true
				break
			}
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")

	var dataBytes []byte

	dataBytes = append(dataBytes, []byte("event: message_start\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"message_start","message":{"id":"123333","type":"message","role":"assistant","model":"claude-3-opus-20240229","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":844,"output_tokens":2}}}`+"\n\n",
		)...)

	if hasToolResult {
		dataBytes = append(dataBytes, []byte("event: content_block_start\n")...)
		dataBytes = append(
			dataBytes,
			[]byte(
				`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`+"\n\n",
			)...)

		dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
		dataBytes = append(
			dataBytes,
			[]byte(
				`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"The current weather in San Francisco is 65 degrees Fahrenheit. It's a nice, moderate temperature typical of the San Francisco Bay Area climate."}}`+"\n\n",
			)...)

		dataBytes = append(dataBytes, []byte("event: content_block_stop\n")...)
		dataBytes = append(
			dataBytes,
			[]byte(`data: {"type":"content_block_stop","index":0}`+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: message_delta\n")...)
		dataBytes = append(
			dataBytes,
			[]byte(
				`data: {"type":"message_delta","delta":{"stop_reason":"end_return","stop_sequence":null},"usage":{"output_tokens":9}}`+"\n\n",
			)...)
	} else {
		dataBytes = append(dataBytes, []byte("event: content_block_start\n")...)
		dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_019ktsPEWabjtYw1iGdjT2Qy","name":"get_weather","input":{}}}`+"\n\n")...)

		for _, t := range testMessagesJsonDeltaContent {
			dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
			dataBytes = append(dataBytes, []byte(fmt.Sprintf(`data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"%s"}}`, t)+"\n\n")...)
		}

		dataBytes = append(dataBytes, []byte("event: content_block_stop\n")...)
		dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_stop","index":0}`+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: message_delta\n")...)
		dataBytes = append(dataBytes, []byte(`data: {"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null},"usage":{"output_tokens":9}}`+"\n\n")...)
	}

	dataBytes = append(dataBytes, []byte("event: message_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_stop"}`+"\n\n")...)

	_, _ = w.Write(dataBytes)
}

func handlerMessagesStreamToolUseWithoutParameters(w http.ResponseWriter, r *http.Request) {
	messagesReq, err := getRequest[anthropic.MessagesRequest](r)
	if err != nil {
		http.Error(w, "request error", http.StatusBadRequest)
		return
	}

	var hasToolResult bool

	for _, m := range messagesReq.Messages {
		for _, c := range m.Content {
			if c.Type == anthropic.MessagesContentTypeToolResult {
				hasToolResult = true
				break
			}
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")

	var dataBytes []byte

	dataBytes = append(dataBytes, []byte("event: message_start\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"message_start","message":{"id":"123333","type":"message","role":"assistant","model":"claude-3-opus-20240229","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":844,"output_tokens":2}}}`+"\n\n",
		)...)

	if hasToolResult {
		dataBytes = append(dataBytes, []byte("event: content_block_start\n")...)
		dataBytes = append(
			dataBytes,
			[]byte(
				`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`+"\n\n",
			)...)

		dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
		dataBytes = append(
			dataBytes,
			[]byte(
				`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"The current weather in San Francisco is 65 degrees Fahrenheit. It's a nice, moderate temperature typical of the San Francisco Bay Area climate."}}`+"\n\n",
			)...)

		dataBytes = append(dataBytes, []byte("event: content_block_stop\n")...)
		dataBytes = append(
			dataBytes,
			[]byte(`data: {"type":"content_block_stop","index":0}`+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: message_delta\n")...)
		dataBytes = append(
			dataBytes,
			[]byte(
				`data: {"type":"message_delta","delta":{"stop_reason":"end_return","stop_sequence":null},"usage":{"output_tokens":9}}`+"\n\n",
			)...)
	} else {
		dataBytes = append(dataBytes, []byte("event: content_block_start\n")...)
		dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_019ktsPEWabjtYw1iGdjT2Qy","name":"get_weather","input":{}}}`+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: content_block_stop\n")...)
		dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_stop","index":0}`+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: message_delta\n")...)
		dataBytes = append(dataBytes, []byte(`data: {"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null},"usage":{"output_tokens":9}}`+"\n\n")...)
	}

	dataBytes = append(dataBytes, []byte("event: message_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_stop"}`+"\n\n")...)

	_, _ = w.Write(dataBytes)
}

func TestCreateMessagesStreamIgnoresCommentsAndUnknownEvents(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handlerMessagesStreamCommentsAndUnknownEvents)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	// A small limit ensures the test fails if comment/unknown lines were
	// counted against the empty-message budget.
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
		anthropic.WithEmptyMessagesLimit(5),
	)

	var received string
	_, err := client.CreateMessagesStream(context.Background(), anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
		},
		OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			received += data.Delta.GetText()
		},
	})
	if err != nil {
		t.Fatalf("CreateMessagesStream unexpected error: %s", err)
	}
	if received != "hello" {
		t.Fatalf("expected content %q, got %q", "hello", received)
	}
}

func handlerMessagesStreamCommentsAndUnknownEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")

	var dataBytes []byte

	dataBytes = append(dataBytes, []byte("event: message_start\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"message_start","message":{"id":"1","type":"message","role":"assistant","content":[],"model":"claude-3-haiku-20240307","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":14,"output_tokens":1}}}`+"\n\n",
		)...)

	dataBytes = append(dataBytes, []byte("event: content_block_start\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`+"\n\n",
		)...)

	// Many SSE comment lines (proxy keep-alives) and unknown/future event
	// types, far exceeding the empty-message limit of 5.
	for i := 0; i < 50; i++ {
		dataBytes = append(dataBytes, []byte(": keep-alive\n\n")...)
		dataBytes = append(dataBytes, []byte("event: some_future_event\n")...)
		dataBytes = append(dataBytes, []byte(`data: {"type":"some_future_event"}`+"\n\n")...)
	}

	dataBytes = append(dataBytes, []byte("event: content_block_delta\n")...)
	dataBytes = append(
		dataBytes,
		[]byte(
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hello"}}`+"\n\n",
		)...)

	dataBytes = append(dataBytes, []byte("event: content_block_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"content_block_stop","index":0}`+"\n\n")...)

	dataBytes = append(dataBytes, []byte("event: message_stop\n")...)
	dataBytes = append(dataBytes, []byte(`data: {"type":"message_stop"}`+"\n\n")...)

	_, _ = w.Write(dataBytes)
}

func handlerMessagesStreamEmptyMessages(numEmptyMessages int, payload string) test.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getRequest[anthropic.MessagesRequest](r)
		if err != nil {
			http.Error(w, "request error", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")

		var dataBytes []byte

		dataBytes = append(dataBytes, []byte("event: message_start\n")...)
		dataBytes = append(
			dataBytes,
			[]byte(
				`data: {"type":"message_start","message":{"id":"123333","type":"message","role":"assistant","model":"claude-3-opus-20240229","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":844,"output_tokens":2}}}`+"\n\n",
			)...)

		for i := 0; i < numEmptyMessages; i++ {
			dataBytes = append(dataBytes, []byte(payload+"\n")...)
		}

		_, _ = w.Write(dataBytes)
	}
}

func TestVertexMessagesStream(t *testing.T) {
	project := "project"
	location := "location"
	model := anthropic.ModelClaude3Haiku20240307
	vertexModel := "claude-3-haiku@20240307"

	baseEndpoint := fmt.Sprintf(
		"/v1/projects/%s/locations/%s/publishers/anthropic/models",
		project,
		location,
	)

	server := test.NewTestServer()
	server.RegisterHandler(baseEndpoint+"/"+vertexModel+":streamRawPredict", handlerMessagesStream)

	ts := server.VertexTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + baseEndpoint
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithVertexAI(project, location),
		anthropic.WithBaseURL(baseUrl),
	)
	var received string
	resp, err := client.CreateMessagesStream(context.Background(), anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: model,
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
		OnContentBlockStop:  func(data anthropic.MessagesEventContentBlockStopData, content anthropic.MessageContent) {},
		OnMessageDelta:      func(data anthropic.MessagesEventMessageDeltaData) {},
		OnMessageStop:       func(data anthropic.MessagesEventMessageStopData) {},
	})
	if err != nil {
		t.Fatalf("CreateMessagesStream error: %s", err)
	}

	expectedContent := strings.Join(testMessagesStreamContent, "")
	if received != expectedContent {
		t.Fatalf(
			"CreateMessagesStream content not match expected: %s, got: %s",
			expectedContent,
			received,
		)
	}
	if resp.GetFirstContentText() != expectedContent {
		t.Fatalf(
			"CreateMessagesStream content not match expected: %s, got: %s",
			expectedContent,
			resp.GetFirstContentText(),
		)
	}

	headers, err := resp.GetRateLimitHeaders()
	if err != nil {
		t.Fatalf("CreateMessagesStream GetRateLimitHeaders error: %s", err)
	}
	t.Logf("CreateMessagesStream rate limit headers: %+v", headers)

	t.Logf("CreateMessagesStream resp: %+v", resp)
}

func TestVertexMessagesStreamError(t *testing.T) {
	project := "project"
	location := "location"
	model := anthropic.ModelClaude3Haiku20240307
	vertexModel := "claude-3-haiku@20240307"

	baseEndpoint := fmt.Sprintf(
		"/v1/projects/%s/locations/%s/publishers/anthropic/models",
		project,
		location,
	)

	server := test.NewTestServer()
	server.RegisterHandler(baseEndpoint+"/"+vertexModel+":streamRawPredict", handlerMessagesStream)

	ts := server.VertexTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + baseEndpoint
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithVertexAI(project, location),
		anthropic.WithBaseURL(baseUrl),
	)
	param := anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: model,
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
