package main

import (
	"fmt"
	"os"
)

func main() {
	telegramChesapeakeChatID, ok := os.LookupEnv("TELEGRAM_CHESAPEAKE_CHAT_ID")
	if !ok {
		fmt.Println("TELEGRAM_CHESAPEAKE_CHAT_ID not set")
		return
	}
	skipChesapeake := os.Getenv("SKIP_CHESAPEAKE") == "true"

	if skipChesapeake {
		fmt.Println("Skipping Chesapeake")
		return
	}
	fmt.Printf("Chesapeake chat ID: %s\n", telegramChesapeakeChatID)
}
