package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

type School struct {
	ID   string
	Name string
}

var (
	telegramToken            string
	telegramChesapeakeChatID string
	schools                  = []School{
		{ID: "d9edb69f-dc06-41a4-8d8d-15c3e47d812f", Name: "Butts Road Intermediate"},
		{ID: "6809b286-dbc7-48c1-bd22-d8db93816941", Name: "Butts Road Primary"},
	}
	logger *slog.Logger
)

func main() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	// check that we have the needed values
	telegramChesapeakeChatID = os.Getenv("TELEGRAM_CHESAPEAKE_CHAT_ID")
	if telegramChesapeakeChatID == "" {
		logger.Error("TELEGRAM_CHESAPEAKE_CHAT_ID not set")
		return
	}
	telegramToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if telegramToken == "" {
		logger.Error("TELEGRAM_BOT_TOKEN not set")
		return
	}

	// get the current time in the America/New_York time zone
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		logger.Error("Failed to load location", "location", "America/New_York", "error", err)
		return
	}
	now := time.Now().In(loc)
	tomorrow := now.AddDate(0, 0, 1)
	// now := time.Date(2025, 9, 3, 12, 0, 0, 0, loc) // for testing

	logger = logger.With(
		slog.Time("now", now),
		slog.Time("tomorrow", tomorrow),
	)

	telegramMessages := strings.Builder{}

	for _, school := range schools {
		lunchMenu, err := getLunchMenuForSchool(tomorrow, school)
		if err != nil {
			logger.Error("Failed to get lunch menu", "school", school.Name, "error", err)
			continue
		}
		logger.Info("Lunch menu", "school", school.Name, "menu", lunchMenu)
		telegramMessages.WriteString(fmt.Sprintf("*%s*:\n%s\n", school.Name, lunchMenu))
	}
	if telegramMessages.Len() == 0 {
		logger.Warn("No lunch menus found for any school")
		return
	}
	telegramMessage := strings.Builder{}
	telegramMessage.WriteString(fmt.Sprintf("Lunch menu for %s:\n\n", tomorrow.Format("01/02/2006")))
	telegramMessage.WriteString(telegramMessages.String())

	if err := sendTelegramMessage(telegramMessage.String()); err != nil {
		logger.Error("Failed to send Telegram message", "error", err)
	}
	fmt.Println(telegramMessage.String())
}

func sendTelegramMessage(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramToken)
	payload := fmt.Sprintf(`{"chat_id":"%s","text":"%s","parse_mode":"Markdown"}`, telegramChesapeakeChatID, message)

	resp, err := http.Post(url, "application/json", strings.NewReader(payload))
	if err != nil {
		logger.Error("Failed to send Telegram message", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to send Telegram message", "status", resp.Status)
		return fmt.Errorf("failed to send Telegram message: %s", resp.Status)
	}

	return nil
}

func getLunchMenuForSchool(date time.Time, school School) (string, error) {
	// implementation goes here
	logger.Info("Getting lunch menu", "date", date, "school", school.Name)
	url := fmt.Sprintf(
		"https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitemsByGrade?SchoolId=%s&ServingDate=%s&ServingLine=Main%%20Line&MealType=Lunch&Grade=02&PersonId=null",
		school.ID,
		// this needs to be CGI escaped
		url.QueryEscape(date.Format("01/02/2006")),
	)
	logger.Info("Constructed URL", "url", url)

	// get url
	resp, err := http.Get(url)
	if err != nil {
		logger.Error("Failed to get lunch menu", "school", school.Name, "error", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to get lunch menu", "school", school.Name, "status", resp.Status)
		return "", fmt.Errorf("failed to get lunch menu: %s", resp.Status)
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read lunch menu", "school", school.Name, "error", err)
		return "", err
	}

	menu := strings.Builder{}

	// write to JSON file
	_, _ = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		description, _ := jsonparser.GetString(value, "MenuItemDescription")
		logger.Info("Menu item", "item", description)
		menu.WriteString(description + "\n")
	}, "ENTREES")

	// check if menu is empty
	if menu.Len() == 0 {
		return "", errors.New("no menu items found for " + school.Name)
	}
	return menu.String(), nil
}
