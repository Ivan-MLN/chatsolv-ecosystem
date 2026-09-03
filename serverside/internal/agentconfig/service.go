package agentconfig

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidInput = errors.New("invalid agent configuration")
	ErrForbidden    = errors.New("agent configuration forbidden")
)

type Personality struct {
	WorkspaceID, AgentID string
	BotName              string   `json:"bot_name"`
	Role                 string   `json:"role"`
	Tone                 string   `json:"tone"`
	CommunicationStyle   string   `json:"communication_style"`
	PrimaryLanguage      string   `json:"primary_language"`
	ResponseLength       string   `json:"response_length"`
	EmojiUsage           string   `json:"emoji_usage"`
	GreetingStyle        string   `json:"greeting_style"`
	ClosingStyle         string   `json:"closing_style"`
	CustomInstructions   string   `json:"custom_instructions"`
	BehaviorRules        []string `json:"behavior_rules"`
	EscalationRules      []string `json:"escalation_rules"`
	ForbiddenTopics      []string `json:"forbidden_topics"`
	FallbackBehavior     string   `json:"fallback_behavior"`
}

type Repository interface {
	Authorize(context.Context, string, string) (string, error)
	SavePersonality(context.Context, Personality) (int64, error)
	GetPersonality(context.Context, string) (Personality, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository} }

func (s *Service) UpdatePersonality(ctx context.Context, userID, agentID string, personality Personality) (int64, error) {
	role, err := s.repository.Authorize(ctx, userID, agentID)
	if err != nil {
		return 0, err
	}
	if role != "owner" && role != "admin" {
		return 0, ErrForbidden
	}
	personality.AgentID = agentID
	if !validPersonality(personality) {
		return 0, ErrInvalidInput
	}
	return s.repository.SavePersonality(ctx, personality)
}

func validPersonality(p Personality) bool {
	return bounded(p.BotName, 1, 100) && bounded(p.Role, 1, 100) &&
		oneOf(p.Tone, "friendly", "professional", "warm", "neutral", "formal", "casual") &&
		oneOf(p.CommunicationStyle, "casual_professional", "formal", "conversational", "concise") &&
		bounded(p.PrimaryLanguage, 2, 10) && oneOf(p.ResponseLength, "short", "medium", "detailed") &&
		oneOf(p.EmojiUsage, "none", "minimal", "moderate") &&
		bounded(p.GreetingStyle, 1, 40) && bounded(p.ClosingStyle, 1, 40) && len(p.CustomInstructions) <= 4000
}
func oneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}
func bounded(value string, min, max int) bool {
	n := len(strings.TrimSpace(value))
	return n >= min && n <= max
}
