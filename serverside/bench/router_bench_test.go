package bench

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"authbackend/internal/auth"
	"authbackend/pkg/response"
)

const benchSecret = "this-is-a-benchmark-secret-key-that-is-at-least-32-bytes-long"

// ---------------------------------------------------------------------
// 1. Fiber Routing & Serialization Benchmark
// ---------------------------------------------------------------------

func BenchmarkHealthRoute(b *testing.B) {
	app := fiber.New()
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resp, err := app.Test(req, -1)
		if err != nil || resp.StatusCode != 200 {
			b.Fatalf("expected 200, got %v, err %v", resp.StatusCode, err)
		}
	}
}

func BenchmarkEnvelopeResponseSerialization(b *testing.B) {
	app := fiber.New()
	app.Get("/api/v1/dashboard", func(c *fiber.Ctx) error {
		return response.OK(c, fiber.StatusOK, "Dashboard retrieved", fiber.Map{
			"workspace_id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
			"agent": fiber.Map{
				"status": "ready",
			},
			"second_brain": fiber.Map{
				"status":            "ready",
				"knowledge_sources": 5,
			},
			"channel": fiber.Map{
				"status": "connected",
			},
			"conversations": fiber.Map{
				"today": 12,
				"open":  3,
			},
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/dashboard", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resp, err := app.Test(req, -1)
		if err != nil || resp.StatusCode != 200 {
			b.Fatalf("expected 200, got %v, err %v", resp.StatusCode, err)
		}
	}
}

// ---------------------------------------------------------------------
// 2. JWT Generation & Verification Benchmark
// ---------------------------------------------------------------------

func BenchmarkJWTTokenGeneration(b *testing.B) {
	jwtMgr := auth.NewJWTManager([]byte(benchSecret), 15*time.Minute)
	userID := uuid.NewString()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := jwtMgr.Generate(userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTTokenVerification(b *testing.B) {
	jwtMgr := auth.NewJWTManager([]byte(benchSecret), 15*time.Minute)
	userID := uuid.NewString()
	tokenStr, _, err := jwtMgr.Generate(userID)
	if err != nil {
		b.Fatal(err)
	}

	app := fiber.New()
	app.Use(auth.RequireAccessToken(jwtMgr))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resp, err := app.Test(req, -1)
		if err != nil || resp.StatusCode != 200 {
			b.Fatalf("expected 200, got %v, err %v", resp.StatusCode, err)
		}
	}
}

// ---------------------------------------------------------------------
// 3. Argon2id Password Hashing & Verification Benchmark
// ---------------------------------------------------------------------

func BenchmarkArgon2idPasswordHashing(b *testing.B) {
	hasher := auth.NewArgon2Hasher(auth.DefaultArgon2Params())
	password := "SecurePassword123!"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Hash(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkArgon2idPasswordVerify(b *testing.B) {
	hasher := auth.NewArgon2Hasher(auth.DefaultArgon2Params())
	password := "SecurePassword123!"
	encodedHash, err := hasher.Hash(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		match, err := hasher.Verify(password, encodedHash)
		if err != nil || !match {
			b.Fatalf("expected match, got %v err %v", match, err)
		}
	}
}

// ---------------------------------------------------------------------
// 4. HMAC-SHA256 Signature Verification Benchmark (Internal Microservices)
// ---------------------------------------------------------------------

func BenchmarkHMACSignatureVerification(b *testing.B) {
	body := []byte(`{"channel_id":"7326b803-fab7-4305-9070-149e0bdf69a4","external_message_id":"wamid_123","external_user_id":"6281234567890","message_type":"text","content":{"text":"Halo dari benchmark"}}`)
	timestamp := time.Now().UTC().Format(time.RFC3339)

	mac := hmac.New(sha256.New, []byte(benchSecret))
	mac.Write([]byte(timestamp + "." + string(body)))
	signature := hex.EncodeToString(mac.Sum(nil))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		macCheck := hmac.New(sha256.New, []byte(benchSecret))
		macCheck.Write([]byte(timestamp + "." + string(body)))
		expected := hex.EncodeToString(macCheck.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expected)) {
			b.Fatal("signature mismatch")
		}
	}
}

// ---------------------------------------------------------------------
// 5. AES-256-GCM Webhook Secret Encryption / Decryption Benchmark
// ---------------------------------------------------------------------

func BenchmarkAESGCMEncryption(b *testing.B) {
	key := sha256.Sum256([]byte(benchSecret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		b.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		b.Fatal(err)
	}

	plaintext := []byte("whsec_benchmark_secret_key_1234567890")
	nonce := make([]byte, gcm.NonceSize())
	_, _ = io.ReadFull(rand.Reader, nonce)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
		if len(ciphertext) == 0 {
			b.Fatal("encryption failed")
		}
	}
}

func BenchmarkAESGCMDecryption(b *testing.B) {
	key := sha256.Sum256([]byte(benchSecret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		b.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		b.Fatal(err)
	}

	plaintext := []byte("whsec_benchmark_secret_key_1234567890")
	nonce := make([]byte, gcm.NonceSize())
	_, _ = io.ReadFull(rand.Reader, nonce)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		nonceSize := gcm.NonceSize()
		extractedNonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
		decrypted, err := gcm.Open(nil, extractedNonce, actualCiphertext, nil)
		if err != nil || !bytes.Equal(decrypted, plaintext) {
			b.Fatal("decryption failed")
		}
	}
}

// ---------------------------------------------------------------------
// 6. UUID Parsing & Validation Benchmark
// ---------------------------------------------------------------------

func BenchmarkUUIDParsing(b *testing.B) {
	raw := "87687adc-28a9-4412-830b-f6c99e7a9e2d"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		u, err := uuid.Parse(raw)
		if err != nil || u == uuid.Nil {
			b.Fatal(err)
		}
	}
}
