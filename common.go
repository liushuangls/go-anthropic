package anthropic

type Model string

const (
	ModelClaudeInstant1Dot2        Model = "claude-instant-1.2"
	ModelClaude2Dot0               Model = "claude-2.0"
	ModelClaude2Dot1               Model = "claude-2.1"
	ModelClaude3Opus20240229       Model = "claude-3-opus-20240229"
	ModelClaude3Sonnet20240229     Model = "claude-3-sonnet-20240229"
	ModelClaude3Dot5Sonnet20240620 Model = "claude-3-5-sonnet-20240620"
	ModelClaude3Haiku20240307      Model = "claude-3-haiku-20240307"
)

type ChatRole string

const (
	RoleUser      ChatRole = "user"
	RoleAssistant ChatRole = "assistant"
)
