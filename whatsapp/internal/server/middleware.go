package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const replayWindow = 5 * time.Minute

// HMAC validates X-ChatSolv-Timestamp and X-ChatSolv-Signature on every
// request. The signature scheme mirrors the backend's InternalHMAC middleware:
//
//	HMAC-SHA256(secret, timestamp + "." + body)
func HMAC(secret string, now func() time.Time, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts := r.Header.Get("X-ChatSolv-Timestamp")
		sig := r.Header.Get("X-ChatSolv-Signature")
		if ts == "" || sig == "" {
			writeError(w, http.StatusUnauthorized, "missing auth headers")
			return
		}

		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid timestamp")
			return
		}
		delta := now().UTC().Sub(t.UTC())
		if delta < -replayWindow || delta > replayWindow {
			writeError(w, http.StatusUnauthorized, "timestamp out of replay window")
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(ts + "." + string(body)))
		expected := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(sig), []byte(expected)) {
			writeError(w, http.StatusUnauthorized, "invalid signature")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logger logs method, path, status code, and latency for every request.
func Logger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"latency_ms", time.Since(start).Milliseconds(),
		)
	})
}

// statusRecorder wraps ResponseWriter to capture the written status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// writeJSON writes code and marshals payload as JSON.
func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError writes a JSON error envelope.
func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"success": false, "message": msg})
}
