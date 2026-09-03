package agentconfig

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	role    string
	saved   Personality
	version int64
}

func (f *fakeRepository) Authorize(context.Context, string, string) (string, error) {
	return f.role, nil
}
func (f *fakeRepository) SavePersonality(_ context.Context, p Personality) (int64, error) {
	f.saved = p
	f.version++
	return f.version, nil
}
func (f *fakeRepository) GetPersonality(context.Context, string) (Personality, error) {
	return f.saved, nil
}
func TestUpdatePersonalityValidatesStructuredConfiguration(t *testing.T) {
	svc := NewService(&fakeRepository{role: "owner"})
	_, err := svc.UpdatePersonality(context.Background(), "u", "a", Personality{BotName: "Naya", Role: "Customer Service", Tone: "hostile", PrimaryLanguage: "id", ResponseLength: "medium", EmojiUsage: "moderate"})
	require.ErrorIs(t, err, ErrInvalidInput)
}
func TestUpdatePersonalityRequiresOwnerOrAdminAndIncrementsVersion(t *testing.T) {
	repo := &fakeRepository{role: "viewer"}
	svc := NewService(repo)
	_, err := svc.UpdatePersonality(context.Background(), "u", "a", validPersonalityFixture())
	require.ErrorIs(t, err, ErrForbidden)
	repo.role = "admin"
	version, err := svc.UpdatePersonality(context.Background(), "u", "a", validPersonalityFixture())
	require.NoError(t, err)
	require.Equal(t, int64(1), version)
	require.Equal(t, "Naya", repo.saved.BotName)
}
func validPersonalityFixture() Personality {
	return Personality{BotName: "Naya", Role: "Customer Service", Tone: "friendly", CommunicationStyle: "casual_professional", PrimaryLanguage: "id", ResponseLength: "medium", EmojiUsage: "moderate", GreetingStyle: "warm", ClosingStyle: "helpful"}
}
