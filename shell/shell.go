package shell

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"llm-chat/llm"
	"llm-chat/utils/printx"
	"os"
	"strings"
)

type Shell struct {
	ShouldStream bool
	UseTurbo     bool
	Temperature  float32
	History      []*llm.Message
	environment  *Environment
}

type Options struct {
	ShouldStream   bool
	UseTurbo       bool
	Temperature    float32
	InitialHistory []*llm.Message
}

//go:embed system_prompt.txt
var defaultSystemPrompt string

func NewShell(env *Environment, options *Options) *Shell {
	shouldStream := false
	useTurbo := false
	temperature := float32(0.0)
	initialHistory := []*llm.Message{{
		Role:    llm.MessageRoleSystem,
		Content: defaultSystemPrompt,
	}}
	if options != nil {
		if options.ShouldStream {
			shouldStream = true
		}
		if options.UseTurbo {
			useTurbo = true
		}
		if options.Temperature > 0.0 {
			temperature = options.Temperature
		}
		if options.InitialHistory != nil {
			initialHistory = options.InitialHistory
		}
	}
	return &Shell{
		ShouldStream: shouldStream,
		UseTurbo:     useTurbo,
		Temperature:  temperature,
		History:      initialHistory,
		environment:  env,
	}
}

func (s *Shell) Run(ctx context.Context) error {
	printx.PrintStandardHeader("Conversation")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("user: ")
ScannerLoop:
	for scanner.Scan() {
		userMessageText := scanner.Text()
		switch userMessageText {
		case "":
			fmt.Print("user: ")
			continue ScannerLoop
		case "exit":
			fmt.Println("\nexiting...")
			break ScannerLoop
		case "help":
			printx.PrintInColor(printx.ColorGray, "This is a simple chat interface.\nTo exit, type \"exit\".")
			fmt.Print("user: ")
			continue ScannerLoop
		default:
			s.History = append(s.History, &llm.Message{
				Role:    llm.MessageRoleUser,
				Content: strings.TrimSpace(userMessageText),
			})
			chatModel := s.environment.Models.DefaultChatModel
			if s.UseTurbo {
				chatModel = s.environment.Models.DefaultLightChatModel
			}
			if s.ShouldStream {
				stream, err := chatModel.MessageStream(ctx, s.History, &llm.MessageOptions{
					Temperature: s.Temperature,
				})
				if err != nil {
					return err
				}
				fmt.Print("bot: ")
				streamResponseText := ""
				for event := range stream {
					if event.Error != nil {
						return event.Error
					}
					streamResponseText += event.Text
					fmt.Print(event.Text)
				}
				s.History = append(s.History, &llm.Message{
					Role: llm.MessageRoleUser,
				})
			} else {
				response, err := chatModel.Message(ctx, s.History, &llm.MessageOptions{
					Temperature: s.Temperature,
				})
				if err != nil {
					return err
				} else if response == nil || response.Content == "" {
					return fmt.Errorf("empty response")
				} else if response.FunctionCall != nil {
					return fmt.Errorf("function call not supported but received: %v", response.FunctionCall)
				} else {
					fmt.Println("bot: ", strings.TrimSpace(response.Content))
					s.History = append(s.History, response)
				}
			}
			fmt.Print("\nuser: ")
		}
	}
	return nil
}
