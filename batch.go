package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

// While in beta, batches may contain up to 10,000 requests and be up to 32 MB in total size.
type BatchRequest struct {
	Requests []InnerRequests `json:"requests"`
}

type InnerRequests struct {
	CustomId string          `json:"custom_id"`
	Params   MessagesRequest `json:"params"`
}

type BatchId string
type BatchResponseType string

const (
	// only option for now.
	BatchResponseTypeMessageBatch BatchResponseType = "message_batch"
)

type ProcessingStatus string

const (
	ProcessingStatusInProgress ProcessingStatus = "in_progress"
	ProcessingStatusCanceling  ProcessingStatus = "canceling"
	ProcessingStatusEnded      ProcessingStatus = "ended"
)

// All times returned in RFC 3339
type BatchResponse struct {
	httpHeader

	BatchRespCore
}

type BatchRespCore struct {
	Id                BatchId           `json:"id"`
	Type              BatchResponseType `json:"type"`
	ProcessingStatus  string            `json:"processing_status"`
	RequestCounts     RequestCounts     `json:"request_counts"`
	EndedAt           *time.Time        `json:"ended_at"`
	CreatedAt         time.Time         `json:"created_at"`
	ExpiresAt         time.Time         `json:"expires_at"`
	ArchivedAt        *time.Time        `json:"archived_at"`
	CancelInitiatedAt *time.Time        `json:"cancel_initiated_at"`
	ResultsUrl        *string           `json:"results_url"`
}

type RequestCounts struct {
	Processing int `json:"processing"`
	Succeeded  int `json:"succeeded"`
	Errored    int `json:"errored"`
	Canceled   int `json:"canceled"`
	Expired    int `json:"expired"`
}

func (c *Client) CreateBatch(
	ctx context.Context,
	request BatchRequest,
) (*BatchResponse, error) {
	var setters []requestSetter
	if len(c.config.BetaVersion) > 0 {
		setters = append(setters, withBetaVersion(c.config.BetaVersion...))
	}

	urlSuffix := "/messages/batches"
	req, err := c.newRequest(ctx, http.MethodPost, urlSuffix, request, setters...)
	if err != nil {
		return nil, err
	}

	var response BatchResponse
	err = c.sendRequest(req, &response)

	return &response, err
}

func (c *Client) RetrieveBatch(
	ctx context.Context,
	batchId BatchId,
) (*BatchResponse, error) {
	var setters []requestSetter
	if len(c.config.BetaVersion) > 0 {
		setters = append(setters, withBetaVersion(c.config.BetaVersion...))
	}

	urlSuffix := "/messages/batches/" + string(batchId)
	// no body as we are doing a GET
	req, err := c.newRequest(ctx, http.MethodGet, urlSuffix, "", setters...)
	if err != nil {
		return nil, err
	}

	var response BatchResponse
	err = c.sendRequest(req, &response)

	return &response, err
}

type RetrieveBatchResponse struct {
	httpHeader

	// Each line in the file is a JSON object containing the result of a
	// single request in the Message Batch. Results are not guaranteed to
	// be in the same order as requests. Use the custom_id field to match
	// results to requests.

	Responses   []MessagesResponse
	RawResponse []byte
}

// Untested - I don't know about this!
func (c *Client) RetrieveBatchResults(
	ctx context.Context,
	batchId BatchId,
) (*RetrieveBatchResponse, error) {
	var setters []requestSetter
	if len(c.config.BetaVersion) > 0 {
		setters = append(setters, withBetaVersion(c.config.BetaVersion...))
	}

	urlSuffix := "/messages/batches/" + string(batchId) + "/results"

	// no body - as we're just doing a GET
	req, err := c.newRequest(ctx, http.MethodGet, urlSuffix, "", setters...)
	if err != nil {
		return nil, err
	}

	var response RetrieveBatchResponse

	res, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	response.SetHeader(res.Header)

	if err := c.handlerRequestError(res); err != nil {
		return nil, err
	}

	response.RawResponse, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// for each line in the response, decode the JSON object into a MessagesResponse
	// and append it to the Responses slice.
	// this is tricky because the content within the response may contain newline control characters (\n)

	for _, line := range bytes.Split(response.RawResponse, []byte("\n")) {
		if len(line) == 0 {
			continue
		}

		var parsed MessagesResponse
		err := json.Unmarshal(line, &parsed)
		if err != nil {
			return nil, err
		}

		response.Responses = append(response.Responses, parsed)
	}

	return &response, err
}

type ListBatchResponse struct {
	httpHeader

	Data    []BatchRespCore `json:"data"`
	HasMore bool            `json:"has_more"`
	FirstId *BatchId        `json:"first_id"`
	LastId  *BatchId        `json:"last_id"`
}

type ListBatchRequest struct {
	BeforeId string `json:"before_id,omitempty"`
	AfterId  string `json:"after_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

func (l ListBatchRequest) validate() error {
	if l.Limit < 0 || l.Limit > 100 {
		return errors.New("limit must be between 1 and 100")
	}

	return nil
}

func (c *Client) ListBatches(
	ctx context.Context,
	lbr ListBatchRequest,
) (*ListBatchResponse, error) {
	var setters []requestSetter
	if len(c.config.BetaVersion) > 0 {
		setters = append(setters, withBetaVersion(c.config.BetaVersion...))
	}

	if err := lbr.validate(); err != nil {
		return nil, err
	}

	urlSuffix := "/messages/batches/"
	// no body as we are doing a GET
	req, err := c.newRequest(ctx, http.MethodGet, urlSuffix, "", setters...)
	if err != nil {
		return nil, err
	}

	var response ListBatchResponse
	err = c.sendRequest(req, &response)

	return &response, err
}

func (c *Client) CancelBatch(
	ctx context.Context,
	batchId BatchId,
) (*BatchResponse, error) {
	var setters []requestSetter
	if len(c.config.BetaVersion) > 0 {
		setters = append(setters, withBetaVersion(c.config.BetaVersion...))
	}

	urlSuffix := "/messages/batches/" + string(batchId) + "/cancel"
	// no body as we are doing a GET
	req, err := c.newRequest(ctx, http.MethodGet, urlSuffix, "", setters...)
	if err != nil {
		return nil, err
	}

	var response BatchResponse
	err = c.sendRequest(req, &response)

	return &response, err
}
