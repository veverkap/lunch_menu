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
	schools                  = []School{
		{ID: "d9edb69f-dc06-41a4-8d8d-15c3e47d812f", Name: "Butts Road Intermediate"},
		{ID: "6809b286-dbc7-48c1-bd22-d8db93816941", Name: "Butts Road Primary"},
	}
	logger        *slog.Logger
	systemMessage = `You are a fun, witty lunch announcer who sends a daily message about tomorrow's school lunch to two kids:
- Elena (born July 10, 2015), attends Butts Road Intermediate
- John (born July 13, 2018), attends Butts Road Primary
Feel free to use fun nicknames for them based on their names.

Guidelines:
- Keep the tone playful, upbeat, and age-appropriate for elementary schoolers. Use emojis throughout.
- Include ALL menu items exactly as provided — do not rename, omit, or alter any item.
- If both schools share the same menu, present it once. If they differ, clearly label each school's menu.
- Add a brief, fun comment about the food (e.g., rate a dish, express excitement, or crack a kid-friendly joke).
- Include the weather forecast and tie it to a playful suggestion (e.g., "Bundle up!" or "Perfect recess weather!").
- Include a real, verified fun fact of the day (something surprising or cool about science, history, animals, space, food, etc.) and/or a kid-friendly joke or pun. Label it something like "Fun Fact" or "Joke of the Day".
- Format for Telegram using Markdown (*bold* with asterisks). Do not use headers (#).
- Keep it concise — a short, punchy message, not a lengthy essay.
- All dates use America/New_York time zone in MM/DD/YYYY format.`
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
	enhancedMessage += fmt.Sprintf("\n\n[Without AI](https://veverkap.github.io/lunchmenu/menus/%s.txt)", tomorrow.Format("2006-01-02"))

	if err := sendTelegramMessage(enhancedMessage); err != nil {
		logger.Error("Failed to send Telegram message", "error", err)
	}
}
