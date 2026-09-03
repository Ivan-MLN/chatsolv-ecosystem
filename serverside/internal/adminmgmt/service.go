package adminmgmt

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID, workspaceID string) ([]WorkspaceAdmin, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, ErrInvalidInput
	}
	role, err := s.repo.Authorize(ctx, userID, workspaceID)
	if err != nil || role == "" {
		return nil, ErrForbidden
	}
	return s.repo.List(ctx, workspaceID)
}

func (s *Service) Create(ctx context.Context, userID, workspaceID string, in WorkspaceAdmin) (WorkspaceAdmin, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Phone) == "" {
		return WorkspaceAdmin{}, ErrInvalidInput
	}
	role, err := s.repo.Authorize(ctx, userID, workspaceID)
	if err != nil || (role != "owner" && role != "admin") {
		return WorkspaceAdmin{}, ErrForbidden
	}

	normPhone := NormalizePhone(in.Phone)
	if !phoneRegex.MatchString(normPhone) {
		return WorkspaceAdmin{}, ErrInvalidInput
	}

	in.WorkspaceID = workspaceID
	in.Phone = normPhone
	in.IsActive = true
	return s.repo.Create(ctx, in)
}

func (s *Service) CreateAdmin(ctx context.Context, userID, workspaceID, name, phone, role string) error {
	_, err := s.Create(ctx, userID, workspaceID, WorkspaceAdmin{
		Name:  name,
		Phone: phone,
		Role:  role,
	})
	return err
}

func (s *Service) Update(ctx context.Context, userID, workspaceID, adminID string, in WorkspaceAdmin) (WorkspaceAdmin, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(adminID) == "" || strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Phone) == "" {
		return WorkspaceAdmin{}, ErrInvalidInput
	}
	role, err := s.repo.Authorize(ctx, userID, workspaceID)
	if err != nil || (role != "owner" && role != "admin") {
		return WorkspaceAdmin{}, ErrForbidden
	}

	normPhone := NormalizePhone(in.Phone)
	if !phoneRegex.MatchString(normPhone) {
		return WorkspaceAdmin{}, ErrInvalidInput
	}

	in.WorkspaceID = workspaceID
	in.ID = adminID
	in.Phone = normPhone
	return s.repo.Update(ctx, in)
}

func (s *Service) Delete(ctx context.Context, userID, workspaceID, adminID string) error {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(adminID) == "" {
		return ErrInvalidInput
	}
	role, err := s.repo.Authorize(ctx, userID, workspaceID)
	if err != nil || (role != "owner" && role != "admin") {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, workspaceID, adminID)
}

func (s *Service) GetNextForRotation(ctx context.Context, workspaceID string) (WorkspaceAdmin, error) {
	return s.repo.GetNextForRotation(ctx, workspaceID)
}

func (s *Service) FindByPhone(ctx context.Context, phone string) (WorkspaceAdmin, error) {
	return s.repo.FindByPhone(ctx, phone)
}

func (s *Service) RecordAssignment(ctx context.Context, adminID string) error {
	return s.repo.RecordAssignment(ctx, adminID)
}
