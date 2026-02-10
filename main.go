package main

import (
	"context"
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
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type School struct {
	ID   string
	Name string
}

var (
	telegramToken            string
	telegramChesapeakeChatID string
	githubToken              string
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
	tomorrow := now.AddDate(0, 0, 1)
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
		telegramMessages.WriteString(fmt.Sprintf("*%s*:\n%s\n", schools[0].Name, firstLunchMenu))
		telegramMessages.WriteString(fmt.Sprintf("*%s*:\n%s\n", schools[1].Name, secondLunchMenu))
	} else {
		logger.Info("Lunch menus are the same between schools")
		telegramMessages.WriteString(fmt.Sprintf("*Lunch Menu*:\n%s\n", firstLunchMenu))
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
		telegramMessages.WriteString(fmt.Sprintf("*Weather*:\n%s\n", weather))
	}

	telegramMessage := strings.Builder{}
	telegramMessage.WriteString(fmt.Sprintf("Lunch menu for %s:\n\n", tomorrow.Format("01/02/2006")))
	telegramMessage.WriteString(telegramMessages.String())
	telegramMessage.WriteString(fmt.Sprintf("\n*Weather*:\n%s\n", weather))

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

}

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
	// implementation goes here
	return chatCompletion.Choices[0].Message.Content, nil
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
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Error("Failed to close response body", "error", cerr)
		}
	}()

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

func getTomorrowsWeather(date time.Time) (string, error) {
	// implementation goes here
	url := "https://wttr.in/Chesapeake?format=j1"
	logger.Info("Getting weather", "date", date, "url", url)

	requestStart := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		logger.Error("Failed to get weather", "error", err)
		return "", err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Error("Failed to close response body", "error", cerr)
		}
	}()
	logger.Info("Weather request completed", "duration", time.Since(requestStart).String())

	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to get weather", "status", resp.Status)
		return "", fmt.Errorf("failed to get weather: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read weather", "error", err)
		return "", err
	}

	// find the forecast for the given date
	var weather string
	_, _ = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		dateStr, _ := jsonparser.GetString(value, "date")
		parsedDate, _ := time.Parse("2006-01-02", dateStr)
		// compare just the date parts
		// we assume the date is in local time, so we set the time to noon to avoid any DST issues
		loc, _ := time.LoadLocation("America/New_York")
		parsedDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 12, 0, 0, 0, loc)
		// logger.Info("Parsed date", "dateStr", dateStr, "parsedDate", parsedDate, "targetDate", date)
		if parsedDate.Equal(date) {
			// we want to get the weather at 6am, 12pm, and 6pm
			hourlyWeather := strings.Builder{}
			_, _ = jsonparser.ArrayEach(value, func(hourlyValue []byte, _ jsonparser.ValueType, _ int, _ error) {
				timeStr, _ := jsonparser.GetString(hourlyValue, "time")
				if timeStr == "600" || timeStr == "1200" || timeStr == "1800" {
					tempF, _ := jsonparser.GetString(hourlyValue, "tempF")
					weatherDesc, _ := jsonparser.GetString(hourlyValue, "weatherDesc", "[0]", "value")
					switch timeStr {
					case "600":
						timeStr = "6 AM"
					case "1200":
						timeStr = "12 PM"
					case "1800":
						timeStr = "6 PM"
					}
					hourlyWeather.WriteString(fmt.Sprintf("%s - %s°F - %s\n", timeStr, tempF, weatherDesc))
				}
			}, "hourly")
			weather = hourlyWeather.String()
		}
	}, "weather")

	if weather != "" {
		return weather, nil
	}
	logger.Warn("No weather found for date", "date", date)
	return "", nil
}
