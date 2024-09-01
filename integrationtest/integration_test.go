package integrationtest

import (
	"os"
	"testing"
)

var (
	APIKey = os.Getenv("ANTHROPIC_KEY")
)

func testAPIKey(t *testing.T) {
	if APIKey == "" {
		t.Fatal("ANTHROPIC_KEY must be set for integration tests")
	}
}
