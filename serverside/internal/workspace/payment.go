package workspace

import (
	"context"
)

type PaymentInput struct {
	PlanID                string                 `json:"plan_id"`
	Provider              string                 `json:"provider"`
	PaymentReference      string                 `json:"payment_reference"`
	ProviderTransactionID string                 `json:"provider_transaction_id"`
	Amount                int64                  `json:"amount"`
	Currency              string                 `json:"currency"`
	Metadata              map[string]interface{} `json:"metadata"`
}

type PaymentRecord struct {
	ID                    string                 `json:"id"`
	WorkspaceID           string                 `json:"workspace_id"`
	SubscriptionID        string                 `json:"subscription_id"`
	PlanID                string                 `json:"plan_id"`
	Provider              string                 `json:"provider"`
	ProviderTransactionID string                 `json:"provider_transaction_id"`
	PaymentReference      string                 `json:"payment_reference"`
	Amount                int64                  `json:"amount"`
	Currency              string                 `json:"currency"`
	Status                string                 `json:"status"`
	Metadata              map[string]interface{} `json:"metadata"`
}

type PaymentRepository interface {
	CreatePaymentRecord(ctx context.Context, p PaymentInput, workspaceID, subscriptionID string) (PaymentRecord, error)
	ConfirmPayment(ctx context.Context, reference, providerTxID string, meta map[string]interface{}) (PaymentRecord, error)
}
