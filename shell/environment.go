package shell

import "llm-chat/llm"

type Environment struct {
	Models    *llm.Models
	ChatModel llm.ChatModel
}
