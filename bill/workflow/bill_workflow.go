package workflow

// if a function is NOT capital case, then this function (AND TYPES) is available locally
// if a function IS capitalized, then we can use it

import (
	"encore-app/bill/activities"
	"encore-app/bill/models"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.temporal.io/sdk/workflow"
)

// Bill Workflow to manage the lifecycle of a Bill

func BillWorkflow(ctx workflow.Context, billID string, currency models.Currency) error {

	bill := models.Bill{
		ID:        billID,
		Currency:  currency,
		Status:    models.BillStatusOpen,
		LineItems: []models.LineItem{},
		CreatedAt: workflow.Now(ctx),
	}

	var isClosed bool = false

	var billResult *models.Bill

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
	}

	saveBillCtx := workflow.WithActivityOptions(ctx, activityOptions)
	err := workflow.ExecuteActivity(saveBillCtx, activities.SaveBill, bill).Get(ctx, &billResult)
	if err != nil {
		return fmt.Errorf("failed to persist initial bill state: %w", err)
	}

	err = workflow.SetUpdateHandler(ctx, "add-line-item", func(ctx workflow.Context, item models.LineItem) (*models.Bill, error) {

		if bill.Status == models.BillStatusClosed {
			return &bill, fmt.Errorf("cannot add line item, bill already closed")
		}

		// ad-hoc currency conversions
		// TODO: call FOREX API, use another activity
		// For now, just use a map
		item.ExchangeRate = decimal.NewFromInt(1)
		item.Amount = item.OriginalAmount
		if item.Currency != bill.Currency {
			rate, exists := models.ExchangeRates[string(item.Currency)][string(bill.Currency)]
			if !exists {
				return &bill, fmt.Errorf("no exchange rate found for %s to %s", item.Currency, bill.Currency)
			}
			convertedMoney, err := (&models.Money{
				Amount:   item.OriginalAmount,
				Currency: item.Currency,
			}).Convert(rate, bill.Currency)

			if err != nil {
				return &bill, fmt.Errorf("failed to convert currency: %w", err)
			}
			item.Amount = convertedMoney.Amount
			item.ExchangeRate = decimal.NewFromFloat(rate)
		}

		bill.LineItems = append(bill.LineItems, item)
		bill.TotalAmount += item.Amount

		upsertLineItemActivityOptions := workflow.ActivityOptions{
			StartToCloseTimeout: time.Second * 10,
		}
		upsertLineItemCtx := workflow.WithActivityOptions(ctx, upsertLineItemActivityOptions)

		err := workflow.ExecuteActivity(upsertLineItemCtx, activities.SaveLineItem, &item).Get(ctx, &item)
		if err != nil {
			return &bill, fmt.Errorf("failed to upsert line item: %w", err)
		}
		saveBillActivityOptions := workflow.ActivityOptions{
			StartToCloseTimeout: time.Second * 10,
		}
		saveBillCtx := workflow.WithActivityOptions(ctx, saveBillActivityOptions)
		var updatedBill models.Bill
		err = workflow.ExecuteActivity(saveBillCtx, activities.SaveBill, bill).Get(ctx, &updatedBill)
		if err != nil {
			return &bill, fmt.Errorf("failed to save bill: %w", err)
		}

		return &bill, nil
	})

	if err != nil {
		return err
	}

	err = workflow.SetUpdateHandler(ctx, "close-bill", func(ctx workflow.Context) (models.Bill, error) {
		if bill.Status == models.BillStatusClosed {
			return bill, fmt.Errorf("cannot close bill, bill already closed")
		}

		var closedBill models.Bill
		closeBillActivityOptions := workflow.ActivityOptions{
			StartToCloseTimeout: time.Second * 10,
		}
		bill.Status = models.BillStatusClosed
		bill.ClosedAt = workflow.Now(ctx)

		closeBillCtx := workflow.WithActivityOptions(ctx, closeBillActivityOptions)
		err := workflow.ExecuteActivity(closeBillCtx, activities.SaveBill, bill).Get(ctx, &closedBill)

		if err != nil {
			return bill, fmt.Errorf("failed to close bill: %w", err)
		}

		isClosed = true

		return bill, nil
	})

	if err != nil {
		return err
	}

	selector := workflow.NewSelector(ctx)
	var maxBillAge = time.Hour * 24 * 30
	timerFuture := workflow.NewTimer(ctx, maxBillAge)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		bill.Status = models.BillStatusClosed
		bill.ClosedAt = workflow.Now(ctx)
	})

	err = workflow.Await(ctx, func() bool {
		return workflow.AllHandlersFinished(ctx) && isClosed
	})
	if err != nil {
		return err
	}
	return nil
}
