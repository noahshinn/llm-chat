package main

import (
	"context"
	"flag"
	"fmt"
	"llm-chat/llm"
	"llm-chat/shell"
	"os"
)

func main() {
	useTurbo := flag.Bool("turbo", false, "Use turbo model")
	openaiApiKey := flag.String("api-key", "", "OpenAI API key")
	disableStreaming := flag.Bool("disable-streaming", false, "Disable streaming")
	flag.Parse()

	env := &shell.Environment{}
	apiKey := *openaiApiKey
	if apiKey == "-" {
		openaiAPIKey := os.Getenv("OPENAI_API_KEY")
		if openaiAPIKey == "" {
			panic(fmt.Errorf("must provide OpenAI API key via --api-key or set OPENAI_API_KEY as an environment variable"))
		}
		apiKey = openaiAPIKey
	}
	env.Models = llm.AllModels(apiKey)
	shell := shell.NewShell(env, &shell.Options{
		ShouldStream: !*disableStreaming,
		UseTurbo:     *useTurbo,
	})
	ctx := context.Background()
	if err := shell.Run(ctx); err != nil {
		panic(err)
	}
}
