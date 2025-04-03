package models

import "time"

type LineItem struct {
	Description    string
	Amount         float64
	ID             string
	BillID         string
	IdempotencyKey string
	CreatedAt      time.Time
}

type AddLineItemRequest struct {
	Description    string
	Amount         float64
	IdempotencyKey string
}
