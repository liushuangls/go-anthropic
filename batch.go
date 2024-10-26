package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type ResultType string

const (
	ResultTypeSucceeded ResultType = "succeeded"
	ResultTypeErrored   ResultType = "errored"
	ResultTypeCanceled  ResultType = "canceled"
	ResultTypeExpired   ResultType = "expired"
)

type BatchId string

type BatchResponseType string

const (
	BatchResponseTypeMessageBatch BatchResponseType = "message_batch"
)

type ProcessingStatus string

const (
	ProcessingStatusInProgress ProcessingStatus = "in_progress"
	ProcessingStatusCanceling  ProcessingStatus = "canceling"
	ProcessingStatusEnded      ProcessingStatus = "ended"
)

// While in beta, batches may contain up to 10,000 requests and be up to 32 MB in total size.
type BatchRequest struct {
	Requests []InnerRequests `json:"requests"`
}

type InnerRequests struct {
	CustomId string          `json:"custom_id"`
	Params   MessagesRequest `json:"params"`
}

// All times returned in RFC 3339
type BatchResponse struct {
	httpHeader

	BatchRespCore
}

type BatchRespCore struct {
	Id                BatchId           `json:"id"`
	Type              BatchResponseType `json:"type"`
	ProcessingStatus  ProcessingStatus  `json:"processing_status"`
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
	req, err := c.newRequest(ctx, http.MethodGet, urlSuffix, nil, setters...)
	if err != nil {
		return nil, err
	}

	var response BatchResponse
	err = c.sendRequest(req, &response)

	return &response, err
}

type BatchResult struct {
	CustomId string `json:"custom_id"`
	Result   struct {
		Type    ResultType       `json:"type"`
		Message MessagesResponse `json:"message"`
	} `json:"result"`
}

type RetrieveBatchResponse struct {
	httpHeader

	// Each line in the file is a JSON object containing the result of a
	// single request in the Message Batch. Results are not guaranteed to
	// be in the same order as requests. Use the custom_id field to match
	// results to requests.

	Responses   []BatchResult
	RawResponse []byte
}

func (c *Client) RetrieveBatchResults(
	ctx context.Context,
	batchId BatchId,
) (*RetrieveBatchResponse, error) {
	var setters []requestSetter
	if len(c.config.BetaVersion) > 0 {
		setters = append(setters, withBetaVersion(c.config.BetaVersion...))
	}

	// The documentation states that the URL should be obtained from the results_url field in the batch response.
	// It clearly states that the URL should 'not be assumed'. However this seems to work fine.
	urlSuffix := "/messages/batches/" + string(batchId) + "/results"
	req, err := c.newRequest(ctx, http.MethodGet, urlSuffix, nil, setters...)
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

	response.Responses, err = decodeRawResponse(response.RawResponse)
	if err != nil {
		return nil, err
	}

	return &response, err
}

func decodeRawResponse(rawResponse []byte) ([]BatchResult, error) {
	// for each line in the response, decode the JSON object into a MessagesResponse
	// and append it to the Responses slice.
	// this is tricky because the content within the response may contain newline control characters (\n)

	// for each line in the response, decode the JSON object into a MessagesResponse
	// and append it to the Responses slice.
	// this is tricky because the content within the response may contain newline control characters (\n)
	// TODO: test this and make sure it works

	var results []BatchResult
	for _, line := range bytes.Split(rawResponse, []byte("\n")) {
		if len(line) == 0 {
			continue
		}

		var parsed BatchResult
		err := json.Unmarshal(line, &parsed)
		if err != nil {
			return nil, err
		}

		results = append(results, parsed)
	}

	return results, nil
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
	if l.Limit < 1 || l.Limit > 100 {
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

	v := url.Values{}
	if lbr.BeforeId != "" {
		v.Set("before_id", lbr.BeforeId)
	}
	if lbr.AfterId != "" {
		v.Set("after_id", lbr.AfterId)
	}
	if lbr.Limit > 0 {
		v.Set("limit", fmt.Sprintf("%d", lbr.Limit))
	}

	// encode the query parameters into the URL
	urlSuffix += "?" + v.Encode()
	req, err := c.newRequest(ctx, http.MethodGet, urlSuffix, nil, setters...)
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
	req, err := c.newRequest(ctx, http.MethodPost, urlSuffix, nil, setters...)
	if err != nil {
		return nil, err
	}

	var response BatchResponse
	err = c.sendRequest(req, &response)

	return &response, err
}
