# go-saga [![Build Status](https://travis-ci.org/itimofeev/go-saga.svg?branch=master)](https://travis-ci.org/itimofeev/go-saga) [![codecov](https://codecov.io/gh/itimofeev/go-saga/branch/master/graph/badge.svg)](https://codecov.io/gh/itimofeev/go-saga)
Saga pattern implementation in Go

This library implements Choreography-based saga pattern. This pattern used when you need to deal with distributed transaction.
Often in microservice architecture we need to do some actions in one service, then send request to second service, then send notification via third service.
Saga allows defining compensation functions for each step that will be automatically applied in case of error on any step.

# Installing
```go get github.com/itimofeev/go-saga```

# Getting started

```
func TestExample(t *testing.T) {
    // defines new saga
    s := NewSaga("saga name")
    
    x := 0 // saga will change x by adding 10 than adding 100
    require.NoError(t, s.AddStep(&Step{
        Name:           "1",
        Func:           func(context.Context) error { x += 10; return nil },
        CompensateFunc: func(context.Context) error { x -= 10; return nil },
    }))
    require.NoError(t, s.AddStep(&Step{
        Name:           "2",
        // suppose function in second step returns error
        Func:           func(context.Context) error { x += 100; return errors.New("err") },
        CompensateFunc: func(context.Context) error { x -= 100; return nil },
    }))
    
    store := New()
    c := NewCoordinator(context.Background(), context.Background(), s, store)
    require.Error(t, c.Play().ExecutionError)
    
    // x is still 0, because saga rolled back all applied steps
    require.Equal(t, 0, x)
}
```

# Store
Coordinator stores all sagas executions using `Store` interface.
```
type Store interface {
	AppendLog(log *Log) error
	GetAllLogsByExecutionID(executionID string) ([]*Log, error)
	GetStepLogsToCompensate(executionID string) ([]*Log, error)
}
```
This library implements only in-memory store to eliminate dependencies.
But it's easy to implement this interface using any DB, for example PostgreSQL.
