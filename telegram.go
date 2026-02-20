package main

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
)

func sendTelegramMessage(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramToken)
	payload := fmt.Sprintf(`{"chat_id":"%s","text":"%s","parse_mode":"Markdown"}`, telegramChesapeakeChatID, message)

	resp, err := http.Post(url, "application/json", strings.NewReader(payload))
	if err != nil {
		logger.Error("Failed to send Telegram message", "error", err, "url", url, "payload", payload)
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Error("Failed to close response body", "error", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to send Telegram message", "status", resp.Status)
		return fmt.Errorf("failed to send Telegram message: %s", resp.Status)
	}

	return nil
}

func sendTelegramImage(photoData []byte) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", telegramToken)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("photo", "menu_image.png")
	if err != nil {
		log.Fatal(err)
	}
	_, err = part.Write(photoData)
	if err != nil {
		log.Fatal(err)
	}

	_ = writer.WriteField("chat_id", telegramChesapeakeChatID)

	err = writer.Close()
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to send Telegram message", "status", resp.Status)
		return fmt.Errorf("failed to send Telegram message: %s", resp.Status)
	}

	return nil
}
