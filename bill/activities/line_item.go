package activities

import (
	"context"
	"encore-app/bill/database"
	"encore-app/bill/models"
	"log"
)

func SaveLineItem(ctx context.Context, item *models.LineItem) (*models.LineItem, error) {

	query := `
	INSERT INTO line_items (
		id, bill_id, description, amount, original_amount, exchange_rate, currency
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (id) DO UPDATE 
	SET description = EXCLUDED.description, 
	    amount = EXCLUDED.amount,
	    original_amount = EXCLUDED.original_amount,
	    exchange_rate = EXCLUDED.exchange_rate,
	    currency = EXCLUDED.currency
	RETURNING id, bill_id, description, amount, original_amount, exchange_rate, currency
	`

	var lineItem models.LineItem
	err := database.BillDB.QueryRow(
		ctx,
		query,
		item.ID,
		item.BillID,
		item.Description,
		item.Amount,
		item.OriginalAmount,
		item.ExchangeRate,
		item.Currency,
	).Scan(
		&lineItem.ID,
		&lineItem.BillID,
		&lineItem.Description,
		&lineItem.Amount,
		&lineItem.OriginalAmount,
		&lineItem.ExchangeRate,
		&lineItem.Currency,
	)
	if err != nil {
		log.Printf("Failed to upsert line item: %v", err)
		return nil, err
	}

	return &lineItem, nil
}
