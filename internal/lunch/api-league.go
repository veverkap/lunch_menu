package lunch

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/buger/jsonparser"
)

type APILeagueClient struct {
	Key    string
	Client *http.Client
}

func NewAPIClient(key string) *APILeagueClient {
	return &APILeagueClient{Key: key, Client: &http.Client{Timeout: 10 * time.Second}}
}

type RiddleResponse struct {
	Riddle string
	Answer string
}

func (r *RiddleResponse) String() string {
	return fmt.Sprintf("Riddle: %s\nAnswer: %s", r.Riddle, r.Answer)
}

func (c *APILeagueClient) GetGif(query string) (string, error) {
	urlEncodedQuery := url.QueryEscape(query)
	url := fmt.Sprintf("https://api.apileague.com/search-gifs?query=%s&number=1&api-key=%s", urlEncodedQuery, c.Key)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return "", err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		slog.Error("Failed to get GIF", "error", err)
		return "", err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("Failed to close response body", "error", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to get GIF", "status", resp.Status)
		return "", fmt.Errorf("failed to get GIF: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read GIF response", "error", err)
		return "", err
	}

	var gifURL string
	_, _ = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		if gifURL == "" {
			u, _ := jsonparser.GetString(value, "url")
			gifURL = u
		}
	}, "images")

	if gifURL == "" {
		slog.Warn("No GIFs found for query", "query", query)
		return "", nil // no error, just no results
	}

	return gifURL, nil
}

func (c *APILeagueClient) GetRiddle(diffID string) (RiddleResponse, error) {
	req, err := http.NewRequest("GET", "https://api.apileague.com/retrieve-random-riddle?difficulty="+diffID+"&api-key="+c.Key, nil)
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return RiddleResponse{}, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		slog.Error("Failed to get riddle", "error", err)
		return RiddleResponse{}, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("Failed to close response body", "error", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to get riddle", "status", resp.Status)
		return RiddleResponse{}, fmt.Errorf("failed to get riddle: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read riddle response", "error", err)
		return RiddleResponse{}, err
	}

	riddle, err := jsonparser.GetString(body, "riddle")
	if err != nil {
		slog.Error("Failed to parse riddle field", "error", err)
		return RiddleResponse{}, err
	}
	answer, err := jsonparser.GetString(body, "answer")
	if err != nil {
		slog.Error("Failed to parse answer field", "error", err)
		return RiddleResponse{}, err
	}

	return RiddleResponse{Riddle: riddle, Answer: answer}, nil
}
