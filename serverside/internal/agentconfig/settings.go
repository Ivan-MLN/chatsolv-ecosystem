package agentconfig

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"
)

type AgentProfile struct {
	WorkspaceID     string `json:"-"`
	AgentID         string `json:"-"`
	DisplayName     string `json:"display_name"`
	AvatarObjectKey string `json:"avatar_object_key,omitempty"`
	Description     string `json:"description"`
	GreetingMessage string `json:"greeting_message"`
	AwayMessage     string `json:"away_message"`
	FallbackMessage string `json:"fallback_message"`
	Language        string `json:"language"`
}

type HandoffRulesConfig struct {
	CustomerRequest bool     `json:"customer_request"`
	LowConfidence   bool     `json:"low_confidence"`
	SeriousComplaint bool    `json:"serious_complaint"`
	Refund          bool     `json:"refund"`
	PaymentIssue    bool     `json:"payment_issue"`
	TimeoutMinutes  int      `json:"timeout_minutes"`
	RotationSystem  string   `json:"rotation_system"`
	CustomTriggers  []string `json:"custom_triggers"`
}

type BusinessProfile struct {
	WorkspaceID         string                 `json:"-"`
	BusinessName        string                 `json:"business_name"`
	Industry            string                 `json:"industry"`
	BusinessType        string                 `json:"business_type,omitempty"`
	BusinessDescription string                 `json:"business_description"`
	TargetCustomer      string                 `json:"target_customer,omitempty"`
	ProductsServices    string                 `json:"products_services,omitempty"`
	CommunicationStyle  string                 `json:"communication_style,omitempty"`
	PrimaryUseCases     []string               `json:"primary_use_cases,omitempty"`
	HandoffRules        *HandoffRulesConfig    `json:"handoff_rules,omitempty"`
	OperatingHours      map[string]interface{} `json:"operating_hours,omitempty"`
	Website             string                 `json:"website,omitempty"`
	Email               string                 `json:"email,omitempty"`
	Phone               string                 `json:"phone,omitempty"`
	Address             string                 `json:"address"`
	BusinessHours       map[string]string      `json:"business_hours,omitempty"`
	Timezone            string                 `json:"timezone"`
	BrandVoice          string                 `json:"brand_voice"`
	CompanyValues       []string               `json:"company_values,omitempty"`
}

type BusinessPolicies struct {
	WorkspaceID     string `json:"-"`
	ShippingPolicy  string `json:"shipping_policy"`
	RefundPolicy    string `json:"refund_policy"`
	ReturnPolicy    string `json:"return_policy"`
	WarrantyPolicy  string `json:"warranty_policy"`
	PaymentPolicy   string `json:"payment_policy"`
	ComplaintPolicy string `json:"complaint_policy"`
}

type SettingsRepository interface {
	AuthorizeWorkspace(context.Context, string, string) (string, error)
	Authorize(context.Context, string, string) (string, error)
	SaveAgentProfile(context.Context, AgentProfile) (int64, error)
	GetAgentProfile(context.Context, string) (AgentProfile, error)
	SaveBusinessProfile(context.Context, BusinessProfile) (int64, error)
	GetBusinessProfile(context.Context, string) (BusinessProfile, error)
	SaveBusinessPolicies(context.Context, BusinessPolicies) (int64, error)
	GetBusinessPolicies(context.Context, string) (BusinessPolicies, error)
}

type SettingsService struct{ repository SettingsRepository }

func NewSettingsService(repository SettingsRepository) *SettingsService {
	return &SettingsService{repository: repository}
}

func (s *SettingsService) UpdateAgentProfile(ctx context.Context, userID, agentID string, profile AgentProfile) (int64, error) {
	role, err := s.repository.Authorize(ctx, userID, agentID)
	if err != nil {
		return 0, err
	}
	if !canWrite(role) {
		return 0, ErrForbidden
	}
	profile.AgentID = agentID
	if !bounded(profile.DisplayName, 1, 100) || !bounded(profile.Language, 2, 10) || len(profile.Description) > 2000 || len(profile.GreetingMessage) > 2000 || len(profile.AwayMessage) > 2000 || len(profile.FallbackMessage) > 2000 {
		return 0, ErrInvalidInput
	}
	return s.repository.SaveAgentProfile(ctx, profile)
}

func (s *SettingsService) GetBusinessProfile(ctx context.Context, workspaceID string) (BusinessProfile, error) {
	return s.repository.GetBusinessProfile(ctx, workspaceID)
}

func (s *SettingsService) UpdateBusinessProfile(ctx context.Context, userID, workspaceID string, profile BusinessProfile) (int64, error) {
	role, err := s.repository.AuthorizeWorkspace(ctx, userID, workspaceID)
	if err != nil {
		return 0, err
	}
	if !canWrite(role) {
		return 0, ErrForbidden
	}
	profile.WorkspaceID = workspaceID
	if !bounded(profile.BusinessName, 1, 160) || !bounded(profile.Industry, 1, 100) || len(profile.BusinessDescription) > 10000 || len(profile.Address) > 4000 || len(profile.BrandVoice) > 4000 || !validTimezone(profile.Timezone) || !validWebsite(profile.Website) {
		return 0, ErrInvalidInput
	}
	return s.repository.SaveBusinessProfile(ctx, profile)
}

func (s *SettingsService) UpdateBusinessPolicies(ctx context.Context, userID, workspaceID string, policies BusinessPolicies) (int64, error) {
	role, err := s.repository.AuthorizeWorkspace(ctx, userID, workspaceID)
	if err != nil {
		return 0, err
	}
	if !canWrite(role) {
		return 0, ErrForbidden
	}
	policies.WorkspaceID = workspaceID
	for _, value := range []string{policies.ShippingPolicy, policies.RefundPolicy, policies.ReturnPolicy, policies.WarrantyPolicy, policies.PaymentPolicy, policies.ComplaintPolicy} {
		if len(value) > 20000 {
			return 0, ErrInvalidInput
		}
	}
	return s.repository.SaveBusinessPolicies(ctx, policies)
}

func canWrite(role string) bool { return role == "owner" || role == "admin" }
func validTimezone(value string) bool {
	if value == "" {
		return true
	}
	_, err := time.LoadLocation(value)
	return strings.TrimSpace(value) != "" && err == nil
}
func validWebsite(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

var _ = errors.Is
