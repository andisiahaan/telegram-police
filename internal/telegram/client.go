package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const apiBaseURL = "https://api.telegram.org/bot%s/%s"

// Client is an HTTP client for the Telegram Bot API.
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a Client with a 10-second timeout.
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Do sends a POST JSON request to the given Telegram API method and returns
// the raw result field. Returns an error if the HTTP call fails or if Telegram
// reports ok: false.
func (c *Client) Do(method string, params map[string]any) (json.RawMessage, error) {
	url := fmt.Sprintf(apiBaseURL, c.token, method)

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("telegram client: marshal params: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("telegram client: http post %s: %w", method, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("telegram client: read body %s: %w", method, err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(raw, &apiResp); err != nil {
		return nil, fmt.Errorf("telegram client: unmarshal response %s: %w", method, err)
	}

	if !apiResp.OK {
		slog.Error("telegram api error", "method", method, "description", apiResp.Description)
		return nil, fmt.Errorf("telegram api: %s: %s", method, apiResp.Description)
	}

	return apiResp.Result, nil
}
