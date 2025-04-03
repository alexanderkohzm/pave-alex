package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type LineItem struct {
	ID             string
	BillID         string
	Description    string
	Amount         int64
	OriginalAmount int64
	ExchangeRate   decimal.Decimal
	Currency       Currency
	CreatedAt      time.Time
}

type AddLineItemRequest struct {
	Description string
	Amount      float64 // assuming frontend passes in a float
	Currency    Currency
}
