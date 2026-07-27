package booklet

import (
	"context"
	"fmt"
	"net/http"
)

type Client struct {
	httpClient     *http.Client
	baseURL        string
	internalSecret string
}

func NewClient(baseURL, secret string) *Client {
	return &Client{
		httpClient:     &http.Client{},
		baseURL:        baseURL,
		internalSecret: secret,
	}
}

func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	url := fmt.Sprintf("%s/internal/users/%s", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Booklet-Internal-Secret", c.internalSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("booklet returned status %d", resp.StatusCode)
	}
	return nil
}
