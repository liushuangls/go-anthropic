package anthropic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	rawPredictSuffix       = ":rawPredict"
	streamRawPredictSuffix = ":streamRawPredict"
	countTokensSuffix      = "/count-tokens:rawPredict"
)

var _ ClientAdapter = (*VertexAdapter)(nil)

type VertexAdapter struct {
}

func (v *VertexAdapter) TranslateError(resp *http.Response, body []byte) (error, bool) {
	switch resp.StatusCode {
	case http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusTooManyRequests:
		var errRes VertexAIErrorResponse
		err := json.Unmarshal(body, &errRes)
		if err != nil {
			// it could be an array
			var errResArr []VertexAIErrorResponse
			err = json.Unmarshal(body, &errResArr)
			if err == nil && len(errResArr) > 0 {
				errRes = errResArr[0]
			}
		}

		if err != nil || errRes.Error == nil {
			reqErr := RequestError{
				StatusCode: resp.StatusCode,
				Err:        err,
				Body:       body,
			}
			return &reqErr, true
		}
		return fmt.Errorf(
			"error, status code: %d, message: %w",
			resp.StatusCode,
			errRes.Error,
		), true
	}
	return nil, false
}

func (v *VertexAdapter) fullMessagesURL(baseUrl string, suffix string, model Model) string {
	// replace the first slash with a colon
	return fmt.Sprintf("%s/%s:%s", baseUrl, model.asVertexModel(), suffix[1:])
}

func (v *VertexAdapter) fullCountURL(baseUrl string, suffix string) string {
	trimmedBaseUrl, _ := strings.CutSuffix(baseUrl, "/")
	return fmt.Sprintf("%s%s", trimmedBaseUrl, suffix)
}

func (v *VertexAdapter) translateUrlSuffix(suffix string, stream bool) (string, error) {
	switch suffix {
	case "/messages":
		if stream {
			return streamRawPredictSuffix, nil
		} else {
			return rawPredictSuffix, nil
		}
	case "/messages/count_tokens":
		return countTokensSuffix, nil
	}

	return "", fmt.Errorf("unknown suffix: %s", suffix)
}

func (v *VertexAdapter) PrepareRequest(
	c *Client,
	method string,
	urlSuffix string,
	body any,
) (string, error) {
	// if the body implements the ModelGetter interface, use the model from the body
	model := Model("")
	if body != nil {
		if vertexAISupport, ok := body.(VertexAISupport); ok {
			model = vertexAISupport.GetModel()
			vertexAISupport.SetAnthropicVersion(c.config.APIVersion)

			var err error
			urlSuffix, err = v.translateUrlSuffix(urlSuffix, vertexAISupport.IsStreaming())
			if err != nil {
				return "", err
			}

			if urlSuffix == countTokensSuffix {
				return v.fullCountURL(c.config.BaseURL, urlSuffix), nil
			}
		} else {
			return "", fmt.Errorf("this call is not supported by the Vertex AI API")
		}
	}

	return v.fullMessagesURL(c.config.BaseURL, urlSuffix, model), nil
}

func (v *VertexAdapter) SetRequestHeaders(c *Client, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+c.config.GetApiKey())
	return nil
}
