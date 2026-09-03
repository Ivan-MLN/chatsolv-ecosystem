package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"strings"
	"time"
)

type Argon2Params struct {
	Memory                uint32
	Iterations            uint32
	Parallelism           uint8
	SaltLength, KeyLength uint32
}

func DefaultArgon2Params() Argon2Params { return Argon2Params{64 * 1024, 3, 2, 16, 32} }

type Argon2Hasher struct{ p Argon2Params }

func NewArgon2Hasher(p Argon2Params) *Argon2Hasher { return &Argon2Hasher{p} }
func (h *Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.p.SaltLength)
	if _, e := rand.Read(salt); e != nil {
		return "", e
	}
	key := argon2.IDKey([]byte(password), salt, h.p.Iterations, h.p.Memory, h.p.Parallelism, h.p.KeyLength)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", h.p.Memory, h.p.Iterations, h.p.Parallelism, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}
func (h *Argon2Hasher) Verify(password, encoded string) (bool, error) {
	var m, t uint32
	var p uint8
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, nil
	}
	if _, e := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); e != nil {
		return false, e
	}
	salt, e := base64.RawStdEncoding.DecodeString(parts[4])
	if e != nil {
		return false, e
	}
	expected, e := base64.RawStdEncoding.DecodeString(parts[5])
	if e != nil {
		return false, e
	}
	actual := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTManager(secret []byte, ttl time.Duration) *JWTManager { return &JWTManager{secret, ttl} }
func (m *JWTManager) Generate(userID string) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(m.ttl)
	c := jwt.RegisteredClaims{Subject: userID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(exp), ID: uuid.NewString()}
	s, e := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
	return s, exp, e
}

func (m *JWTManager) Parse(value string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(value, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected JWT signing method")
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid || claims.Subject == "" {
		return "", fmt.Errorf("invalid access token")
	}
	return claims.Subject, nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func tokenHash(s string) string {
	x := sha256.Sum256([]byte(s))
	return base64.RawURLEncoding.EncodeToString(x[:])
}
