package api

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "go.temporal.io/sdk/client"
    "pave-alex/bill/workflow"
    temporalClient "pave-alex/temporal"
)


type CreateBillRequest struct {
    Currency string // "USD" or "GEL"
}

type BillResponse struct {
    ID          string
    Currency    string
    Status      string
    TotalAmount float64
    Items       []workflow.LineItem
}

type AddLineItemRequest struct {
    Description string
    Amount      float64
}

//encore:api public method=POST path=/bills 
func CreateBill(ctx context.Context, req *CreateBillRequest) (*BillResponse, error) {


	if req.Currency != "USD" && req.Currency != "GEL" {
		return nil, fmt.Errorf("invalid currency, must be either USD or GEL")
	}

	billID := uuid.New().String() 

	workflowOptions := client.StartWorkflowOptions {
		ID: "bill-" + billID, 
		TaskQueue: "bill-task-queue",
	}

	_, err := temporalClient.Client.ExecuteWorkflow(ctx, workflowOptions, workflow.BillWorkflow, billID, req.Currency)
    if err != nil {
        return nil, err
    }

	return &BillResponse{
        ID:       billID,
        Currency: req.Currency,
        Status:   "OPEN",
        Items:    []workflow.LineItem{},
    }, nil
}

// AddLineItem adds a line item to an existing bill
//encore:api public method=POST path=/bills/:billID/items
func AddLineItem(ctx context.Context, billID string, req *AddLineItemRequest) (*BillResponse, error) {
    workflowID := "bill-" + billID
    
    // Create line item
    item := workflow.LineItem{
        ID:          uuid.New().String(),
        Description: req.Description,
        Amount:      req.Amount,
    }
    
    // Send signal to workflow
    err := temporalClient.Client.SignalWorkflow(ctx, workflowID, "", "add-item-signal", item)
    if err != nil {
        return nil, err
    }
    
    // Query workflow for current state
    var bill workflow.Bill
    we, err := temporalClient.Client.QueryWorkflow(ctx, workflowID, "", "get-bill-details", &bill)
    if err != nil {
        return nil, err
    }

    err = we.Get(&bill)
    if err != nil {
        fmt.Println("[AddLineItem] Getting Bill Detail:", err)
        return nil, err
    }


    
    return convertToResponse(bill), nil
}

// CloseBill closes an open bill
//encore:api public method=POST path=/bills/:billID/close
func CloseBill(ctx context.Context, billID string) (*BillResponse, error) {
    workflowID := "bill-" + billID
    
    // Send close signal to workflow
    err := temporalClient.Client.SignalWorkflow(ctx, workflowID, "", "close-bill-signal", nil)
    if err != nil {
        return nil, err
    }
    
    // Query workflow for current state
    var bill workflow.Bill
    _, err = temporalClient.Client.QueryWorkflow(ctx, workflowID, "", "get-bill-details", &bill)
    if err != nil {
        return nil, err
    }
    
    return convertToResponse(bill), nil
}

// GetBill gets a bill by ID
//encore:api public method=GET path=/bills/:billID
func GetBill(ctx context.Context, billID string) (*BillResponse, error) {

    workflowID := "bill-" + billID

    // Query workflow for current state
    var bill workflow.Bill
    we, err := temporalClient.Client.QueryWorkflow(ctx, workflowID, "", "get-bill-details", &bill)
    if err != nil {
        fmt.Println("[GetBill] Error querying workflow:", err)
        return nil, err
    }

    err = we.Get(&bill)

    if err != nil {
        fmt.Println("[GetBill] Getting workflow:", err)
        return nil, err
    }

    response := convertToResponse(bill)

    return response, nil
}

// Helper function to convert workflow Bill to API response
func convertToResponse(bill workflow.Bill) *BillResponse {
    return &BillResponse{
        ID:          bill.ID,
        Currency:    bill.Currency,
        Status:      bill.Status,
        TotalAmount: bill.TotalAmount,
        Items:       bill.Items,
    }
}