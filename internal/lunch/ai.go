package lunch

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type AIClient struct {
	GitHubToken string
	Client      openai.Client
}

func NewAIClient(githubToken string) *AIClient {
	return &AIClient{
		GitHubToken: githubToken,
		Client: openai.NewClient(
			option.WithAPIKey(githubToken),
			option.WithBaseURL("https://models.github.ai/inference"),
		),
	}
}

func (c *AIClient) SprinkleAIOnIt(message string) (string, error) {
	slog.Info("Enhancing message with AI", "message", message)
	chatCompletion, err := c.Client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemMessage),
			openai.UserMessage(message),
		},
		Model: "openai/gpt-5-mini",
	})
	if err != nil {
		slog.Error("Failed to get chat completion", "error", err)
		return message, err
	}
	if len(chatCompletion.Choices) == 0 {
		err := fmt.Errorf("no choices returned from chat completion")
		slog.Error("Failed to get chat completion", "error", err)
		return message, err
	}
	return chatCompletion.Choices[0].Message.Content, nil
}
