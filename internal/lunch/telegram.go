package lunch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

type TelegramClient struct {
	ChatID string
	Token  string
}

func NewTelegramClient(token, chatID string) *TelegramClient {
	return &TelegramClient{Token: token, ChatID: chatID}
}

func (c *TelegramClient) SendMessage(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.Token)
	payloadBytes, err := json.Marshal(map[string]string{
		"chat_id":    c.ChatID,
		"text":       message,
		"parse_mode": "Markdown",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	resp, err := retryablehttp.Post(url, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		slog.Error("Failed to send Telegram message", "error", err, "chat_id", c.ChatID)
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("Failed to close response body", "error", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to send Telegram message", "status", resp.Status)
		return fmt.Errorf("failed to send Telegram message: %s", resp.Status)
	}

	return nil
}
