package handoff

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"authbackend/internal/adminmgmt"
)

type fakeHandoffRepo struct {
	handoffs map[string]HandoffRequest
	events   []ConversationEvent
}

func (f *fakeHandoffRepo) Authorize(_ context.Context, userID, workspaceID string) (string, error) {
	if userID == "unauthorized" {
		return "", ErrForbidden
	}
	return "owner", nil
}

func (f *fakeHandoffRepo) CreateHandoff(_ context.Context, h HandoffRequest) (HandoffRequest, error) {
	h.ID = "h-1"
	if f.handoffs == nil {
		f.handoffs = make(map[string]HandoffRequest)
	}
	f.handoffs[h.ShortCode] = h
	f.handoffs[h.ID] = h
	return h, nil
}

func (f *fakeHandoffRepo) GetByShortCode(_ context.Context, code string) (HandoffRequest, error) {
	if h, ok := f.handoffs[code]; ok {
		return h, nil
	}
	return HandoffRequest{}, ErrHandoffNotFound
}

func (f *fakeHandoffRepo) GetByID(_ context.Context, workspaceID, id string) (HandoffRequest, error) {
	if h, ok := f.handoffs[id]; ok && h.WorkspaceID == workspaceID {
		return h, nil
	}
	return HandoffRequest{}, ErrHandoffNotFound
}

func (f *fakeHandoffRepo) List(_ context.Context, workspaceID string, limit int) ([]HandoffRequest, error) {
	var list []HandoffRequest
	for _, h := range f.handoffs {
		if h.WorkspaceID == workspaceID {
			list = append(list, h)
		}
	}
	return list, nil
}

func (f *fakeHandoffRepo) AcceptAtomic(_ context.Context, handoffID, adminID string) (HandoffRequest, error) {
	if h, ok := f.handoffs[handoffID]; ok {
		if h.Status == "accepted" {
			return HandoffRequest{}, ErrHandoffAlreadyClaimed
		}
		h.Status = "accepted"
		h.AssignedAdminID = &adminID
		now := time.Now()
		h.AcceptedAt = &now
		f.handoffs[handoffID] = h
		f.handoffs[h.ShortCode] = h
		return h, nil
	}
	return HandoffRequest{}, ErrHandoffNotFound
}

func (f *fakeHandoffRepo) Resolve(_ context.Context, handoffID string) (HandoffRequest, error) {
	if h, ok := f.handoffs[handoffID]; ok {
		h.Status = "resolved"
		now := time.Now()
		h.ResolvedAt = &now
		f.handoffs[handoffID] = h
		f.handoffs[h.ShortCode] = h
		return h, nil
	}
	return HandoffRequest{}, ErrHandoffNotFound
}

func (f *fakeHandoffRepo) Reassign(_ context.Context, handoffID, adminID string, timeout time.Time) (HandoffRequest, error) {
	if h, ok := f.handoffs[handoffID]; ok {
		h.AssignedAdminID = &adminID
		h.TimeoutAt = timeout
		return h, nil
	}
	return HandoffRequest{}, ErrHandoffNotFound
}

func (f *fakeHandoffRepo) SetConversationModeAndAdmin(_ context.Context, conversationID, workspaceID, mode string, adminID *string, handoffID *string, reason string) error {
	return nil
}

func (f *fakeHandoffRepo) RecordEvent(_ context.Context, event ConversationEvent) error {
	f.events = append(f.events, event)
	return nil
}

func (f *fakeHandoffRepo) ListEvents(_ context.Context, workspaceID, conversationID string) ([]ConversationEvent, error) {
	return f.events, nil
}

type fakeAdminService struct {
	admin adminmgmt.WorkspaceAdmin
}

func (f *fakeAdminService) GetNextForRotation(_ context.Context, workspaceID string) (adminmgmt.WorkspaceAdmin, error) {
	return f.admin, nil
}

func (f *fakeAdminService) FindByPhone(_ context.Context, phone string) (adminmgmt.WorkspaceAdmin, error) {
	if f.admin.Phone == phone {
		return f.admin, nil
	}
	return adminmgmt.WorkspaceAdmin{}, adminmgmt.ErrAdminNotFound
}

func (f *fakeAdminService) RecordAssignment(_ context.Context, adminID string) error {
	return nil
}

type fakeWASender struct {
	sent []string
}

func (f *fakeWASender) SendTextMessage(_ context.Context, channelID, recipientPhone, text string) error {
	f.sent = append(f.sent, recipientPhone+": "+text)
	return nil
}

func TestHandoffTriggerAcceptAndResolve(t *testing.T) {
	repo := &fakeHandoffRepo{handoffs: make(map[string]HandoffRequest)}
	adminSvc := &fakeAdminService{
		admin: adminmgmt.WorkspaceAdmin{
			ID:          "admin-1",
			WorkspaceID: "ws-1",
			Name:        "Nael",
			Phone:       "628123456789",
		},
	}
	wa := &fakeWASender{}
	service := NewService(repo, adminSvc, wa, nil)

	// 1. Customer Triggers Handoff
	res, err := service.TriggerHandoff(context.Background(), "chan-1", "ws-1", "conv-1", "62899999999", "CUSTOMER_REQUEST")
	require.NoError(t, err)
	require.True(t, res.AdminAssigned)
	require.Equal(t, "Nael", res.AdminName)
	shortCode := res.Handoff.ShortCode
	require.NotEmpty(t, shortCode)

	// 2. Admin 1 accepts via /acc CS-XXXX
	reply, err := service.AcceptByCommand(context.Background(), "chan-1", "628123456789", shortCode)
	require.NoError(t, err)
	require.Contains(t, reply, "berhasil Anda ambil alih")

	// 3. Second Admin attempts /acc -> should be rejected atomically
	reply2, err := service.AcceptByCommand(context.Background(), "chan-1", "628999111222", shortCode)
	require.Error(t, err)
	require.Contains(t, reply2, "tidak terdaftar")

	// 4. Admin finishes session via /done
	doneReply, err := service.ResolveByCommand(context.Background(), "chan-1", "628123456789", shortCode)
	require.NoError(t, err)
	require.Contains(t, doneReply, "telah diselesaikan")
}
