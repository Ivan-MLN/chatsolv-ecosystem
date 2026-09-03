package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Server wires middleware, routes, and the underlying net/http server.
type Server struct {
	http *http.Server
	log  *slog.Logger
}

// New builds a Server. Call Start to begin accepting connections.
func New(addr, secret string, h *Handler, log *slog.Logger) *Server {
	mux := http.NewServeMux()

	// /health is unauthenticated — fast liveness probe.
	mux.HandleFunc("GET /health", h.Health)

	// All /internal/v1/ routes require a valid HMAC signature.
	internal := http.NewServeMux()
	internal.HandleFunc("POST /internal/v1/channels/connect", h.Connect)
	internal.HandleFunc("POST /internal/v1/channels/disconnect", h.Disconnect)
	internal.HandleFunc("GET /internal/v1/channels/status", h.Status)
	internal.HandleFunc("GET /internal/v1/channels/profile", h.Profile)
	internal.HandleFunc("POST /internal/v1/messages/send", h.SendMessage)
	mux.Handle("/internal/v1/", HMAC(secret, time.Now, internal))

	handler := Logger(log, mux)

	return &Server{
		http: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		log: log,
	}
}

// Start begins listening. It blocks until the server stops.
func (s *Server) Start() error {
	s.log.Info("whatsapp service listening", "addr", s.http.Addr)
	return s.http.ListenAndServe()
}

// Shutdown performs a graceful shutdown within the given context deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
