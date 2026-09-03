package workspace

import "time"

type Status string

const (
	StatusProvisioning Status = "provisioning"
	StatusActive       Status = "active"
	StatusSuspended    Status = "suspended"
	StatusDeleting     Status = "deleting"
	StatusDeleted      Status = "deleted"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

type SubscriptionStatus string

const (
	SubscriptionInactive       SubscriptionStatus = "inactive"
	SubscriptionPendingPayment SubscriptionStatus = "pending_payment"
	SubscriptionActive         SubscriptionStatus = "active"
	SubscriptionPastDue        SubscriptionStatus = "past_due"
	SubscriptionSuspended      SubscriptionStatus = "suspended"
	SubscriptionCancelled      SubscriptionStatus = "cancelled"
	SubscriptionExpired        SubscriptionStatus = "expired"
	SubscriptionTrialing       SubscriptionStatus = "trialing" // legacy mapping
)

type Workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	OwnerUserID string    `json:"-"`
	Status      Status    `json:"status"`
	Timezone    string    `json:"timezone"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Membership struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	Role        Role      `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Subscription struct {
	ID                 string             `json:"id"`
	WorkspaceID        string             `json:"workspace_id"`
	Status             SubscriptionStatus `json:"status"`
	PlanID             string             `json:"plan_id"`
	BillingCycle       string             `json:"billing_cycle"`
	Currency           string             `json:"currency"`
	Amount             int64              `json:"amount"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	PaymentReference   string             `json:"payment_reference,omitempty"`
	CancelAtPeriodEnd  bool               `json:"cancel_at_period_end"`
	TrialStartedAt     *time.Time         `json:"trial_started_at,omitempty"`
	TrialEndsAt        *time.Time         `json:"trial_ends_at,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type Entitlement struct {
	ID              string `json:"id"`
	SubscriptionID  string `json:"subscription_id"`
	WorkspaceID     string `json:"workspace_id"`
	MaxAgents       int    `json:"max_agents"`
	MaxChannels     int    `json:"max_channels"`
	MaxStorageMB    int64  `json:"max_storage_mb"`
	MaxDocuments    int    `json:"max_documents"`
	MonthlyMessages int64  `json:"monthly_messages"`
	PublicAPI       bool   `json:"public_api"`
	Webhooks        bool   `json:"webhooks"`
}

type CreateInput struct {
	Name, Slug, Timezone string
}

type UpdateInput struct {
	Name, Timezone string
}

type CreateResult struct {
	Workspace    Workspace    `json:"workspace"`
	Membership   Membership   `json:"membership"`
	Subscription Subscription `json:"subscription"`
	Entitlement  Entitlement  `json:"entitlements"`
}

type Detail struct {
	Workspace  Workspace  `json:"workspace"`
	Membership Membership `json:"membership"`
}

type AccessModeInfo struct {
	Mode                 string `json:"mode"`
	SubscriptionRequired bool   `json:"subscription_required"`
	Unlimited            bool   `json:"unlimited"`
}

type LimitsInfo struct {
	Agents                any   `json:"agents"`                  // int or "unlimited"
	Channels              any   `json:"channels"`                // int or "unlimited"
	MessagesMonthly       any   `json:"messages_monthly"`        // int64 or "unlimited"
	KnowledgeDocuments    any   `json:"knowledge_documents"`     // int or "unlimited"
	KnowledgeStorageBytes any   `json:"knowledge_storage_bytes"` // int64 or "unlimited"
	PublicAPI             bool  `json:"public_api"`
	Webhooks              bool  `json:"webhooks"`
}

type UsageInfo struct {
	Agents                int64 `json:"agents"`
	Channels              int64 `json:"channels"`
	Messages              int64 `json:"messages"`
	KnowledgeDocuments    int64 `json:"knowledge_documents"`
	KnowledgeStorageBytes int64 `json:"knowledge_storage_bytes"`
}

type SubscriptionDetail struct {
	Subscription Subscription    `json:"subscription"`
	Entitlement  Entitlement     `json:"entitlements"`
	Access       *AccessModeInfo `json:"access,omitempty"`
	Limits       *LimitsInfo     `json:"limits,omitempty"`
	Usage        *UsageInfo      `json:"usage,omitempty"`
}

type AgentSeed struct{ ID, WorkspaceID, Name, Status string }
type BrainSeed struct{ ID, WorkspaceID, AgentID, Status string }
type OutboxSeed struct{ ID, WorkspaceID, Type, AggregateType, AggregateID string }
