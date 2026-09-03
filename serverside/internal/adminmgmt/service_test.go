package adminmgmt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeAdminRepo struct {
	admins   map[string]WorkspaceAdmin
	assigned string
}

func (f *fakeAdminRepo) Authorize(_ context.Context, userID, workspaceID string) (string, error) {
	if userID == "unauthorized" {
		return "", ErrForbidden
	}
	return "owner", nil
}

func (f *fakeAdminRepo) List(_ context.Context, workspaceID string) ([]WorkspaceAdmin, error) {
	var list []WorkspaceAdmin
	for _, a := range f.admins {
		if a.WorkspaceID == workspaceID {
			list = append(list, a)
		}
	}
	return list, nil
}

func (f *fakeAdminRepo) GetByID(_ context.Context, workspaceID, adminID string) (WorkspaceAdmin, error) {
	if a, ok := f.admins[adminID]; ok && a.WorkspaceID == workspaceID {
		return a, nil
	}
	return WorkspaceAdmin{}, ErrAdminNotFound
}

func (f *fakeAdminRepo) FindByPhone(_ context.Context, phone string) (WorkspaceAdmin, error) {
	for _, a := range f.admins {
		if a.Phone == phone {
			return a, nil
		}
	}
	return WorkspaceAdmin{}, ErrAdminNotFound
}

func (f *fakeAdminRepo) Create(_ context.Context, admin WorkspaceAdmin) (WorkspaceAdmin, error) {
	admin.ID = "admin-1"
	if f.admins == nil {
		f.admins = make(map[string]WorkspaceAdmin)
	}
	f.admins[admin.ID] = admin
	return admin, nil
}

func (f *fakeAdminRepo) Update(_ context.Context, admin WorkspaceAdmin) (WorkspaceAdmin, error) {
	f.admins[admin.ID] = admin
	return admin, nil
}

func (f *fakeAdminRepo) Delete(_ context.Context, workspaceID, adminID string) error {
	delete(f.admins, adminID)
	return nil
}

func (f *fakeAdminRepo) GetNextForRotation(_ context.Context, workspaceID string) (WorkspaceAdmin, error) {
	for _, a := range f.admins {
		if a.WorkspaceID == workspaceID && a.IsActive && a.Status == "online" {
			return a, nil
		}
	}
	return WorkspaceAdmin{}, ErrAdminNotFound
}

func (f *fakeAdminRepo) RecordAssignment(_ context.Context, adminID string) error {
	f.assigned = adminID
	return nil
}

func TestAdminManagementLifecycle(t *testing.T) {
	repo := &fakeAdminRepo{admins: make(map[string]WorkspaceAdmin)}
	service := NewService(repo)

	// Create
	created, err := service.Create(context.Background(), "user-1", "ws-1", WorkspaceAdmin{
		Name:   "Nael",
		Phone:  "08123456789",
		Status: "online",
	})
	require.NoError(t, err)
	require.Equal(t, "admin-1", created.ID)
	require.Equal(t, "628123456789", created.Phone) // Normalized

	// List
	list, err := service.List(context.Background(), "user-1", "ws-1")
	require.NoError(t, err)
	require.Len(t, list, 1)

	// Find by phone
	found, err := service.FindByPhone(context.Background(), "628123456789")
	require.NoError(t, err)
	require.Equal(t, "Nael", found.Name)

	// Rotation
	rot, err := service.GetNextForRotation(context.Background(), "ws-1")
	require.NoError(t, err)
	require.Equal(t, "Nael", rot.Name)

	// Forbidden user
	_, err = service.List(context.Background(), "unauthorized", "ws-1")
	require.ErrorIs(t, err, ErrForbidden)
}
