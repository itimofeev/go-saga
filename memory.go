package saga

import (
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

func (s *store) GetAllLogsByExecutionID(executionID string) ([]*Log, error) {
	res, ok := s.m[executionID]
	if ok {
		return res, nil
	}
	return nil, errors.New("no logs found")
}

func (s *store) AppendLog(log *Log) error {
	s.m[log.ExecutionID] = append(s.m[log.ExecutionID], log)
	return nil
}
