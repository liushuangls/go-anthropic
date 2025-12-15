package integrationtest

import (
	"os"
	"testing"
)

var (
	APIKey  = os.Getenv("ANTHROPIC_KEY")
	BaseURL = os.Getenv("ANTHROPIC_BASE_URL")
)

func testAPIKey(t *testing.T) {
	if APIKey == "" {
		t.Fatal("ANTHROPIC_KEY must be set for integration tests")
	}
}

var (
	VertexAPIKey      = os.Getenv("VERTEX_KEY")
	VertexAPILocation = os.Getenv("VERTEX_LOCATION")
	VertexAPIProject  = os.Getenv("VERTEX_PROJECT")
)

func testVertexAPIKey(t *testing.T) {
	if VertexAPIKey == "" {
		t.Fatal("VERTEX_KEY must be set for integration tests")
	}
	if VertexAPILocation == "" {
		t.Fatal("VERTEX_LOCATION must be set for integration tests")
	}
	if VertexAPIProject == "" {
		t.Fatal("VERTEX_PROJECT must be set for integration tests")
	}
}
