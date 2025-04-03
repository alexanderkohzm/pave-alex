package models

import (
	"time"
)

type CreateBillRequest struct {
	Currency string // "USD" or "GEL"
}

type BillResponse struct {
	ID          string
	Currency    string
	Status      string
	TotalAmount float64
	Items       []LineItem
}

type Bill struct {
	ID       string
	Currency string // This shouldn't be a string, it should be a CurrencyType
	Status   string
	Items    []LineItem
	// We probably should NOT use float. Need to think about another way
	// Follow up question -> what happens if we need to handle ethereum if it has 18 decimal places
	// use BigNumber or the Decimal Package (this is usually just using BigNumber but just has a nice wrapper)
	// How do we choose to save our TotalAmount? Do we use string or decimal?
	// What are the pros and cons?
	// https://pkg.go.dev/github.com/shopspring/decimal
	// JSON -> converts Integer into a Number type (JSON doesn't actually have Integer)
	// So that's kind of why we want to use string
	// We store it as a DECIMAL on the DB level (btw Decimal is stored as a strong). We basically just want to do calculations for Decimal on the DB level
	TotalAmount float64 // TODO - think about how to represent money. What are the issues?
	CreatedAt   time.Time
	ClosedAt    time.Time
	// https://github.com/bojanz/currency
	// Need to figure out how to perform currency conversions
	// Need to know the trade offs - why do I pass around
	// How do we handle the precision?
	// Minor units --> integer string. Alternative is numeric string
}
