package worker

import (
    "context"
    "encore.dev/rlog"
    "go.temporal.io/sdk/worker"
    "encore.app/bill/workflow"
    temporalClient "encore.app/temporal"
)

//encore:service
type Service struct{} 

// Initialize the worker when the package loads
var _ = initWorker()

func initWorker() error {
    svc := &Service{}
    return svc.Start(context.Background())
}

// Starts Temporal Worker 
//encore:api private 
func (s *Service) Start(ctx context.Context) error {
    // Create a worker 
    w := worker.New(temporalClient.Client, "bill-task-queue", worker.Options{})

    w.RegisterWorkflow(workflow.BillWorkflow) 

    err := w.Start() 
    if err != nil {
        rlog.Error("failed to start worker", "err", err) 
        return err 
    }
    rlog.Info("Temporal Worker started")
    return nil 
}