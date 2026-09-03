package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"authbackend/generated/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresRepository) CreatePaymentRecord(ctx context.Context, p PaymentInput, workspaceID, subscriptionID string) (PaymentRecord, error) {
	wsUUID, err := parseDBUUID(workspaceID)
	if err != nil {
		return PaymentRecord{}, err
	}
	subUUID, err := parseDBUUID(subscriptionID)
	if err != nil {
		return PaymentRecord{}, err
	}

	paymentID := uuid.New()
	metaJSON, _ := json.Marshal(p.Metadata)

	row, err := sqlc.New(r.pool).CreatePayment(ctx, sqlc.CreatePaymentParams{
		ID:                    pgtype.UUID{Bytes: paymentID, Valid: true},
		WorkspaceID:           wsUUID,
		SubscriptionID:        subUUID,
		Provider:              p.Provider,
		ProviderTransactionID: pgtype.Text{String: p.ProviderTransactionID, Valid: p.ProviderTransactionID != ""},
		PaymentReference:      p.PaymentReference,
		Amount:                p.Amount,
		Currency:              p.Currency,
		Status:                "pending",
		Metadata:              metaJSON,
	})
	if err != nil {
		return PaymentRecord{}, err
	}

	return PaymentRecord{
		ID:                    uuidString(row.ID),
		WorkspaceID:           uuidString(row.WorkspaceID),
		SubscriptionID:        uuidString(row.SubscriptionID),
		PlanID:                p.PlanID,
		Provider:              row.Provider,
		ProviderTransactionID: row.ProviderTransactionID.String,
		PaymentReference:      row.PaymentReference,
		Amount:                row.Amount,
		Currency:              row.Currency,
		Status:                row.Status,
	}, nil
}

func (r *PostgresRepository) ConfirmPayment(ctx context.Context, reference, providerTxID string, meta map[string]interface{}) (PaymentRecord, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return PaymentRecord{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := sqlc.New(tx)
	pay, err := q.GetPaymentByReference(ctx, reference)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PaymentRecord{}, errors.New("payment not found")
		}
		return PaymentRecord{}, err
	}

	// Idempotency: If already paid, return existing record without double-activating or duplicating
	if pay.Status == "paid" {
		return PaymentRecord{
			ID:                    uuidString(pay.ID),
			WorkspaceID:           uuidString(pay.WorkspaceID),
			SubscriptionID:        uuidString(pay.SubscriptionID),
			PlanID:                "chatsolv_starter",
			Provider:              pay.Provider,
			ProviderTransactionID: pay.ProviderTransactionID.String,
			PaymentReference:      pay.PaymentReference,
			Amount:                pay.Amount,
			Currency:              pay.Currency,
			Status:                pay.Status,
		}, nil
	}

	metaJSON, _ := json.Marshal(meta)
	updatedPay, err := q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{
		ID:      pay.ID,
		Status:  "paid",
		Column3: providerTxID,
		Column4: metaJSON,
	})
	if err != nil {
		return PaymentRecord{}, err
	}

	now := time.Now().UTC()
	periodEnd := now.Add(30 * 24 * time.Hour)
	_, err = q.UpdateSubscriptionStatus(ctx, sqlc.UpdateSubscriptionStatusParams{
		WorkspaceID: pay.WorkspaceID,
		Status:      "active",
		Column3:     "chatsolv_starter",
		Column4:     pgtype.Timestamptz{Time: now, Valid: true},
		Column5:     pgtype.Timestamptz{Time: periodEnd, Valid: true},
		Column6:     reference,
	})
	if err != nil {
		return PaymentRecord{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return PaymentRecord{}, err
	}

	return PaymentRecord{
		ID:                    uuidString(updatedPay.ID),
		WorkspaceID:           uuidString(updatedPay.WorkspaceID),
		SubscriptionID:        uuidString(updatedPay.SubscriptionID),
		PlanID:                "chatsolv_starter",
		Provider:              updatedPay.Provider,
		ProviderTransactionID: updatedPay.ProviderTransactionID.String,
		PaymentReference:      updatedPay.PaymentReference,
		Amount:                updatedPay.Amount,
		Currency:              updatedPay.Currency,
		Status:                updatedPay.Status,
	}, nil
}
