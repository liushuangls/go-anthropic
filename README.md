# go-anthropic

[![Go Reference](https://pkg.go.dev/badge/github.com/liushuangls/go-anthropic.svg)](https://pkg.go.dev/github.com/liushuangls/go-anthropic)

Anthropic Claude API wrapper for Go (Unofficial). Support:

- Completions
- Streaming Completions
- Messages
- Streaming Messages

## Installation

```
go get github.com/liushuangls/go-anthropic
```

Currently, go-anthropic requires Go version 1.21 or greater.

## Usage

### Messages example usage:

```go
package main

import (
	"errors"
	"fmt"

	"github.com/liushuangls/go-anthropic"
)

func main() {
	client := anthropic.NewClient("your anthropic apikey")
	resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaudeInstant1Dot2,
		Messages: []anthropic.Message{
			{Role: anthropic.RoleUser, Content: "What is your name?"},
		},
		MaxTokens: 1000,
	})
	if err != nil {
		var e anthropic.APIError
		if errors.As(err, &e) {
			fmt.Printf("Messages error, type: %s, message: %s", e.Type, e.Message)
		} else {
			fmt.Printf("Messages error: %v\n", err)
        }
		return
	}
	fmt.Println(resp.Content[0].Text)
}
```

### Messages stream example usage:

```go
package main

import (
	"errors"
	"fmt"

	"github.com/liushuangls/go-anthropic"
)

func main() {
	client := anthropic.NewClient("your anthropic apikey")
	resp, err := client.CreateMessagesSteam(context.Background(),  anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				{Role: anthropic.RoleUser, Content: "What is your name?"},
			},
			MaxTokens:   1000,
		},
		OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			fmt.Printf("Stream Content: %s\n", data.Delta.Text)
		},
	})
	if err != nil {
		var e anthropic.APIError
		if errors.As(err, &e) {
			fmt.Printf("Messages stream error, type: %s, message: %s", e.Type, e.Message)
		} else {
			fmt.Printf("Messages stream error: %v\n", err)
        }
		return
	}
	fmt.Println(resp.Content[0].Text)
}
```

## Acknowledgments
The following project had particular influence on go-anthropic is design.

- [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai)

