package agentconfig

import (
	"authbackend/generated/sqlc"
	"authbackend/internal/brain/obsidian"
	"authbackend/internal/hermes"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SyncService struct {
	pool     *pgxpool.Pool
	brain    obsidian.SecondBrain
	provider hermes.AgentProvider
}

func NewSyncService(pool *pgxpool.Pool, brain obsidian.SecondBrain, provider hermes.AgentProvider) *SyncService {
	return &SyncService{pool: pool, brain: brain, provider: provider}
}

func (s *SyncService) Sync(ctx context.Context, agentID string) error {
	id, err := syncUUID(agentID)
	if err != nil {
		return err
	}
	q := sqlc.New(s.pool)
	resource, err := q.GetAgentSyncResource(ctx, id)
	if err != nil {
		return err
	}
	personality, err := q.GetPersonality(ctx, id)
	if err != nil {
		return err
	}
	markdown := renderPersonality(Personality{
		BotName: personality.BotName, Role: personality.Role, Tone: personality.Tone,
		CommunicationStyle: personality.CommunicationStyle, PrimaryLanguage: personality.PrimaryLanguage,
		ResponseLength: personality.ResponseLength, EmojiUsage: personality.EmojiUsage,
		GreetingStyle: personality.GreetingStyle, ClosingStyle: personality.ClosingStyle,
		CustomInstructions: personality.CustomInstructions, FallbackBehavior: personality.FallbackBehavior,
	})
	if !resource.VaultKey.Valid || !resource.ProviderAgentID.Valid {
		return fmt.Errorf("agent resources are not ready")
	}
	if err = s.brain.WriteNote(ctx, resource.VaultKey.String, obsidian.Note{Path: "bot/personality.md", Content: markdown}); err != nil {
		return err
	}
	if err = s.provider.UpdateAgent(ctx, resource.ProviderAgentID.String, hermes.UpdateAgentInput{SystemPrompt: markdown}); err != nil {
		return err
	}
	return q.CompleteAgentSync(ctx, id)
}

func renderPersonality(p Personality) string {
	return fmt.Sprintf("---\nbot_name: %q\nrole: %q\ntone: %q\ncommunication_style: %q\nprimary_language: %q\nresponse_length: %q\nemoji_usage: %q\n---\n\n# Agent Personality\n\n## Greeting\n%s\n\n## Closing\n%s\n\n## Custom Instructions\n%s\n\n## Fallback Behavior\n%s\n", p.BotName, p.Role, p.Tone, p.CommunicationStyle, p.PrimaryLanguage, p.ResponseLength, p.EmojiUsage, cleanMarkdown(p.GreetingStyle), cleanMarkdown(p.ClosingStyle), cleanMarkdown(p.CustomInstructions), cleanMarkdown(p.FallbackBehavior))
}

func cleanMarkdown(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "\x00", "")
}

func syncUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}
