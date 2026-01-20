package main

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func GenerateDescription(command string) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	prompt := fmt.Sprintf("Generate a very short (5-10 words max) description of what this command does: %s\n\nRespond with ONLY the description, no explanation or additional text.", command)

	message, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: 50,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate description: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("no response from Claude")
	}

	return message.Content[0].Text, nil
}
