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
			anthropic.WithBetaVersion(anthropic.BetaMessageBatches20240924),
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
			Id: anthropic.BatchId(
				"batch_id_" + strconv.FormatInt(time.Now().Unix(), 10),
			),
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

func TestRetrieveBatch(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages/batches/batch_id_1234", handleRetrieveBatchEndpoint)
	server.RegisterHandler("/v1/messages/batches/batch_id_not_found", handleRetrieveBatchEndpoint)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
		anthropic.WithBetaVersion(anthropic.BetaMessageBatches20240924),
	)

	t.Run("retrieve batch success", func(t *testing.T) {
		resp, err := client.RetrieveBatch(context.Background(), "batch_id_1234")
		if err != nil {
			t.Fatalf("RetrieveBatch error: %s", err)
		}
		t.Logf("Retrieve Batch resp: %+v", resp)
	})

	t.Run("retrieve batch failure", func(t *testing.T) {
		_, err := client.RetrieveBatch(context.Background(), "batch_id_not_found")
		if err == nil {
			t.Fatalf("RetrieveBatch expected error, got nil")
		}
	})
}

func handleRetrieveBatchEndpoint(w http.ResponseWriter, r *http.Request) {
	var resBytes []byte

	// retrieving batches only accepts GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	batchId := strings.TrimPrefix(r.URL.Path, "/v1/messages/batches/")
	if batchId == "" {
		http.Error(w, "missing batch id", http.StatusBadRequest)
		return
	}

	if batchId == "batch_id_not_found" {
		http.Error(w, "batch not found", http.StatusNotFound)
		return
	}

	res := anthropic.BatchResponse{
		BatchRespCore: forgeBatchResponse(batchId),
	}
	resBytes, _ = json.Marshal(res)
	_, _ = w.Write(resBytes)
}

func TestRetrieveBatchResults(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler(
		"/v1/messages/batches/batch_id_1234/results",
		handleRetrieveBatchResultsEndpoint,
	)
	server.RegisterHandler(
		"/v1/messages/batches/batch_id_not_found/results",
		handleRetrieveBatchResultsEndpoint,
	)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
		anthropic.WithBetaVersion(anthropic.BetaMessageBatches20240924),
	)

	t.Run("retrieve batch results success", func(t *testing.T) {
		resp, err := client.RetrieveBatchResults(context.Background(), "batch_id_1234")
		if err != nil {
			t.Fatalf("RetrieveBatchResults error: %s", err)
		}
		t.Logf("Retrieve Batch Results resp: %+v", resp)
	})

	t.Run("retrieve batch results failure", func(t *testing.T) {
		_, err := client.RetrieveBatchResults(context.Background(), "batch_id_not_found")
		if err == nil {
			t.Fatalf("RetrieveBatchResults expected error, got nil")
		}
	})
}

func handleRetrieveBatchResultsEndpoint(w http.ResponseWriter, r *http.Request) {
	var resBytes []byte

	// retrieving batch results only accepts GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	batchId := strings.TrimPrefix(r.URL.Path, "/v1/messages/batches/")
	batchId = strings.TrimSuffix(batchId, "/results")
	if batchId == "" {
		http.Error(w, "missing batch id", http.StatusBadRequest)
		return
	}

	if batchId == "batch_id_not_found" {
		http.Error(w, "batch not found", http.StatusNotFound)
		return
	}

	res := anthropic.RetrieveBatchResultsResponse{
		Responses: []anthropic.BatchResult{
			{
				CustomId: "custom_id_1234",
				Result:   forgeBatchResult("batch_id_1234"),
			},
		},
	}
	resBytes, _ = json.Marshal(res)
	_, _ = w.Write(resBytes)
}

func TestRetrieveBatchResultsJSONL(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler(
		"/v1/messages/batches/batch_id_jsonl/results",
		handleRetrieveBatchResultsJSONLEndpoint,
	)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
		anthropic.WithBetaVersion(anthropic.BetaMessageBatches20240924),
	)

	resp, err := client.RetrieveBatchResults(context.Background(), "batch_id_jsonl")
	if err != nil {
		t.Fatalf("RetrieveBatchResults error: %s", err)
	}

	if len(resp.Responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(resp.Responses))
	}

	for i, r := range resp.Responses {
		wantID := "custom_id_" + strconv.Itoa(i)
		if r.CustomId != wantID {
			t.Fatalf("response %d: expected custom_id %q, got %q", i, wantID, r.CustomId)
		}
		if r.Result.Type != anthropic.ResultTypeSucceeded {
			t.Fatalf("response %d: expected result type succeeded, got %q", i, r.Result.Type)
		}
	}

	// The middle record is larger than bufio.Scanner's default 64KB buffer,
	// so it must decode intact without truncation.
	bigText := resp.Responses[1].Result.Result.GetFirstContentText()
	if len(bigText) < 64*1024 {
		t.Fatalf("expected large record text to exceed 64KB, got %d bytes", len(bigText))
	}
}

func handleRetrieveBatchResultsJSONLEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/x-jsonl")

	makeLine := func(i int, text string) []byte {
		res := anthropic.BatchResult{
			CustomId: "custom_id_" + strconv.Itoa(i),
			Result: anthropic.BatchResultCore{
				Type: anthropic.ResultTypeSucceeded,
				Result: anthropic.MessagesResponse{
					ID:   "msg_" + strconv.Itoa(i),
					Type: anthropic.MessagesResponseTypeMessage,
					Role: anthropic.RoleAssistant,
					Content: []anthropic.MessageContent{
						{
							Type: anthropic.MessagesContentTypeText,
							Text: toPtr(text),
						},
					},
					Model:      anthropic.ModelClaude3Haiku20240307,
					StopReason: anthropic.MessagesStopReasonEndTurn,
				},
			},
		}
		b, _ := json.Marshal(res)
		return append(b, '\n')
	}

	var body []byte
	body = append(body, makeLine(0, "first")...)
	body = append(body, makeLine(1, strings.Repeat("A", 70*1024))...)
	body = append(body, makeLine(2, "third")...)

	_, _ = w.Write(body)
}

func TestListBatches(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages/batches", handleListBatchesEndpoint)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
		anthropic.WithBetaVersion(anthropic.BetaMessageBatches20240924),
	)

	t.Run("list batches success", func(t *testing.T) {
		resp, err := client.ListBatches(context.Background(), anthropic.ListBatchesRequest{
			Limit:    toPtr(10),
			BeforeId: nil,
			AfterId:  nil,
		})
		if err != nil {
			t.Fatalf("ListBatches error: %s", err)
		}
		t.Logf("List Batches resp: %+v", resp)
	})

	t.Run("list failure: limit too high", func(t *testing.T) {
		_, err := client.ListBatches(context.Background(), anthropic.ListBatchesRequest{
			Limit:    toPtr(101),
			BeforeId: nil,
			AfterId:  nil,
		})
		if err == nil {
			t.Fatalf("ListBatches expected error, got nil")
		}
	})

	t.Run("list batches with before_id and after_id", func(t *testing.T) {
		_, err := client.ListBatches(context.Background(), anthropic.ListBatchesRequest{
			Limit:    toPtr(10),
			BeforeId: toPtr("batch_id_1234"),
			AfterId:  toPtr("batch_id_567"),
		})
		if err != nil {
			t.Fatalf("ListBatches error: %s", err)
		}
	})
}

func handleListBatchesEndpoint(w http.ResponseWriter, r *http.Request) {
	var resBytes []byte

	// listing batches only accepts GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	res := anthropic.ListBatchesResponse{
		Data: []anthropic.BatchRespCore{
			forgeBatchResponse("batch_id_1234"),
			forgeBatchResponse("batch_id_567"),
		},
		HasMore: false,
		FirstId: nil,
		LastId:  nil,
	}

	resBytes, _ = json.Marshal(res)
	_, _ = w.Write(resBytes)
}

func TestCancelBatch(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages/batches/batch_id_1234/cancel", handleCancelBatchEndpoint)
	server.RegisterHandler(
		"/v1/messages/batches/batch_id_not_found/cancel",
		handleCancelBatchEndpoint,
	)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
		anthropic.WithBetaVersion(anthropic.BetaMessageBatches20240924),
	)

	t.Run("cancel batch success", func(t *testing.T) {
		resp, err := client.CancelBatch(context.Background(), "batch_id_1234")
		if err != nil {
			t.Fatalf("CancelBatch error: %s", err)
		}
		t.Logf("Cancel Batch resp: %+v", resp)
	})

	t.Run("cancel batch failure", func(t *testing.T) {
		resp, err := client.CancelBatch(context.Background(), "batch_id_not_found")
		if err == nil {
			t.Fatalf("CancelBatch expected error, got nil")
		}
		t.Logf("Cancel Batch resp: %+v", resp)
	})
}

func handleCancelBatchEndpoint(w http.ResponseWriter, r *http.Request) {
	var resBytes []byte

	// canceling batches only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	batchId := strings.TrimPrefix(r.URL.Path, "/v1/messages/batches/")
	batchId = strings.TrimSuffix(batchId, "/cancel")
	if batchId == "" {
		http.Error(w, "missing batch id", http.StatusBadRequest)
		return
	}

	if batchId == "batch_id_not_found" {
		http.Error(w, "batch not found", http.StatusNotFound)
		return
	}

	res := anthropic.BatchResponse{
		BatchRespCore: forgeBatchResponse(batchId),
	}
	resBytes, _ = json.Marshal(res)
	_, _ = w.Write(resBytes)
}

func forgeBatchResult(customId string) anthropic.BatchResultCore {
	return anthropic.BatchResultCore{
		Type: anthropic.ResultTypeSucceeded,
		Result: anthropic.MessagesResponse{
			ID:   customId,
			Type: anthropic.MessagesResponseTypeMessage,
			Role: anthropic.RoleAssistant,
			Content: []anthropic.MessageContent{
				{
					Type: anthropic.MessagesContentTypeText,
					Text: toPtr("My name is Claude."),
				},
			},
			Model:        anthropic.ModelClaude3Haiku20240307,
			StopReason:   anthropic.MessagesStopReasonEndTurn,
			StopSequence: "",
			Usage: anthropic.MessagesUsage{
				InputTokens:  10,
				OutputTokens: 10,
			},
		},
	}
}

func forgeBatchResponse(batchId string) anthropic.BatchRespCore {
	t1 := time.Now().Add(-time.Hour * 2)
	return anthropic.BatchRespCore{
		Id:               anthropic.BatchId(batchId),
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
	}
}

func TestBatchResultErrored(t *testing.T) {
	line := `{"custom_id":"req-1","result":{"type":"errored","error":{"type":"error","error":{"type":"invalid_request_error","message":"bad request"}}}}`

	var parsed anthropic.BatchResult
	if err := json.Unmarshal([]byte(line), &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if parsed.Result.Type != anthropic.ResultTypeErrored {
		t.Fatalf("unexpected result type: %s", parsed.Result.Type)
	}
	if parsed.Result.Error == nil {
		t.Fatalf("expected error envelope, got nil")
	}
	if parsed.Result.Error.Error == nil ||
		parsed.Result.Error.Error.Type != anthropic.ErrTypeInvalidRequest {
		t.Fatalf("unexpected error payload: %+v", parsed.Result.Error)
	}
	if parsed.Result.Error.Error.Message != "bad request" {
		t.Fatalf("unexpected error message: %q", parsed.Result.Error.Error.Message)
	}
}
