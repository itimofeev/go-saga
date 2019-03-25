package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"time"
)

func NewCoordinator(ctx context.Context, saga *Saga, logStore Store, coordinatorID ...string) *ExecutionCoordinator {
	c := &ExecutionCoordinator{
		ctx:      ctx,
		saga:     saga,
		logStore: logStore,
	}
	if len(coordinatorID) > 0 {
		c.ExecutionID = coordinatorID[0]
	} else {
		c.ExecutionID = RandString()
	}
	return c
}

type ExecutionCoordinator struct {
	ExecutionID string

	aborted          bool
	executionError   error
	compensateErrors []error

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

	f := c.saga.steps[i].Func

	params := []reflect.Value{reflect.ValueOf(c.ctx)}
	resp := getFuncValue(f).Call(params)
	err := isReturnError(resp)

	marshaledResp, marshalErr := marshalResp(resp[:len(resp)-1])
	checkErr(marshalErr)

	stepLog := &Log{
		ExecutionID: c.ExecutionID,
		Name:        c.saga.Name,
		Time:        time.Now(),
		Type:        LogTypeSagaStepExec,
		StepNumber:  &i,
		StepName:    &c.saga.steps[i].Name,
		StepPayload: marshaledResp,
	}

	if err != nil {
		errStr := err.Error()
		stepLog.StepError = &errStr
	}

	checkErr(c.logStore.AppendLog(stepLog))

	if err != nil {
		c.executionError = err
		c.abort()
	}
}

func marshalResp(resp []reflect.Value) ([]byte, error) {
	slice := make([]interface{}, 0, len(resp))
	for _, value := range resp {
		slice = append(slice, value.Interface())
	}

	return json.Marshal(slice)
}

func (c *ExecutionCoordinator) abort() {
	toCompensateLogs, err := c.logStore.GetStepLogsToCompensate(c.ExecutionID)
	checkErr(err, "c.logStore.GetAllLogsByExecutionID(c.ExecutionID)")

	stepsToCompensate := len(toCompensateLogs)
	checkErr(c.logStore.AppendLog(&Log{
		ExecutionID: c.ExecutionID,
		Name:        c.saga.Name,
		Time:        time.Now(),
		Type:        LogTypeSagaAbort,
		StepNumber:  &stepsToCompensate,
	}))

	c.aborted = true
	for i := 0; i < stepsToCompensate; i++ {
		toCompensateLog := toCompensateLogs[i]

		compensateFuncRaw := c.saga.steps[*toCompensateLog.StepNumber].CompensateFunc
		compensateFuncValue := getFuncValue(compensateFuncRaw)
		compensateRuncType := reflect.TypeOf(compensateFuncRaw)

		types := make([]reflect.Type, 0, compensateRuncType.NumIn())
		for i := 1; i < compensateRuncType.NumIn(); i++ {
			types = append(types, compensateRuncType.In(i))
		}
		unmarshal, err := unmarshalParams(types, toCompensateLog.StepPayload)
		checkErr(err, "unmarshalParams()")

		params := make([]reflect.Value, 0)
		params = append(params, reflect.ValueOf(c.ctx))
		params = append(params, unmarshal...)

		if err := c.compensateStep(*toCompensateLog.StepNumber, params, compensateFuncValue); err != nil {
			c.compensateErrors = append(c.compensateErrors, err)
		}
	}
}

func unmarshalParams(types []reflect.Type, payload []byte) ([]reflect.Value, error) {
	rawVals := make([]interface{}, 0, len(types))
	for _, typ := range types {
		rawVals = append(rawVals, reflect.New(typ).Interface())
	}

	checkErr(json.Unmarshal(payload, &rawVals), "json.Unmarshal(payload, &rawVals)")
	res := make([]reflect.Value, 0, len(types))

	for i := 0; i < len(rawVals); i++ {
		objV := reflect.ValueOf(rawVals[i])

		if reflect.TypeOf(rawVals[i]).Kind() == reflect.Ptr && objV.Type() != types[i] {
			objV = objV.Elem()
		}

		res = append(res, objV)
	}
	return res, nil
}

func (c *ExecutionCoordinator) compensateStep(i int, params []reflect.Value, compensateFunc reflect.Value) error {
	checkErr(c.logStore.AppendLog(&Log{
		ExecutionID: c.ExecutionID,
		Name:        c.saga.Name,
		Time:        time.Now(),
		Type:        LogTypeSagaStepCompensate,
		StepNumber:  &i,
		StepName:    &c.saga.steps[i].Name,
	}))

	res := compensateFunc.Call(params)
	if err := isReturnError(res); err != nil {
		return err
	}
	return nil
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
