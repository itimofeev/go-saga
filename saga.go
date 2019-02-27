package saga

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

func NewSaga(name string) *Saga {
	return &Saga{
		Name: name,
	}
}

type StepOptions struct {
}

type Step struct {
	Name           string
	Func           interface{}
	CompensateFunc interface{}
	Options        *StepOptions
}

type Result struct {
	ExecutionError   error
	CompensateErrors []error
}

type Saga struct {
	Name  string
	steps []*Step
}

func (saga *Saga) AddStep(step *Step) error {
	if err := checkStep(step); err != nil {
		return err
	}
	saga.steps = append(saga.steps, step)
	return nil
}

func checkStep(step *Step) error {
	funcType := reflect.TypeOf(step.Func)
	if funcType.Kind() != reflect.Func {
		return fmt.Errorf("func field is not a func, but %s", funcType.Kind())
	}

	compensateType := reflect.TypeOf(step.CompensateFunc)
	if compensateType.Kind() != reflect.Func {
		return fmt.Errorf("func field is not a func, but %s", compensateType.Kind())
	}

	if funcType.NumIn() == 0 {
		return errors.New("func must have at least one parameter context.Context")
	}
	if funcType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		return errors.New("first parameter of a func must be of type context.Context")
	}
	if funcType.NumOut() == 0 {
		return errors.New("func must have at least one out value of type error")
	}
	if !funcType.Out(funcType.NumOut() - 1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.New("last out parameter of func must be of type error")
	}

	if compensateType.NumIn() == 0 {
		return errors.New("compensate must have at least one parameter context.Context")
	}
	if compensateType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		return errors.New("first parameter of a compensate must be of type context.Context")
	}
	if compensateType.NumOut() != 1 {
		return errors.New("compensate must must return single value of type error")
	}

	return nil
}
