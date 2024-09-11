package llm

import (
	"context"
	"fmt"
)

type Provider interface {
	Chat(ctx context.Context, model string, temperature int32, messages []Message, tools []Tool) (<-chan Message, <-chan string, error)
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Text        string                 `json:"text,omitempty"`
	ContentType string                 `json:"type,omitempty"`
	Id          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Content     string                 `json:"content,omitempty"`
	ToolUseId   string                 `json:"tool_use_id,omitempty"`
	PartialJson *string                `json:"partial_json,omitempty"`
}

func (m *Content) MergeContentDelta(mc Content) {
	switch mc.ContentType {
	case "text":
		m.Text += mc.Text
	case "text_delta":
		m.Text += mc.Text
	// case "tool_result":
	// 	m.MessageContentToolResult = mc.MessageContentToolResult
	// case "tool_use":
	// 	m.MessageContentToolUse = &MessageContentToolUse{
	// 		ID:   mc.MessageContentToolUse.ID,
	// 		Name: mc.MessageContentToolUse.Name,
	// 	}
	case "input_json_delta":
		if m.PartialJson == nil {
			m.PartialJson = mc.PartialJson
		} else {
			*m.PartialJson += *mc.PartialJson
		}
	}
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"input_schema"`
}

func GetProvider(provider, key string) (Provider, error) {
	switch provider {
	case "anthropic":
		return NewAnthropic(key), nil
	default:
		return nil, fmt.Errorf("Unimplemented")
	}
}
