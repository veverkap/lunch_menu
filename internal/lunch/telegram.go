package lunch

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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
	payload := fmt.Sprintf(`{"chat_id":"%s","text":"%s","parse_mode":"Markdown"}`, c.ChatID, message)

	resp, err := retryablehttp.Post(url, "application/json", strings.NewReader(payload))
	if err != nil {
		slog.Error("Failed to send Telegram message", "error", err, "url", url, "payload", payload)
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
