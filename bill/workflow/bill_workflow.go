package workflow 
// if a function is NOT capital case, then this function (AND TYPES) is available locally
// if a function IS capitalized, then we can use it 

import (
	"pave-alex/bill"
	"fmt"
	"go.temporal.io/sdk/workflow"
)

// Bill Workflow to manage the lifecycle of a Bill

func BillWorkflow(ctx workflow.Context, billID string, currency string) error {

	var b bill.Bill
	b = bill.Bill {
		ID: billID, 
		Currency: currency,  // ENUM/CONSTANT 
		Status: "OPEN", // ENUM/CONSTANT 
		Items: []bill.LineItem{},
		CreatedAt: workflow.Now(ctx),
	}

	err := workflow.SetUpdateHandler(ctx, "add-line-item", func(ctx workflow.Context, item bill.LineItem)(bill.Bill, error) {
		if b.Status == "CLOSED" {
			return b, fmt.Errorf("Cannot add line item, bill already closed")
		}

		// let's say we use activity to handle our bill data 
		// what happens if you update it but it fails? 
		// if it fails, then you will need to handle it. If it fails to save, we need
		// to know exactly what to do. 
		
		// maybe  

		b.Items = append(b.Items, item)
		b.TotalAmount += item.Amount 

		// DB = the PERSISTED STATE
		// This bill.items etc is the LIVE state
		// We need to make sure that DB and the LIVE state are eventually consistent 
		// e.g. this is why we would want to use upsert 

		// can persist to DB here with Activity 
		// workflow.ExecuteActivity(ctx, PersistBillToDB, bill)
		return b, nil 
	})

	if err != nil {
		return err 
	}

	err = workflow.SetUpdateHandler(ctx, "close-bill", func(ctx workflow.Context)(bill.Bill, error) {
		if b.Status == "CLOSED" {
			return b, fmt.Errorf("Cannot close bill, bill already closed")
		}
		b.Status = "CLOSED"
		b.ClosedAt = workflow.Now(ctx)
		return b, nil 
	})
	if err != nil {
		return err
	}

	// Register query handler for getting bill details 
	err = workflow.SetQueryHandler(ctx, "get-bill-details", func() (bill.Bill, error) {
		fmt.Println("Getting bill...", b)
		return b, nil 
	})

	if err != nil {
		return err 
	}

	err = workflow.Await(ctx, func() bool {
		return b.Status == "CLOSED"
	})
	if err != nil {
		return err
	}
	return nil

	for b.Status == "OPEN" {
		selector := workflow.NewSelector(ctx) 
		// select. add future -> by default, based on the input, it's just 30 days
		// the bill can close by itself, but you can also early terminate it  
			// 	selector.AddFuture(timerFuture, func(f workflow.Future) {
			// 	if isEnded {
			// 		return
			// 	}
			// 	workflow.GetLogger(ctx).Info("Billing period ended, closing the bill", "BillId", workflowInput.BillId)

			// 	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
			// 	runID := workflow.GetInfo(ctx).WorkflowExecution.RunID

			// 	workflow.SignalExternalWorkflow(ctx, workflowID, runID, activity.CloseBillSignal, activity.CloseBillInput{BillId: workflowInput.BillId})
			// 	isEnded = true
			// })

		// selector.AddReceive(addItemChannel, func(c workflow.ReceiveChannel, more bool) {
		// 	var item LineItem 
		// 	c.Receive(ctx, &item) 

		// 	// Add items to bill 
		// 	bill.Items = append(bill.Items, item) 
		// 	bill.TotalAmount += item.Amount
		// })

		// selector.AddReceive(closeBillChannel, func(c workflow.ReceiveChannel, more bool) {
		// 	c.Receive(ctx, nil) 
		// 	bill.Status = "CLOSED"
		// 	bill.ClosedAt = workflow.Now(ctx)
		// })

		selector.Select(ctx)
	}

	// Should the bill end by itself? 
	// for example, should it close in 30 days after it has been created?
	// This shows how a workflow can make a long timer 

	return nil
}
