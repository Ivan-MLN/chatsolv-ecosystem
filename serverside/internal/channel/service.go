package channel

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrForbidden        = errors.New("channel forbidden")
	ErrNotFound         = errors.New("channel or workspace not found")
	ErrQuotaExceeded    = errors.New("channel quota exceeded")
	ErrInvalidInput     = errors.New("invalid channel input")
	ErrConnectionFailed = errors.New("whatsapp connection failed")
)

type Channel struct {
	ID                string `json:"id"`
	WorkspaceID       string `json:"workspace_id"`
	AgentID           string `json:"agent_id"`
	Type              string `json:"type"`
	DisplayName       string `json:"display_name"`
	PhoneNumber       string `json:"phone_number,omitempty"`
	Status            string `json:"status"`
	ServiceInstanceID string `json:"-"`
}
type Repository interface {
	Authorize(context.Context, string, string) (string, string, error)
	AuthorizeMutation(context.Context, string, string) error
	Count(context.Context, string) (int64, error)
	Max(context.Context, string) (int64, error)
	Create(context.Context, Channel) (Channel, error)
	List(context.Context, string, string) ([]Channel, error)
	UpdateStatus(context.Context, string, string, string, string) error
	Delete(context.Context, string, string) error
}
type WhatsAppProfile struct {
	PushName        string   `json:"push_name"`
	Phone           string   `json:"phone"`
	JID             string   `json:"jid"`
	LID             string   `json:"lid,omitempty"`
	Status          string   `json:"status,omitempty"`
	ProfilePicture  string   `json:"profile_picture,omitempty"`
	IsBusiness      bool     `json:"is_business"`
	Description     string   `json:"description,omitempty"`
	Address         string   `json:"address,omitempty"`
	Email           string   `json:"email,omitempty"`
	Categories      []string `json:"categories,omitempty"`
	BusinessHoursTZ string   `json:"business_hours_tz,omitempty"`
}

type BotClient interface {
	Connect(context.Context, string, string) (ConnectResponse, error)
	Disconnect(context.Context, string) error
	GetProfile(context.Context, string) (WhatsAppProfile, error)
	SendTextMessage(context.Context, string, string, string) error
}
type ConnectResponse struct {
	SessionID, Status, QR, PairingCode string `json:"-"`
}
type Service struct {
	repo Repository
	bot  BotClient
}

func NewService(repo Repository, bot BotClient) *Service { return &Service{repo, bot} }
func (s *Service) List(ctx context.Context, userID, workspaceID string) ([]Channel, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.List(ctx, workspaceID, userID)
}
func (s *Service) GetProfile(ctx context.Context, userID, channelID string) (WhatsAppProfile, error) {
	if strings.TrimSpace(channelID) == "" {
		return WhatsAppProfile{}, ErrInvalidInput
	}
	if err := s.repo.AuthorizeMutation(ctx, userID, channelID); err != nil {
		return WhatsAppProfile{}, err
	}
	return s.bot.GetProfile(ctx, channelID)
}
func (s *Service) SetStatus(ctx context.Context, userID, channelID, newStatus string) (Channel, error) {
	if strings.TrimSpace(channelID) == "" {
		return Channel{}, ErrInvalidInput
	}
	if err := s.repo.AuthorizeMutation(ctx, userID, channelID); err != nil {
		return Channel{}, err
	}
	if newStatus != "connected" && newStatus != "suspended" {
		return Channel{}, ErrInvalidInput
	}
	if err := s.repo.UpdateStatus(ctx, channelID, newStatus, "", ""); err != nil {
		return Channel{}, err
	}
	return Channel{ID: channelID, Status: newStatus}, nil
}

func (s *Service) Restart(ctx context.Context, userID, channelID string) (Channel, error) {
	if strings.TrimSpace(channelID) == "" {
		return Channel{}, ErrInvalidInput
	}
	if err := s.repo.AuthorizeMutation(ctx, userID, channelID); err != nil {
		return Channel{}, err
	}
	_ = s.bot.Disconnect(ctx, channelID)
	// Try reconnecting
	pairing, err := s.bot.Connect(ctx, channelID, "")
	if err != nil {
		_ = s.repo.UpdateStatus(ctx, channelID, "error", "", "")
		return Channel{ID: channelID, Status: "error"}, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	_ = s.repo.UpdateStatus(ctx, channelID, pairing.Status, "", pairing.SessionID)
	return Channel{ID: channelID, Status: pairing.Status}, nil
}

func (s *Service) Delete(ctx context.Context, userID, channelID string) error {
	if strings.TrimSpace(channelID) == "" {
		return ErrInvalidInput
	}
	if err := s.repo.AuthorizeMutation(ctx, userID, channelID); err != nil {
		return err
	}
	// Best-effort disconnect session on bot service, but don't fail deletion if bot service session is already gone or unreachable
	_ = s.bot.Disconnect(ctx, channelID)
	return s.repo.Delete(ctx, channelID, userID)
}
func (s *Service) ConnectWhatsApp(ctx context.Context, userID, workspaceID, displayName string, phoneNumber string) (Channel, ConnectResponse, error) {
	return s.ConnectWhatsAppWithBypass(ctx, userID, workspaceID, displayName, phoneNumber, false)
}

func (s *Service) ConnectWhatsAppWithBypass(ctx context.Context, userID, workspaceID, displayName string, phoneNumber string, unlimited bool) (Channel, ConnectResponse, error) {
	displayName = strings.TrimSpace(displayName)
	phoneNumber = strings.TrimSpace(phoneNumber)
	if workspaceID == "" || displayName == "" || len(displayName) > 120 {
		return Channel{}, ConnectResponse{}, ErrInvalidInput
	}
	role, agentID, err := s.repo.Authorize(ctx, userID, workspaceID)
	if err != nil {
		return Channel{}, ConnectResponse{}, err
	}
	if role != "owner" && role != "admin" {
		return Channel{}, ConnectResponse{}, ErrForbidden
	}
	if !unlimited {
		count, err := s.repo.Count(ctx, workspaceID)
		if err != nil {
			return Channel{}, ConnectResponse{}, err
		}
		max, err := s.repo.Max(ctx, workspaceID)
		if err != nil {
			return Channel{}, ConnectResponse{}, err
		}
		if count >= max {
			return Channel{}, ConnectResponse{}, ErrQuotaExceeded
		}
	}
	created, err := s.repo.Create(ctx, Channel{WorkspaceID: workspaceID, AgentID: agentID, Type: "whatsapp", DisplayName: displayName, PhoneNumber: phoneNumber, Status: "connecting"})
	if err != nil {
		return Channel{}, ConnectResponse{}, err
	}
	pairing, err := s.bot.Connect(ctx, created.ID, phoneNumber)
	if err != nil {
		_ = s.bot.Disconnect(ctx, created.ID)
		if cleanupErr := s.repo.Delete(ctx, created.ID, userID); cleanupErr != nil {
			_ = s.repo.UpdateStatus(ctx, created.ID, "error", phoneNumber, "")
		}
		return Channel{}, ConnectResponse{}, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	_ = s.repo.UpdateStatus(ctx, created.ID, pairing.Status, phoneNumber, pairing.SessionID)
	created.Status = pairing.Status
	created.ServiceInstanceID = pairing.SessionID
	return created, pairing, nil
}

type HMACBotClient struct {
	baseURL, secret string
	client          *http.Client
	now             func() time.Time
}

func NewHMACBotClient(baseURL, secret string) *HMACBotClient {
	return &HMACBotClient{baseURL, secret, &http.Client{Timeout: 15 * time.Second}, time.Now}
}
func (c *HMACBotClient) Connect(ctx context.Context, channelID string, phoneNumber string) (ConnectResponse, error) {
	payload, _ := json.Marshal(map[string]string{"channel_id": channelID, "phone_number": phoneNumber})
	timestamp := c.now().UTC().Format(time.RFC3339)
	mac := hmac.New(sha256.New, []byte(c.secret))
	_, _ = mac.Write([]byte(timestamp + "." + string(payload)))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/channels/connect", bytes.NewReader(payload))
	if err != nil {
		return ConnectResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ChatSolv-Timestamp", timestamp)
	req.Header.Set("X-ChatSolv-Signature", hex.EncodeToString(mac.Sum(nil)))
	res, err := c.client.Do(req)
	if err != nil {
		return ConnectResponse{}, err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return ConnectResponse{}, fmt.Errorf("bot service returned %d", res.StatusCode)
	}
	var envelope struct {
		Data struct {
			SessionID   string `json:"session_id"`
			Status      string `json:"status"`
			QR          string `json:"qr"`
			PairingCode string `json:"pairing_code"`
		} `json:"data"`
	}
	if err = json.NewDecoder(res.Body).Decode(&envelope); err != nil {
		return ConnectResponse{}, err
	}
	return ConnectResponse{
		SessionID:   envelope.Data.SessionID,
		Status:      envelope.Data.Status,
		QR:          envelope.Data.QR,
		PairingCode: envelope.Data.PairingCode,
	}, nil
}
