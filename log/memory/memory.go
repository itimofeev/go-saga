package memory

import (
	"errors"
	"github.com/itimofeev/go-saga/log"
)

func New() log.Store {
	return &store{
		m: make(map[string][]*log.Log),
	}
}

type store struct {
	m map[string][]*log.Log
}

func (s *store) GetAllLogsByExecutionID(executionID string) ([]*log.Log, error) {
	res, ok := s.m[executionID]
	if ok {
		return res, nil
	}
	return nil, errors.New("no logs found")
}

func (s *store) AppendLog(log *log.Log) error {
	s.m[log.ExecutionID] = append(s.m[log.ExecutionID], log)
	return nil
}
