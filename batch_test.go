package anthropic_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/liushuangls/go-anthropic/v2/internal/test"
)

func TestCreateBatch(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages/batches", handleCreateBatchEndpoint)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
		anthropic.WithBetaVersion(anthropic.BetaMessageBatches20240924),
	)

	t.Run("create batch success", func(t *testing.T) {
		resp, err := client.CreateBatch(context.Background(), anthropic.BatchRequest{
			Requests: []anthropic.InnerRequests{
				{
					CustomId: "custom-identifier-not-real-this-is-a-test",
					Params: anthropic.MessagesRequest{
						Model:       anthropic.ModelClaude3Haiku20240307,
						MultiSystem: anthropic.NewMultiSystemMessages("you are an assistant", "you are snarky"),
						MaxTokens:   10,
						Messages: []anthropic.Message{
							anthropic.NewUserTextMessage("What is your name?"),
							anthropic.NewAssistantTextMessage("My name is Claude."),
							anthropic.NewUserTextMessage("What is your favorite color?"),
						},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("CreateBatch error: %s", err)
		}
		t.Logf("Create Batch resp: %+v", resp)
	})

	t.Run("fails with missing beta version header", func(t *testing.T) {
		clientWithoutBeta := anthropic.NewClient(
			test.GetTestToken(),
			anthropic.WithBaseURL(baseUrl),
		)
		_, err := clientWithoutBeta.CreateBatch(context.Background(), anthropic.BatchRequest{})
		if err == nil {
			t.Fatalf("CreateBatch expected error, got nil")
		}
	})

}

func handleCreateBatchEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// Creating batches only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	var completeReq anthropic.BatchRequest
	if completeReq, err = getRequest[anthropic.BatchRequest](r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}

	betaHeaders := r.Header.Get("Anthropic-Beta")
	if !strings.Contains(betaHeaders, string(anthropic.BetaMessageBatches20240924)) {
		http.Error(w, "missing beta version header", http.StatusBadRequest)
		return
	}

	custId := completeReq.Requests[0].CustomId
	if custId == "" {
		// I think this should be a bad request. TODO check docs
		http.Error(w, "custom_id is required", http.StatusBadRequest)
		return
	}

	t1 := time.Now().Add(-time.Hour * 2)

	res := anthropic.BatchResponse{
		BatchRespCore: anthropic.BatchRespCore{
			Id:               anthropic.BatchId("batch_id_" + strconv.FormatInt(time.Now().Unix(), 10)),
			Type:             anthropic.BatchResponseTypeMessageBatch,
			ProcessingStatus: anthropic.ProcessingStatusInProgress,
			RequestCounts: anthropic.RequestCounts{
				Processing: 1,
				Succeeded:  2,
				Canceled:   3,
				Errored:    4,
				Expired:    5,
			},
			EndedAt:           nil,
			CreatedAt:         t1,
			ExpiresAt:         t1.Add(time.Hour * 4),
			ArchivedAt:        nil,
			CancelInitiatedAt: nil,
			ResultsUrl:        nil,
		},
	}
	resBytes, _ = json.Marshal(res)
	_, _ = w.Write(resBytes)
}
