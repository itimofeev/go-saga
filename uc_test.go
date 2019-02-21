package saga

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestName77(t *testing.T) {
	s := budget{}

	require.Empty(t, s.calcBudget())

	s.addTransfer("", "one", 10)
	s.addTransfer("", "two", 10)
	s.addTransfer("one", "two", 10)

	fmt.Println(s.calcBudget())

	sagas := NewSaga(context.Background(), "", New())

	sagas.AddStep("", func(ctx context.Context) (interface{}, error) {
		return s.addTransfer("", "one", 10)
	}, func(ctx context.Context, t1 *Transfer) error {
		fmt.Println("hello!")
		return s.removeTransfer(t1)
	})

	sagas.AddStep("", func(ctx context.Context) (interface{}, error) {
		t2, _ := s.addTransfer("two", "one", 20)
		return t2, errors.New("hello")
	}, func(ctx context.Context, t2 *Transfer) error {
		return s.removeTransfer(t2)
	})

	fmt.Println(s.calcBudget())
}

type Transfer struct {
	ID     string
	From   string
	To     string
	amount int
}

type budget struct {
	s []*Transfer
}

func (s *budget) calcBudget() map[string]int {
	m := make(map[string]int)
	for _, transfer := range s.s {
		fromAmount := m[transfer.From]
		toAmount := m[transfer.To]

		m[transfer.From] = fromAmount - transfer.amount
		m[transfer.To] = toAmount + transfer.amount
	}
	delete(m, "")
	return m
}

func (s *budget) addTransfer(from string, to string, amount int) (*Transfer, error) {
	transfer := &Transfer{
		From:   from,
		To:     to,
		amount: amount,
		ID:     RandString(),
	}
	s.s = append(s.s, transfer)
	return transfer, nil
}

func (s *budget) removeTransfer(transfer *Transfer) error {
	for i, t := range s.s {
		if t.ID == transfer.ID {
			s.s = append(s.s[:i], s.s[i+1:]...)
		}
	}
	return nil
}
