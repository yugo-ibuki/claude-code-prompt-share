package models

import "time"

// JSONLMessage represents a single line in the JSONL file
type JSONLMessage struct {
	Type             string                 `json:"type"`
	ParentUUID       *string                `json:"parentUuid,omitempty"`
	IsSidechain      bool                   `json:"isSidechain,omitempty"`
	UserType         string                 `json:"userType,omitempty"`
	CWD              string                 `json:"cwd,omitempty"`
	SessionID        string                 `json:"sessionId,omitempty"`
	Version          string                 `json:"version,omitempty"`
	GitBranch        string                 `json:"gitBranch,omitempty"`
	Message          *MessageContent        `json:"message,omitempty"`
	UUID             string                 `json:"uuid,omitempty"`
	Timestamp        time.Time              `json:"timestamp,omitempty"`
	ThinkingMetadata map[string]interface{} `json:"thinkingMetadata,omitempty"`
	Todos            []interface{}          `json:"todos,omitempty"`
	RequestID        string                 `json:"requestId,omitempty"`
}

// MessageContent represents the actual message content
type MessageContent struct {
	Role       string                   `json:"role"`
	Content    interface{}              `json:"content"` // Can be string or array
	Model      string                   `json:"model,omitempty"`
	ID         string                   `json:"id,omitempty"`
	Type       string                   `json:"type,omitempty"`
	StopReason *string                  `json:"stop_reason,omitempty"`
	Usage      map[string]interface{}   `json:"usage,omitempty"`
}

// Session represents a conversation session
type Session struct {
	ID          string
	ProjectPath string
	ProjectName string
	Messages    []ConversationMessage
	StartTime   time.Time
	EndTime     time.Time
}

// ConversationMessage represents a user or assistant message
type ConversationMessage struct {
	UUID      string
	Role      string // "user" or "assistant"
	Content   string
	Timestamp time.Time
	IsAgent   bool
}

// Project represents a Claude Code project
type Project struct {
	EncodedPath string
	DecodedPath string
	Sessions    []SessionInfo
}

// SessionInfo represents basic session information for listing
type SessionInfo struct {
	ID          string
	ProjectPath string
	ProjectName string
	StartTime   time.Time
	EndTime     time.Time
	MessageCount int
	FirstMessage string
}
