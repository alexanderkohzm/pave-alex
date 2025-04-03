package activities

import (
	"context"
	"database/sql"
	"encore-app/bill/database"
	"encore-app/bill/models"
	"errors"
	"fmt"
	"log"
	"strings"
)

func UpsertLineItem(ctx context.Context, item *models.LineItem) (*models.LineItem, error) {

	var existingBillID string
	err := database.BillDB.QueryRow(ctx, `
	SELECT bill_id from line_items WHERE id = $1
	`, item.ID).Scan(&existingBillID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf(`db error - failed when checking line item's bill id: %w`, err)
	}

	if existingBillID != item.BillID && existingBillID != "" {
		return nil, fmt.Errorf(`db error - line item's bill id does not match the bill id in the request`)
	}

	var existingIdempotencyKey string
	// perform idempotency check - if the idempotency key is already in the database, we don't need to insert it
	err = database.BillDB.QueryRow(ctx, `
	SELECT idempotency_key from line_items WHERE idempotency_key = $1
	`, item.IdempotencyKey).Scan(&existingIdempotencyKey)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf(`db error - failed when checking line item's idempotency key: %w`, err)
	}

	if strings.TrimSpace(existingIdempotencyKey) == strings.TrimSpace(item.IdempotencyKey) {
		log.Printf("Line item already processed: %s", item.ID)
		return nil, fmt.Errorf(`line item has already been processed`)
	}

	transaction, err := database.BillDB.Stdlib().BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()

	// Note: we do NOT let the consumer of our API
	// update the bill_id. This is a conscious decision because
	// it is unusual for a line item to be moved from one bill to another if this is
	// within ONE bill workflow
	// if we really want to move a line item from one bill to another, it should be
	// encapsulated in a service that specializes in that

	// The trade off is that we will be making multiple queries to our DB
	// to find the line item
	query := `
	INSERT INTO line_items (id, bill_id, description, amount, idempotency_key)
	VALUES ($1, $2, $3, $4, $5) 
	ON CONFLICT (id) DO UPDATE 
	SET description = EXCLUDED.description, 
		amount = EXCLUDED.amount
	RETURNING id, bill_id, description, amount, idempotency_key
	`

	var lineItem models.LineItem
	err = transaction.QueryRowContext(ctx, query, item.ID, item.BillID, item.Description, item.Amount, item.IdempotencyKey).
		Scan(
			&lineItem.ID,
			&lineItem.BillID,
			&lineItem.Description,
			&lineItem.Amount,
			&lineItem.IdempotencyKey,
		)

	if err != nil {
		log.Printf("Failed to upsert line item: %v", err)
		return nil, err
	}

	// now we need to update the bill total amount
	updateBillQuery := `
		UPDATE bills 
		SET total_amount = (
			SELECT COALESCE(SUM(amount), 0.0)
			FROM line_items 
			WHERE bill_id = $1
		)
		WHERE id = $1
	`
	_, err = transaction.ExecContext(ctx, updateBillQuery, item.BillID)
	if err != nil {
		log.Printf("Failed to update bill total amount: %v", err)
		return nil, err
	}

	// commit transaction if all passes
	err = transaction.Commit()
	if err != nil {
		return nil, fmt.Errorf(`db error - failed to commit transaction: %w`, err)
	}
	return &lineItem, nil
}
