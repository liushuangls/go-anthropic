package test

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/liushuangls/go-anthropic/v2"
)

const testAPI = "this-is-my-secure-token-do-not-steal!!"

func GetTestToken() string {
	return testAPI
}

type ServerTest struct {
	handlers map[string]Handler
}

type Handler func(w http.ResponseWriter, r *http.Request)

func NewTestServer() *ServerTest {
	return &ServerTest{handlers: make(map[string]Handler)}
}

func (ts *ServerTest) RegisterHandler(path string, handler Handler) {
	ts.handlers[path] = handler
}

// AnthropicTestServer Creates a mocked Anthropic server which can pretend to handle requests during testing.
func (ts *ServerTest) AnthropicTestServer() *httptest.Server {
	return httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("received request at path %q\n", r.URL.Path)

			// check auth
			if r.Header.Get("X-Api-Key") != GetTestToken() {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			handlerCall, ok := ts.handlers[r.URL.Path]
			if !ok {
				log.Printf("path %q not found\n", r.URL.Path)
				http.Error(w, "the resource path doesn't exist", http.StatusNotFound)
				return
			}
			handlerCall(w, r)
			log.Printf("request handled successfully\n")
		}),
	)
}

// VertexTestServer Creates a mocked Vertex Anthropic server which can pretend to handle requests during testing.
func (ts *ServerTest) VertexTestServer() *httptest.Server {
	return httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("received request at path %q\n", r.URL.Path)

			expectedAuth := "Bearer " + GetTestToken()
			// check auth
			if r.Header.Get("Authorization") != expectedAuth {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)

				var errResArr []anthropic.VertexAIErrorResponse
				errResArr = append(errResArr, anthropic.VertexAIErrorResponse{
					Error: &anthropic.VertexAPIError{
						Code:    401,
						Message: "Unauthorized",
						Status:  "UNAUTHORIZED",
					},
				})

				body, _ := json.Marshal(errResArr)
				w.Write(body)

				return
			}

			handlerCall, ok := ts.handlers[r.URL.Path]
			if !ok {
				log.Printf("path %q not found\n", r.URL.Path)
				http.Error(w, "the resource path doesn't exist", http.StatusNotFound)
				return
			}
			handlerCall(w, r)
			log.Printf("request handled successfully\n")
		}),
	)
}
