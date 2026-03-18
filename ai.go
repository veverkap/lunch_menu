package main

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func sprinkleAIOnIt(message string) (string, error) {
	logger.Info("Enhancing message with AI", "message", message)
	client := openai.NewClient(
		option.WithAPIKey(githubToken),
		option.WithBaseURL("https://models.github.ai/inference"),
	)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemMessage),
			openai.UserMessage(message),
		},
		Model: "openai/gpt-5-mini",
	})
	if err != nil {
		logger.Error("Failed to get chat completion", "error", err)
		return message, err
	}
	if len(chatCompletion.Choices) == 0 {
		err := fmt.Errorf("no choices returned from chat completion")
		logger.Error("Failed to get chat completion", "error", err)
		return message, err
	}
	return chatCompletion.Choices[0].Message.Content, nil
}
