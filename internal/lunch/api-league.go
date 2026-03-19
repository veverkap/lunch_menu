package lunch

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type APILeagueClient struct {
	Key string
}

func NewAPIClient(key string) *APILeagueClient {
	return &APILeagueClient{Key: key}
}

type GifResponse struct {
	Images []struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"images"`
}

type RiddleResponse struct {
	Riddle string `json:"riddle"`
	Answer string `json:"answer"`
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
	resp, err := http.DefaultClient.Do(req)
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

	var gifResp GifResponse
	if err := json.NewDecoder(resp.Body).Decode(&gifResp); err != nil {
		slog.Error("Failed to decode GIF response", "error", err)
		return "", err
	}

	if len(gifResp.Images) == 0 {
		slog.Warn("No GIFs found for query", "query", query)
		return "", nil // no error, just no results
	}

	return gifResp.Images[0].URL, nil
}

func (c *APILeagueClient) GetRiddle(diffID string) (RiddleResponse, error) {
	req, err := http.NewRequest("GET", "https://api.apileague.com/retrieve-random-riddle?difficulty="+diffID+"&api-key="+c.Key, nil)
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return RiddleResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)

	resp, err := http.DefaultClient.Do(req)
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

	var riddleResp RiddleResponse
	if err := json.NewDecoder(resp.Body).Decode(&riddleResp); err != nil {
		slog.Error("Failed to decode riddle response", "error", err)
		return RiddleResponse{}, err
	}

	return riddleResp, nil
}
