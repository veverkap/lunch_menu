package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

func getLunchMenuForSchool(date time.Time, school School) (string, error) {
	logger.Info("Getting lunch menu", "date", date, "school", school.Name)
	url := fmt.Sprintf(
		"https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitemsByGrade?SchoolId=%s&ServingDate=%s&ServingLine=Main%%20Line&MealType=Lunch&Grade=02&PersonId=null",
		school.ID,
		url.QueryEscape(date.Format("01/02/2006")),
	)
	logger.Info("Constructed URL", "url", url)

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read lunch menu", "school", school.Name, "error", err)
		return "", err
	}

	menu := strings.Builder{}

	_, _ = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		description, _ := jsonparser.GetString(value, "MenuItemDescription")
		logger.Info("Menu item", "item", description)
		menu.WriteString(description + "\n")
	}, "ENTREES")

	if menu.Len() == 0 {
		return "", errors.New("no menu items found for " + school.Name)
	}
	return menu.String(), nil
}
