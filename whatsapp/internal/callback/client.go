package callback

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Client sends HMAC-signed JSON POST requests to the ChatSolv backend.
// The signature scheme matches the backend's internal HMAC middleware:
//
//	HMAC-SHA256(secret, timestamp + "." + body)
type Client struct {
	baseURL string
	secret  string
	http    *http.Client
	log     *slog.Logger
	now     func() time.Time // injectable for testing
}

// New creates a Client pointed at baseURL.
func New(baseURL, secret string, timeout time.Duration, log *slog.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		secret:  secret,
		http:    &http.Client{Timeout: timeout},
		log:     log,
		now:     time.Now,
	}
}

// Post marshals payload to JSON and sends it as an HMAC-signed POST to path.
func (c *Client) Post(ctx context.Context, path string, payload any) error {
	return c.PostWithResponse(ctx, path, payload, nil)
}

// PostWithResponse marshals payload, sends an HMAC-signed POST to path,
// checks for 2xx status, and optionally decodes the JSON response body into target.
func (c *Client) PostWithResponse(ctx context.Context, path string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	ts := c.now().UTC().Format(time.RFC3339)
	mac := hmac.New(sha256.New, []byte(c.secret))
	mac.Write([]byte(ts + "." + string(body)))
	sig := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ChatSolv-Timestamp", ts)
	req.Header.Set("X-ChatSolv-Signature", sig)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("backend returned HTTP %d for %s", resp.StatusCode, path)
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
