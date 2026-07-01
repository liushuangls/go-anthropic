package anthropic

type Model string

const (
	// Deprecated: claude-2.0 was retired on July 21, 2025; requests now fail. Use ModelClaudeOpus4Dot8.
	ModelClaude2Dot0 Model = "claude-2.0"
	// Deprecated: claude-2.1 was retired on July 21, 2025; requests now fail. Use ModelClaudeOpus4Dot8.
	ModelClaude2Dot1 Model = "claude-2.1"
	// Deprecated: claude-3-opus-20240229 was retired on January 5, 2026; requests now fail. Use ModelClaudeOpus4Dot8.
	ModelClaude3Opus20240229 Model = "claude-3-opus-20240229"
	// Deprecated: claude-3-sonnet-20240229 was retired on July 21, 2025; requests now fail. Use ModelClaudeSonnet4Dot6 or ModelClaudeSonnet5.
	ModelClaude3Sonnet20240229 Model = "claude-3-sonnet-20240229"
	// Deprecated: claude-3-5-sonnet-20240620 was retired on October 28, 2025; requests now fail. Use ModelClaudeSonnet4Dot6 or ModelClaudeSonnet5.
	ModelClaude3Dot5Sonnet20240620 Model = "claude-3-5-sonnet-20240620"
	// Deprecated: claude-3-5-sonnet-20241022 was retired on October 28, 2025; requests now fail. Use ModelClaudeSonnet4Dot6 or ModelClaudeSonnet5.
	ModelClaude3Dot5Sonnet20241022 Model = "claude-3-5-sonnet-20241022"
	// Deprecated: the claude-3-5-sonnet-latest alias pointed at a now-retired Claude 3.5 Sonnet snapshot; requests now fail. Use ModelClaudeSonnet4Dot6 or ModelClaudeSonnet5.
	ModelClaude3Dot5SonnetLatest Model = "claude-3-5-sonnet-latest"
	// Deprecated: claude-3-haiku-20240307 was retired on April 20, 2026; requests now fail. Use ModelClaudeHaiku4Dot5.
	ModelClaude3Haiku20240307 Model = "claude-3-haiku-20240307"
	// Deprecated: the claude-3-5-haiku-latest alias pointed at a now-retired Claude 3.5 Haiku snapshot; requests now fail. Use ModelClaudeHaiku4Dot5.
	ModelClaude3Dot5HaikuLatest Model = "claude-3-5-haiku-latest"
	// Deprecated: claude-3-5-haiku-20241022 was retired on February 19, 2026; requests now fail. Use ModelClaudeHaiku4Dot5.
	ModelClaude3Dot5Haiku20241022  Model = "claude-3-5-haiku-20241022"
	ModelClaudeHaiku4Dot5          Model = "claude-haiku-4-5"
	ModelClaudeHaiku4Dot5V20251001 Model = "claude-haiku-4-5-20251001"
	// Deprecated: the claude-3-7-sonnet-latest alias pointed at a now-retired Claude 3.7 Sonnet snapshot; requests now fail. Use ModelClaudeSonnet4Dot6 or ModelClaudeSonnet5.
	ModelClaude3Dot7SonnetLatest Model = "claude-3-7-sonnet-latest"
	// Deprecated: claude-3-7-sonnet-20250219 was retired on February 19, 2026; requests now fail. Use ModelClaudeSonnet4Dot6 or ModelClaudeSonnet5.
	ModelClaude3Dot7Sonnet20250219 Model = "claude-3-7-sonnet-20250219"
	// Deprecated: the claude-opus-4-0 alias pointed at claude-opus-4-20250514, retired June 15, 2026; requests now fail. Use ModelClaudeOpus4Dot8.
	ModelClaudeOpus4Dot0 Model = "claude-opus-4-0"
	// Deprecated: claude-opus-4-20250514 was retired on June 15, 2026; requests now fail. Use ModelClaudeOpus4Dot8.
	ModelClaudeOpus4V20250514 Model = "claude-opus-4-20250514"
	// Deprecated: claude-opus-4-1 is scheduled to retire on August 5, 2026. Use ModelClaudeOpus4Dot8.
	ModelClaudeOpus4Dot1 Model = "claude-opus-4-1"
	// Deprecated: claude-opus-4-1-20250805 is scheduled to retire on August 5, 2026. Use ModelClaudeOpus4Dot8.
	ModelClaudeOpus4Dot1V20250805 Model = "claude-opus-4-1-20250805"
	// Deprecated: the claude-sonnet-4-0 alias pointed at claude-sonnet-4-20250514, retired June 15, 2026; requests now fail. Use ModelClaudeSonnet4Dot6 or ModelClaudeSonnet5.
	ModelClaudeSonnet4Dot0 Model = "claude-sonnet-4-0"
	// Deprecated: claude-sonnet-4-20250514 was retired on June 15, 2026; requests now fail. Use ModelClaudeSonnet4Dot6 or ModelClaudeSonnet5.
	ModelClaudeSonnet4V20250514     Model = "claude-sonnet-4-20250514"
	ModelClaudeSonnet4Dot5          Model = "claude-sonnet-4-5"
	ModelClaudeSonnet4Dot5V20250929 Model = "claude-sonnet-4-5-20250929"
	ModelClaudeOpus4Dot5            Model = "claude-opus-4-5"
	ModelClaudeOpus4Dot5V20251101   Model = "claude-opus-4-5-20251101"
	ModelClaudeSonnet4Dot6          Model = "claude-sonnet-4-6"
	ModelClaudeOpus4Dot6            Model = "claude-opus-4-6"
	ModelClaudeOpus4Dot7            Model = "claude-opus-4-7"
	ModelClaudeOpus4Dot8            Model = "claude-opus-4-8"
	ModelClaudeSonnet5              Model = "claude-sonnet-5"
	ModelClaudeFable5               Model = "claude-fable-5"
	ModelClaudeMythos5              Model = "claude-mythos-5"
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
	case ModelClaudeHaiku4Dot5, ModelClaudeHaiku4Dot5V20251001:
		return "claude-haiku-4-5@20251001"
	case ModelClaudeOpus4Dot0, ModelClaudeOpus4V20250514:
		return "claude-opus-4@20250514"
	case ModelClaudeOpus4Dot1, ModelClaudeOpus4Dot1V20250805:
		return "claude-opus-4-1@20250805"
	case ModelClaudeSonnet4Dot0, ModelClaudeSonnet4V20250514:
		return "claude-sonnet-4@20250514"
	case ModelClaudeSonnet4Dot5, ModelClaudeSonnet4Dot5V20250929:
		return "claude-sonnet-4-5@20250929"
	case ModelClaudeOpus4Dot5V20251101, ModelClaudeOpus4Dot5:
		return "claude-opus-4-5@20251101"
	case ModelClaudeSonnet4Dot6:
		return "claude-sonnet-4-6"
	case ModelClaudeOpus4Dot6:
		return "claude-opus-4-6"
	case ModelClaudeOpus4Dot7:
		return "claude-opus-4-7"
	case ModelClaudeOpus4Dot8:
		return "claude-opus-4-8"
	case ModelClaudeSonnet5:
		return "claude-sonnet-5"
	case ModelClaudeFable5:
		return "claude-fable-5"
	case ModelClaudeMythos5:
		return "claude-mythos-5"
	default:
		return string(m)
	}
}
