package bill

import (
	"context"
	"encore-app/bill/database"
	"encore-app/bill/models"
	"encore-app/bill/workflow"
	"fmt"
	"log"

	temporalClient "encore-app/bill/temporal"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

// add a LIST bill API -> query by status, query by currency

//encore:api public method=POST path=/bills
func CreateBill(ctx context.Context, req *models.CreateBillRequest) (*models.BillResponse, error) {

	if !req.Currency.Validate() {
		return nil, fmt.Errorf("create Bill - invalid currency, must be either USD or GEL")
	}

	billID := uuid.New().String()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "bill-" + billID,
		TaskQueue: "bill-task-queue",
	}

	_, err := temporalClient.Client.ExecuteWorkflow(ctx, workflowOptions, workflow.BillWorkflow, billID, req.Currency)

	if err != nil {
		log.Printf("Failed to start bill workflow: %v", err)
		return nil, fmt.Errorf("could not create bill at this time")
	}

	return &models.BillResponse{
		ID:        billID,
		Currency:  req.Currency,
		Status:    "OPEN",
		LineItems: []models.LineItem{},
	}, nil
}

// AddLineItem adds a line item to an existing bill
//
//encore:api public method=POST path=/bills/:billID/items
func AddLineItem(ctx context.Context, billID string, req *models.AddLineItemRequest) (*models.BillResponse, error) {

	workflowID := "bill-" + billID

	convertedMoney, err := models.NewMoney(req.Amount, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("AddLineItem - failed to convert amount to money: %w", err)
	}

	item := models.LineItem{
		ID:             uuid.New().String(),
		Description:    req.Description,
		OriginalAmount: convertedMoney.Amount,
		BillID:         billID,
		Currency:       req.Currency,
	}

	updateHandle, err := temporalClient.Client.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
		WorkflowID:   workflowID,
		RunID:        "",
		UpdateName:   "add-line-item",
		Args:         []any{item},
		WaitForStage: client.WorkflowUpdateStageAccepted,
	})

	if err != nil {
		return nil, fmt.Errorf("AddLineItem - failed to update workflow: %w", err)
	}

	var updatedBill models.Bill
	err = updateHandle.Get(ctx, &updatedBill)
	if err != nil {
		return nil, fmt.Errorf("AddLineItem - failed to get update Bill: %w", err)
	}

	return convertToResponse(updatedBill), nil
}

// CloseBill closes an open bill
//
//encore:api public method=POST path=/bills/:billID/close
func CloseBill(ctx context.Context, billID string) (*models.BillResponse, error) {
	workflowID := "bill-" + billID

	updateHandle, err := temporalClient.Client.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
		WorkflowID:   workflowID,
		RunID:        "",
		UpdateName:   "close-bill",
		WaitForStage: client.WorkflowUpdateStageAccepted,
	})

	if err != nil {
		return nil, fmt.Errorf("close bill - failed to update workflow: %w", err)
	}

	var updatedBill models.Bill
	err = updateHandle.Get(ctx, &updatedBill)
	if err != nil {
		return nil, fmt.Errorf("close Bill - failed to get update result: %w", err)
	}
	return convertToResponse(updatedBill), nil
}

// GetBill gets a bill by ID
//
//encore:api public method=GET path=/bills/:billID
func GetBill(ctx context.Context, billID string) (*models.BillResponse, error) {

	var bill models.Bill
	err := database.BillDB.QueryRow(ctx, `
	SELECT id, currency, status, total_amount, created_at, closed_at
	FROM bills 
	WHERE id = $1
	`, billID).Scan(
		&bill.ID,
		&bill.Currency,
		&bill.Status,
		&bill.TotalAmount,
		&bill.CreatedAt,
		&bill.ClosedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bill: %w", err)
	}
	return convertToResponse(bill), nil
}

// Helper function to convert workflow Bill to API response
func convertToResponse(b models.Bill) *models.BillResponse {
	return &models.BillResponse{
		ID:          b.ID,
		Currency:    b.Currency,
		Status:      models.BillStatus(b.Status),
		TotalAmount: b.TotalAmount,
		LineItems:   b.LineItems,
	}
}
