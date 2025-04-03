package workflow

// if a function is NOT capital case, then this function (AND TYPES) is available locally
// if a function IS capitalized, then we can use it

import (
	"encore-app/bill/activities"
	"encore-app/bill/models"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Bill Workflow to manage the lifecycle of a Bill

func BillWorkflow(ctx workflow.Context, billID string, currency string) error {

	bill := models.Bill{
		ID:        billID,
		Currency:  currency, // ENUM/CONSTANT
		Status:    "OPEN",   // ENUM/CONSTANT
		Items:     []models.LineItem{},
		CreatedAt: workflow.Now(ctx),
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
	}

	var billResult *models.Bill

	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	err := workflow.ExecuteActivity(ctx, activities.CreateBill, bill).Get(ctx, &billResult)
	if err != nil {
		return fmt.Errorf("failed to persist initial bill state: %w", err)
	}

	err = workflow.SetUpdateHandler(ctx, "add-line-item", func(ctx workflow.Context, item models.LineItem) (*models.Bill, error) {

		if bill.Status == "CLOSED" {
			return &bill, fmt.Errorf("cannot add line item, bill already closed")
		}
		// Persist the state of the bill by upserting the line item
		var updatedBill models.Bill

		updateBillActivityOptions := workflow.ActivityOptions{
			StartToCloseTimeout: time.Second * 10,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 3},
		}
		updateBillCtx := workflow.WithActivityOptions(ctx, updateBillActivityOptions)
		err := workflow.ExecuteActivity(updateBillCtx, activities.UpsertLineItem, &item).Get(ctx, &updatedBill)
		if err != nil {
			return &bill, fmt.Errorf("failed to upsert line item: %w", err)
		}

		// Update the live state in temporal
		bill.Items = append(bill.Items, item)
		bill.TotalAmount += item.Amount

		// should perform some check to compare live state and persisted state
		// if they are different, log the error
		// Reflect Deep Equal is not a good idea because it will compare pointers
		// potentially write a custom function to compare the two bills

		return &bill, nil
	})

	if err != nil {
		return err
	}

	err = workflow.SetUpdateHandler(ctx, "close-bill", func(ctx workflow.Context) (models.Bill, error) {
		if bill.Status == "CLOSED" {
			return bill, fmt.Errorf("cannot close bill, bill already closed")
		}

		var closedBill models.Bill
		closeBillActivityOptions := workflow.ActivityOptions{
			StartToCloseTimeout: time.Second * 10,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 3},
		}
		closeBillCtx := workflow.WithActivityOptions(ctx, closeBillActivityOptions)
		err := workflow.ExecuteActivity(closeBillCtx, activities.CloseBill, bill.ID).Get(ctx, &closedBill)

		if err != nil {
			return bill, fmt.Errorf("failed to close bill: %w", err)
		}

		bill.Status = "CLOSED"
		bill.ClosedAt = workflow.Now(ctx)
		return bill, nil
	})

	if err != nil {
		return err
	}

	err = workflow.Await(ctx, func() bool {
		return bill.Status == "CLOSED"
	})
	if err != nil {
		return err
	}
	return nil

	// for bill.Status == "OPEN" {
	// 	selector := workflow.NewSelector(ctx)
	// 	// select. add future -> by default, based on the input, it's just 30 days
	// 	// the bill can close by itself, but you can also early terminate it
	// 	// 	selector.AddFuture(timerFuture, func(f workflow.Future) {
	// 	// 	if isEnded {
	// 	// 		return
	// 	// 	}
	// 	// 	workflow.GetLogger(ctx).Info("Billing period ended, closing the bill", "BillId", workflowInput.BillId)

	// 	// 	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	// 	// 	runID := workflow.GetInfo(ctx).WorkflowExecution.RunID

	// 	// 	workflow.SignalExternalWorkflow(ctx, workflowID, runID, activity.CloseBillSignal, activity.CloseBillInput{BillId: workflowInput.BillId})
	// 	// 	isEnded = true
	// 	// })

	// 	// selector.AddReceive(addItemChannel, func(c workflow.ReceiveChannel, more bool) {
	// 	// 	var item LineItem
	// 	// 	c.Receive(ctx, &item)

	// 	// 	// Add items to bill
	// 	// 	bill.Items = append(bill.Items, item)
	// 	// 	bill.TotalAmount += item.Amount
	// 	// })

	// 	// selector.AddReceive(closeBillChannel, func(c workflow.ReceiveChannel, more bool) {
	// 	// 	c.Receive(ctx, nil)
	// 	// 	bill.Status = "CLOSED"
	// 	// 	bill.ClosedAt = workflow.Now(ctx)
	// 	// })

	// 	selector.Select(ctx)
	// }

	// Should the bill end by itself?
	// for example, should it close in 30 days after it has been created?
	// This shows how a workflow can make a long timer

	// return nil
}
