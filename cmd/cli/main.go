package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/veverkap/lunchmenu/internal/lunch"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config, err := lunch.NewConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		return
	}
	logger = logger.With(
		slog.Time("now", config.Now),
		slog.Time("tomorrow", config.Tomorrow),
		slog.Any("location", config.Location.String()),
	)

	slog.SetDefault(logger)

	logger.Debug("Config loaded successfully")

	apiLeagueClient := lunch.NewAPIClient(config.APILeagueKey)
	if apiLeagueClient == nil {
		slog.Error("Failed to create API client")
		return
	}
	logger.Debug("API client created successfully")

	aiClient := lunch.NewAIClient(config.GitHubToken, config.SystemMessage)
	if aiClient == nil {
		slog.Error("Failed to create AI client")
		return
	}
	logger.Debug("AI client created successfully")

	telegramClient := lunch.NewTelegramClient(config.TelegramToken, config.TelegramChesapeakeChatID)
	if telegramClient == nil {
		slog.Error("Failed to create Telegram client")
		return
	}
	logger.Debug("Telegram client created successfully")

	telegramMessages := strings.Builder{}

	slog.Info("Getting lunch menus for schools", "schools", config.Schools)
	firstLunchMenu, firstLunchItems, err := lunch.GetLunchMenuForSchool(config.Tomorrow, config.Schools[0])
	if err != nil {
		slog.Error("Failed to get lunch menu", "school", config.Schools[0].Name, "error", err)
		return // no point in continuing if we can't get the first school's menu
	}
	secondLunchMenu, secondLunchItems, err := lunch.GetLunchMenuForSchool(config.Tomorrow, config.Schools[1])
	if err != nil {
		slog.Error("Failed to get lunch menu", "school", config.Schools[1].Name, "error", err)
		return // no point in continuing if we can't get the second school's menu
	}

	// merge the two slices of menu items and get a random one for the riddle
	allLunchMenuItems := append(firstLunchItems, secondLunchItems...)

	randomMenuItem := allLunchMenuItems[config.Now.UnixNano()%int64(len(allLunchMenuItems))]

	slog.Info("Random menu item for riddle", "item", randomMenuItem)

	gif, err := apiLeagueClient.GetGif(randomMenuItem)
	if err != nil {
		slog.Error("Failed to get GIF", "error", err)
	}

	if firstLunchMenu != secondLunchMenu {
		slog.Info("Lunch menus differ between schools")
		fmt.Fprintf(&telegramMessages, "*%s*:\n%s\n", config.Schools[0].Name, firstLunchMenu)
		fmt.Fprintf(&telegramMessages, "*%s*:\n%s\n", config.Schools[1].Name, secondLunchMenu)
	} else {
		slog.Info("Lunch menus are the same between schools")
		fmt.Fprintf(&telegramMessages, "*Lunch Menu*:\n%s\n", firstLunchMenu)
	}
	if telegramMessages.Len() == 0 {
		slog.Warn("No lunch menus found for any school")
		return
	}
	slog.Info("Getting tomorrow's weather", "date", config.Tomorrow)
	weather, err := lunch.GetTomorrowsWeather(config.Tomorrow)
	if err != nil {
		slog.Error("Failed to get weather", "error", err)
	} else if weather != "" {
		fmt.Fprintf(&telegramMessages, "*Weather*:\n%s\n", weather)
	}

	telegramMessage := strings.Builder{}
	fmt.Fprintf(&telegramMessage, "Lunch menu for %s:\n\n", config.Tomorrow.Format("01/02/2006"))
	telegramMessage.WriteString(telegramMessages.String())

	slog.Info("Final Telegram message", "message", telegramMessage.String())

	// write this to menus/YYYY-MM-DD.txt
	filename := fmt.Sprintf("menus/%s.txt", config.Tomorrow.Format("2006-01-02"))

	menuWritten := false
	if err := os.WriteFile(filename, []byte(telegramMessage.String()), 0644); err != nil {
		slog.Error("Failed to write menu to file", "file", filename, "error", err)
	} else {
		menuWritten = true
	}
	enhancedMessage := telegramMessage.String()

	if aiMessage, err := aiClient.SprinkleAIOnIt(telegramMessage.String()); err != nil {
		slog.Error("Failed to enhance message with AI", "error", err)
	} else {
		enhancedMessage = aiMessage
	}
	slog.Info("Enhanced message", "message", enhancedMessage)

	riddle, err := apiLeagueClient.GetRiddle("easy")
	if err == nil {
		riddleFilename := fmt.Sprintf("riddles/%s.txt", config.Tomorrow.Format("2006-01-02"))
		if err := os.WriteFile(riddleFilename, []byte(riddle.String()), 0644); err != nil {
			slog.Error("Failed to write riddle to file", "file", riddleFilename, "error", err, "riddle", riddle.String())
		}
		enhancedMessage += fmt.Sprintf("\n\n*Riddle of the Day*:\n%s\n[Answer](https://veverkap.github.io/lunchmenu/riddles/%s.txt)", riddle.Riddle, config.Tomorrow.Format("2006-01-02"))
	}

	if menuWritten {
		enhancedMessage += fmt.Sprintf("\n\n[Without AI](https://veverkap.github.io/lunchmenu/menus/%s.txt)", config.Tomorrow.Format("2006-01-02"))
	}

	if gif != "" {
		enhancedMessage += fmt.Sprintf("\n\n*GIF of the Day*:\n%s", gif)
	}

	slog.Info("Final enhanced Telegram message", "message", enhancedMessage)

	slog.Info("Sending Telegram message")

	if err := telegramClient.SendMessage(enhancedMessage); err != nil {
		slog.Error("Failed to send Telegram message", "error", err)
	}
}
