package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	telegramChesapeakeChatID, ok := os.LookupEnv("TELEGRAM_CHESAPEAKE_CHAT_ID")
	if !ok {
		logger.Error("TELEGRAM_CHESAPEAKE_CHAT_ID not set")
		return
	}
	skipChesapeake := os.Getenv("SKIP_CHESAPEAKE") == "true"

	if skipChesapeake {
		logger.Info("Skipping Chesapeake")
		return
	}
	logger.Info("Chesapeake chat ID: %s", telegramChesapeakeChatID)
}
