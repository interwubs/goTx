package aonTx

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Operator interface {
	Append(op *ChainOperation)
	Do(op *ChainOperation) error
	ExecuteAll() error
}

type Chain struct {
	RetryOptions

	ops     []*ChainOperation
	async   bool
	retries bool

	lock           sync.Mutex
	completedCount int
	completedErr   error
	errs           error
}

type ChainOperation struct {
	tryFunc     func() error
	secondaryOp *ChainOperation
}

func NewOperation(try UpdateFunc, secondaryOp *ChainOperation) *ChainOperation {
	return &ChainOperation{tryFunc: try, secondaryOp: secondaryOp}
}

func NewChain(async bool) *Chain {
	return &Chain{
		ops:     make([]*ChainOperation, 0),
		async:   async,
		retries: false,
		RetryOptions: RetryOptions{
			MaxRetries: 3,
			Backoff: &ExponentialBackoff{
				InitialInterval: 1 * time.Second,
				MaxInterval:     30 * time.Second,
				Multiplier:      2,
				RandomFactor:    0.2,
			},
		},
	}
}

func (t *Chain) Append(operation *ChainOperation) {
	t.ops = append(t.ops, operation)
}

func (t *Chain) Do(operation *ChainOperation) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.execute(operation)
}

func (t *Chain) execute(operation *ChainOperation) error {
	t.completedErr = nil

	var err error
	if t.retries {
		err = Retry(operation.tryFunc, t.RetryOptions)
	} else {
		err = operation.tryFunc()
	}

	if err != nil {
		if operation.secondaryOp != nil {
			return t.execute(operation.secondaryOp)
		} else {
			t.handleCompletion(err)
			return t.completedErr
		}
	}

	t.handleCompletion(nil)

	if t.completedErr != nil {
		return t.completedErr
	}

	return nil
}

func (t *Chain) ExecuteAll() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.executeAll()
}

func (t *Chain) executeAll() error {
	t.completedCount = 0
	t.completedErr = nil

	for _, o := range t.ops {
		op := o
		if t.async {
			go func() {
				err := t.execute(op)
				if err != nil {
					t.handleCompletion(err)
				}
			}()
		} else {
			err := t.execute(op)
			if err != nil {
				t.handleCompletion(err)
			}
		}
	}

	if t.completedErr != nil {
		t.doSecondary()
		return t.completedErr
	}

	return nil
}

func (t *Chain) handleCompletion(err error) {
	if err != nil {
		t.completedErr = err
		t.errs = errors.Wrap(t.completedErr, err.Error())
	}

	t.completedCount++

	if t.completedErr != nil {
		t.doSecondary()
	}
}

func (t *Chain) handleChildOpCompletion(err error) {
	if err != nil {
		t.completedErr = err
		t.errs = errors.Wrap(t.completedErr, err.Error())
	}

	t.completedCount++

	if t.completedErr != nil {
		t.doSecondary()
	}
}

func (t *Chain) doSecondary() {
	for i := t.completedCount - 1; i >= 0; i-- {
		rollbackOp := t.ops[i]
		err := rollbackOp.tryFunc()
		if err != nil {
			panic(err)
		}
	}
}
