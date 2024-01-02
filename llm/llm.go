package llm

import (
	"context"
)

type ChatModelID string

const (
	ChatModelGPT35Turbo     ChatModelID = "gpt-3.5-turbo-0613"
	ChatModelGPT35Turbo_16K ChatModelID = "gpt-3.5-turbo-16k-0613"
	ChatModelGPT4           ChatModelID = "gpt-4-0613"
	ChatModelGPT4Turbo      ChatModelID = "gpt-4-1106-preview"
	ChatModelGPT4_32K       ChatModelID = "gpt-4-32k-0613"
)

type Models struct {
	DefaultChatModel            ChatModel
	DefaultLongContextChatModel ChatModel
	DefaultLightChatModel       ChatModel
	DefaultCheapChatModel       ChatModel
	ChatModels                  map[ChatModelID]ChatModel
}

func AllModels(api_key string) *Models {
	return &Models{
		DefaultChatModel:            NewOpenAIChatModel(ChatModelGPT4Turbo, api_key),
		DefaultLongContextChatModel: NewOpenAIChatModel(ChatModelGPT4Turbo, api_key),
		DefaultLightChatModel:       NewOpenAIChatModel(ChatModelGPT4Turbo, api_key),
		DefaultCheapChatModel:       NewOpenAIChatModel(ChatModelGPT35Turbo, api_key),
		ChatModels: map[ChatModelID]ChatModel{
			ChatModelGPT35Turbo:     NewOpenAIChatModel(ChatModelGPT35Turbo, api_key),
			ChatModelGPT4:           NewOpenAIChatModel(ChatModelGPT4, api_key),
			ChatModelGPT4Turbo:      NewOpenAIChatModel(ChatModelGPT4Turbo, api_key),
			ChatModelGPT35Turbo_16K: NewOpenAIChatModel(ChatModelGPT35Turbo_16K, api_key),
			ChatModelGPT4_32K:       NewOpenAIChatModel(ChatModelGPT4_32K, api_key),
		},
	}
}

type FunctionDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Items       *ArrayItems `json:"items,omitempty"`
}

type ArrayItems struct {
	Type string `json:"type"`
}

type FunctionCall struct {
	Name      string
	Arguments string
}

type Message struct {
	Role         MessageRole   `json:"role"`
	Content      string        `json:"content"`
	Name         string        `json:"name"`
	FunctionCall *FunctionCall `json:"function_call"`
}

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleFunction  MessageRole = "function"
)

type StreamEvent struct {
	Text  string
	Error error
}

type MessageOptions struct {
	Temperature   float32        `json:"temperature"`
	MaxTokens     int            `json:"max_tokens"`
	StopSequences []string       `json:"stop_sequences"`
	Functions     []*FunctionDef `json:"functions"`
	FunctionCall  string         `json:"function_call"`
}

const FunctionCallNone = "none"
const FunctionCallAuto = "auto"

type ChatModel interface {
	MessageStream(ctx context.Context, messages []*Message, options *MessageOptions) (chan StreamEvent, error)
	Message(ctx context.Context, messages []*Message, options *MessageOptions) (*Message, error)
	ContextLength() int
}
