package integrationtest

import (
	"context"
	"math/rand"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
)

func randomString(l int) string {
	const charset = "1234567890abcdef"
	b := make([]byte, l)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func TestIntegrationBatch(t *testing.T) {
	testAPIKey(t)
	ctx := context.Background()

	myId := "rand_id_" + randomString(5)
	createBatchRequest := anthropic.BatchRequest{
		Requests: []anthropic.InnerRequests{
			{
				CustomId: myId,
				Params: anthropic.MessagesRequest{
					Model: anthropic.ModelClaude3Haiku20240307,
					MultiSystem: anthropic.NewMultiSystemMessages(
						"you are an assistant",
						"you are snarky",
					),
					MaxTokens: 10,
					Messages: []anthropic.Message{
						anthropic.NewUserTextMessage("What is your name?"),
						anthropic.NewAssistantTextMessage("My name is Claude."),
						anthropic.NewUserTextMessage("What is your favorite color?"),
					},
				},
			},
		},
	}

	betaOpts := anthropic.WithBetaVersion(
		anthropic.BetaTools20240404,
		anthropic.BetaMaxTokens35Sonnet20240715,
		anthropic.BetaMessageBatches20240924,
	)
	client := anthropic.NewClient(APIKey, betaOpts)

	// this will be set by the CreateBatch call below, and used in later tests
	var batchID anthropic.BatchId

	t.Run("CreateBatch on real API", func(t *testing.T) {
		resp, err := client.CreateBatch(ctx, createBatchRequest)
		if err != nil {
			t.Fatalf("CreateBatch error: %s", err)
		}
		t.Logf("CreateBatch resp: %+v", resp)

		// Save batchID for later tests
		batchID = resp.Id
	})

	t.Run("RetrieveBatch on real API", func(t *testing.T) {
		resp, err := client.RetrieveBatch(ctx, batchID)
		if err != nil {
			t.Fatalf("RetrieveBatch error: %s", err)
		}
		t.Logf("RetrieveBatch resp: %+v", resp)
	})

	var completedBatch *anthropic.BatchId
	t.Run("ListBatches on real API", func(t *testing.T) {
		req := anthropic.ListBatchRequest{
			Limit: 99,
		}
		resp, err := client.ListBatches(ctx, req)
		if err != nil {
			t.Fatalf("ListBatches error: %s", err)
		}
		t.Logf("ListBatches resp: %+v", resp)

		for _, batch := range resp.Data {
			if batch.ProcessingStatus == "ended" && batch.CancelInitiatedAt == nil {
				completedBatch = &batch.Id
				break
			}
		}
	})

	if completedBatch == nil {
		// We probably need a better way to test this, but for now we'll skip if there's no completed batch
		t.Skip("No completed batch to test RetrieveBatchResults")
	} else {
		// This test should be run after the first batch has completed.
		// You should have a completed batch in your account to run this test!
		// You can have a batch you run to completion by commenting out the CancelBatch call below.
		t.Run("RetrieveBatchResults on real API", func(t *testing.T) {
			resp, err := client.RetrieveBatchResults(ctx, *completedBatch)
			if err != nil {
				t.Fatalf("RetrieveBatchResults error: %s", err)
			}
			t.Logf("RetrieveBatchResults resp: %+v", resp)

			if len(resp.Responses) == 0 {
				t.Fatalf("RetrieveBatchResults returned no responses")
			}

			if resp.Responses[0].CustomId == "" {
				t.Fatalf("RetrieveBatchResults returned a response with no CustomId. Parse error?")
			}
		})
	}

	t.Run("CancelBatch on real API", func(t *testing.T) {
		resp, err := client.CancelBatch(ctx, batchID)
		if err != nil {
			t.Fatalf("CancelBatch error: %s", err)
		}
		t.Logf("CancelBatch resp: %+v", resp)
	})
}
