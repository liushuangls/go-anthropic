package anthropic_test

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
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
	"anthropic-ratelimit-requests-limit":          "100",
	"anthropic-ratelimit-requests-remaining":      "99",
	"anthropic-ratelimit-requests-reset":          "2024-06-04T07:13:19Z",
	"anthropic-ratelimit-tokens-limit":            "10000",
	"anthropic-ratelimit-tokens-remaining":        "9900",
	"anthropic-ratelimit-tokens-reset":            "2024-06-04T07:13:19Z",
	"anthropic-ratelimit-input-tokens-limit":      "100",
	"anthropic-ratelimit-input-tokens-remaining":  "100",
	"anthropic-ratelimit-input-tokens-reset":      "2024-06-04T07:13:19Z",
	"anthropic-ratelimit-output-tokens-limit":     "100",
	"anthropic-ratelimit-output-tokens-remaining": "100",
	"anthropic-ratelimit-output-tokens-reset":     "2024-06-04T07:13:19Z",
	"retry-after": "", // retry-after is optional and may not be present.
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

	t.Run("create messages success", func(t *testing.T) {
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

		t.Logf("CreateMessages resp: %+v", resp)
	})

	t.Run("create messages success with single system message", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
			System:    "test system message",
		})
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		t.Logf("CreateMessages resp: %+v", resp)
	})

	t.Run("create messages success with single multi-system message", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens:   1000,
			MultiSystem: anthropic.NewMultiSystemMessages("test single multi-system message"),
		})
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		t.Logf("CreateMessages resp: %+v", resp)
	})

	t.Run("create messages success with multi-system messages", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
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
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		t.Logf("CreateMessages resp: %+v", resp)
	})

}

func TestNewUserTextMessage(t *testing.T) {
	m := anthropic.NewUserTextMessage("What is your name?")
	if m.Role != anthropic.RoleUser {
		t.Fatalf("Role mismatch. got %s, want %s", m.Role, anthropic.RoleUser)
	}

	if m.Content[0].Type != anthropic.MessagesContentTypeText {
		t.Fatalf(
			"Content type mismatch. got %s, want %s",
			m.Content[0].Type,
			anthropic.MessagesContentTypeText,
		)
	}

	if *m.Content[0].Text != "What is your name?" {
		t.Fatalf("Content text mismatch. got %s, want %s", *m.Content[0].Text, "What is your name?")
	}
}

func TestNewAssistantTextMessage(t *testing.T) {
	m := anthropic.NewAssistantTextMessage("My name is Claude.")
	if m.Role != anthropic.RoleAssistant {
		t.Fatalf("Role mismatch. got %s, want %s", m.Role, anthropic.RoleAssistant)
	}

	if m.Content[0].Type != anthropic.MessagesContentTypeText {
		t.Fatalf(
			"Content type mismatch. got %s, want %s",
			m.Content[0].Type,
			anthropic.MessagesContentTypeText,
		)
	}

	if *m.Content[0].Text != "My name is Claude." {
		t.Fatalf("Content text mismatch. got %s, want %s", *m.Content[0].Text, "My name is Claude.")
	}
}

func TestGetFirstContent(t *testing.T) {
	t.Run("returns empty content", func(t *testing.T) {
		m := anthropic.Message{}
		c := m.GetFirstContent()
		if c.Type != "" {
			t.Fatalf("Content type mismatch. got %s, want %s", c.Type, "")
		}
		if c.Text != nil {
			t.Fatalf("Content text mismatch. got %s, want %s", *c.Text, "")
		}
	})

	t.Run("returns single content", func(t *testing.T) {
		m := anthropic.NewAssistantTextMessage("My name is Claude.")
		c := m.GetFirstContent()
		if c.Type != anthropic.MessagesContentTypeText {
			t.Fatalf(
				"Content type mismatch. got %s, want %s",
				c.Type,
				anthropic.MessagesContentTypeText,
			)
		}

		if *c.Text != "My name is Claude." {
			t.Fatalf("Content text mismatch. got %s, want %s", *c.Text, "My name is Claude.")
		}
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
		if c.Type != anthropic.MessagesContentTypeText {
			t.Fatalf(
				"Content type mismatch. got %s, want %s",
				c.Type,
				anthropic.MessagesContentTypeText,
			)
		}

		if *c.Text != "My name is Claude." {
			t.Fatalf("Content text mismatch. got %s, want %s", *c.Text, "My name is Claude.")
		}
	})
}

func TestGetFirstContentText(t *testing.T) {
	t.Run("returns empty text", func(t *testing.T) {
		m := anthropic.MessagesResponse{}
		if m.GetFirstContentText() != "" {
			t.Fatalf("Content text mismatch. got %s, want %s", m.GetFirstContentText(), "")
		}
	})

	t.Run("returns text", func(t *testing.T) {
		m := anthropic.MessagesResponse{
			Content: []anthropic.MessageContent{
				anthropic.NewTextMessageContent("test string"),
			},
		}
		if m.GetFirstContentText() != "test string" {
			t.Fatalf("Content text mismatch. got %s, want %s", m.GetFirstContentText(), "")
		}
	})
}

func TestGetText(t *testing.T) {
	t.Run("returns empty text", func(t *testing.T) {
		c := anthropic.MessageContent{}
		if c.GetText() != "" {
			t.Fatalf("Content text mismatch. got %s, want %s", c.GetText(), "")
		}
	})

	t.Run("returns text", func(t *testing.T) {
		c := anthropic.NewTextMessageContent("My name is Claude.")
		if c.GetText() != "My name is Claude." {
			t.Fatalf("Content text mismatch. got %s, want %s", c.GetText(), "My name is Claude.")
		}
	})
}

func TestConcatText(t *testing.T) {
	t.Run("concatenates text when text content text present", func(t *testing.T) {
		mc := anthropic.NewTextMessageContent("original")
		mc.ConcatText(" added")

		if mc.GetText() != "original added" {
			t.Fatalf("Content text mismatch. got %s, want %s", mc.GetText(), "original added")
		}
	})

	t.Run("concatenates text when text content text not present", func(t *testing.T) {
		mc := anthropic.MessageContent{}
		mc.ConcatText("added")

		if mc.GetText() != "added" {
			t.Fatalf("Content text mismatch. got %s, want %s", mc.GetText(), "added")
		}
	})
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
		Model: anthropic.ModelClaude3Haiku20240307,
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
					anthropic.NewImageMessageContent(
						anthropic.NewMessageContentSource(
							anthropic.MessagesContentSourceTypeBase64,
							imageMediaType,
							imageData,
						),
					),
					anthropic.NewTextMessageContent("Describe these images."),
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

func TestMessagesVisionV2(t *testing.T) {
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

	imageUrl := "https://encrypted-tbn2.gstatic.com/images?q=tbn:ANd9GcQotvvFFhgI3W0wmWTyZ6MIbP3c5kLX6McyarVRFBLFf9T6eGuQoFQiP-S2RA181pSZkurOOMSRHKMqPP0dvumCu31mf29w-Hdlll_7Cg"
	resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaudeSonnet4V20250514,
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewImageUrlMessageContent(imageUrl),
					anthropic.NewTextMessageContent("Describe these images."),
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

	request.Messages = append(
		request.Messages,
		anthropic.NewToolResultsMessage(toolUse.ID, "65 degrees", false),
	)

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
	expectedSuccess := map[string]any{
		"anthropic-ratelimit-requests-limit":          100,
		"anthropic-ratelimit-requests-remaining":      99,
		"anthropic-ratelimit-requests-reset":          "2024-06-04T07:13:19Z",
		"anthropic-ratelimit-tokens-limit":            10000,
		"anthropic-ratelimit-tokens-remaining":        9900,
		"anthropic-ratelimit-tokens-reset":            "2024-06-04T07:13:19Z",
		"anthropic-ratelimit-input-tokens-limit":      100,
		"anthropic-ratelimit-input-tokens-remaining":  100,
		"anthropic-ratelimit-input-tokens-reset":      "2024-06-04T07:13:19Z",
		"anthropic-ratelimit-output-tokens-limit":     100,
		"anthropic-ratelimit-output-tokens-remaining": 100,
		"anthropic-ratelimit-output-tokens-reset":     "2024-06-04T07:13:19Z",
		"retry-after": -1, // if not present, should be -1
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

			bs2, err := json.Marshal(c.expected)
			if err != nil {
				t.Fatal(err)
			}

			var rlHeader2 anthropic.RateLimitHeaders
			if err := json.Unmarshal(bs2, &rlHeader2); err != nil {
				t.Fatal(err)
			}

			if rlHeaders != rlHeader2 {
				t.Fatalf("rate limit headers mismatch. got %s, want %s", string(bs), string(bs2))
			}
		})
	}

	t.Run("returns error for missing rate limit headers", func(t *testing.T) {
		invalidHeaders := map[string]string{}
		resp, err := getRespWithHeaders(invalidHeaders)
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		_, err = resp.GetRateLimitHeaders()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestMessagesWithCaching(t *testing.T) {
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
			Model: anthropic.ModelClaude3Haiku20240307,
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
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		t.Logf("CreateMessages resp: %+v", resp)
	})

	t.Run("caches a multi-system message", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
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
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		t.Logf("CreateMessages resp: %+v", resp)
	})

}

func TestSetCacheControl(t *testing.T) {
	t.Run("sets cache control", func(t *testing.T) {
		mc := anthropic.MessageContent{
			Type: anthropic.MessagesContentTypeText,
			Text: toPtr("hello"),
		}

		mc.SetCacheControl(anthropic.CacheControlTypeEphemeral)
		if mc.CacheControl == nil {
			t.Fatal("expected cache control to be set")
		}

		if mc.CacheControl.Type != anthropic.CacheControlTypeEphemeral {
			t.Fatalf(
				"expected cache control type to be %s, got %s",
				anthropic.CacheControlTypeEphemeral,
				mc.CacheControl.Type,
			)
		}

		if mc.CacheControl.TTL != "" {
			t.Fatalf("expected cache control ttl to be empty, got %s", mc.CacheControl.TTL)
		}
	})

	t.Run("sets cache control with 1h ttl", func(t *testing.T) {
		mc := anthropic.MessageContent{
			Type: anthropic.MessagesContentTypeText,
			Text: toPtr("hello"),
		}

		mc.SetCacheControlTTL(anthropic.CacheControlTypeEphemeral, anthropic.CacheControlTTL1h)
		if mc.CacheControl == nil {
			t.Fatal("expected cache control to be set")
		}

		if mc.CacheControl.Type != anthropic.CacheControlTypeEphemeral {
			t.Fatalf(
				"expected cache control type to be %s, got %s",
				anthropic.CacheControlTypeEphemeral,
				mc.CacheControl.Type,
			)
		}

		if mc.CacheControl.TTL != anthropic.CacheControlTTL1h {
			t.Fatalf(
				"expected cache control ttl to be %s, got %s",
				anthropic.CacheControlTTL1h,
				mc.CacheControl.TTL,
			)
		}
	})

	t.Run("sets cache control with 5m ttl", func(t *testing.T) {
		mc := anthropic.MessageContent{
			Type: anthropic.MessagesContentTypeText,
			Text: toPtr("hello"),
		}

		mc.SetCacheControlTTL(anthropic.CacheControlTypeEphemeral, anthropic.CacheControlTTL5m)
		if mc.CacheControl == nil {
			t.Fatal("expected cache control to be set")
		}

		if mc.CacheControl.Type != anthropic.CacheControlTypeEphemeral {
			t.Fatalf(
				"expected cache control type to be %s, got %s",
				anthropic.CacheControlTypeEphemeral,
				mc.CacheControl.Type,
			)
		}

		if mc.CacheControl.TTL != anthropic.CacheControlTTL5m {
			t.Fatalf(
				"expected cache control ttl to be %s, got %s",
				anthropic.CacheControlTTL5m,
				mc.CacheControl.TTL,
			)
		}
	})
}

func TestMessageCacheControlMarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		cacheControl anthropic.MessageCacheControl
		expected     string
	}{
		{
			name: "marshals without ttl",
			cacheControl: anthropic.MessageCacheControl{
				Type: anthropic.CacheControlTypeEphemeral,
			},
			expected: `{"type":"ephemeral"}`,
		},
		{
			name: "marshals with 1h ttl",
			cacheControl: anthropic.MessageCacheControl{
				Type: anthropic.CacheControlTypeEphemeral,
				TTL:  anthropic.CacheControlTTL1h,
			},
			expected: `{"type":"ephemeral","ttl":"1h"}`,
		},
		{
			name: "marshals with 5m ttl",
			cacheControl: anthropic.MessageCacheControl{
				Type: anthropic.CacheControlTypeEphemeral,
				TTL:  anthropic.CacheControlTTL5m,
			},
			expected: `{"type":"ephemeral","ttl":"5m"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs, err := json.Marshal(tt.cacheControl)
			if err != nil {
				t.Fatal(err)
			}

			if string(bs) != tt.expected {
				t.Fatalf(
					"marshalled MessageCacheControl mismatch. \ngot %s, \nwant %s",
					string(bs),
					tt.expected,
				)
			}
		})
	}
}

func TestMessagesRequest_MarshalJSON(t *testing.T) {
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
		if err != nil {
			t.Fatal(err)
		}

		expected := `{"system":"test","model":"claude-3-haiku-20240307","messages":[{"role":"user","content":[{"type":"text","text":"What is your name?"}]}],"max_tokens":1000}`
		if string(bs) != expected {
			t.Fatalf(
				"marshalled MessagesRequest mismatch. \ngot %s, \nwant %s",
				string(bs),
				expected,
			)
		}
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
		if err != nil {
			t.Fatal(err)
		}

		expected := `{"system":[{"type":"text","text":"test"}],"model":"claude-3-haiku-20240307","messages":[{"role":"user","content":[{"type":"text","text":"What is your name?"}]}],"max_tokens":1000}`
		if string(bs) != expected {
			t.Fatalf(
				"marshalled MessagesRequest mismatch. \ngot %s, \nwant %s",
				string(bs),
				expected,
			)
		}
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
		if err != nil {
			t.Fatal(err)
		}

		expected := `{"model":"claude-3-haiku-20240307","messages":[{"role":"user","content":[{"type":"text","text":"What is your name?"}]}],"max_tokens":1000}`
		if string(bs) != expected {
			t.Fatalf(
				"marshalled MessagesRequest mismatch. \ngot %s, \nwant %s",
				string(bs),
				expected,
			)
		}
	})
}

func TestUsageHeaders(t *testing.T) {
	resp, err := getRespWithHeaders(rateLimitHeaders)
	if err != nil {
		t.Fatalf("CreateMessages error: %v", err)
	}

	usage := resp.Usage
	if usage.InputTokens != 10 {
		t.Fatalf("InputTokens mismatch. got %d, want 10", usage.InputTokens)
	}

	if usage.OutputTokens != 10 {
		t.Fatalf("OutputTokens mismatch. got %d, want 10", usage.OutputTokens)
	}

	if usage.CacheCreationInputTokens != 0 {
		t.Fatalf(
			"CacheCreationInputTokens mismatch. got %d, want 0",
			usage.CacheCreationInputTokens,
		)
	}

	if usage.CacheReadInputTokens != 0 {
		t.Fatalf("CacheReadInputTokens mismatch. got %d, want 0", usage.CacheReadInputTokens)
	}
}

func TestVertexMessages(t *testing.T) {
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
	server.RegisterHandler(
		baseEndpoint+"/"+vertexModel+":rawPredict",
		handleMessagesEndpoint(rateLimitHeaders),
	)

	ts := server.VertexTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + baseEndpoint
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithVertexAI(project, location),
		anthropic.WithBaseURL(baseUrl),
		anthropic.WithEmptyMessagesLimit(100),
		anthropic.WithHTTPClient(http.DefaultClient),
	)

	t.Run("create messages success", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: model,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
		})
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		t.Logf("CreateMessages resp: %+v", resp)
	})

	t.Run("create messages success with single system message", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: model,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
			System:    "test system message",
		})
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		t.Logf("CreateMessages resp: %+v", resp)
	})

	t.Run("create messages success with single multi-system message", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: model,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens:   1000,
			MultiSystem: anthropic.NewMultiSystemMessages("test single multi-system message"),
		})
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		t.Logf("CreateMessages resp: %+v", resp)
	})

	t.Run("create messages success with multi-system messages", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: model,
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
		if err != nil {
			t.Fatalf("CreateMessages error: %v", err)
		}

		t.Logf("CreateMessages resp: %+v", resp)
	})

}

func TestVertexUnauthorized(t *testing.T) {
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
	server.RegisterHandler(
		baseEndpoint+"/"+vertexModel+":rawPredict",
		handleMessagesEndpoint(rateLimitHeaders),
	)

	ts := server.VertexTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + baseEndpoint
	client := anthropic.NewClient(
		"wrong-token",
		anthropic.WithVertexAI(project, location),
		anthropic.WithBaseURL(baseUrl),
		anthropic.WithEmptyMessagesLimit(100),
		anthropic.WithHTTPClient(http.DefaultClient),
	)

	t.Run("create messages auth error", func(t *testing.T) {
		resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: model,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens: 1000,
		})
		if err == nil {
			t.Fatalf("CreateMessages expected error: %v", err)
		}

		if respErr, ok := err.(*anthropic.RequestError); ok {
			if respErr.StatusCode != http.StatusUnauthorized {
				t.Fatalf(
					"CreateMessages expected status code %d, got %d",
					http.StatusUnauthorized,
					respErr.StatusCode,
				)
			}

			// // get the []VertexAIErrorResponse
			// var errs []anthropic.VertexAIErrorResponse
			// if respErr.Err != nil {
			// 	if (vErr, ok:= respErr.Err.(*anthropic.VertexAIError); ok {
			// }
			// if respErr. != "Unauthorized" {
			// 	t.Fatalf("CreateMessages expected message 'Unauthorized', got %s", respErr.Message)
			// }
		} else {
			ve := &anthropic.VertexAPIError{}
			if !errors.As(err, &ve) {
				t.Fatalf("CreateMessages expected VertexAIError, got %v", err)
			}

			if ve.Code != http.StatusUnauthorized {
				t.Fatalf("CreateMessages expected status code %d, got %d", http.StatusUnauthorized, ve.Code)
			}
			if ve.Message != "Unauthorized" {
				t.Fatalf("CreateMessages expected message 'Unauthorized', got %s", ve.Message)
			}
		}

		t.Logf("CreateMessages resp: %+v", resp)
	})
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
		if messagesReq, err = getRequest[anthropic.MessagesRequest](r); err != nil {
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
					anthropic.NewTextMessageContent(
						"The current weather in San Francisco is 65 degrees Fahrenheit. It's a nice, moderate temperature typical of the San Francisco Bay Area climate.",
					),
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

func TestMessagesWebSearch(t *testing.T) {

	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", func(w http.ResponseWriter, r *http.Request) {
		var req anthropic.MessagesRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		// Check if this is a request with web search tool results
		hasWebSearchResult := false
		for _, msg := range req.Messages {
			for _, content := range msg.Content {
				if content.Type == anthropic.MessagesContentTypeWebSearchToolResult {
					hasWebSearchResult = true
					break
				}
			}
		}

		var responseJSON string

		if hasWebSearchResult {
			// Final response with citations - build JSON manually to handle citations field correctly
			responseJSON = `{
				"id": "msg_01ABC123",
				"type": "message",
				"role": "assistant",
				"model": "` + string(req.Model) + `",
				"content": [{
					"type": "text",
					"text": "Go is a statically typed, compiled programming language designed at Google",
					"citations": [{
						"type": "web_search_result_location",
						"url": "https://en.wikipedia.org/wiki/Go_(programming_language)",
						"title": "Go (programming language) - Wikipedia",
						"cited_text": "Go is a statically typed, compiled programming language designed at Google",
						"encrypted_index": "encrypted_index_abc123"
					}]
				}],
				"stop_reason": "end_turn",
				"usage": {
					"input_tokens": 100,
					"output_tokens": 200,
					"server_tool_use": {
						"web_search_requests": 1
					}
				}
			}`
		} else {
			// Initial response with server_tool_use and web_search_tool_result
			responseJSON = `{
				"id": "msg_01ABC123",
				"type": "message",
				"role": "assistant",
				"model": "` + string(req.Model) + `",
				"content": [
					{
						"type": "text",
						"text": "I'll search for information about the Go programming language."
					},
					{
						"type": "server_tool_use",
						"id": "srvtoolu_01ABC123",
						"name": "web_search",
						"input": {"query": "what is go programming language"}
					},
					{
						"type": "web_search_tool_result",
						"tool_use_id": "srvtoolu_01ABC123",
						"content": [{
							"type": "web_search_result",
							"url": "https://en.wikipedia.org/wiki/Go_(programming_language)",
							"title": "Go (programming language) - Wikipedia",
							"encrypted_content": "encrypted_content_xyz789",
							"page_age": "2025-06-01"
						}]
					}
				],
				"stop_reason": "pause_turn",
				"usage": {
					"input_tokens": 100,
					"output_tokens": 200
				}
			}`
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	})

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
			anthropic.NewUserTextMessage("What is the Go programming language?"),
		},
		MaxTokens: 1000,
		Tools: []anthropic.ToolDefinition{
			{
				Type: "web_search_20250305",
				Name: "web_search",
			},
		},
	}

	// First request - Claude decides to search
	resp, err := client.CreateMessages(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StopReason != anthropic.MessagesStopReasonPauseTurn {
		t.Fatalf("Expected stop_reason pause_turn, got: %s", resp.StopReason)
	}

	// Find the web search result to send back
	var webSearchResult *anthropic.MessageContent
	for i, c := range resp.Content {
		if c.Type == anthropic.MessagesContentTypeWebSearchToolResult {
			webSearchResult = &resp.Content[i]
		}
	}

	if webSearchResult == nil {
		t.Fatal("web_search_tool_result not found in response")
	}

	// Add the assistant's response to the conversation
	request.Messages = append(request.Messages, anthropic.Message{
		Role:    anthropic.RoleAssistant,
		Content: resp.Content,
	})

	// Second request - Continue with the search results
	resp, err = client.CreateMessages(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Final response: %+v", resp)

	// Verify we got citations in the final response
	var hasCitations bool
	for _, content := range resp.Content {
		if len(content.Citations) > 0 {
			hasCitations = true
			for _, citation := range content.Citations {
				if citation.Type == anthropic.CitationTypeWebSearchResultLocation {
					t.Logf("Citation: %s - %s", *citation.Url, *citation.Title)
				}
			}
		}
	}

	if !hasCitations {
		t.Fatal("Expected citations in final response")
	}

	// Verify usage tracking
	if resp.Usage.ServerToolUse == nil {
		t.Fatal("Expected ServerToolUse in usage tracking")
	}

	if resp.Usage.ServerToolUse.WebSearchRequests != 1 {
		t.Fatalf(
			"Expected 1 web search request, got: %d",
			resp.Usage.ServerToolUse.WebSearchRequests,
		)
	}
}

func TestCitationTypePageLocation(t *testing.T) {
	if anthropic.CitationTypePageLocation != "page_location" {
		t.Fatalf("unexpected value: %q", anthropic.CitationTypePageLocation)
	}

	body := `{"type":"text","text":"x","citations":[{"type":"page_location","cited_text":"y","document_index":0}]}`
	var mc anthropic.MessageContent
	if err := json.Unmarshal([]byte(body), &mc); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(mc.Citations) != 1 {
		t.Fatalf("expected 1 citation, got %d", len(mc.Citations))
	}
	if mc.Citations[0].Type != anthropic.CitationTypePageLocation {
		t.Fatalf("unexpected citation type: %q", mc.Citations[0].Type)
	}
}

func TestCitationPageNumberFields(t *testing.T) {
	body := `{
		"type": "text",
		"text": "x",
		"citations": [
			{
				"type": "page_location",
				"cited_text": "y",
				"document_index": 0,
				"start_page_number": 3,
				"end_page_number": 5
			}
		]
	}`
	var mc anthropic.MessageContent
	if err := json.Unmarshal([]byte(body), &mc); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(mc.Citations) != 1 {
		t.Fatalf("expected 1 citation, got %d", len(mc.Citations))
	}
	c := mc.Citations[0]
	if c.StartPage == nil || *c.StartPage != 3 {
		t.Fatalf("unexpected start_page_number: %v", c.StartPage)
	}
	if c.EndPage == nil || *c.EndPage != 5 {
		t.Fatalf("unexpected end_page_number: %v", c.EndPage)
	}
}
