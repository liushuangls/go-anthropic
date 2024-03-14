# go-anthropic

[![Go Reference](https://pkg.go.dev/badge/github.com/liushuangls/go-anthropic.svg)](https://pkg.go.dev/github.com/liushuangls/go-anthropic)
[![Go Report Card](https://goreportcard.com/badge/github.com/liushuangls/go-anthropic)](https://goreportcard.com/report/github.com/liushuangls/go-anthropic)
[![codecov](https://codecov.io/gh/liushuangls/go-anthropic/graph/badge.svg?token=O6JSAOZORX)](https://codecov.io/gh/liushuangls/go-anthropic)
[![Sanity check](https://github.com/liushuangls/go-anthropic/actions/workflows/pr.yml/badge.svg)](https://github.com/liushuangls/go-anthropic/actions/workflows/pr.yml)

Anthropic Claude API wrapper for Go (Unofficial). Support:

- Completions
- Streaming Completions
- Messages
- Streaming Messages
- Vision

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
			anthropic.NewUserTextMessage("What is your name?"),
		},
		MaxTokens: 1000,
	})
	if err != nil {
		var e *anthropic.APIError
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
	resp, err := client.CreateMessagesStream(context.Background(),  anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage("What is your name?"),
			},
			MaxTokens:   1000,
		},
		OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			fmt.Printf("Stream Content: %s\n", data.Delta.Text)
		},
	})
	if err != nil {
		var e *anthropic.APIError
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

### Messages Vision example usage:

```go
package main

import (
	"errors"
	"fmt"

	"github.com/liushuangls/go-anthropic"
)

func main() {
	client := anthropic.NewClient("your anthropic apikey")

	imagePath := "xxx"
	imageMediaType := "image/jpeg"
	imageFile, err := os.Open(imagePath)
	if err != nil {
		panic(err)
	}
	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		panic(err)
	}

	resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Opus20240229,
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewImageMessageContent(anthropic.MessageContentImageSource{
						Type:      "base64",
						MediaType: imageMediaType,
						Data:      imageData,
					}),
					anthropic.NewTextMessageContent("Describe this image."),
				},
			},
		},
		MaxTokens: 1000,
	})
	if err != nil {
		var e *anthropic.APIError
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

## Acknowledgments
The following project had particular influence on go-anthropic is design.

- [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai)

