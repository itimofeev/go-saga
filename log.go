package saga

import (
	"context"
	"time"
)

// noinspection ALL
const (
	LogTypeStartSaga          = "StartSaga"
	LogTypeSagaStepExec       = "SagaStepExec"
	LogTypeSagaAbort          = "SagaAbort"
	LogTypeSagaStepCompensate = "SagaStepCompensate"
	LogTypeSagaComplete       = "SagaComplete"
)

type Log struct {
	ExecutionID  string
	Name         string
	Type         string
	Time         time.Time
	StepNumber   *int
	StepName     *string
	StepError    *string
	StepPayload  []byte
	StepDuration time.Duration
}

type Store interface {
	AppendLog(ctx context.Context, log *Log) error
	GetAllLogsByExecutionID(ctx context.Context, executionID string) ([]*Log, error)
	GetStepLogsToCompensate(ctx context.Context, executionID string) ([]*Log, error)
}
