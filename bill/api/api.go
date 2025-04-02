package api

import (
    "context"
    "fmt"
    "log"
    "github.com/google/uuid"
    "go.temporal.io/sdk/client"
    "pave-alex/bill"
    "pave-alex/bill/workflow"
    temporalClient "pave-alex/temporal"
)

// add a LIST bill API -> query by status, query by currency

//encore:api public method=POST path=/bills 
func CreateBill(ctx context.Context, req *bill.CreateBillRequest) (*bill.BillResponse, error) {

	if req.Currency != "USD" && req.Currency != "GEL" {
		return nil, fmt.Errorf("Create Bill - invalid currency, must be either USD or GEL")
	}

	billID := uuid.New().String() 

	workflowOptions := client.StartWorkflowOptions {
		ID: "bill-" + billID, 
		TaskQueue: "bill-task-queue",
	}

	_, err := temporalClient.Client.ExecuteWorkflow(ctx, workflowOptions, workflow.BillWorkflow, billID, req.Currency)

    if err != nil {
        log.Printf("Failed to start bill workflow: %v", err)
        return nil, fmt.Errorf("could not create bill at this time")
    }

	return &bill.BillResponse{
        ID:       billID,
        Currency: req.Currency,
        Status:   "OPEN",
        Items:    []bill.LineItem{},
    }, nil
}

// AddLineItem adds a line item to an existing bill
//encore:api public method=POST path=/bills/:billID/items
func AddLineItem(ctx context.Context, billID string, req *bill.AddLineItemRequest) (*bill.BillResponse, error) {

    // WHY do we add the prefix? 
    workflowID := "bill-" + billID

    item := bill.LineItem {
        ID: uuid.New().String(),
        Description: req.Description, 
        Amount: req.Amount, 
    }

        // Call Temporal Update API (sync, gets result)
    updateHandle, err := temporalClient.Client.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
        WorkflowID: workflowID,
        RunID:      "",
        UpdateName: "add-line-item",
        Args:       []interface{}{item},
        WaitForStage: client.WorkflowUpdateStageAccepted, 
    })
    
    if err != nil {
        return nil, fmt.Errorf("AddLineItem - failed to update workflow: %w", err) 
    }

    var updatedBill bill.Bill 
    err = updateHandle.Get(ctx, &updatedBill) 
    if err != nil {
        return nil, fmt.Errorf("AddLineItem - failed to get update result: %w", err) 
    }

    return convertToResponse(updatedBill), nil 
}

// CloseBill closes an open bill
//encore:api public method=POST path=/bills/:billID/close
func CloseBill(ctx context.Context, billID string) (*bill.BillResponse, error) {
    workflowID := "bill-" + billID

    updateHandle, err := temporalClient.Client.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
        WorkflowID: workflowID, 
        RunID: "", 
        UpdateName: "close-bill",
        WaitForStage: client.WorkflowUpdateStageAccepted, 
    })

    if err != nil {
        return nil, fmt.Errorf("Close bill - failed to update workflow: %w", err)
    }

    var updatedBill bill.Bill 
    err = updateHandle.Get(ctx, &updatedBill) 
    if err != nil {
        return nil, fmt.Errorf("Close Bill - failed to get update result: %w", err)
    }
    return convertToResponse(updatedBill), nil 
}

// GetBill gets a bill by ID
//encore:api public method=GET path=/bills/:billID
func GetBill(ctx context.Context, billID string) (*bill.BillResponse, error) {

    workflowID := "bill-" + billID

    // Query workflow for current state
    var bill bill.Bill
    we, err := temporalClient.Client.QueryWorkflow(ctx, workflowID, "", "get-bill-details", &bill)
    if err != nil {
        fmt.Println("Get Bill - Error querying workflow:", err)
        return nil, err
    }

    err = we.Get(&bill)

    if err != nil {
        fmt.Println("Get Bill - Getting workflow error:", err)
        return nil, err
    }

    response := convertToResponse(bill)

    return response, nil
}

// Helper function to convert workflow Bill to API response
func convertToResponse(b bill.Bill) *bill.BillResponse {
    return &bill.BillResponse{
        ID:          b.ID,
        Currency:    b.Currency,
        Status:      b.Status,
        TotalAmount: b.TotalAmount,
        Items:       b.Items,
    }
}