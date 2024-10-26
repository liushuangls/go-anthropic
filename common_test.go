package anthropic_test

import (
	"encoding/json"
	"io"
	"net/http"
)

func toPtr(s string) *string {
	return &s
}

func getRequest[T any](r *http.Request) (req T, err error) {
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
