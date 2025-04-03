package activities

import (
	"context"
	"encore-app/bill/database"
	"encore-app/bill/models"
	"fmt"
)

func CreateBill(ctx context.Context, bill models.Bill) (*models.Bill, error) {
	const query = `
		INSERT INTO bills (id, currency, status, total_amount, created_at, closed_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`
	err := database.BillDB.QueryRow(ctx, query, bill.ID, bill.Currency, bill.Status, bill.TotalAmount, bill.CreatedAt, bill.ClosedAt).Scan(&bill.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

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

func CloseBill(ctx context.Context, id string) (*models.Bill, error) {
	query := `
		UPDATE bills
		SET status = 'closed',
		    closed_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, currency, status, total_amount, created_at, closed_at;
	`

	var bill models.Bill
	err := database.BillDB.QueryRow(ctx, query, id).Scan(
		&bill.ID,
		&bill.Currency,
		&bill.Status,
		&bill.TotalAmount,
		&bill.CreatedAt,
		&bill.ClosedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to close bill: %w", err)
	}
	return &bill, nil
}
