package anthropic_test

import (
	"context"
	"encoding/json"
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

	t.Run("create complete success", func(t *testing.T) {
		client := anthropic.NewClient(test.GetTestToken(), anthropic.WithBaseURL(baseUrl))
		resp, err := client.CreateComplete(context.Background(), anthropic.CompleteRequest{
			Model:             anthropic.ModelClaude3Haiku20240307,
			Prompt:            "\n\nHuman: What is your name?\n\nAssistant:",
			MaxTokensToSample: 1000,
		})
		if err != nil {
			t.Fatalf("CreateComplete error: %v", err)
		}

		t.Logf("Create Complete resp: %+v", resp)
	})

	t.Run("create complete failure", func(t *testing.T) {
		client := anthropic.NewClient("invalid token", anthropic.WithBaseURL(baseUrl))
		_, err := client.CreateComplete(context.Background(), anthropic.CompleteRequest{
			Model:             anthropic.ModelClaude3Haiku20240307,
			Prompt:            "\n\nHuman: What is your name?\n\nAssistant:",
			MaxTokensToSample: 1000,
		})
		if err == nil {
			t.Fatalf("CreateComplete expected error, got nil")
		}
	})
}

func TestSetTemperature(t *testing.T) {
	cr := anthropic.CompleteRequest{
		Model:             anthropic.ModelClaude3Haiku20240307,
		Prompt:            "\n\nHuman: What is your name?\n\nAssistant:",
		MaxTokensToSample: 1000,
	}

	temp := float32(0.5)

	cr.SetTemperature(temp)
	if *cr.Temperature != temp {
		t.Fatalf("SetTemperature failed: %v", cr.Temperature)
	}
}

func TestSetTopP(t *testing.T) {
	cr := anthropic.CompleteRequest{
		Model:             anthropic.ModelClaude3Haiku20240307,
		Prompt:            "\n\nHuman: What is your name?\n\nAssistant:",
		MaxTokensToSample: 1000,
	}

	topP := float32(0.5)

	cr.SetTopP(topP)
	if *cr.TopP != topP {
		t.Fatalf("SetTopP failed: %v", cr.TopP)
	}
}

func TestSetTopK(t *testing.T) {
	cr := anthropic.CompleteRequest{
		Model:             anthropic.ModelClaude3Haiku20240307,
		Prompt:            "\n\nHuman: What is your name?\n\nAssistant:",
		MaxTokensToSample: 1000,
	}

	topK := 5

	cr.SetTopK(topK)
	if *cr.TopK != topK {
		t.Fatalf("SetTopK failed: %v", cr.TopK)
	}
}

func handleCompleteEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	var completeReq anthropic.CompleteRequest
	if completeReq, err = getRequest[anthropic.CompleteRequest](r); err != nil {
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
