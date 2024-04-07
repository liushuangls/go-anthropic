package anthropic_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/liushuangls/go-anthropic/v2/internal/test"
)

func TestComplete(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/complete", handleCompleteEndpoint)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(test.GetTestToken(), anthropic.WithBaseURL(baseUrl))
	resp, err := client.CreateComplete(context.Background(), anthropic.CompleteRequest{
		Model:             anthropic.ModelClaudeInstant1Dot2,
		Prompt:            "\n\nHuman: What is your name?\n\nAssistant:",
		MaxTokensToSample: 1000,
	})
	if err != nil {
		t.Fatalf("CreateComplete error: %v", err)
	}

	t.Logf("Create Complete resp: %+v", resp)
}

func handleCompleteEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	var completeReq anthropic.CompleteRequest
	if completeReq, err = getCompleteRequest(r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}

	res := anthropic.CompleteResponse{
		Type:       "completion",
		ID:         strconv.Itoa(int(time.Now().Unix())),
		Completion: "hello",
		Model:      completeReq.Model,
	}
	resBytes, _ = json.Marshal(res)
	_, _ = w.Write(resBytes)
}

func getCompleteRequest(r *http.Request) (req anthropic.CompleteRequest, err error) {
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
