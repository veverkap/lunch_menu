package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/buger/jsonparser"
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
			openai.UserMessage(message),
			openai.SystemMessage(systemMessage),
		},
		Model: "openai/gpt-5-mini",
	})
	if err != nil {
		logger.Error("Failed to get chat completion", "error", err)
		return message, err
	}
	return chatCompletion.Choices[0].Message.Content, nil
}

func generateMenuImage(menuMessage string) ([]byte, error) {
	logger.Info("Generating menu image with Gemini")

	prompt := fmt.Sprintf("Generate a fun, colorful, cartoon-style illustration representing this school lunch menu and weather. Do not include any text in the image.\n\n%s", menuMessage)

	requestBody := fmt.Sprintf(`{
		"contents": [{"role": "user", "parts": [{"text": %q}]}],
		"generationConfig": {"responseModalities": ["IMAGE"]}
	}`, prompt)

	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-image:streamGenerateContent?key=%s", geminiAPIKey)

	resp, err := http.Post(apiURL, "application/json", bytes.NewReader([]byte(requestBody)))
	if err != nil {
		logger.Error("Failed to call Gemini API", "error", err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Error("Failed to close response body", "error", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read Gemini response", "error", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Gemini API error", "status", resp.Status, "body", string(body))
		return nil, fmt.Errorf("gemini API error: %s", resp.Status)
	}

	imageData, err := jsonparser.GetString(body, "[0]", "candidates", "[0]", "content", "parts", "[0]", "inlineData", "data")
	if err != nil {
		logger.Error("Failed to parse image data from Gemini response", "error", err)
		return nil, fmt.Errorf("failed to parse image data: %w", err)
	}

	imageBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		logger.Error("Failed to decode base64 image data", "error", err)
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	if err := os.WriteFile(fmt.Sprintf("imgs/%s.png", tomorrow.Format("2006-01-02")), imageBytes, 0644); err != nil {
		logger.Error("Failed to write menu image to file", "file", fmt.Sprintf("imgs/%s.png", tomorrow.Format("2006-01-02")), "error", err)
	}

	logger.Info("Generated menu image", "size", len(imageBytes))
	return imageBytes, nil
}
