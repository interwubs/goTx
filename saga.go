package aonTx

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

type (
	UpdateFunc     func() error
	CompensateFunc func() error
)

type Transactor interface {
	Append(txFunc UpdateFunc, rollbackFunc CompensateFunc)
	Execute(txFunc UpdateFunc, rollbackFunc CompensateFunc) error
	ExecuteAll() error
}

type SagaTx struct {
	txFuncs       []UpdateFunc
	rollbackFuncs []CompensateFunc
	async         bool

	retries bool
	RetryOptions

	lock           sync.Mutex
	completedCount int
	completedErr   error
	errs           error
}

func NewTx(async bool) *SagaTx {
	return &SagaTx{
		txFuncs:       make([]UpdateFunc, 0),
		rollbackFuncs: make([]CompensateFunc, 0),
		async:         async,
		retries:       false,
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

func (t *SagaTx) AppendUnrecoverableErrors(errs ...error) {
	if t.UnrecoverableErrors != nil {
		t.UnrecoverableErrors = append(t.UnrecoverableErrors, errs...)
		return
	}

	t.UnrecoverableErrors = errs
}

func (t *SagaTx) AppendFunc(txFunc UpdateFunc, rollbackFunc CompensateFunc) {
	t.txFuncs = append(t.txFuncs, txFunc)
	t.rollbackFuncs = append(t.rollbackFuncs, rollbackFunc)
}

func (t *SagaTx) ExecuteFunc(txFunc UpdateFunc, rollbackFunc CompensateFunc) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.completedErr = nil

	var err error
	if t.retries {
		err = Retry(txFunc, t.RetryOptions)
	} else {
		err = txFunc()
	}

	if err != nil {
		t.AppendFunc(func() error { return nil }, rollbackFunc)
	}
	t.handleCompletion(err)
	if t.completedErr != nil {
		return t.completedErr
	}

	return nil
}

func (t *SagaTx) ExecuteFuncs() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.completedCount = 0
	t.completedErr = nil

	for _, txFunc := range t.txFuncs {
		txF := txFunc
		if t.async {
			go func(txF UpdateFunc) {
				var err error
				if t.retries {
					err = Retry(txF, t.RetryOptions)
				} else {
					err = txF()
				}
				t.handleCompletion(err)
			}(txF)
		} else {
			var err error
			if t.retries {
				err = Retry(txF, t.RetryOptions)
			} else {
				err = txF()
			}
			t.handleCompletion(err)
		}

		if t.completedErr != nil {
			return t.completedErr
		}
	}

	return nil
}

func (t *SagaTx) handleCompletion(err error) {
	if err != nil {
		t.completedErr = err
		t.errs = errors.Wrap(t.completedErr, err.Error())
	}

	t.completedCount++

	if t.completedErr != nil {
		t.rollback()
	}
}

func (t *SagaTx) rollback() {
	for i := t.completedCount - 1; i >= 0; i-- {
		rollbackFunc := t.rollbackFuncs[i]
		err := rollbackFunc()
		if err != nil {
			panic(err)
		}
	}
}
