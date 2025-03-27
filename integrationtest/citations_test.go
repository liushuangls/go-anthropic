package integrationtest

import (
	"context"
	"embed"
	"encoding/base64"
	"io"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
)

//go:embed sources/apology.pdf
var sources embed.FS

func TestIntegrationCitations(t *testing.T) {
	testAPIKey(t)
	client := anthropic.NewClient(APIKey)
	ctx := context.Background()

	// Only works with Claude 3.7 Sonnet, 3.5 Sonnet, and 3.5 Haiku
	// Citations not available through Vertex API

	// Create a textDocument with citations enabled
	textDocument := anthropic.NewTextDocumentMessageContent(
		"The sky is blue. The grass is green. Water boils at 100 degrees Celsius. The sky is blue because of Rayleigh scattering.",
		"Facts Document",
		"This textDocument contains simple scientific facts.",
		true, // Enable citations
	)

	customDocument := anthropic.NewCustomContentDocumentMessageContent(
		stringSliceToCustomContentDocument([]string{
			"The Tao that can be told is not the eternal Tao.",
			"The name that can be named is not the eternal name.",
			"The nameless is the beginning of heaven and earth.",
			"The named is the mother of ten thousand things.",
		}),
		"Tao Te Ching",
		"This textDocument contains the opening of the Tao Te Ching.",
		true, // Enable citations
	)

	pdfPath := "sources/apology.pdf"
	pdfFile, err := sources.Open(pdfPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	fileBytes, err := io.ReadAll(pdfFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	pdfDocument := anthropic.NewPDFDocumentMessageContent(
		base64.StdEncoding.EncodeToString(fileBytes),
		"Plato's Apology",
		"The trial and apology of socrates by plato",
		true)

	// Create a request that uses the textDocument and asks about its content
	request := anthropic.MessagesRequest{
		Model:     anthropic.ModelClaude3Dot7Sonnet20250219, // Only works with Claude 3.7 Sonnet, 3.5 Sonnet, and 3.5 Haiku
		MaxTokens: 1024,
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					textDocument,
					anthropic.NewTextMessageContent("Why is the color of the sky blue according to the Facts Document? Use citations to back up your answer."),
				},
			},
		},
	}

	// Create a request that uses the textDocument and asks about its content
	requestCustomDocument := anthropic.MessagesRequest{
		Model:     anthropic.ModelClaude3Dot7Sonnet20250219,
		MaxTokens: 1024,
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					customDocument,
					anthropic.NewTextMessageContent("Is that which can be named the eternal name? Use citations to back up your answer."),
				},
			},
		},
	}

	// Create a request that uses the textDocument and asks about its content
	requestPDFDocument := anthropic.MessagesRequest{
		Model:     anthropic.ModelClaude3Dot7Sonnet20250219,
		MaxTokens: 1024,
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					pdfDocument,
					anthropic.NewTextMessageContent("Does Socrates admit to corrupting the youth? Please be brief and use citations to back up your answer."),
				},
			},
		},
	}

	t.Run("CreateMessages with citations on Claude 3.7 Sonnet", func(t *testing.T) {
		resp, err := client.CreateMessages(ctx, request)
		if err != nil {
			t.Fatalf("CreateMessages error: %s", err)
		}

		t.Logf("Response content: %+v", resp.Content)

		// Verify we got citations in the response
		hasCitations := false
		for _, content := range resp.Content {
			if len(content.Citations) > 0 {
				hasCitations = true
				t.Logf("Found citations: %+v", content.Citations)
				break
			}
		}

		if !hasCitations {
			t.Errorf("Expected citations in the response, but none were found")
		}
	})

	t.Run("CreateMessages with citations on Claude 3.7 Sonnet (Custom Document)", func(t *testing.T) {
		resp, err := client.CreateMessages(ctx, requestCustomDocument)
		if err != nil {
			t.Fatalf("CreateMessages error: %s", err)
		}

		t.Logf("Response content: %+v", resp.Content)

		// Verify we got citations in the response
		hasCitations := false
		for _, content := range resp.Content {
			if len(content.Citations) > 0 {
				hasCitations = true
				t.Logf("Found citations: %+v", content.Citations)
				break
			}
		}

		if !hasCitations {
			t.Errorf("Expected citations in the response, but none were found")
		}
	})

	t.Run("CreateMessages with citations on Claude 3.7 Sonnet (PDF Document)", func(t *testing.T) {
		resp, err := client.CreateMessages(ctx, requestPDFDocument)
		if err != nil {
			t.Fatalf("CreateMessages error: %s", err)
		}

		t.Logf("Response content: %+v", resp.Content)

		// Verify we got citations in the response
		hasCitations := false
		for _, content := range resp.Content {
			if len(content.Citations) > 0 {
				hasCitations = true
				t.Logf("Found citations: %+v", content.Citations)
				break
			}
		}

		if !hasCitations {
			t.Errorf("Expected citations in the response, but none were found")
		}
	})

	t.Run("CreateMessagesStream with citations on Claude 3.7 Sonnet", func(t *testing.T) {
		streamRequest := anthropic.MessagesStreamRequest{
			MessagesRequest: request,
		}

		// Track if we received any citation delta events
		receivedCitationDelta := false

		// Add verbose logging for all event types
		streamRequest.OnMessageStart = func(data anthropic.MessagesEventMessageStartData) {
			t.Logf("OnMessageStart event received: %+v", data.Type)
		}

		streamRequest.OnContentBlockStart = func(data anthropic.MessagesEventContentBlockStartData) {
			t.Logf("OnContentBlockStart event received: index=%d, type=%s",
				data.Index, data.ContentBlock.Type)
		}

		// Add callback to detect citation delta events in the content block delta
		streamRequest.OnContentBlockDelta = func(data anthropic.MessagesEventContentBlockDeltaData) {
			t.Logf("OnContentBlockDelta event received: index=%d, delta_type=%s",
				data.Index, data.Delta.Type)

			if data.Delta.Type == anthropic.MessagesContentTypeCitationsDelta {
				receivedCitationDelta = true
				t.Logf("Received citation delta: %+v", data.Delta.Citation)
			}
		}

		streamRequest.OnContentBlockStop = func(data anthropic.MessagesEventContentBlockStopData, content anthropic.MessageContent) {
			t.Logf("OnContentBlockStop event received: index=%d, content_type=%s",
				data.Index, content.Type)
			if len(content.Citations) > 0 {
				t.Logf("Citations in stopped block: %+v", content.Citations)
			}
		}

		streamRequest.OnMessageDelta = func(data anthropic.MessagesEventMessageDeltaData) {
			t.Logf("OnMessageDelta event received: type=%s", data.Type)
		}

		streamRequest.OnMessageStop = func(data anthropic.MessagesEventMessageStopData) {
			t.Logf("OnMessageStop event received")
		}

		resp, err := client.CreateMessagesStream(ctx, streamRequest)
		if err != nil {
			t.Fatalf("CreateMessagesStream error: %s", err)
		}

		t.Logf("Stream response content: %+v", resp.Content)

		// Check for citations in the final merged content
		hasCitations := false
		for _, content := range resp.Content {
			if len(content.Citations) > 0 {
				hasCitations = true
				t.Logf("Found citations in stream response: %+v", content.Citations)
				break
			}
		}

		if !hasCitations {
			t.Errorf("Expected citations in the stream response, but none were found")
		}

		if !receivedCitationDelta {
			t.Errorf("Expected to receive citation delta events, but none were detected")
		}
	})

	// Test with Claude 3.5 Haiku
	t.Run("CreateMessages with citations on Claude 3.5 Haiku", func(t *testing.T) {
		haikuRequest := request
		haikuRequest.Model = anthropic.ModelClaude3Dot5HaikuLatest

		resp, err := client.CreateMessages(ctx, haikuRequest)
		if err != nil {
			t.Fatalf("CreateMessages error with Claude 3.5 Haiku: %s", err)
		}

		t.Logf("Claude 3.5 Haiku response content: %+v", resp.Content)

		// Verify we got citations in the response
		hasCitations := false
		for _, content := range resp.Content {
			if len(content.Citations) > 0 {
				hasCitations = true
				t.Logf("Found citations from Claude 3.5 Haiku: %+v", content.Citations)
				break
			}
		}

		if !hasCitations {
			t.Errorf("Expected citations in the Claude 3.5 Haiku response, but none were found")
		}
	})

	// Test with custom content textDocument
	t.Run("CreateMessages with custom content textDocument", func(t *testing.T) {
		customContent := []anthropic.MessageContent{
			anthropic.NewTextMessageContent("The Earth orbits the Sun."),
			anthropic.NewTextMessageContent("The Moon orbits the Earth."),
		}

		customDoc := anthropic.NewCustomContentDocumentMessageContent(
			customContent,
			"Astronomy Facts",
			"This textDocument contains basic astronomy facts.",
			true, // Enable citations
		)

		customRequest := anthropic.MessagesRequest{
			Model:     anthropic.ModelClaude3Dot7Sonnet20250219,
			MaxTokens: 1024,
			Messages: []anthropic.Message{
				{
					Role: anthropic.RoleUser,
					Content: []anthropic.MessageContent{
						customDoc,
						anthropic.NewTextMessageContent("What orbits the Earth according to the textDocument? Please cite your sources."),
					},
				},
			},
		}

		resp, err := client.CreateMessages(ctx, customRequest)
		if err != nil {
			t.Fatalf("CreateMessages error with custom textDocument: %s", err)
		}

		t.Logf("Custom textDocument response content: %+v", resp.Content)

		// Verify we got citations in the response
		hasCitations := false
		for _, content := range resp.Content {
			if len(content.Citations) > 0 {
				hasCitations = true
				t.Logf("Found citations in custom textDocument response: %+v", content.Citations)
				break
			}
		}

		if !hasCitations {
			t.Errorf("Expected citations in the custom textDocument response, but none were found")
		}
	})
}

func stringSliceToCustomContentDocument(chunks []string) []anthropic.MessageContent {
	if len(chunks) == 0 {
		return nil
	}

	documents := make([]anthropic.MessageContent, 0, len(chunks))
	for _, chunk := range chunks {
		documents = append(documents, anthropic.MessageContent{
			Type: "text",
			Text: &chunk,
		})
	}
	return documents
}
