package workspace

import (
	"authbackend/internal/access"
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("workspace not found")
	ErrForbidden    = errors.New("workspace action forbidden")
	ErrInvalidInput = errors.New("invalid workspace input")
	ErrSlugExists   = errors.New("workspace slug already exists")
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Repository interface {
	CreateWithOwnerTrial(context.Context, Workspace, Membership, Subscription, Entitlement, AgentSeed, BrainSeed, OutboxSeed) error
	GetForMember(context.Context, string, string) (Workspace, Membership, error)
	Update(context.Context, Workspace) error
}

type SubscriptionRepository interface {
	GetSubscriptionForMember(context.Context, string, string) (SubscriptionDetail, error)
}

type Service struct {
	repository Repository
	resolver   *access.Resolver
	now        func() time.Time
}

func NewService(repository Repository, resolver *access.Resolver, now func() time.Time) *Service {
	return &Service{repository: repository, resolver: resolver, now: now}
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (CreateResult, error) {
	name := strings.TrimSpace(input.Name)
	slug := normalizeSlug(input.Slug)
	timezone := strings.TrimSpace(input.Timezone)
	if name == "" || len(name) > 120 || len(slug) > 63 || !slugPattern.MatchString(slug) || !validTimezone(timezone) {
		return CreateResult{}, ErrInvalidInput
	}
	now := s.now().UTC()
	workspaceID := uuid.NewString()
	subscriptionID := uuid.NewString()
	w := Workspace{ID: workspaceID, Name: name, Slug: slug, OwnerUserID: userID, Status: StatusProvisioning, Timezone: timezone, CreatedAt: now, UpdatedAt: now}
	m := Membership{ID: uuid.NewString(), WorkspaceID: workspaceID, UserID: userID, Role: RoleOwner, CreatedAt: now, UpdatedAt: now}
	
	// Subscription-first model: status starts as inactive, pending_payment, or active for developer
	initialStatus := SubscriptionInactive
	if s.resolver != nil {
		res, err := s.resolver.Resolve(ctx, userID, "")
		if err == nil && res.PlatformRole == "developer" {
			initialStatus = SubscriptionActive
		}
	}

	subscription := Subscription{
		ID:                 subscriptionID,
		WorkspaceID:        workspaceID,
		Status:             initialStatus,
		PlanID:             "chatsolv_starter",
		BillingCycle:       "monthly",
		Currency:           "IDR",
		Amount:             459000,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.Add(30 * 24 * time.Hour),
		CancelAtPeriodEnd:  false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	entitlement := Entitlement{
		ID:              uuid.NewString(),
		SubscriptionID:  subscriptionID,
		WorkspaceID:     workspaceID,
		MaxAgents:       1,
		MaxChannels:     1,
		MaxStorageMB:    2048,
		MaxDocuments:    200,
		MonthlyMessages: 20_000,
		PublicAPI:       true,
		Webhooks:        true,
	}
	agent := AgentSeed{ID: uuid.NewString(), WorkspaceID: workspaceID, Name: "Default Agent", Status: "pending"}
	brain := BrainSeed{ID: uuid.NewString(), WorkspaceID: workspaceID, AgentID: agent.ID, Status: "pending"}
	event := OutboxSeed{ID: uuid.NewString(), WorkspaceID: workspaceID, Type: "workspace.provision", AggregateType: "workspace", AggregateID: workspaceID}
	if err := s.repository.CreateWithOwnerTrial(ctx, w, m, subscription, entitlement, agent, brain, event); err != nil {
		return CreateResult{}, err
	}
	if s.resolver != nil {
		s.resolver.InvalidateCache(workspaceID)
	}
	return CreateResult{Workspace: w, Membership: m, Subscription: subscription, Entitlement: entitlement}, nil
}

func (s *Service) Get(ctx context.Context, userID, workspaceID string) (Detail, error) {
	workspace, membership, err := s.repository.GetForMember(ctx, workspaceID, userID)
	if err != nil {
		return Detail{}, err
	}
	return Detail{Workspace: workspace, Membership: membership}, nil
}

func (s *Service) Update(ctx context.Context, userID, workspaceID string, input UpdateInput) (Workspace, error) {
	workspace, membership, err := s.repository.GetForMember(ctx, workspaceID, userID)
	if err != nil {
		return Workspace{}, err
	}
	if membership.Role != RoleOwner && membership.Role != RoleAdmin {
		return Workspace{}, ErrForbidden
	}
	if strings.TrimSpace(input.Name) != "" {
		workspace.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.Timezone) != "" {
		workspace.Timezone = strings.TrimSpace(input.Timezone)
	}
	if workspace.Name == "" || len(workspace.Name) > 120 || !validTimezone(workspace.Timezone) {
		return Workspace{}, ErrInvalidInput
	}
	workspace.UpdatedAt = s.now().UTC()
	if err := s.repository.Update(ctx, workspace); err != nil {
		return Workspace{}, err
	}
	return workspace, nil
}

func (s *Service) Subscription(ctx context.Context, userID, workspaceID string) (SubscriptionDetail, error) {
	repository, ok := s.repository.(SubscriptionRepository)
	if !ok {
		return SubscriptionDetail{}, errors.New("subscription repository unavailable")
	}
	subDetail, err := repository.GetSubscriptionForMember(ctx, workspaceID, userID)
	if err != nil {
		return SubscriptionDetail{}, err
	}

	if s.resolver != nil {
		accessRes, err := s.resolver.Resolve(ctx, userID, workspaceID)
		if err == nil {
			if accessRes.PlatformRole == "developer" {
				subDetail.Access = &AccessModeInfo{
					Mode:                 "developer",
					SubscriptionRequired: false,
					Unlimited:            true,
				}
				subDetail.Limits = &LimitsInfo{
					Agents:                "unlimited",
					Channels:              "unlimited",
					MessagesMonthly:       "unlimited",
					KnowledgeDocuments:    "unlimited",
					KnowledgeStorageBytes: "unlimited",
					PublicAPI:             true,
					Webhooks:              true,
				}
			} else {
				subDetail.Access = &AccessModeInfo{
					Mode:                 "subscription",
					SubscriptionRequired: true,
					Unlimited:            false,
				}
				subDetail.Limits = &LimitsInfo{
					Agents:                subDetail.Entitlement.MaxAgents,
					Channels:              subDetail.Entitlement.MaxChannels,
					MessagesMonthly:       subDetail.Entitlement.MonthlyMessages,
					KnowledgeDocuments:    subDetail.Entitlement.MaxDocuments,
					KnowledgeStorageBytes: subDetail.Entitlement.MaxStorageMB * 1024 * 1024,
					PublicAPI:             subDetail.Entitlement.PublicAPI,
					Webhooks:              subDetail.Entitlement.Webhooks,
				}
			}
		}
	}
	return subDetail, nil
}

func normalizeSlug(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), "-")
}

func validTimezone(value string) bool {
	if value == "" {
		return false
	}
	_, err := time.LoadLocation(value)
	return err == nil
}
