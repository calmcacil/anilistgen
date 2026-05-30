package mdblist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	apiBase   = "https://mdblist.com/api"
	maxRetry  = 3
	rateLimit = time.Second
)

// List represents an MDBList list.
type List struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Public      int    `json:"public"`
	URL         string `json:"url"`
	ItemsCount  int    `json:"items_count"`
}

// Client manages communication with the MDBList API.
type Client struct {
	http      *http.Client
	apiKey    string
	lastCall  time.Time
}

// New creates a new MDBList client.
func New(apiKey string) *Client {
	return &Client{
		http:   &http.Client{Timeout: 30 * time.Second},
		apiKey: apiKey,
	}
}

// throttle ensures we don't exceed MDBList rate limits.
func (c *Client) throttle() {
	elapsed := time.Since(c.lastCall)
	if elapsed < rateLimit {
		time.Sleep(rateLimit - elapsed)
	}
	c.lastCall = time.Now()
}

// ListLists returns all lists belonging to the authenticated user.
func (c *Client) ListLists(ctx context.Context) ([]List, error) {
	c.throttle()

	u := fmt.Sprintf("%s/lists?apikey=%s", apiBase, url.QueryEscape(c.apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MDBList API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var lists []List
	// MDBList response varies — try both array and object-wrapped
	if err := json.Unmarshal(body, &lists); err != nil {
		var wrapped struct {
			Lists []List `json:"lists"`
		}
		if err2 := json.Unmarshal(body, &wrapped); err2 != nil {
			return nil, fmt.Errorf("parse lists response: %w (raw: %s)", err, string(body))
		}
		lists = wrapped.Lists
	}

	return lists, nil
}

// FindListByTitle searches the user's lists for one with a matching title.
func (c *Client) FindListByTitle(ctx context.Context, title string) (*List, error) {
	lists, err := c.ListLists(ctx)
	if err != nil {
		return nil, err
	}
	for _, l := range lists {
		if strings.EqualFold(l.Title, title) {
			return &l, nil
		}
	}
	return nil, nil
}

// CreateList creates a new public MDBList list and returns it.
func (c *Client) CreateList(ctx context.Context, title, description string, items []string) (*List, error) {
	c.throttle()

	payload := map[string]any{
		"title":       title,
		"description": description,
		"public":      1,
		"items":       items,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	u := fmt.Sprintf("%s/list?apikey=%s", apiBase, url.QueryEscape(c.apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var lastErr error
	for attempt := range maxRetry {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<(attempt+1)) * time.Second)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request: %w", err)
			continue
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited (attempt %d)", attempt+1)
			continue
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			lastErr = fmt.Errorf("MDBList error (HTTP %d): %s", resp.StatusCode, string(respBody))
			continue
		}

		var list List
		if err := json.Unmarshal(respBody, &list); err != nil {
			return nil, fmt.Errorf("parse create response: %w", err)
		}
		return &list, nil
	}

	return nil, fmt.Errorf("giving up after %d retries: %w", maxRetry, lastErr)
}

// UpdateList replaces all items in an existing list.
func (c *Client) UpdateList(ctx context.Context, listID string, items []string) error {
	c.throttle()

	payload := map[string]any{
		"items": items,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	u := fmt.Sprintf("%s/list/%s?apikey=%s", apiBase, url.PathEscape(listID), url.QueryEscape(c.apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("update request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MDBList update error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
