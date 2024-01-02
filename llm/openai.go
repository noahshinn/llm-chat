package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const OPENAI_API_URL = "https://api.openai.com/v1"
const OPENAI_CHAT_COMPLETIONS_ENDPOINT = "/chat/completions"

type OpenAIModel struct {
	modelID ChatModelID
	apiKey  string
}

func NewOpenAIChatModel(modelID ChatModelID, apiKey string) *OpenAIModel {
	return &OpenAIModel{modelID: modelID, apiKey: apiKey}
}

func (m *OpenAIModel) Message(ctx context.Context, messages []*Message, options *MessageOptions) (*Message, error) {
	args := m.buildArgs(messages, options)
	if response, err := apiResponse(ctx, m.apiKey, OPENAI_CHAT_COMPLETIONS_ENDPOINT, args); err != nil {
		return nil, err
	} else {
		return parseMessageResponse(response)
	}
}

func (m *OpenAIModel) MessageStream(ctx context.Context, messages []*Message, options *MessageOptions) (chan StreamEvent, error) {
	args := m.buildArgs(messages, options)
	args["stream"] = true
	response, err := apiRequest(ctx, m.apiKey, OPENAI_CHAT_COMPLETIONS_ENDPOINT, args)
	if err != nil {
		return nil, err
	}
	responseChan := make(chan StreamEvent)
	go func() {
		defer close(responseChan)
		scanner := bufio.NewScanner(response.Body)
		separator := "\n\n"
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			} else if i := bytes.Index(data, []byte(separator)); i >= 0 {
				return i + len(separator), data[0:i], nil
			} else if atEOF {
				return len(data), data, nil
			}
			return 0, nil, nil
		})
		for scanner.Scan() {
			chunk := scanner.Text()
			if data, ok := strings.CutPrefix(chunk, "data:"); !ok {
				responseChan <- StreamEvent{Error: fmt.Errorf("invalid SSE event from OpenAI: %s", chunk)}
			} else {
				data := strings.Trim(data, " \n\r\t")
				if data == "[DONE]" {
					break
				}
				deltaText, err := parseStreamChunk(data)
				if err != nil {
					responseChan <- StreamEvent{Error: err}
				}
				responseChan <- StreamEvent{Text: deltaText}
			}
		}
		if err := scanner.Err(); err != nil {
			responseChan <- StreamEvent{Error: err}
		}
	}()
	return responseChan, nil
}

func (m *OpenAIModel) ContextLength() int {
	switch m.modelID {
	case ChatModelGPT35Turbo:
		return 4096
	case ChatModelGPT35Turbo_16K:
		return 16000
	case ChatModelGPT4:
		return 8000
	case ChatModelGPT4Turbo:
		return 128000
	case ChatModelGPT4_32K:
		return 32000
	default:
		return 4096
	}
}

func (m *OpenAIModel) buildArgs(messages []*Message, options *MessageOptions) map[string]any {
	jsonMessages := []map[string]string{}
	for _, message := range messages {
		jsonMessage := map[string]string{
			"role":    string(message.Role),
			"content": message.Content,
		}
		if message.Name != "" {
			jsonMessage["name"] = message.Name
		}
		jsonMessages = append(jsonMessages, jsonMessage)
	}
	args := map[string]any{
		"model":       m.modelID,
		"messages":    jsonMessages,
		"temperature": options.Temperature,
	}
	if options.MaxTokens > 0 {
		args["max_tokens"] = options.MaxTokens
	}
	if len(options.StopSequences) > 0 {
		args["stop"] = options.StopSequences
	}
	if len(options.Functions) > 0 {
		args["functions"] = options.Functions
	}
	if options.FunctionCall != "" {
		if options.FunctionCall == FunctionCallNone || options.FunctionCall == FunctionCallAuto {
			args["function_call"] = options.FunctionCall
		} else {
			args["function_call"] = map[string]string{
				"name": options.FunctionCall,
			}
		}
	}
	return args
}

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func parseMessageResponse(response map[string]any) (*Message, error) {
	if choices, ok := response["choices"].([]any); !ok {
		return nil, &Error{Message: "invalid response, no choices"}
	} else if len(choices) != 1 {
		return nil, &Error{Message: "invalid response, expected 1 choice"}
	} else if choice, ok := choices[0].(map[string]any); !ok {
		return nil, &Error{Message: "invalid response, choice is not a map"}
	} else if message, ok := choice["message"].(map[string]any); !ok {
		return nil, &Error{Message: "invalid response, message is not a map"}
	} else if content, ok := message["content"].(string); ok {
		return &Message{
			Role:    MessageRole(message["role"].(string)),
			Content: content,
		}, nil
	} else if functionCall, ok := message["function_call"].(map[string]any); ok {
		if name, ok := functionCall["name"].(string); !ok {
			return nil, &Error{Message: "invalid response, function call has no name"}
		} else if functionCallArgs, ok := functionCall["arguments"].(string); !ok {
			return nil, &Error{Message: "invalid response, function call args is not a map"}
		} else {
			return &Message{
				Role: MessageRoleFunction,
				Name: name,
				FunctionCall: &FunctionCall{
					Name:      name,
					Arguments: functionCallArgs,
				},
			}, nil
		}
	}
	return nil, &Error{Message: "invalid response, no content or function call"}
}

func parseStreamChunk(data string) (content string, err error) {
	var mapData map[string]any
	if err := json.Unmarshal([]byte(data), &mapData); err != nil {
		return "", err
	}
	if choices, ok := mapData["choices"].([]any); !ok {
		return "", &Error{Message: "invalid response, no choices"}
	} else if len(choices) != 1 {
		return "", &Error{Message: "invalid response, expected 1 choice"}
	} else if choice, ok := choices[0].(map[string]any); !ok {
		return "", &Error{Message: "invalid response, choice is not a map"}
	} else if delta, ok := choice["delta"].(map[string]any); !ok {
		return "", &Error{Message: "invalid response, delta is not a map"}
	} else if content, ok := delta["content"].(string); ok {
		return content, nil
	} else if finishReason, ok := choice["finish_reason"].(string); ok && finishReason == "stop" {
		return "", nil
	} else {
		return "", &Error{Message: "invalid response, no content"}
	}
}

func apiRequest(ctx context.Context, apiKey string, endpoint string, args map[string]any) (*http.Response, error) {
	if encoded, err := json.Marshal(args); err != nil {
		return nil, err
	} else if request, err := http.NewRequestWithContext(ctx, "POST", OPENAI_API_URL+endpoint, bytes.NewBuffer(encoded)); err != nil {
		return nil, err
	} else {
		request.Header.Set("Content-Type", "application/json; charset=utf-8")
		request.Header.Set("Authorization", "Bearer "+apiKey)
		client := &http.Client{}
		return client.Do(request)
	}
}

func apiResponse(ctx context.Context, apiKey string, endpoint string, args map[string]any) (map[string]any, error) {
	if response, err := apiRequest(ctx, apiKey, endpoint, args); err != nil {
		return nil, err
	} else if responseBody, err := io.ReadAll(response.Body); err != nil {
		return nil, err
	} else {
		result := map[string]any{}
		if err := json.Unmarshal(responseBody, &result); err != nil {
			return nil, err
		}
		if err, ok := result["error"].(map[string]any); ok {
			response := Error{Message: "OpenAI error"}
			if value, ok := err["code"].(string); ok {
				response.Code = value
			}
			if value, ok := err["message"].(string); ok {
				response.Message = value
			}
			return nil, &response
		}
		return result, nil
	}
}
