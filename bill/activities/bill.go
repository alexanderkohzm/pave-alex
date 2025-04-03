package activities

import (
	"context"
	"encore-app/bill/database"
	"encore-app/bill/models"
	"fmt"
)

func GetBillByID(ctx context.Context, id string) (*models.Bill, error) {
	var bill models.Bill
	err := database.BillDB.QueryRow(ctx, `
	SELECT id, currency, status, total_amount, created_at, closed_at
	FROM bills 
	wHERE id = $1
	`, id).Scan(
		&bill.ID,
		&bill.Currency,
		&bill.Status,
		&bill.TotalAmount,
		&bill.CreatedAt,
		&bill.ClosedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bill: %w", err)
	}
	return &bill, nil
}

func SaveBill(ctx context.Context, bill models.Bill) (*models.Bill, error) {
	query := `
		INSERT INTO bills (id, currency, status, total_amount, created_at, closed_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE
		SET
			currency = EXCLUDED.currency,
			status = EXCLUDED.status,
			total_amount = EXCLUDED.total_amount,
			closed_at = EXCLUDED.closed_at
		RETURNING id, currency, status, total_amount, created_at, closed_at;
	`
	var saved models.Bill
	err := database.BillDB.QueryRow(ctx, query,
		bill.ID,
		bill.Currency,
		bill.Status,
		bill.TotalAmount,
		bill.CreatedAt,
		bill.ClosedAt,
	).Scan(
		&saved.ID,
		&saved.Currency,
		&saved.Status,
		&saved.TotalAmount,
		&saved.CreatedAt,
		&saved.ClosedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save bill: %w", err)
	}
	return &saved, nil
}
