package activities 

import (
	"context"
	"pave-alex/bill"
	"time"
	"database/sql"
)

type BillDB struct {
	DB *sql.DB
}

func NewBillDB(db *sql.DB) *BillDB {
	return &BillDB{DB: db}
}

// Idempotent - we don't care if it's the first or 5th time, we just want to reflect the state in the DB
func (billDb *BillDB)PersistBill(ctx context.Context, b bill.Bill, db *sql.DB,) error {
	_, err := billDb.DB.Exec(ctx, `
			INSERT INTO bills (id, currency, status, total_amount, created_at, closed_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE
			SET currency = EXCLUDED.currency,
					status = EXCLUDED.status,
					total_amount = EXCLUDED.total_amount,
					closed_at = EXCLUDED.closed_at
	`, b.ID, b.Currency, b.Status, b.TotalAmount, b.CreatedAt, nullTime(b.ClosedAt))
	return err
}

func GetBillByID(ctx context.Context, id string, db *sql.DB) (bill.Bill, error) {
	var b bill.Bill 
	err := db.QueryRow(ctx, `
	SELECT id, currency, status, total_amount, created_at, closed_at
	FROM bills 
	wHERE id = $1
	`, id).Scan(
		&b.ID,
		&b.Currency,
		&b.Status,
		&b.TotalAmount,
		&b.CreatedAt, 
		&b.ClosedAt,
	)
	return b, err
}

func nullTime(t time.Time) *time.Time {
	if t.IsZero() {
			return nil
	}
	return &t
}