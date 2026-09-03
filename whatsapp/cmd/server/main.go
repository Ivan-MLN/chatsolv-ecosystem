package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"chatsolv-whatsapp/internal/callback"
	"chatsolv-whatsapp/internal/config"
	"chatsolv-whatsapp/internal/server"
	"chatsolv-whatsapp/internal/whatsapp"
	"github.com/joho/godotenv"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"google.golang.org/protobuf/proto"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "error", err)
		os.Exit(1)
	}

	// Fetch and sync the latest WhatsApp Web client version dynamically
	if latestVer, err := whatsmeow.GetLatestVersion(context.Background(), nil); err == nil && latestVer != nil {
		store.SetWAVersion(*latestVer)
		slog.Info("synced latest whatsapp web version", "version", latestVer.String())
	}

	// Set realistic desktop browser device properties (Ubuntu / Chrome)
	// matching standard WhatsApp Web clients
	store.DeviceProps.Os = proto.String("Ubuntu")
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
	store.SetOSInfo("Ubuntu", [3]uint32{22, 4, 4})

	log := buildLogger(cfg.LogLevel)

	// Callback client sends HMAC-signed events to the ChatSolv backend.
	cbClient := callback.New(cfg.BackendURL, cfg.InternalServiceSecret, cfg.CallbackTimeout, log)
	cbHandler := callback.NewHandler(cbClient, log)

	// Manager owns all whatsmeow sessions, one per channel.
	mgr := whatsapp.NewManager(cfg.DBRoot, log, cbHandler.Handle)
	cbHandler.SetSender(mgr)

	// HTTP server exposes the internal control API.
	h := server.NewHandler(adaptManager(mgr), log)
	srv := server.New(":"+cfg.Port, cfg.InternalServiceSecret, h, log)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Error("server stopped unexpectedly", "error", err)
	case s := <-sig:
		log.Info("shutdown signal received", "signal", s)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	mgr.DisconnectAll()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "error", err)
	}
	log.Info("service stopped")
}

func buildLogger(level string) *slog.Logger {
	l := slog.LevelInfo
	if level == "debug" {
		l = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}

// managerAdapter bridges *whatsapp.Manager to server.SessionManager.
// It converts whatsapp.ConnectResult to server.ConnectResult without
// creating an import cycle between the server and whatsapp packages.
type managerAdapter struct{ m *whatsapp.Manager }

func adaptManager(m *whatsapp.Manager) server.SessionManager {
	return &managerAdapter{m}
}

func (a *managerAdapter) Connect(ctx context.Context, channelID string, phoneNumber string) (server.ConnectResult, error) {
	r, err := a.m.Connect(ctx, channelID, phoneNumber)
	if err != nil {
		return server.ConnectResult{}, err
	}
	return server.ConnectResult{SessionID: r.SessionID, Status: r.Status, QR: r.QR, PairingCode: r.PairingCode}, nil
}

func (a *managerAdapter) Disconnect(channelID string) error {
	return a.m.Disconnect(channelID)
}

func (a *managerAdapter) IsConnected(channelID string) bool {
	return a.m.IsConnected(channelID)
}

func (a *managerAdapter) GetProfile(ctx context.Context, channelID string) (whatsapp.WhatsAppProfile, error) {
	return a.m.GetProfile(ctx, channelID)
}

func (a *managerAdapter) SendTextMessage(ctx context.Context, channelID string, recipient string, text string) error {
	return a.m.SendTextMessage(ctx, channelID, recipient, text)
}
