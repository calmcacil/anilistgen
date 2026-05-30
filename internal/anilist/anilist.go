package anilist

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	apiURL    = "https://graphql.anilist.co"
	maxRetry  = 3
	pageSize  = 100
)

// Show represents an anime show from the AniList API.
type Show struct {
	ID       int      `json:"id"`
	IDMal    *int     `json:"idMal"`
	Title    Title    `json:"title"`
	Format   string   `json:"format"`
	Episodes *int     `json:"episodes"`
	Genres   []string `json:"genres"`
	Status   string   `json:"status"`
}

// Title holds the english and romaji titles.
type Title struct {
	English *string `json:"english"`
	Romaji  *string `json:"romaji"`
}

// DisplayTitle returns the English title if available, falling back to romaji.
func (s Show) DisplayTitle() string {
	if s.Title.English != nil && *s.Title.English != "" {
		return *s.Title.English
	}
	if s.Title.Romaji != nil {
		return *s.Title.Romaji
	}
	return fmt.Sprintf("Anime #%d", s.ID)
}

// graphqlResponse is the top-level response from AniList.
type graphqlResponse struct {
	Data struct {
		Page struct {
			Media []Show `json:"media"`
		} `json:"Page"`
	} `json:"data"`
}

// Client fetches data from the AniList GraphQL API.
type Client struct {
	http *http.Client
}

// New creates a new AniList client.
func New() *Client {
	return &Client{
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchSeason returns all TV/ONA anime for the given season and year.
func (c *Client) FetchSeason(ctx context.Context, season string, year int) ([]Show, error) {
	query := `query($s: MediaSeason, $y: Int, $page: Int, $perPage: Int) {
		Page(page: $page, perPage: $perPage) {
			media(
				season: $s, seasonYear: $y,
				type: ANIME,
				sort: POPULARITY_DESC,
				format_in: [TV, ONA]
			) {
				id
				idMal
				title { romaji english }
				format
				episodes
				genres
				status
			}
		}
	}`

	payload := map[string]any{
		"query": query,
		"variables": map[string]any{
			"s":       season,
			"y":       year,
			"page":    1,
			"perPage": pageSize,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	var resp graphqlResponse
	if err := c.doRequest(ctx, body, &resp); err != nil {
		return nil, fmt.Errorf("fetch %s %d: %w", season, year, err)
	}

	return resp.Data.Page.Media, nil
}

func (c *Client) doRequest(ctx context.Context, body []byte, dst any) error {
	var lastErr error
	for attempt := range maxRetry {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL,
			io.NopCloser(strings.NewReader(string(body))))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request: %w", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited (attempt %d)", attempt+1)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
			continue
		}

		return json.NewDecoder(resp.Body).Decode(dst)
	}

	return fmt.Errorf("giving up after %d retries: %w", maxRetry, lastErr)
}
