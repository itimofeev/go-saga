package saga

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
	Err error
}

type Saga struct {
	Name  string
	steps []*Step
}

func (saga *Saga) AddStep(step *Step) {
	// FIXME check that f and compensate are correct and return an error
	saga.steps = append(saga.steps, step)
}
