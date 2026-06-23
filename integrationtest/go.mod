module github.com/liushuangls/go-anthropic/integrationtest

go 1.24

replace github.com/liushuangls/go-anthropic/v2 => ..

require (
	github.com/henvic/httpretty v0.1.4
	github.com/liushuangls/go-anthropic/v2 v2.15.0
	golang.org/x/oauth2 v0.30.0
)

require cloud.google.com/go/compute/metadata v0.3.0 // indirect
