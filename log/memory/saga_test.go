package memory

import (
	"context"
	"errors"
	"github.com/itimofeev/go-saga"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/require"
	"testing"
)

type mock struct {
	callCounter int
	err         error
}

func (t *mock) f(ctx context.Context) error {
	t.callCounter++
	return t.err
}

func TestName(t *testing.T) {
	s := saga.NewSaga(context.Background(), "err4", New())

	m := &mock{}
	m2 := &mock{}
	comp := &mock{}

	s.AddStep("first", m.f, comp.f)
	s.AddStep("second", m2.f, comp.f)
	s.Play()

	require.Equal(t, m.callCounter, 1)
	require.Equal(t, m2.callCounter, 1)
	require.Equal(t, comp.callCounter, 0)
}

func TestNameErr(t *testing.T) {
	s := saga.NewSaga(context.Background(), "err3", New())

	m := &mock{err: errors.New("hello")}
	comp := &mock{}

	s.AddStep("single", m.f, comp.f)
	s.Play()

	require.Equal(t, m.callCounter, 1)
	require.Equal(t, comp.callCounter, 1)
}

func TestNameErr2(t *testing.T) {
	s := saga.NewSaga(context.Background(), "err2", New())

	m := &mock{}
	comp := &mock{}
	m2 := &mock{err: errors.New("hello")}

	s.AddStep("first", m.f, comp.f)
	s.AddStep("second", m2.f, comp.f)
	s.Play()

	require.Equal(t, m.callCounter, 1)
	require.Equal(t, m2.callCounter, 1)
	require.Equal(t, comp.callCounter, 2)
}

func TestNameErr3(t *testing.T) {
	logStore := New()
	s := saga.NewSaga(context.Background(), "hello", logStore)

	m := &mock{err: errors.New("hello")}
	comp := &mock{}
	m2 := &mock{}

	s.AddStep("first", m.f, comp.f)
	s.AddStep("second", m2.f, comp.f)
	s.Play()

	require.Equal(t, m.callCounter, 1)
	require.Equal(t, m2.callCounter, 0)
	require.Equal(t, comp.callCounter, 1)

	logs, _ := logStore.GetAllLogsByExecutionID(s.ExecutionID)
	litter.Dump(logs)
}
