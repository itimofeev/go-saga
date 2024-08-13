package saga

import (
	"context"
	"errors"
)

func New() Store {
	return &store{
		m: make(map[string][]*Log),
	}
}

type store struct {
	m map[string][]*Log
}

func (s *store) GetAllLogsByExecutionID(_ context.Context, executionID string) ([]*Log, error) {
	res, ok := s.m[executionID]
	if ok {
		return res, nil
	}
	return nil, errors.New("no logs found")
}

func (s *store) GetStepLogsToCompensate(_ context.Context, executionID string) ([]*Log, error) {
	logs, ok := s.m[executionID]
	if !ok {
		return nil, errors.New("no logs found")
	}
	var res []*Log
	for i := len(logs) - 1; i >= 0; i-- {
		if logs[i].Type == LogTypeSagaStepExec && logs[i].StepError == nil {
			res = append(res, logs[i])
		}
	}
	return res, nil
}

func (s *store) AppendLog(_ context.Context, log *Log) error {
	s.m[log.ExecutionID] = append(s.m[log.ExecutionID], log)
	return nil
}
