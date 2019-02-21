package saga

import (
	"context"
	"errors"
	"github.com/itimofeev/go-saga/log"
	llog "log"
	"math/rand"
	"reflect"
	"time"
)

func NewSaga(ctx context.Context, name string, store log.Store) *Saga {
	return &Saga{
		ctx:         ctx,
		Name:        name,
		ExecutionID: RandString(),
		logStore:    store,
	}
}

type Step struct {
	name           string
	number         int
	directFunc     interface{}
	compensateFunc interface{}
}

type Saga struct {
	ExecutionID string
	Name        string

	params       [][]reflect.Value
	toCompensate []reflect.Value
	aborted      bool

	steps []*Step

	ctx context.Context

	logStore log.Store
}

func (saga *Saga) Play() {
	checkErr(saga.logStore.AppendLog(&log.Log{
		ExecutionID: saga.ExecutionID,
		Name:        saga.Name,
		Time:        time.Now(),
		Type:        log.LogTypeStartSaga,
	}))

	for i := 0; i < len(saga.steps); i++ {
		saga.execStep(i)
	}

	checkErr(saga.logStore.AppendLog(&log.Log{
		ExecutionID: saga.ExecutionID,
		Name:        saga.Name,
		Time:        time.Now(),
		Type:        log.LogTypeSagaComplete,
	}))
}

func (saga *Saga) AddStep(name string, f interface{}, compensate interface{}) {
	// FIXME check that f and compensate are correct and return an error
	saga.steps = append(saga.steps, &Step{
		name:           name,
		number:         len(saga.steps),
		directFunc:     f,
		compensateFunc: compensate,
	})
}

func (saga *Saga) abort() {
	stepsToCompensate := len(saga.toCompensate)
	checkErr(saga.logStore.AppendLog(&log.Log{
		ExecutionID: saga.ExecutionID,
		Name:        saga.Name,
		Time:        time.Now(),
		Type:        log.LogTypeSagaAbort,
		StepNumber:  &stepsToCompensate,
	}))

	saga.aborted = true
	for i := stepsToCompensate - 1; i >= 0; i-- {
		saga.compensateStep(i)
	}
}

func (saga *Saga) compensateStep(i int) {
	checkErr(saga.logStore.AppendLog(&log.Log{
		ExecutionID: saga.ExecutionID,
		Name:        saga.Name,
		Time:        time.Now(),
		Type:        log.LogTypeSagaStepCompensate,
		StepNumber:  &i,
		StepName:    &saga.steps[i].name,
	}))

	params := make([]reflect.Value, 0)
	params = append(params, reflect.ValueOf(saga.ctx))
	params = addParams(params, saga.params[i])
	value := saga.toCompensate[i]
	res := value.Call(params)
	if isReturnError(res) {
		panic(res[0])
	}
}

func (saga *Saga) execStep(i int) {
	if saga.aborted {
		return
	}

	checkErr(saga.logStore.AppendLog(&log.Log{
		ExecutionID: saga.ExecutionID,
		Name:        saga.Name,
		Time:        time.Now(),
		Type:        log.LogTypeSagaStepExec,
		StepNumber:  &i,
		StepName:    &saga.steps[i].name,
	}))

	f := saga.steps[i].directFunc
	compensate := saga.steps[i].compensateFunc

	params := []reflect.Value{reflect.ValueOf(saga.ctx)}
	resp := getFuncValue(f).Call(params)

	saga.toCompensate = append(saga.toCompensate, getFuncValue(compensate))
	saga.params = append(saga.params, resp)

	if isReturnError(resp) {
		saga.abort()
	}
}

func isReturnError(result []reflect.Value) bool {
	if len(result) > 0 && !result[len(result)-1].IsNil() {
		return true
	}
	return false
}

func addParams(values []reflect.Value, returned interface{}) []reflect.Value {
	if returned == nil {
		return values
	}

	t := reflect.TypeOf(returned)
	v := reflect.ValueOf(returned)
	if t.Kind() != reflect.Slice {
		values = append(values, reflect.ValueOf(returned))
	} else if v.Len() > 1 { // expect that this is error
		for i := 1; i < v.Len(); i++ {
			values = append(values, v.Index(i))
		}
	}
	return values
}

func getFuncValue(obj interface{}) reflect.Value {
	funcValue := reflect.ValueOf(obj)
	if funcValue.Kind() != reflect.Func {
		checkErr(errors.New("registered object must be a func"))
	}
	if funcValue.Type().NumIn() < 1 ||
		funcValue.Type().In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		checkErr(errors.New("first argument must use context.ctx"))
	}
	return funcValue
}

func checkErr(err error, msg ...string) {
	if err != nil {
		if err != nil {
			llog.Panicln(msg, err)
		}
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
