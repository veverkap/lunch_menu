package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type School struct {
	ID   string
	Name string
}

var (
	tomorrow                 time.Time
	telegramToken            string
	telegramChesapeakeChatID string
	githubToken              string
	geminiAPIKey             string
	schools                  = []School{
		{ID: "d9edb69f-dc06-41a4-8d8d-15c3e47d812f", Name: "Butts Road Intermediate"},
		{ID: "6809b286-dbc7-48c1-bd22-d8db93816941", Name: "Butts Road Primary"},
	}
	logger        *slog.Logger
	systemMessage = `You are a witty assistant who sends a message daily about the lunch for the next day to two children - Elena, a girl (born on July 10, 2015) who attends Butts Road Intermediate and John a boy (born on July 13 2018) who attends Butts Road Primary. Feel free to make up nicknames for them based on their names.

Make sure to include the weather, the lunch menu and some comment. 

Use emojis.

All dates are in the America/New_York time zone and format MM/DD/YYYY.

Do not change the contents of the menu.  

It must be accurately communicated.`
)

func main() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// check that we have the needed values
	telegramChesapeakeChatID = os.Getenv("TELEGRAM_CHESAPEAKE_CHAT_ID")
	if telegramChesapeakeChatID == "" {
		logger.Error("TELEGRAM_CHESAPEAKE_CHAT_ID not set")
		return
	}
	logger.Debug("TELEGRAM_CHESAPEAKE_CHAT_ID set")

	telegramToken = os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		logger.Error("TELEGRAM_TOKEN not set")
		return
	}
	logger.Debug("TELEGRAM_TOKEN set")

	githubToken = os.Getenv("GH_TOKEN")
	if githubToken == "" {
		logger.Error("GH_TOKEN not set")
		return
	}
	logger.Debug("GH_TOKEN set")

	geminiAPIKey = os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		logger.Error("GEMINI_API_KEY not set")
		return
	}

	// get the current time in the America/New_York time zone
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		logger.Error("Failed to load location", "location", "America/New_York", "error", err)
		return
	}
	// we want just the date, so set the time to noon to avoid any DST issues
	now := time.Now().In(loc)
	now = time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, loc)
	logger.Debug("Current time", "now", now)
	// get tomorrow's date
	tomorrow = now.AddDate(0, 0, 1)
	// now := time.Date(2025, 9, 3, 12, 0, 0, 0, loc) // for testing
	logger.Debug("Tomorrow's date", "tomorrow", tomorrow)

	logger = logger.With(
		slog.Time("now", now),
		slog.Time("tomorrow", tomorrow),
	)

	telegramMessages := strings.Builder{}
	logger.Info("Getting lunch menus for schools", "schools", schools)

	firstLunchMenu, err := getLunchMenuForSchool(tomorrow, schools[0])
	if err != nil {
		logger.Error("Failed to get lunch menu", "school", schools[0].Name, "error", err)
		return // no point in continuing if we can't get the first school's menu
	}
	secondLunchMenu, err := getLunchMenuForSchool(tomorrow, schools[1])
	if err != nil {
		logger.Error("Failed to get lunch menu", "school", schools[1].Name, "error", err)
		return
	}

	if firstLunchMenu != secondLunchMenu {
		logger.Info("Lunch menus differ between schools")
		fmt.Fprintf(&telegramMessages, "*%s*:\n%s\n", schools[0].Name, firstLunchMenu)
		fmt.Fprintf(&telegramMessages, "*%s*:\n%s\n", schools[1].Name, secondLunchMenu)
	} else {
		logger.Info("Lunch menus are the same between schools")
		fmt.Fprintf(&telegramMessages, "*Lunch Menu*:\n%s\n", firstLunchMenu)
	}

	if telegramMessages.Len() == 0 {
		logger.Warn("No lunch menus found for any school")
		return
	}
	logger.Info("Getting tomorrow's weather", "date", tomorrow)
	weather, err := getTomorrowsWeather(tomorrow)
	if err != nil {
		logger.Error("Failed to get weather", "error", err)
	} else if weather != "" {
		fmt.Fprintf(&telegramMessages, "*Weather*:\n%s\n", weather)
	}

	telegramMessage := strings.Builder{}
	fmt.Fprintf(&telegramMessage, "Lunch menu for %s:\n\n", tomorrow.Format("01/02/2006"))
	telegramMessage.WriteString(telegramMessages.String())
	fmt.Fprintf(&telegramMessage, "\n*Weather*:\n%s\n", weather)

	logger.Info("Final Telegram message", "message", telegramMessage.String())

	// write this to menus/YYYY-MM-DD.txt
	filename := fmt.Sprintf("menus/%s.txt", tomorrow.Format("2006-01-02"))
	if err := os.WriteFile(filename, []byte(telegramMessage.String()), 0644); err != nil {
		logger.Error("Failed to write menu to file", "file", filename, "error", err)
	}

	enhancedMessage, err := sprinkleAIOnIt(telegramMessage.String())
	if err != nil {
		logger.Error("Failed to enhance message with AI", "error", err)
	}

	logger.Info("Enhanced message", "message", enhancedMessage)
	enhancedMessage += fmt.Sprintf("\n\n[Without AI](https://veverkap.github.io/lunch_menu/menus/%s.txt)", tomorrow.Format("2006-01-02"))

	if err := sendTelegramMessage(enhancedMessage); err != nil {
		logger.Error("Failed to send Telegram message", "error", err)
	}

	img, err := generateMenuImage(telegramMessage.String())
	if err != nil {
		logger.Error("Failed to generate menu image", "error", err)
	}

	if err := sendTelegramImage(img); err != nil {
		logger.Error("Failed to send Telegram image", "error", err)
	}
}


