package anthropic

import (
        "encoding/json"
        "testing"

        "github.com/stretchr/testify/assert"
)

func TestCitationsMessageContent(t *testing.T) {
        // Test creating a text document with citations enabled
        doc := NewTextDocumentMessageContent(
                "The grass is green. The sky is blue.",
                "My Document",
                "This is a trustworthy document.",
                true,
        )

        assert.Equal(t, MessagesContentTypeDocument, doc.Type)
        assert.Equal(t, "My Document", doc.Title)
        assert.Equal(t, "This is a trustworthy document.", doc.Context)
        assert.NotNil(t, doc.DocumentCitations)
        assert.True(t, doc.DocumentCitations.Enabled)
        assert.Equal(t, string(MessagesContentSourceTypeText), string(doc.Source.Type))
        assert.Equal(t, "text/plain", doc.Source.MediaType)
        assert.Equal(t, "The grass is green. The sky is blue.", doc.Source.Data)

        // Test creating a custom content document
        content := []MessageContent{
                {Type: MessagesContentTypeText, Text: strPtr("First chunk")},
                {Type: MessagesContentTypeText, Text: strPtr("Second chunk")},
        }
        customDoc := NewCustomContentDocumentMessageContent(
                content,
                "Custom Document",
                "Document with custom chunks",
                true,
        )

        assert.Equal(t, MessagesContentTypeDocument, customDoc.Type)
        assert.Equal(t, "Custom Document", customDoc.Title)
        assert.Equal(t, "Document with custom chunks", customDoc.Context)
        assert.NotNil(t, customDoc.DocumentCitations)
        assert.True(t, customDoc.DocumentCitations.Enabled)
        assert.Equal(t, string(MessagesContentSourceTypeContent), string(customDoc.Source.Type))
        assert.Equal(t, content, customDoc.Source.Content)

        // Test merging citations delta
        textContent := MessageContent{
                Type: MessagesContentTypeText,
                Text: strPtr("Some text"),
        }

        citation := Citation{
                Type:           CitationTypeCharLocation,
                CitedText:     "The grass is green.",
                DocumentIndex: 0,
                DocumentTitle: "My Document",
                StartCharIndex: intPtr(0),
                EndCharIndex:   intPtr(20),
        }

        delta := MessageContent{
                Type:     MessagesContentTypeCitationsDelta,
                Citation: &citation,
        }

        textContent.MergeContentDelta(delta)
        assert.Len(t, textContent.Citations, 1)
        assert.Equal(t, citation, textContent.Citations[0])

        // Test JSON marshaling/unmarshaling
        jsonData := `{
                "type": "text",
                "text": "the grass is green",
                "citations": [
                        {
                                "type": "char_location",
                                "cited_text": "The grass is green.",
                                "document_index": 0,
                                "document_title": "My Document",
                                "start_char_index": 0,
                                "end_char_index": 20
                        }
                ]
        }`

        var msgContent MessageContent
        err := json.Unmarshal([]byte(jsonData), &msgContent)
        assert.NoError(t, err)
        
        // Print the unmarshaled content for debugging
        debugJson, _ := json.MarshalIndent(msgContent, "", "  ")
        t.Logf("Unmarshaled content: %s", string(debugJson))
        
        assert.Equal(t, MessagesContentTypeText, msgContent.Type)
        assert.Equal(t, "the grass is green", *msgContent.Text)
        
        // Initialize Citations if nil
        if msgContent.Citations == nil {
                msgContent.Citations = make([]Citation, 0)
        }
        
        assert.Len(t, msgContent.Citations, 1)
        assert.Equal(t, CitationTypeCharLocation, msgContent.Citations[0].Type)
        assert.Equal(t, "The grass is green.", msgContent.Citations[0].CitedText)
}

func strPtr(s string) *string {
        return &s
}

func intPtr(i int) *int {
        return &i
}