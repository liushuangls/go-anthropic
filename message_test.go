package anthropic_test

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/liushuangls/go-anthropic/v2/internal/test"
	"github.com/liushuangls/go-anthropic/v2/internal/test/checks"
	"github.com/liushuangls/go-anthropic/v2/jsonschema"
)

//go:embed internal/test/sources/*
var sources embed.FS

var rateLimitHeaders = map[string]string{
	"anthropic-ratelimit-requests-limit":     "100",
	"anthropic-ratelimit-requests-remaining": "99",
	"anthropic-ratelimit-requests-reset":     "2024-06-04T07:13:19Z",
	"anthropic-ratelimit-tokens-limit":       "10000",
	"anthropic-ratelimit-tokens-remaining":   "9900",
	"anthropic-ratelimit-tokens-reset":       "2024-06-04T07:13:19Z",
	"retry-after":                            "100",
}

func TestMessages(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handleMessagesEndpoint(rateLimitHeaders))

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
		anthropic.WithAPIVersion(anthropic.APIVersion20230601),
		anthropic.WithEmptyMessagesLimit(100),
		anthropic.WithHTTPClient(http.DefaultClient),
	)
	resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaudeInstant1Dot2,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage("What is your name?"),
		},
		MaxTokens: 1000,
	})
	if err != nil {
		t.Fatalf("CreateMessages error: %v", err)
	}

	t.Logf("CreateMessages resp: %+v", resp)
}

func TestMessagesTokenError(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handleMessagesEndpoint(rateLimitHeaders))

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken()+"1",
		anthropic.WithBaseURL(baseUrl),
	)
	_, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaudeInstant1Dot2,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage("What is your name?"),
		},
		MaxTokens: 1000,
	})
	checks.HasError(t, err, "should error")

	var e *anthropic.RequestError
	if !errors.As(err, &e) {
		t.Log("should request error")
	}

	t.Logf("CreateMessages error: %s", err)
}

func TestMessagesVision(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handleMessagesEndpoint(rateLimitHeaders))

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
	)

	imagePath := "internal/test/sources/ant.jpg"
	imageMediaType := "image/jpeg"
	imageFile, err := sources.Open(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Opus20240229,
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewImageMessageContent(anthropic.MessageContentImageSource{
						Type:      "base64",
						MediaType: imageMediaType,
						Data:      imageData,
					}),
					anthropic.NewTextMessageContent("Describe this image."),
				},
			},
		},
		MaxTokens: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("CreateMessages resp: %+v", resp)
}

func TestMessagesToolUse(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handleMessagesEndpoint(rateLimitHeaders))

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
	)

	request := anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Haiku20240307,
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
	}

	resp, err := client.CreateMessages(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	request.Messages = append(request.Messages, anthropic.Message{
		Role:    anthropic.RoleAssistant,
		Content: resp.Content,
	})

	var toolUse *anthropic.MessageContentToolUse

	for _, c := range resp.Content {
		if c.Type == anthropic.MessagesContentTypeToolUse {
			toolUse = c.MessageContentToolUse
			t.Logf("ToolUse: %+v", toolUse)
		} else {
			t.Logf("Content: %+v", c)
		}
	}

	if toolUse == nil {
		t.Fatal("tool use not found")
	}

	request.Messages = append(request.Messages, anthropic.NewToolResultsMessage(toolUse.ID, "65 degrees", false))

	resp, err = client.CreateMessages(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("CreateMessages resp: %+v", resp)

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

func TestMessagesRateLimitHeaders(t *testing.T) {

	t.Run("parses valid rate limit headers", func(t *testing.T) {

		server := test.NewTestServer()
		server.RegisterHandler("/v1/messages", handleMessagesEndpoint(rateLimitHeaders))

		ts := server.AnthropicTestServer()
		ts.Start()
		defer ts.Close()

		baseUrl := ts.URL + "/v1"
		client := anthropic.NewClient(
			test.GetTestToken(),
			anthropic.WithBaseURL(baseUrl),
		)

		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
		})
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		rlHeaders, err := resp.GetRateLimitHeaders()
		if err != nil {
			t.Fatalf("GetRateLimitHeaders error: %v", err)
		}

		bs, err := json.Marshal(rlHeaders)
		if err != nil {
			t.Fatal(err)
		}

		var expectedHeaders = map[string]any{
			"anthropic-ratelimit-requests-limit":     100,
			"anthropic-ratelimit-requests-remaining": 99,
			"anthropic-ratelimit-requests-reset":     "2024-06-04T07:13:19Z",
			"anthropic-ratelimit-tokens-limit":       10000,
			"anthropic-ratelimit-tokens-remaining":   9900,
			"anthropic-ratelimit-tokens-reset":       "2024-06-04T07:13:19Z",
			"retry-after":                            100,
		}

		bs2, err := json.Marshal(expectedHeaders)
		if err != nil {
			t.Fatal(err)
		}

		if string(bs) != string(bs2) {
			t.Fatalf("rate limit headers mismatch. got %s, want %s", string(bs), string(bs2))
		}
	})

	t.Run("returns error for missing rate limit headers", func(t *testing.T) {

		invalidHeaders := map[string]string{}

		server := test.NewTestServer()
		server.RegisterHandler("/v1/messages", handleMessagesEndpoint(invalidHeaders))

		ts := server.AnthropicTestServer()
		ts.Start()
		defer ts.Close()

		baseUrl := ts.URL + "/v1"
		client := anthropic.NewClient(
			test.GetTestToken(),
			anthropic.WithBaseURL(baseUrl),
		)

		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
		})
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		_, err = resp.GetRateLimitHeaders()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// Allows for injection of custom rate limit headers in the response to test client parsing.
func handleMessagesEndpoint(headers map[string]string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var resBytes []byte

		// completions only accepts POST requests
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

		var messagesReq anthropic.MessagesRequest
		if messagesReq, err = getMessagesRequest(r); err != nil {
			http.Error(w, "could not read request", http.StatusInternalServerError)
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

		res := anthropic.MessagesResponse{
			Type: "completion",
			ID:   strconv.Itoa(int(time.Now().Unix())),
			Role: anthropic.RoleAssistant,
			Content: []anthropic.MessageContent{
				anthropic.NewTextMessageContent("hello"),
			},
			StopReason: anthropic.MessagesStopReasonEndTurn,
			Model:      messagesReq.Model,
			Usage: anthropic.MessagesUsage{
				InputTokens:  10,
				OutputTokens: 10,
			},
		}

		if len(messagesReq.Tools) > 0 {
			if hasToolResult {
				res.Content = []anthropic.MessageContent{
					anthropic.NewTextMessageContent("The current weather in San Francisco is 65 degrees Fahrenheit. It's a nice, moderate temperature typical of the San Francisco Bay Area climate."),
				}
			} else {
				m := map[string]any{
					"location": "San Francisco, CA",
					"unit":     "celsius",
				}
				bs, _ := json.Marshal(m)
				res.Content = []anthropic.MessageContent{
					anthropic.NewTextMessageContent("Okay, let me check the weather in San Francisco:"),
					anthropic.NewToolUseMessageContent("toolu_01Ex86JyJAe8RSbFRCTM3pQo", "get_weather", bs),
				}
				res.StopReason = anthropic.MessagesStopReasonToolUse
			}
		}

		resBytes, _ = json.Marshal(res)
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		_, _ = w.Write(resBytes)
	}
}

func getMessagesRequest(r *http.Request) (req anthropic.MessagesRequest, err error) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(reqBody, &req)
	if err != nil {
		return
	}
	return
}
