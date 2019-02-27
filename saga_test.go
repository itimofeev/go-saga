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
	s := NewSaga("err4")

	m := &mock{}
	m2 := &mock{}
	comp := &mock{}

	s.AddStep(&Step{Name: "first", Func: m.f, CompensateFunc: comp.f})
	s.AddStep(&Step{Name: "second", Func: m2.f, CompensateFunc: comp.f})

	c := NewCoordinator(context.Background(), s, New())
	require.Nil(t, c.Play().Err)

	require.Equal(t, m.callCounter, 1)
	require.Equal(t, m2.callCounter, 1)
	require.Equal(t, comp.callCounter, 0)
}

func TestNameErr(t *testing.T) {
	s := NewSaga("err3")

	m := &mock{err: errors.New("hello")}
	comp := &mock{}

	s.AddStep(&Step{Name: "single", Func: m.f, CompensateFunc: comp.f})

	c := NewCoordinator(context.Background(), s, New())
	require.Error(t, c.Play().Err)

	require.Equal(t, m.callCounter, 1)
	require.Equal(t, comp.callCounter, 1)
}

func TestNameErr2(t *testing.T) {
	s := NewSaga("err2")

	m := &mock{}
	comp := &mock{}
	m2 := &mock{err: errors.New("hello")}

	s.AddStep(&Step{Name: "first", Func: m.f, CompensateFunc: comp.f})
	s.AddStep(&Step{Name: "second", Func: m2.f, CompensateFunc: comp.f})

	c := NewCoordinator(context.Background(), s, New())
	c.Play()

	require.Equal(t, m.callCounter, 1)
	require.Equal(t, m2.callCounter, 1)
	require.Equal(t, comp.callCounter, 2)
}

func TestNameErr3(t *testing.T) {
	logStore := New()
	s := NewSaga("hello")

	m := &mock{err: errors.New("hello")}
	comp := &mock{}
	m2 := &mock{}

	s.AddStep(&Step{Name: "first", Func: m.f, CompensateFunc: comp.f})
	s.AddStep(&Step{Name: "second", Func: m2.f, CompensateFunc: comp.f})

	c := NewCoordinator(context.Background(), s, logStore)
	c.Play()

	require.Equal(t, m.callCounter, 1)
	require.Equal(t, m2.callCounter, 0)
	require.Equal(t, comp.callCounter, 1)

	logs, _ := logStore.GetAllLogsByExecutionID(c.ExecutionID)
	litter.Dump(logs)
}

func TestNameErr4(t *testing.T) {
	logStore := New()
	s := NewSaga("hello")

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

	s.AddStep(&Step{Name: "first", Func: f1, CompensateFunc: f2})

	c := NewCoordinator(context.Background(), s, New())
	err := c.Play()

	require.EqualError(t, err.Err, "some error")
	require.Equal(t, callCount1, 1)
	require.Equal(t, callCount2, 1)

	logs, _ := logStore.GetAllLogsByExecutionID(c.ExecutionID)
	litter.Dump(logs)
}
