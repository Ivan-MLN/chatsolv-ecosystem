package server

import (
	"bytes"
	"chatsolv-whatsapp/internal/whatsapp"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockSessionManager struct {
	connectCalled    bool
	phoneNumber      string
	disconnectCalled bool
	connected        bool
	sentRecipient    string
	sentText         string
	result           ConnectResult
	err              error
}

func (m *mockSessionManager) Connect(ctx context.Context, channelID, phoneNumber string) (ConnectResult, error) {
	m.connectCalled = true
	m.phoneNumber = phoneNumber
	return m.result, m.err
}

func (m *mockSessionManager) Disconnect(channelID string) error {
	m.disconnectCalled = true
	return m.err
}

func (m *mockSessionManager) IsConnected(channelID string) bool {
	return m.connected
}

func (m *mockSessionManager) GetProfile(ctx context.Context, channelID string) (whatsapp.WhatsAppProfile, error) {
	return whatsapp.WhatsAppProfile{PushName: "Test Name", Phone: "628123456789"}, m.err
}

func (m *mockSessionManager) SendTextMessage(ctx context.Context, channelID, recipient, text string) error {
	m.sentRecipient = recipient
	m.sentText = text
	return m.err
}

func sign(secret, body string, t time.Time) (string, string) {
	ts := t.UTC().Format(time.RFC3339)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + body))
	return ts, hex.EncodeToString(mac.Sum(nil))
}

func TestServerHealth(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mgr := &mockSessionManager{}
	h := NewHandler(mgr, log)
	srv := New(":0", "secret-key-that-is-at-least-32-bytes-long", h, log)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	srv.http.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, true, body["success"])
	require.Equal(t, "ok", body["status"])
}

func TestServerConnectRequiresHMAC(t *testing.T) {
	secret := "secret-key-that-is-at-least-32-bytes-long"
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mgr := &mockSessionManager{result: ConnectResult{SessionID: "ch1", Status: "waiting_pairing", QR: "qr123"}}
	h := NewHandler(mgr, log)
	srv := New(":0", secret, h, log)

	// Unauthenticated request -> 401
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/channels/connect", bytes.NewReader([]byte(`{"channel_id":"ch1"}`)))
	srv.http.Handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	// Authenticated request -> 200
	payload := `{"channel_id":"ch1"}`
	ts, sig := sign(secret, payload, time.Now())
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal/v1/channels/connect", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ChatSolv-Timestamp", ts)
	req.Header.Set("X-ChatSolv-Signature", sig)
	srv.http.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, mgr.connectCalled)
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			SessionID string `json:"session_id"`
			Status    string `json:"status"`
			QR        string `json:"qr"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.True(t, body.Success)
	require.Equal(t, "waiting_pairing", body.Data.Status)
	require.Equal(t, "qr123", body.Data.QR)
}

func TestServerConnectPassesPhoneNumberForPairingCode(t *testing.T) {
	secret := "secret-key-that-is-at-least-32-bytes-long"
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mgr := &mockSessionManager{result: ConnectResult{SessionID: "ch1", Status: "waiting_pairing", PairingCode: "ABCD-EFGH"}}
	srv := New(":0", secret, NewHandler(mgr, log), log)
	payload := `{"channel_id":"ch1","phone_number":"628123456789"}`
	ts, sig := sign(secret, payload, time.Now())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/channels/connect", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ChatSolv-Timestamp", ts)
	req.Header.Set("X-ChatSolv-Signature", sig)

	srv.http.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "628123456789", mgr.phoneNumber)
}

func TestServerDisconnect(t *testing.T) {
	secret := "secret-key-that-is-at-least-32-bytes-long"
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mgr := &mockSessionManager{}
	h := NewHandler(mgr, log)
	srv := New(":0", secret, h, log)

	payload := `{"channel_id":"ch1"}`
	ts, sig := sign(secret, payload, time.Now())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/channels/disconnect", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ChatSolv-Timestamp", ts)
	req.Header.Set("X-ChatSolv-Signature", sig)
	srv.http.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, mgr.disconnectCalled)
}

func TestServerSendMessage(t *testing.T) {
	secret := "secret-key-that-is-at-least-32-bytes-long"
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mgr := &mockSessionManager{}
	h := NewHandler(mgr, log)
	srv := New(":0", secret, h, log)

	payload := `{"channel_id":"ch1","recipient":"628123456789","text":"Halo ini pesan eskalasi"}`
	ts, sig := sign(secret, payload, time.Now())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/messages/send", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ChatSolv-Timestamp", ts)
	req.Header.Set("X-ChatSolv-Signature", sig)
	srv.http.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "628123456789", mgr.sentRecipient)
	require.Equal(t, "Halo ini pesan eskalasi", mgr.sentText)
}
