package shell

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"llm-chat/llm"
	"llm-chat/utils/printx"
	"llm-chat/utils/stringsx"
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
			printx.PrintInColor(printx.ColorGray, "\nexiting...")
			break ScannerLoop
		case "help":
			printx.PrintInColor(printx.ColorGray, "This is a simple chat interface.\nTo exit, type \"exit\".")
			fmt.Print("user: ")
			continue ScannerLoop
		default:
			handleError := func(err error) {
				printx.PrintInColor(printx.ColorRed, fmt.Sprintf("error: %v", err))
				print("user: ")
			}
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
					handleError(err)
					continue ScannerLoop
				}
				fmt.Print("bot: ")
				streamResponseText := ""
				for event := range stream {
					if event.Error != nil {
						handleError(event.Error)
						continue ScannerLoop
					}
					streamResponseText += event.Text
					fmt.Print(event.Text)
				}
				s.History = append(s.History, &llm.Message{
					Role:    llm.MessageRoleUser,
					Content: strings.TrimSpace(streamResponseText),
				})
				fmt.Println()
			} else {
				response, err := chatModel.Message(ctx, s.History, &llm.MessageOptions{
					Temperature: s.Temperature,
				})
				if err != nil {
					handleError(err)
					continue ScannerLoop
				} else if response == nil || response.Content == "" {
					handleError(fmt.Errorf("empty response"))
					continue ScannerLoop
				} else if response.FunctionCall != nil {
					handleError(fmt.Errorf("function call not supported but received: %v", response.FunctionCall))
					continue ScannerLoop
				} else {
					content := strings.TrimSpace(response.Content)
					display, err := stringsx.AddPossibleSyntaxHighlighting(content)
					if err != nil {
						display = content
					}
					fmt.Println("bot: ", display)
					s.History = append(s.History, response)
				}
			}
			fmt.Print("user: ")
		}
	}
	return nil
}
