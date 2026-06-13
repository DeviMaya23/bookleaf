package kinde

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/devi/bookleaf/internal/platform/config"
	"github.com/devi/bookleaf/internal/usecase"
)

// tokenExpiryBuffer is subtracted from the token's reported expiry so that the
// cached token is refetched slightly before it actually expires.
const tokenExpiryBuffer = 60 * time.Second

// Client is a Kinde Management API client authenticated via M2M client credentials.
type Client struct {
	httpClient   *http.Client
	tokenURL     string
	clientID     string
	clientSecret string
	audience     string
	apiBaseURL   string

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

var _ usecase.KindeClient = (*Client)(nil)

// NewClient builds a Kinde Management API client from the given Kinde config.
func NewClient(cfg config.KindeConfig) *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		tokenURL:     cfg.M2MTokenURL,
		clientID:     cfg.M2MClientID,
		clientSecret: cfg.M2MClientSecret,
		audience:     cfg.ManagementAudience,
		apiBaseURL:   strings.TrimRight(cfg.IssuerURL, "/") + "/api/v1",
	}
}

// DeleteUser permanently deletes the given user (and their profile) from Kinde.
// A response indicating the user no longer exists is treated as success, so this
// method is safe to call again for an already-deleted user.
func (c *Client) DeleteUser(ctx context.Context, kindeUserID string) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return fmt.Errorf("get m2m token: %w", err)
	}

	endpoint := fmt.Sprintf("%s/user?id=%s&is_delete_profile=true", c.apiBaseURL, url.QueryEscape(kindeUserID))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build delete user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete kinde user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	if isUserNotFoundError(body) {
		return nil
	}

	return fmt.Errorf("delete kinde user: unexpected status %d: %s", resp.StatusCode, string(body))
}

// kindeErrorResponse is the shape of Kinde Management API error responses:
// {"errors":[{"code":"...","message":"..."}]}
type kindeErrorResponse struct {
	Errors []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func isUserNotFoundError(body []byte) bool {
	var parsed kindeErrorResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return false
	}
	for _, e := range parsed.Errors {
		if strings.Contains(strings.ToUpper(e.Code), "NOT_FOUND") ||
			strings.Contains(strings.ToUpper(e.Message), "NOT FOUND") {
			return true
		}
	}
	return false
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiresAt.Add(-tokenExpiryBuffer)) {
		return c.accessToken, nil
	}

	token, expiresIn, err := c.fetchToken(ctx)
	if err != nil {
		return "", err
	}

	c.accessToken = token
	c.expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	return c.accessToken, nil
}

func (c *Client) fetchToken(ctx context.Context) (string, int, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("audience", c.audience)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", 0, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("fetch m2m token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("fetch m2m token: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", 0, fmt.Errorf("decode m2m token response: %w", err)
	}

	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}
