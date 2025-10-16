package anthropic

type Model string

const (
	ModelClaude2Dot0                Model = "claude-2.0"
	ModelClaude2Dot1                Model = "claude-2.1"
	ModelClaude3Opus20240229        Model = "claude-3-opus-20240229"
	ModelClaude3Sonnet20240229      Model = "claude-3-sonnet-20240229"
	ModelClaude3Dot5Sonnet20240620  Model = "claude-3-5-sonnet-20240620"
	ModelClaude3Dot5Sonnet20241022  Model = "claude-3-5-sonnet-20241022"
	ModelClaude3Dot5SonnetLatest    Model = "claude-3-5-sonnet-latest"
	ModelClaude3Haiku20240307       Model = "claude-3-haiku-20240307"
	ModelClaude3Dot5HaikuLatest     Model = "claude-3-5-haiku-latest"
	ModelClaude3Dot5Haiku20241022   Model = "claude-3-5-haiku-20241022"
	ModelClaudeHaiku4Dot5           Model = "claude-haiku-4-5"
	ModelClaudeHaiku4Dot5V20251015  Model = "claude-haiku-4-5-20251015"
	ModelClaude3Dot7SonnetLatest    Model = "claude-3-7-sonnet-latest"
	ModelClaude3Dot7Sonnet20250219  Model = "claude-3-7-sonnet-20250219"
	ModelClaudeOpus4Dot0            Model = "claude-opus-4-0"
	ModelClaudeOpus4V20250514       Model = "claude-opus-4-20250514"
	ModelClaudeSonnet4Dot0          Model = "claude-sonnet-4-0"
	ModelClaudeSonnet4V20250514     Model = "claude-sonnet-4-20250514"
	ModelClaudeSonnet4Dot5          Model = "claude-sonnet-4-5"
	ModelClaudeSonnet4Dot5V20250929 Model = "claude-sonnet-4-5-20250929"
)

type ChatRole string

const (
	RoleUser      ChatRole = "user"
	RoleAssistant ChatRole = "assistant"
)

func (m Model) asVertexModel() string {
	switch m {
	case ModelClaude3Opus20240229:
		return "claude-3-opus@20240229"
	case ModelClaude3Sonnet20240229:
		return "claude-3-sonnet@20240229"
	case ModelClaude3Dot5Sonnet20240620:
		return "claude-3-5-sonnet@20240620"
	case ModelClaude3Dot5Sonnet20241022:
		return "claude-3-5-sonnet-v2@20241022"
	case ModelClaude3Dot7Sonnet20250219:
		return "claude-3-7-sonnet@20250219"
	case ModelClaude3Haiku20240307:
		return "claude-3-haiku@20240307"
	case ModelClaude3Dot5Haiku20241022:
		return "claude-3-5-haiku@20241022"
	case ModelClaudeHaiku4Dot5, ModelClaudeHaiku4Dot5V20251015:
		return "claude-haiku-4-5@20251015"
	case ModelClaudeOpus4Dot0, ModelClaudeOpus4V20250514:
		return "claude-opus-4@20250514"
	case ModelClaudeSonnet4Dot0, ModelClaudeSonnet4V20250514:
		return "claude-sonnet-4@20250514"
	case ModelClaudeSonnet4Dot5, ModelClaudeSonnet4Dot5V20250929:
		return "claude-sonnet-4-5@20250929"
	default:
		return string(m)
	}
}
