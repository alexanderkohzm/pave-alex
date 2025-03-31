package workflow 

import (
	"time"
	"fmt"
	"go.temporal.io/sdk/workflow"
)

type LineItem struct{
	Description string 
	Amount float64 
	ID string 
}

type Bill struct {
    ID          string
    Currency    string
    Status      string
    Items       []LineItem
    TotalAmount float64
    CreatedAt   time.Time
    ClosedAt    time.Time
}

// Bill Workflow to manage the lifecycle of a Bill

func BillWorkflow(ctx workflow.Context, billID string, currency string) error {

	var bill Bill
	// Initialize the bill 
	bill = Bill {
		ID: billID, 
		Currency: currency, 
		Status: "OPEN",
		Items: []LineItem{},
		CreatedAt: workflow.Now(ctx),
	}

	addItemChannel := workflow.GetSignalChannel(ctx, "add-item-signal")
	closeBillChannel := workflow.GetSignalChannel(ctx, "close-bill-signal")

	// Register query handler for getting bill details 
	err := workflow.SetQueryHandler(ctx, "get-bill-details", func() (Bill, error) {
		fmt.Println("Getting bill...", bill)
		return bill, nil 
	})

	if err != nil {
		return err 
	}

	// keep workflow running until bill is closed
	for bill.Status == "OPEN" {
		selector := workflow.NewSelector(ctx) 

		selector.AddReceive(addItemChannel, func(c workflow.ReceiveChannel, more bool) {
			var item LineItem 
			c.Receive(ctx, &item) 

			// Add items to bill 
			bill.Items = append(bill.Items, item) 
			bill.TotalAmount += item.Amount
		})

		selector.AddReceive(closeBillChannel, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil) 
			bill.Status = "CLOSED"
			bill.ClosedAt = workflow.Now(ctx)
		})

		selector.Select(ctx)
	}

	return nil
}
