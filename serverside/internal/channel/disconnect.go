package channel

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (c *HMACBotClient) GetProfile(ctx context.Context, channelID string) (WhatsAppProfile, error) {
	timestamp := c.now().UTC().Format(time.RFC3339)
	mac := hmac.New(sha256.New, []byte(c.secret))
	_, _ = mac.Write([]byte(timestamp + "."))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/v1/channels/profile?channel_id="+channelID, nil)
	if err != nil {
		return WhatsAppProfile{}, err
	}
	req.Header.Set("X-ChatSolv-Timestamp", timestamp)
	req.Header.Set("X-ChatSolv-Signature", hex.EncodeToString(mac.Sum(nil)))
	res, err := c.client.Do(req)
	if err != nil {
		return WhatsAppProfile{}, err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return WhatsAppProfile{}, fmt.Errorf("bot service returned %d", res.StatusCode)
	}
	var envelope struct {
		Data WhatsAppProfile `json:"data"`
	}
	if err = json.NewDecoder(res.Body).Decode(&envelope); err != nil {
		return WhatsAppProfile{}, err
	}
	return envelope.Data, nil
}

func (c *HMACBotClient) Disconnect(ctx context.Context, channelID string) error {
	payload, _ := json.Marshal(map[string]string{"channel_id": channelID})
	timestamp := c.now().UTC().Format(time.RFC3339)
	mac := hmac.New(sha256.New, []byte(c.secret))
	_, _ = mac.Write([]byte(timestamp + "." + string(payload)))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/channels/disconnect", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ChatSolv-Timestamp", timestamp)
	req.Header.Set("X-ChatSolv-Signature", hex.EncodeToString(mac.Sum(nil)))
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("bot service returned %d", res.StatusCode)
	}
	return nil
}

func (c *HMACBotClient) SendTextMessage(ctx context.Context, channelID, recipient, text string) error {
	payload, _ := json.Marshal(map[string]string{
		"channel_id": channelID,
		"recipient":  recipient,
		"text":       text,
	})
	timestamp := c.now().UTC().Format(time.RFC3339)
	mac := hmac.New(sha256.New, []byte(c.secret))
	_, _ = mac.Write([]byte(timestamp + "." + string(payload)))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/messages/send", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ChatSolv-Timestamp", timestamp)
	req.Header.Set("X-ChatSolv-Signature", hex.EncodeToString(mac.Sum(nil)))
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("bot service returned %d", res.StatusCode)
	}
	return nil
}
