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
	"github.com/liushuangls/go-anthropic/v2/jsonschema"
	"github.com/stretchr/testify/require"
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
	"retry-after":                            "", // retry-after is optional and may not be present.
}

func TestMessages(t *testing.T) {
	is := require.New(t)

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

	t.Run("create messages success", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
		})
		is.NoError(err)

		t.Logf("CreateMessages resp: %+v", resp)
	})

	t.Run("create messages success with single system message", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
			System:    "test system message",
		})
		is.NoError(err)

		t.Logf("CreateMessages resp: %+v", resp)
	})

	t.Run("create messages success with single multi-system message", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens:   1000,
			MultiSystem: anthropic.NewMultiSystemMessages("test single multi-system message"),
		})
		is.NoError(err)

		t.Logf("CreateMessages resp: %+v", resp)
	})

	t.Run("create messages success with multi-system messages", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
			MultiSystem: anthropic.NewMultiSystemMessages(
				"test multi-system messages",
				"here",
				"are",
				"some",
				"more",
				"messages",
				"for",
				"testing",
			),
		})
		is.NoError(err)

		t.Logf("CreateMessages resp: %+v", resp)
	})

}

func TestNewUserTextMessage(t *testing.T) {
	is := require.New(t)
	m := anthropic.NewUserTextMessage("What is your name?")

	is.Equal(anthropic.RoleUser, m.Role)
	is.Equal(anthropic.MessagesContentTypeText, m.Content[0].Type)
	is.Equal("What is your name?", *m.Content[0].Text)
}

func TestNewAssistantTextMessage(t *testing.T) {
	is := require.New(t)
	m := anthropic.NewAssistantTextMessage("My name is Claude.")

	is.Equal(anthropic.RoleAssistant, m.Role)
	is.Equal(anthropic.MessagesContentTypeText, m.Content[0].Type)
	is.Equal("My name is Claude.", *m.Content[0].Text)
}

func TestGetFirstContent(t *testing.T) {
	is := require.New(t)

	t.Run("returns empty content", func(t *testing.T) {
		m := anthropic.Message{}
		c := m.GetFirstContent()

		is.Equal(anthropic.MessagesContentType(""), c.Type)
		is.Nil(c.Text, "Text should be nil")
	})

	t.Run("returns single content", func(t *testing.T) {
		m := anthropic.NewAssistantTextMessage("My name is Claude.")
		c := m.GetFirstContent()

		is.Equal(anthropic.MessagesContentTypeText, c.Type)
		is.Equal("My name is Claude.", *c.Text)
	})

	t.Run("returns first content when multiple content present", func(t *testing.T) {
		m := anthropic.Message{
			Role: anthropic.RoleAssistant,
			Content: []anthropic.MessageContent{
				anthropic.NewTextMessageContent("My name is Claude."),
				anthropic.NewTextMessageContent("What is your name?"),
			},
		}
		c := m.GetFirstContent()

		is.Equal(anthropic.MessagesContentTypeText, c.Type)
		is.Equal("My name is Claude.", *c.Text)
	})
}

func TestGetFirstContentText(t *testing.T) {
	is := require.New(t)

	t.Run("returns empty text", func(t *testing.T) {
		m := anthropic.MessagesResponse{}

		is.Equal("", m.GetFirstContentText())
	})

	t.Run("returns text", func(t *testing.T) {
		m := anthropic.MessagesResponse{
			Content: []anthropic.MessageContent{
				anthropic.NewTextMessageContent("test string"),
			},
		}

		is.Equal("test string", m.GetFirstContentText())
	})
}

func TestGetText(t *testing.T) {
	is := require.New(t)
	t.Run("returns empty text", func(t *testing.T) {
		c := anthropic.MessageContent{}

		is.Equal("", c.GetText())
	})

	t.Run("returns text", func(t *testing.T) {
		c := anthropic.NewTextMessageContent("My name is Claude.")

		is.Equal("My name is Claude.", c.GetText())
	})
}

func TestConcatText(t *testing.T) {
	is := require.New(t)
	t.Run("concatenates text when text content text present", func(t *testing.T) {
		mc := anthropic.NewTextMessageContent("original")
		mc.ConcatText(" added")

		is.Equal("original added", mc.GetText())
	})

	t.Run("concatenates text when text content text not present", func(t *testing.T) {
		mc := anthropic.MessageContent{}
		mc.ConcatText("added")

		is.Equal("added", mc.GetText())
	})
}

func TestMessagesTokenError(t *testing.T) {
	is := require.New(t)
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

	is.Error(err, "should be an invalid token error")
	is.Contains(err.Error(), "401")

	var e *anthropic.RequestError
	if !errors.As(err, &e) {
		t.Log("should request error")
	}

	t.Logf("CreateMessages error: %s", err)
}

func TestMessagesVision(t *testing.T) {
	is := require.New(t)
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
	is.NoError(err)

	imageData, err := io.ReadAll(imageFile)
	is.NoError(err)

	resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Opus20240229,
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewImageMessageContent(anthropic.NewMessageContentImageSource("base64", imageMediaType, imageData)),
					anthropic.NewImageMessageContent(anthropic.MessageContentImageSource{
						Type:      "base64",
						MediaType: imageMediaType,
						Data:      imageData,
					}),
					anthropic.NewTextMessageContent("Describe these images."),
				},
			},
		},
		MaxTokens: 1000,
	})
	is.NoError(err)

	t.Logf("CreateMessages resp: %+v", resp)
}

func TestMessagesToolUse(t *testing.T) {
	is := require.New(t)
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
	is.NoError(err)

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

	is.NotNil(toolUse, "Tool use not found")

	request.Messages = append(request.Messages, anthropic.NewToolResultsMessage(toolUse.ID, "65 degrees", false))

	resp, err = client.CreateMessages(context.Background(), request)
	is.NoError(err)

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
	is.True(hasDegrees, "Expected response to contain '65 degrees'")
}

func TestMessagesRateLimitHeaders(t *testing.T) {
	is := require.New(t)
	expectedSuccess := map[string]any{
		"anthropic-ratelimit-requests-limit":     100,
		"anthropic-ratelimit-requests-remaining": 99,
		"anthropic-ratelimit-requests-reset":     "2024-06-04T07:13:19Z",
		"anthropic-ratelimit-tokens-limit":       10000,
		"anthropic-ratelimit-tokens-remaining":   9900,
		"anthropic-ratelimit-tokens-reset":       "2024-06-04T07:13:19Z",
		"retry-after":                            -1, // if not present, should be -1
	}

	requestFailRateLimitHeaders := make(map[string]string)
	for k, v := range rateLimitHeaders {
		requestFailRateLimitHeaders[k] = v
	}
	requestFailRateLimitHeaders["retry-after"] = "10"

	expectedFail := make(map[string]any)
	for k, v := range expectedSuccess {
		expectedFail[k] = v
	}
	expectedFail["retry-after"] = 10

	headerConfigs := []struct {
		name     string
		header   map[string]string
		expected map[string]any
	}{
		{
			name:     "parses rate limit headers for successful request",
			header:   rateLimitHeaders,
			expected: expectedSuccess,
		},
		{
			name:     "parses rate limit headers for failed request",
			header:   requestFailRateLimitHeaders,
			expected: expectedFail,
		},
	}

	for _, c := range headerConfigs {
		t.Run(c.name, func(t *testing.T) {
			server := test.NewTestServer()
			server.RegisterHandler("/v1/messages", handleMessagesEndpoint(c.header))

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
			is.NoError(err)

			rlHeaders, err := resp.GetRateLimitHeaders()
			is.NoError(err)

			bs, err := json.Marshal(rlHeaders)
			is.NoError(err)

			bs2, err := json.Marshal(c.expected)
			is.NoError(err)

			is.Equal(string(bs2), string(bs), "rate limit headers mismatch")
		})
	}

	t.Run("returns error for missing rate limit headers", func(t *testing.T) {
		invalidHeaders := map[string]string{}
		resp, err := getRespWithHeaders(invalidHeaders)
		is.NoError(err)

		_, err = resp.GetRateLimitHeaders()
		is.Error(err)
	})
}

func TestMessagesWithCaching(t *testing.T) {
	is := require.New(t)
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

	t.Run("caches single message", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				{
					Role: anthropic.RoleUser,
					Content: []anthropic.MessageContent{
						{
							Type: anthropic.MessagesContentTypeText,
							Text: toPtr("Is there a doctor on board?"),
							CacheControl: &anthropic.MessageCacheControl{
								Type: anthropic.CacheControlTypeEphemeral,
							},
						},
					},
				},
			},
			MaxTokens: 1000,
		})
		is.NoError(err)

		t.Logf("CreateMessages resp: %+v", resp)
	})

	t.Run("caches a multi-system message", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			MultiSystem: []anthropic.MessageSystemPart{
				{
					Type: "text",
					Text: "You are on a plane. You hear a voice over the intercom: 'Is there a doctor on board?'",
				},
				{
					Type: "text",
					Text: "<the entire contents of the safety card, the safety demonstration, and your medical training>",
					CacheControl: &anthropic.MessageCacheControl{
						Type: anthropic.CacheControlTypeEphemeral,
					},
				},
			},
			Messages: []anthropic.Message{
				{
					Role: anthropic.RoleUser,
					Content: []anthropic.MessageContent{
						{
							Type: anthropic.MessagesContentTypeText,
							Text: toPtr("Is there a doctor on board?"),
						},
					},
				},
			},
			MaxTokens: 1000,
		})
		is.NoError(err)

		t.Logf("CreateMessages resp: %+v", resp)
	})

}

func TestSetCacheControl(t *testing.T) {
	is := require.New(t)
	mc := anthropic.MessageContent{
		Type: anthropic.MessagesContentTypeText,
		Text: toPtr("hello"),
	}

	t.Run("sets cache control", func(t *testing.T) {
		mc.SetCacheControl(anthropic.CacheControlTypeEphemeral)

		is.NotNil(mc.CacheControl, "expected cache control to be set")
		is.Equal(anthropic.CacheControlTypeEphemeral, mc.CacheControl.Type)
	})
}

func TestMessagesRequest_MarshalJSON(t *testing.T) {
	is := require.New(t)
	t.Run("marshals MessagesRequest with system", func(t *testing.T) {
		req := anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
			System:    "test",
		}

		bs, err := json.Marshal(req)
		is.NoError(err)

		expected := `{"system":"test","model":"claude-3-haiku-20240307","messages":[{"role":"user","content":[{"type":"text","text":"What is your name?"}]}],"max_tokens":1000}`
		is.Equal(expected, string(bs), "marshalled MessagesRequest mismatch")
	})

	t.Run("marshals MessagesRequest with multi system", func(t *testing.T) {
		req := anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
			MultiSystem: []anthropic.MessageSystemPart{
				{
					Type: "text",
					Text: "test",
				},
			},
		}

		bs, err := json.Marshal(req)
		is.NoError(err)

		expected := `{"system":[{"type":"text","text":"test"}],"model":"claude-3-haiku-20240307","messages":[{"role":"user","content":[{"type":"text","text":"What is your name?"}]}],"max_tokens":1000}`
		is.Equal(expected, string(bs), "marshalled MessagesRequest mismatch")
	})

	t.Run("marshals MessagesRequest with no system", func(t *testing.T) {
		req := anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
		}

		bs, err := json.Marshal(req)
		is.NoError(err)

		expected := `{"model":"claude-3-haiku-20240307","messages":[{"role":"user","content":[{"type":"text","text":"What is your name?"}]}],"max_tokens":1000}`
		is.Equal(expected, string(bs), "marshalled MessagesRequest mismatch")
	})
}

func TestUsageHeaders(t *testing.T) {
	is := require.New(t)

	resp, err := getRespWithHeaders(rateLimitHeaders)
	is.NoError(err)

	usage := resp.Usage
	is.Equal(10, usage.InputTokens)
	is.Equal(10, usage.OutputTokens)
	is.Equal(0, usage.CacheCreationInputTokens)
	is.Equal(0, usage.CacheReadInputTokens)
}

func getRespWithHeaders(headers map[string]string) (anthropic.MessagesResponse, error) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handleMessagesEndpoint(headers))

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
	)

	return client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Haiku20240307,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage("What is your name?"),
		},
		MaxTokens: 1000,
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

func toPtr[T any](s T) *T {
	return &s
}
