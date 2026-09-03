package access

import (
	"authbackend/generated/sqlc"
	"authbackend/internal/auth"
	"authbackend/pkg/response"
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSubscriptionRequired = errors.New("active subscription required")
	ErrQuotaExceeded        = errors.New("commercial quota exceeded")
	ErrAgentLimitReached    = errors.New("agent limit reached")
	ErrChannelLimitReached  = errors.New("channel limit reached")
	ErrDocumentLimitReached = errors.New("document limit reached")
	ErrStorageLimitReached  = errors.New("storage limit reached")
	ErrMessageLimitReached  = errors.New("message limit reached")
)

const (
	AccessResultLocalsKey = "chatsolv_access_result"
)

func FromLocals(c *fiber.Ctx) (AccessResult, bool) {
	result, ok := c.Locals(AccessResultLocalsKey).(AccessResult)
	return result, ok
}

type AccessMode string

const (
	AccessModeDeveloper    AccessMode = "developer"
	AccessModeSubscription AccessMode = "subscription"
)

type Entitlements struct {
	MaxAgents            int   `json:"max_agents"`        // -1 for unlimited
	MaxChannels          int   `json:"max_channels"`      // -1 for unlimited
	MaxDocuments         int   `json:"max_documents"`     // -1 for unlimited
	MaxStorageBytes      int64 `json:"max_storage_bytes"` // -1 for unlimited
	MonthlyMessages      int64 `json:"monthly_messages"`  // -1 for unlimited
	PublicAPI            bool  `json:"public_api"`
	Webhooks             bool  `json:"webhooks"`
	SubscriptionRequired bool  `json:"subscription_required"`
	IsUnlimited          bool  `json:"is_unlimited"`
}

type AccessResult struct {
	UserID             string       `json:"user_id"`
	WorkspaceID        string       `json:"workspace_id"`
	PlatformRole       string       `json:"platform_role"`  // "developer" | "user"
	WorkspaceRole      string       `json:"workspace_role"` // "owner" | "admin" | "member" | "viewer"
	AccessMode         AccessMode   `json:"access_mode"`
	SubscriptionStatus string       `json:"subscription_status"` // "active", "inactive", etc.
	HasActiveSub       bool         `json:"has_active_sub"`
	Entitlements       Entitlements `json:"entitlements"`
}

type Resolver struct {
	pool  *pgxpool.Pool
	mu    sync.RWMutex
	cache map[string]cachedAccess
}

type cachedAccess struct {
	result    AccessResult
	expiresAt time.Time
}

func NewResolver(pool *pgxpool.Pool) *Resolver {
	return &Resolver{
		pool:  pool,
		cache: make(map[string]cachedAccess),
	}
}

func (r *Resolver) InvalidateCache(workspaceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k := range r.cache {
		if strings.HasSuffix(k, ":"+workspaceID) {
			delete(r.cache, k)
		}
	}
}

func (r *Resolver) InvalidateUserCache(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	prefix := userID + ":"
	for k := range r.cache {
		if strings.HasPrefix(k, prefix) {
			delete(r.cache, k)
		}
	}
}

func (r *Resolver) Resolve(ctx context.Context, userID, workspaceID string) (AccessResult, error) {
	cacheKey := userID + ":" + workspaceID
	r.mu.RLock()
	cached, ok := r.cache[cacheKey]
	r.mu.RUnlock()
	if ok && time.Now().Before(cached.expiresAt) {
		return cached.result, nil
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return AccessResult{}, errors.New("invalid user id")
	}

	q := sqlc.New(r.pool)
	u, err := q.GetUserByID(ctx, userUUID)
	if err != nil {
		return AccessResult{}, err
	}

	platformRole := u.PlatformRole
	if platformRole == "" {
		platformRole = "user"
	}

	res := AccessResult{
		UserID:       userID,
		WorkspaceID:  workspaceID,
		PlatformRole: platformRole,
	}

	// Platform roles only change commercial entitlements. Every workspace-scoped
	// request must still prove membership before either the developer bypass or
	// subscription checks are evaluated.
	if workspaceID != "" {
		wsUUID, err := parseUUID(workspaceID)
		if err != nil {
			return AccessResult{}, errors.New("invalid workspace id")
		}
		member, err := q.GetWorkspaceForMember(ctx, sqlc.GetWorkspaceForMemberParams{
			ID:     wsUUID,
			UserID: userUUID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return AccessResult{}, errors.New("workspace member not found")
			}
			return AccessResult{}, err
		}
		res.WorkspaceRole = member.Role
	}

	if platformRole == "developer" {
		res.AccessMode = AccessModeDeveloper
		res.SubscriptionStatus = "active"
		res.HasActiveSub = true
		res.Entitlements = Entitlements{
			MaxAgents:            -1,
			MaxChannels:          -1,
			MaxDocuments:         -1,
			MaxStorageBytes:      -1,
			MonthlyMessages:      -1,
			PublicAPI:            true,
			Webhooks:             true,
			SubscriptionRequired: false,
			IsUnlimited:          true,
		}

		r.mu.Lock()
		r.cache[cacheKey] = cachedAccess{result: res, expiresAt: time.Now().Add(1 * time.Minute)}
		r.mu.Unlock()
		return res, nil
	}

	// Normal user
	res.AccessMode = AccessModeSubscription
	res.Entitlements.SubscriptionRequired = true
	res.Entitlements.IsUnlimited = false

	if workspaceID == "" {
		r.mu.Lock()
		r.cache[cacheKey] = cachedAccess{result: res, expiresAt: time.Now().Add(1 * time.Minute)}
		r.mu.Unlock()
		return res, nil
	}

	wsUUID, _ := parseUUID(workspaceID)

	sub, err := q.GetWorkspaceSubscription(ctx, wsUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			res.SubscriptionStatus = "inactive"
			res.HasActiveSub = false
			return res, nil
		}
		return AccessResult{}, err
	}

	res.SubscriptionStatus = sub.Status
	// Subscription is considered active ONLY when status is "active" (or legacy "trialing" if unexpired)
	if sub.Status == "active" {
		res.HasActiveSub = true
	} else if sub.Status == "trialing" && sub.TrialEndsAt.Valid && sub.TrialEndsAt.Time.After(time.Now()) {
		res.HasActiveSub = true
	} else {
		res.HasActiveSub = false
	}

	res.Entitlements = Entitlements{
		MaxAgents:            int(sub.MaxAgents),
		MaxChannels:          int(sub.MaxChannels),
		MaxDocuments:         int(sub.MaxDocuments),
		MaxStorageBytes:      sub.MaxStorageMb * 1024 * 1024,
		MonthlyMessages:      sub.MonthlyMessages,
		PublicAPI:            sub.PublicApi,
		Webhooks:             sub.Webhooks,
		SubscriptionRequired: true,
		IsUnlimited:          false,
	}

	r.mu.Lock()
	r.cache[cacheKey] = cachedAccess{result: res, expiresAt: time.Now().Add(1 * time.Minute)}
	r.mu.Unlock()
	return res, nil
}

// Middleware to enforce active subscription or developer bypass on routes
func (r *Resolver) RequireActiveSubscription() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := auth.AuthenticatedUserID(c)
		if userID == "" {
			return response.Fail(c, fiber.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		}

		workspaceID := c.Query("workspace_id")
		if workspaceID == "" {
			workspaceID = c.Params("workspaceID")
		}
		if workspaceID == "" {
			workspaceID = c.Get("X-Workspace-ID")
		}

		access, err := r.Resolve(c.UserContext(), userID, workspaceID)
		if err != nil {
			if strings.Contains(err.Error(), "member not found") {
				return response.Fail(c, fiber.StatusForbidden, "You do not have access to this workspace", "WORKSPACE_ACCESS_DENIED")
			}
			return response.Fail(c, fiber.StatusInternalServerError, "Failed to resolve access", "INTERNAL_ERROR")
		}

		c.Locals(AccessResultLocalsKey, access)

		if access.PlatformRole == "developer" {
			return c.Next()
		}

		if !access.HasActiveSub {
			return response.Fail(c, fiber.StatusForbidden, "Langganan aktif diperlukan untuk menggunakan fitur ini.", "SUBSCRIPTION_REQUIRED")
		}

		return c.Next()
	}
}

func parseUUID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}
