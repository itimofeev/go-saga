package saga

import "time"

//noinspection ALL
const (
	LogTypeStartSaga          = "StartSaga"
	LogTypeSagaStepExec       = "SagaStepExec"
	LogTypeSagaAbort          = "SagaAbort"
	LogTypeSagaStepCompensate = "SagaStepCompensate"
	LogTypeSagaComplete       = "SagaComplete"
)

type Log struct {
	ExecutionID string
	Name        string
	Type        string
	Time        time.Time
	StepNumber  *int
	StepName    *string
	StepError   *string
	StepPayload []byte
}

type Store interface {
	AppendLog(log *Log) error
	GetAllLogsByExecutionID(executionID string) ([]*Log, error)
	GetStepLogsToCompensate(executionID string) ([]*Log, error)
}
