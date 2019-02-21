package saga

import (
	"context"
	"errors"
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
	s := NewSaga(context.Background(), "err4", New())

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
	s := NewSaga(context.Background(), "err3", New())

	m := &mock{err: errors.New("hello")}
	comp := &mock{}

	s.AddStep("single", m.f, comp.f)
	s.Play()

	require.Equal(t, m.callCounter, 1)
	require.Equal(t, comp.callCounter, 1)
}

func TestNameErr2(t *testing.T) {
	s := NewSaga(context.Background(), "err2", New())

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
	s := NewSaga(context.Background(), "hello", logStore)

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

func TestNameErr4(t *testing.T) {
	logStore := New()
	s := NewSaga(context.Background(), "hello", logStore)

	callCount1 := 0
	callCount2 := 0

	f1 := func(ctx context.Context) (string, error) {
		callCount1++
		return "hello", errors.New("some error")
	}
	f2 := func(ctx context.Context, s string) error {
		callCount2++
		require.Equal(t, "hello", s)
		return nil
	}

	s.AddStep("first", f1, f2)
	s.Play()

	require.Equal(t, callCount1, 1)
	require.Equal(t, callCount2, 1)

	logs, _ := logStore.GetAllLogsByExecutionID(s.ExecutionID)
	litter.Dump(logs)
}
