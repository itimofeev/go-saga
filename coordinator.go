package saga

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"time"
)

func NewCoordinator(ctx context.Context, saga *Saga, logStore Store) *ExecutionCoordinator {
	return &ExecutionCoordinator{
		ctx:      ctx,
		saga:     saga,
		logStore: logStore,
	}
}

type ExecutionCoordinator struct {
	ExecutionID string

	returnedValuesFromFunc [][]reflect.Value
	toCompensate           []reflect.Value
	aborted                bool
	executionError         error
	compensateErrors       []error

	ctx context.Context

	saga *Saga

	logStore Store
}

func (c *ExecutionCoordinator) Play() *Result {
	checkErr(c.logStore.AppendLog(&Log{
		ExecutionID: c.ExecutionID,
		Name:        c.saga.Name,
		Time:        time.Now(),
		Type:        LogTypeStartSaga,
	}))

	for i := 0; i < len(c.saga.steps); i++ {
		c.execStep(i)
	}

	checkErr(c.logStore.AppendLog(&Log{
		ExecutionID: c.ExecutionID,
		Name:        c.saga.Name,
		Time:        time.Now(),
		Type:        LogTypeSagaComplete,
	}))
	return &Result{ExecutionError: c.executionError, CompensateErrors: c.compensateErrors}
}

func (c *ExecutionCoordinator) execStep(i int) {
	if c.aborted {
		return
	}

	checkErr(c.logStore.AppendLog(&Log{
		ExecutionID: c.ExecutionID,
		Name:        c.saga.Name,
		Time:        time.Now(),
		Type:        LogTypeSagaStepExec,
		StepNumber:  &i,
		StepName:    &c.saga.steps[i].Name,
	}))

	f := c.saga.steps[i].Func
	compensate := c.saga.steps[i].CompensateFunc

	params := []reflect.Value{reflect.ValueOf(c.ctx)}
	resp := getFuncValue(f).Call(params)

	c.toCompensate = append(c.toCompensate, getFuncValue(compensate))
	c.returnedValuesFromFunc = append(c.returnedValuesFromFunc, resp)

	if err := isReturnError(resp); err != nil {
		c.executionError = err
		c.abort()
	}
}

func (c *ExecutionCoordinator) abort() {
	stepsToCompensate := len(c.toCompensate)
	checkErr(c.logStore.AppendLog(&Log{
		ExecutionID: c.ExecutionID,
		Name:        c.saga.Name,
		Time:        time.Now(),
		Type:        LogTypeSagaAbort,
		StepNumber:  &stepsToCompensate,
	}))

	c.aborted = true
	for i := stepsToCompensate - 1; i >= 0; i-- {
		if err := c.compensateStep(i); err != nil {
			c.compensateErrors = append(c.compensateErrors, err)
		}
	}
}

func (c *ExecutionCoordinator) compensateStep(i int) error {
	checkErr(c.logStore.AppendLog(&Log{
		ExecutionID: c.ExecutionID,
		Name:        c.saga.Name,
		Time:        time.Now(),
		Type:        LogTypeSagaStepCompensate,
		StepNumber:  &i,
		StepName:    &c.saga.steps[i].Name,
	}))

	params := make([]reflect.Value, 0)
	params = append(params, reflect.ValueOf(c.ctx))
	params = addParams(params, c.returnedValuesFromFunc[i])
	compensateFunc := c.toCompensate[i]
	res := compensateFunc.Call(params)
	if err := isReturnError(res); err != nil {
		return err
	}
	return nil
}

func addParams(values []reflect.Value, returned []reflect.Value) []reflect.Value {
	if len(returned) > 1 { // expect that this is error
		for i := 0; i < len(returned)-1; i++ {
			values = append(values, returned[i])
		}
	}
	return values
}

func isReturnError(result []reflect.Value) error {
	if len(result) > 0 && !result[len(result)-1].IsNil() {
		return result[len(result)-1].Interface().(error)
	}
	return nil
}

func getFuncValue(obj interface{}) reflect.Value {
	funcValue := reflect.ValueOf(obj)
	checkOK(funcValue.Kind() == reflect.Func, fmt.Sprintf("registered object must be a func but was %s", funcValue.Kind()))

	checkOK(funcValue.Type().NumIn() >= 1 && funcValue.Type().In(0) == reflect.TypeOf((*context.Context)(nil)).Elem(), "invalid func")
	return funcValue
}

func checkErr(err error, msg ...string) {
	if err != nil {
		log.Panicln(msg, err)
	}
}

func checkOK(ok bool, msg ...string) {
	if !ok {
		log.Panicln(msg)
	}
}

// RandString simply generates random string of length n
func RandString() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
