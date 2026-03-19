package lunch

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

type School struct {
	ID   string
	Name string
}

func GetLunchMenuForSchool(date time.Time, school School) (string, []string, error) {
	slog.Info("Getting lunch menu", "date", date, "school", school.Name)
	url := fmt.Sprintf(
		"https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitemsByGrade?SchoolId=%s&ServingDate=%s&ServingLine=Main%%20Line&MealType=Lunch&Grade=02&PersonId=null",
		school.ID,
		url.QueryEscape(date.Format("01/02/2006")),
	)
	slog.Info("Constructed URL", "url", url)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		slog.Error("Failed to get lunch menu", "school", school.Name, "error", err)
		return "", nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("Failed to close response body", "error", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to get lunch menu", "school", school.Name, "status", resp.Status)
		return "", nil, fmt.Errorf("failed to get lunch menu: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read lunch menu", "school", school.Name, "error", err)
		return "", nil, err
	}

	menu := strings.Builder{}
	menuItems := []string{}

	_, _ = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		description, _ := jsonparser.GetString(value, "MenuItemDescription")
		slog.Info("Menu item", "item", description)
		menu.WriteString(description + "\n")
		menuItems = append(menuItems, description)
	}, "ENTREES")

	if menu.Len() == 0 {
		return "", nil, errors.New("no menu items found for " + school.Name)
	}
	return menu.String(), menuItems, nil
}
