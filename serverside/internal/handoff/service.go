package handoff

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"authbackend/internal/adminmgmt"
)

type WhatsAppMessageSender interface {
	SendTextMessage(ctx context.Context, channelID, recipientPhone, text string) error
}

type AdminService interface {
	GetNextForRotation(ctx context.Context, workspaceID string) (adminmgmt.WorkspaceAdmin, error)
	FindByPhone(ctx context.Context, phone string) (adminmgmt.WorkspaceAdmin, error)
	RecordAssignment(ctx context.Context, adminID string) error
}

type Service struct {
	repo         Repository
	adminService AdminService
	waSender     WhatsAppMessageSender
	log          *slog.Logger
}

func NewService(repo Repository, adminService AdminService, waSender WhatsAppMessageSender, log *slog.Logger) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{
		repo:         repo,
		adminService: adminService,
		waSender:     waSender,
		log:          log,
	}
}

type TriggerResult struct {
	Handoff      HandoffRequest `json:"handoff"`
	AdminAssigned bool          `json:"admin_assigned"`
	AdminName    string         `json:"admin_name,omitempty"`
	CustomerNote string         `json:"customer_note"`
}

func (s *Service) TriggerHandoff(ctx context.Context, channelID, workspaceID, conversationID, customerPhone, reason string) (TriggerResult, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(conversationID) == "" {
		return TriggerResult{}, ErrInvalidInput
	}
	if reason == "" {
		reason = "CUSTOMER_REQUEST"
	}

	shortCode := generateShortCode()
	hReq := HandoffRequest{
		ShortCode:      shortCode,
		WorkspaceID:    workspaceID,
		ConversationID: conversationID,
		CustomerPhone:  customerPhone,
		Reason:         reason,
		Status:         "pending",
		RequestedAt:    time.Now(),
		TimeoutAt:      time.Now().Add(2 * time.Minute),
	}

	var assignedAdmin *adminmgmt.WorkspaceAdmin
	if s.adminService != nil {
		admin, err := s.adminService.GetNextForRotation(ctx, workspaceID)
		if err == nil && admin.ID != "" {
			assignedAdmin = &admin
			hReq.AssignedAdminID = &admin.ID
			now := time.Now()
			hReq.AssignedAt = &now
			hReq.Status = "assigned"
		}
	}

	created, err := s.repo.CreateHandoff(ctx, hReq)
	if err != nil {
		return TriggerResult{}, err
	}

	// Update conversation mode to waiting_for_admin
	var adminIDPtr *string
	if assignedAdmin != nil {
		adminIDPtr = &assignedAdmin.ID
	}
	_ = s.repo.SetConversationModeAndAdmin(ctx, conversationID, workspaceID, "waiting_for_admin", adminIDPtr, &created.ID, reason)

	// Record audit event
	_ = s.repo.RecordEvent(ctx, ConversationEvent{
		WorkspaceID:    workspaceID,
		ConversationID: conversationID,
		EventType:      "conversation.handoff.requested",
		ActorType:      "customer",
		Payload: map[string]any{
			"reason":     reason,
			"short_code": shortCode,
		},
	})

	customerNote := "Baik Kak, pesan Kakak sudah kami teruskan ke tim admin kami. Mohon tunggu sebentar ya."

	// If admin assigned, notify admin via WhatsApp
	if assignedAdmin != nil {
		_ = s.adminService.RecordAssignment(ctx, assignedAdmin.ID)
		_ = s.repo.RecordEvent(ctx, ConversationEvent{
			WorkspaceID:    workspaceID,
			ConversationID: conversationID,
			EventType:      "conversation.handoff.assigned",
			ActorType:      "system",
			ActorID:        &assignedAdmin.ID,
			Payload: map[string]any{
				"admin_name":  assignedAdmin.Name,
				"admin_phone": assignedAdmin.Phone,
				"short_code":  shortCode,
			},
		})

		if s.waSender != nil && channelID != "" {
			notifText := fmt.Sprintf(
				"*ChatSolv — Pelanggan Membutuhkan Bantuan*\n\nPelanggan: %s\nTopik: %s\n\nBalas:\n*/acc %s*\n\nuntuk mengambil alih percakapan.",
				customerPhone,
				formatReason(reason),
				shortCode,
			)
			go func() {
				bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if sendErr := s.waSender.SendTextMessage(bgCtx, channelID, assignedAdmin.Phone, notifText); sendErr != nil {
					s.log.Error("failed to notify admin on whatsapp", "error", sendErr, "admin_phone", assignedAdmin.Phone)
				}
			}()
		}
	} else {
		customerNote = "Saat ini semua admin sedang menangani pelanggan lain. Pesan Kakak sudah masuk antrean dan akan segera dibantu oleh admin yang tersedia."
	}

	return TriggerResult{
		Handoff:       created,
		AdminAssigned: assignedAdmin != nil,
		AdminName:     func() string { if assignedAdmin != nil { return assignedAdmin.Name }; return "" }(),
		CustomerNote:  customerNote,
	}, nil
}

func (s *Service) NotifyHandoff(ctx context.Context, channelID, workspaceID, conversationID, customerPhone, reason string) error {
	_, err := s.TriggerHandoff(ctx, channelID, workspaceID, conversationID, customerPhone, reason)
	return err
}

func (s *Service) AcceptByCommand(ctx context.Context, channelID, senderPhone, shortCode string) (string, error) {
	shortCode = strings.ToUpper(strings.TrimSpace(shortCode))
	if shortCode == "" {
		return "Format perintah tidak valid. Gunakan: /acc CS-XXXX", ErrInvalidInput
	}

	normSender := adminmgmt.NormalizePhone(senderPhone)
	admin, err := s.adminService.FindByPhone(ctx, normSender)
	if err != nil || admin.ID == "" {
		return "Nomor WhatsApp Anda tidak terdaftar sebagai admin aktif di ChatSolv.", ErrForbidden
	}

	hReq, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return fmt.Sprintf("Kode permintaan %s tidak ditemukan atau sudah kadaluarsa.", shortCode), ErrHandoffNotFound
	}

	if hReq.WorkspaceID != admin.WorkspaceID {
		return "Anda tidak memiliki akses ke percakapan workspace ini.", ErrForbidden
	}

	if hReq.Status == "accepted" {
		if hReq.AssignedAdminID != nil && *hReq.AssignedAdminID == admin.ID {
			return "Percakapan ini sudah sedang Anda tangani.", nil
		}
		return "Mohon maaf, percakapan ini sudah diambil oleh admin lain.", ErrHandoffAlreadyClaimed
	}

	// Atomic claim to prevent race condition
	accepted, err := s.repo.AcceptAtomic(ctx, hReq.ID, admin.ID)
	if err != nil {
		return "Mohon maaf, percakapan ini baru saja diambil oleh admin lain.", ErrHandoffAlreadyClaimed
	}

	// Set conversation mode to HUMAN
	_ = s.repo.SetConversationModeAndAdmin(ctx, accepted.ConversationID, accepted.WorkspaceID, "human", &admin.ID, &accepted.ID, accepted.Reason)

	// Record audit event
	_ = s.repo.RecordEvent(ctx, ConversationEvent{
		WorkspaceID:    accepted.WorkspaceID,
		ConversationID: accepted.ConversationID,
		EventType:      "conversation.handoff.accepted",
		ActorType:      "admin",
		ActorID:        &admin.ID,
		Payload: map[string]any{
			"admin_name":  admin.Name,
			"admin_phone": admin.Phone,
			"short_code":  accepted.ShortCode,
		},
	})

	// Inform customer that human admin has taken over
	if s.waSender != nil && channelID != "" {
		custMsg := fmt.Sprintf("Halo Kak, saya %s dari tim customer service. Ada yang bisa saya bantu secara langsung?", admin.Name)
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = s.waSender.SendTextMessage(bgCtx, channelID, accepted.CustomerPhone, custMsg)
		}()
	}

	return fmt.Sprintf("Percakapan dengan pelanggan %s berhasil Anda ambil alih.\nKetik /done jika pelayanan sudah selesai.", accepted.CustomerPhone), nil
}

func (s *Service) ResolveByCommand(ctx context.Context, channelID, senderPhone, shortCode string) (string, error) {
	normSender := adminmgmt.NormalizePhone(senderPhone)
	admin, err := s.adminService.FindByPhone(ctx, normSender)
	if err != nil || admin.ID == "" {
		return "Nomor WhatsApp Anda tidak terdaftar sebagai admin di ChatSolv.", ErrForbidden
	}

	var hReq HandoffRequest
	if shortCode != "" {
		shortCode = strings.ToUpper(strings.TrimSpace(shortCode))
		hReq, err = s.repo.GetByShortCode(ctx, shortCode)
	} else {
		// Find latest active handoff for this admin
		list, lErr := s.repo.List(ctx, admin.WorkspaceID, 20)
		if lErr == nil {
			for _, item := range list {
				if item.Status == "accepted" && item.AssignedAdminID != nil && *item.AssignedAdminID == admin.ID {
					hReq = item
					break
				}
			}
		}
	}

	if hReq.ID == "" {
		return "Tidak ada percakapan aktif yang sedang Anda tangani saat ini.", ErrHandoffNotFound
	}

	resolved, err := s.repo.Resolve(ctx, hReq.ID)
	if err != nil {
		return "Gagal menyelesaikan sesi penanganan.", err
	}

	// Return conversation mode to AGENT
	_ = s.repo.SetConversationModeAndAdmin(ctx, resolved.ConversationID, resolved.WorkspaceID, "agent", nil, nil, "")

	// Record audit event
	_ = s.repo.RecordEvent(ctx, ConversationEvent{
		WorkspaceID:    resolved.WorkspaceID,
		ConversationID: resolved.ConversationID,
		EventType:      "conversation.handoff.resolved",
		ActorType:      "admin",
		ActorID:        &admin.ID,
		Payload: map[string]any{
			"admin_name":  admin.Name,
			"short_code":  resolved.ShortCode,
		},
	})
	_ = s.repo.RecordEvent(ctx, ConversationEvent{
		WorkspaceID:    resolved.WorkspaceID,
		ConversationID: resolved.ConversationID,
		EventType:      "conversation.ai.resumed",
		ActorType:      "system",
		Payload:        map[string]any{},
	})

	// Inform customer that AI CS is active again
	if s.waSender != nil && channelID != "" {
		custMsg := "Terima kasih Kak. ChatSolv siap membantu kembali jika ada yang ingin ditanyakan lagi ya 😊"
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = s.waSender.SendTextMessage(bgCtx, channelID, resolved.CustomerPhone, custMsg)
		}()
	}

	return fmt.Sprintf("Sesi penanganan percakapan %s telah diselesaikan. ChatSolv telah kembali aktif.", resolved.CustomerPhone), nil
}

func (s *Service) ManualDashboardTakeover(ctx context.Context, userID, workspaceID, conversationID string) error {
	role, err := s.repo.Authorize(ctx, userID, workspaceID)
	if err != nil || (role != "owner" && role != "admin") {
		return ErrForbidden
	}

	_ = s.repo.SetConversationModeAndAdmin(ctx, conversationID, workspaceID, "human", nil, nil, "MANUAL_DASHBOARD_TAKEOVER")

	_ = s.repo.RecordEvent(ctx, ConversationEvent{
		WorkspaceID:    workspaceID,
		ConversationID: conversationID,
		EventType:      "conversation.handoff.accepted",
		ActorType:      "admin",
		ActorID:        &userID,
		Payload: map[string]any{
			"source": "dashboard_manual",
		},
	})
	return nil
}

func (s *Service) ReturnToAI(ctx context.Context, userID, workspaceID, conversationID string) error {
	role, err := s.repo.Authorize(ctx, userID, workspaceID)
	if err != nil || (role != "owner" && role != "admin") {
		return ErrForbidden
	}

	_ = s.repo.SetConversationModeAndAdmin(ctx, conversationID, workspaceID, "agent", nil, nil, "")

	_ = s.repo.RecordEvent(ctx, ConversationEvent{
		WorkspaceID:    workspaceID,
		ConversationID: conversationID,
		EventType:      "conversation.ai.resumed",
		ActorType:      "admin",
		ActorID:        &userID,
		Payload: map[string]any{
			"source": "dashboard_manual",
		},
	})
	return nil
}

func (s *Service) List(ctx context.Context, userID, workspaceID string, limit int) ([]HandoffRequest, error) {
	role, err := s.repo.Authorize(ctx, userID, workspaceID)
	if err != nil || role == "" {
		return nil, ErrForbidden
	}
	return s.repo.List(ctx, workspaceID, limit)
}

func (s *Service) ListEvents(ctx context.Context, userID, workspaceID, conversationID string) ([]ConversationEvent, error) {
	role, err := s.repo.Authorize(ctx, userID, workspaceID)
	if err != nil || role == "" {
		return nil, ErrForbidden
	}
	return s.repo.ListEvents(ctx, workspaceID, conversationID)
}

func formatReason(reason string) string {
	switch reason {
	case "CUSTOMER_REQUEST":
		return "Permintaan Berbicara dengan Admin"
	case "REFUND":
		return "Permintaan Refund / Pengembalian Dana"
	case "PAYMENT_ISSUE":
		return "Kendala Pembayaran / Bukti Transfer"
	case "SERIOUS_COMPLAINT":
		return "Komplain / Keluhan Pelanggan"
	case "LOW_CONFIDENCE":
		return "Informasi Belum Tersedia di AI"
	default:
		return reason
	}
}
