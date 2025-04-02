package activities

import (
	"context"
	"pave-alex/bill"
)

// Note - we can't import db from db.go because infrastructure resources would be used outside the service
// so we're passing it in

func PersistLineItems(ctx context.Context, items []bill.LineItem, billID string, db *sql.DB) error {
	for _, item := range items {
			_, err := db.Exec(ctx, `
					INSERT INTO line_items (id, bill_id, description, amount)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT (id) DO NOTHING
			`, item.ID, billID, item.Description, item.Amount)
			if err != nil {
					return err
			}
	}
	return nil
}