package workspace

import (
	"authbackend/internal/auth"
	"authbackend/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type createPaymentRequest struct {
	PlanID   string `json:"plan_id" validate:"required"`
	Provider string `json:"provider" validate:"omitempty"`
}

type verifyWebhookRequest struct {
	PaymentReference      string                 `json:"payment_reference" validate:"required"`
	ProviderTransactionID string                 `json:"provider_transaction_id" validate:"omitempty"`
	Status                string                 `json:"status" validate:"required"`
	Metadata              map[string]interface{} `json:"metadata"`
}

func (h *Handler) CreateCheckout(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceID")
	if workspaceID == "" {
		workspaceID = c.Query("workspace_id")
	}
	if workspaceID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "workspace_id is required", "VALIDATION_ERROR")
	}

	var request createPaymentRequest
	if err := h.bind(c, &request); err != nil {
		return err
	}

	// Server is authoritative for price and plan
	amount := int64(459000)
	currency := "IDR"
	planID := "chatsolv_starter"
	provider := request.Provider
	if provider == "" {
		provider = "xendit"
	}
	reference := "CS-" + uuid.NewString()[:12]

	subDetail, err := h.service.Subscription(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID)
	if err != nil {
		return mapError(c, err)
	}

	repo, ok := h.service.repository.(PaymentRepository)
	if !ok {
		return response.Fail(c, fiber.StatusInternalServerError, "Payment service unavailable", "INTERNAL_ERROR")
	}

	record, err := repo.CreatePaymentRecord(c.UserContext(), PaymentInput{
		PlanID:                planID,
		Provider:              provider,
		PaymentReference:      reference,
		ProviderTransactionID: "",
		Amount:                amount,
		Currency:              currency,
		Metadata:              map[string]interface{}{"created_by": auth.AuthenticatedUserID(c)},
	}, workspaceID, subDetail.Subscription.ID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to create checkout transaction", "INTERNAL_ERROR")
	}

	return response.OK(c, fiber.StatusOK, "Checkout created", fiber.Map{
		"payment_reference": record.PaymentReference,
		"plan_id":           record.PlanID,
		"amount":            record.Amount,
		"currency":          record.Currency,
		"status":            record.Status,
		"provider":          record.Provider,
		"checkout_url":      "https://checkout.chatsolv.com/pay/" + record.PaymentReference,
	})
}

func (h *Handler) PaymentWebhook(c *fiber.Ctx) error {
	var request verifyWebhookRequest
	if err := h.bind(c, &request); err != nil {
		return err
	}

	if request.Status != "paid" && request.Status != "SUCCESS" {
		return response.OK(c, fiber.StatusOK, "Event ignored", fiber.Map{"status": "ignored"})
	}

	repo, ok := h.service.repository.(PaymentRepository)
	if !ok {
		return response.Fail(c, fiber.StatusInternalServerError, "Payment repository unavailable", "INTERNAL_ERROR")
	}

	record, err := repo.ConfirmPayment(c.UserContext(), request.PaymentReference, request.ProviderTransactionID, request.Metadata)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "PAYMENT_VERIFICATION_FAILED")
	}

	if h.service.resolver != nil {
		h.service.resolver.InvalidateCache(record.WorkspaceID)
	}

	return response.OK(c, fiber.StatusOK, "Payment processed successfully", fiber.Map{
		"payment_reference": record.PaymentReference,
		"status":            record.Status,
		"workspace_id":      record.WorkspaceID,
	})
}
