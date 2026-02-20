package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

func getTomorrowsWeather(date time.Time) (string, error) {
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

	var weather string
	_, _ = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		dateStr, _ := jsonparser.GetString(value, "date")
		parsedDate, _ := time.Parse("2006-01-02", dateStr)
		loc, _ := time.LoadLocation("America/New_York")
		parsedDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 12, 0, 0, 0, loc)
		if parsedDate.Equal(date) {
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
