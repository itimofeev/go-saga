package saga

import (
	"context"
	"encoding/json"
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
	ExecutionID  string    `gorm:"index;type:varchar(255)"`
	Name         string    `gorm:"index;type:varchar(255)"`
	Type         string    `gorm:"type:varchar(255)"`
	Time         time.Time `gorm:"index"`
	StepNumber   *int
	StepName     *string `gorm:"type:varchar(255)"`
	StepError    *string
	StepPayload  json.RawMessage `gorm:"type:json"`
	StepDuration time.Duration
}

type Store interface {
	AppendLog(ctx context.Context, log *Log) error
	GetAllLogsByExecutionID(ctx context.Context, executionID string) ([]*Log, error)
	GetStepLogsToCompensate(ctx context.Context, executionID string) ([]*Log, error)
}
