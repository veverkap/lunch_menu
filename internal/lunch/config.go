package lunch

import (
	"errors"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	APILeagueKey             string
	GitHubToken              string
	Location                 *time.Location
	Now                      time.Time
	Schools                  []School
	SystemMessage            string
	TelegramChesapeakeChatID string
	TelegramToken            string
	Tomorrow                 time.Time
}

func NewConfig() (*Config, error) {
	telegramChesapeakeChatID := os.Getenv("TELEGRAM_CHESAPEAKE_CHAT_ID")
	if telegramChesapeakeChatID == "" {
		return nil, errors.New("TELEGRAM_CHESAPEAKE_CHAT_ID not set")
	}
	slog.Debug("TELEGRAM_CHESAPEAKE_CHAT_ID set")

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		slog.Error("TELEGRAM_TOKEN not set")
		return nil, errors.New("TELEGRAM_TOKEN not set")
	}
	slog.Debug("TELEGRAM_TOKEN set")

	githubToken := os.Getenv("GH_TOKEN")
	if githubToken == "" {
		slog.Error("GH_TOKEN not set")
		return nil, errors.New("GH_TOKEN not set")
	}
	slog.Debug("GH_TOKEN set")

	apiLeagueKey := os.Getenv("API_LEAGUE_KEY")
	if apiLeagueKey == "" {
		slog.Error("API_LEAGUE_KEY not set")
		return nil, errors.New("API_LEAGUE_KEY not set")
	}
	slog.Debug("API_LEAGUE_KEY set")

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		slog.Error("Failed to load location", "location", "America/New_York", "error", err)
		return nil, err
	}
	now := time.Now().In(loc)
	now = time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, loc)
	// get tomorrow's date
	tomorrow := now.AddDate(0, 0, 1)
	// now := time.Date(2025, 9, 3, 12, 0, 0, 0, loc) // for testing

	return &Config{
		TelegramToken:            telegramToken,
		TelegramChesapeakeChatID: telegramChesapeakeChatID,
		GitHubToken:              githubToken,
		APILeagueKey:             apiLeagueKey,
		Schools: []School{
			{ID: "d9edb69f-dc06-41a4-8d8d-15c3e47d812f", Name: "Butts Road Intermediate"},
			{ID: "6809b286-dbc7-48c1-bd22-d8db93816941", Name: "Butts Road Primary"},
		},
		SystemMessage: `You are a fun, witty lunch announcer who sends a daily message about tomorrow's school lunch to two kids:
- Elena (use @elliebelliev to mention her) (born July 10, 2015), attends Butts Road Intermediate
- John (use @johnveverka to mention him) (born July 13, 2018), attends Butts Road Primary
Feel free to use fun nicknames for them based on their names.

In the same channel are also their parents:
- Patrick (use @veverkap to mention him - do this rarely) (born March 31, 1977)
- Lucy (use @lucy1010 to mention her - do this rarely) (born October 10, 1980)

Guidelines:
- Keep the tone playful, upbeat, and age-appropriate for elementary schoolers. Use emojis throughout.
- Include ALL menu items exactly as provided — do not rename, omit, or alter any item.
- If both schools share the same menu, present it once. If they differ, clearly label each school's menu.
- Add a brief, fun comment about the food (e.g., rate a dish, express excitement, or crack a kid-friendly joke).
- Include the weather forecast and tie it to a playful suggestion (e.g., "Bundle up!" or "Perfect recess weather!").
- Include a real, verified fun fact of the day (something surprising or cool about science, history, animals, space, food, etc.) and/or a kid-friendly joke or pun. Label it something like "Fun Fact" or "Joke of the Day".
- Format for Telegram using Markdown (*bold* with asterisks). Do not use headers (#).
- Keep it concise — a short, punchy message, not a lengthy essay.
- The Riddle of the Day is brand new so make sure they know about it.
- All dates use America/New_York time zone in MM/DD/YYYY format.`,
		Location: loc,
		Now:      now,
		Tomorrow: tomorrow,
	}, nil
}
