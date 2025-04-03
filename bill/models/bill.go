package models

import (
	"time"
)

type BillStatus string

const (
	BillStatusOpen   BillStatus = "OPEN"
	BillStatusClosed BillStatus = "CLOSED"
)

var validStatuses = []BillStatus{BillStatusOpen, BillStatusClosed}

func (s BillStatus) Validate() bool {
	for _, status := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

type CreateBillRequest struct {
	Currency Currency // "USD" or "GEL"
}

type BillResponse struct {
	ID          string
	Currency    Currency
	Status      BillStatus
	TotalAmount int64
	LineItems   []LineItem
}

type Bill struct {
	ID          string
	Currency    Currency
	Status      BillStatus
	LineItems   []LineItem
	TotalAmount int64
	CreatedAt   time.Time
	ClosedAt    time.Time
}
