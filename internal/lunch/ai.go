package lunch

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type AIClient struct {
	GitHubToken   string
	Client        openai.Client
	SystemMessage string
}

func NewAIClient(githubToken, systemMessage string) *AIClient {
	return &AIClient{
		GitHubToken: githubToken,
		Client: openai.NewClient(
			option.WithAPIKey(githubToken),
			option.WithBaseURL("https://models.github.ai/inference"),
		),
		SystemMessage: systemMessage,
	}
}

func (c *AIClient) SprinkleAIOnIt(message string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	slog.InfoContext(ctx, "Enhancing message with AI")
	chatCompletion, err := c.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(c.SystemMessage),
			openai.UserMessage(message),
		},
		Model: "openai/gpt-5-mini",
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get chat completion", "error", err)
		return message, err
	}
	if len(chatCompletion.Choices) == 0 {
		err := fmt.Errorf("no choices returned from chat completion")
		slog.ErrorContext(ctx, "Failed to get chat completion", "error", err)
		return message, err
	}
	return chatCompletion.Choices[0].Message.Content, nil
}
