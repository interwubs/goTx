package aonTx

import (
	"sync"
	"time"
)

type (
	TxFunc       func() error
	RollbackFunc func() error
)

type Transaction interface {
	AddFunc(txFunc TxFunc, rollbackFunc RollbackFunc)
	ExecuteFunc(txFunc TxFunc, rollbackFunc RollbackFunc) error
	Execute() error
}

type Tx struct {
	txFuncs       []TxFunc
	rollbackFuncs []RollbackFunc
	async         bool

	retries bool
	RetryOptions

	lock           sync.Mutex
	completedCount int
	completedErr   error
}

func NewTx(async bool) *Tx {
	return &Tx{
		txFuncs:       make([]TxFunc, 0),
		rollbackFuncs: make([]RollbackFunc, 0),
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

func (t *Tx) AppendUnrecoverableErrors(errs ...error) {
	if t.UnrecoverableErrors != nil {
		t.UnrecoverableErrors = append(t.UnrecoverableErrors, errs...)
		return
	}

	t.UnrecoverableErrors = errs
}

func (t *Tx) AppendFunc(txFunc TxFunc, rollbackFunc RollbackFunc) {
	t.txFuncs = append(t.txFuncs, txFunc)
	t.rollbackFuncs = append(t.rollbackFuncs, rollbackFunc)
}

func (t *Tx) ExecuteFunc(txFunc TxFunc, rollbackFunc RollbackFunc) error {
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

func (t *Tx) Execute() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.completedCount = 0
	t.completedErr = nil

	for _, txFunc := range t.txFuncs {
		txF := txFunc
		if t.async {
			go func(txF TxFunc) {
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

func (t *Tx) handleCompletion(err error) {
	if err != nil {
		t.completedErr = err
	}

	t.completedCount++

	if t.completedErr != nil {
		t.rollback()
	}
}

func (t *Tx) rollback() {
	for i := t.completedCount - 1; i >= 0; i-- {
		rollbackFunc := t.rollbackFuncs[i]
		err := rollbackFunc()
		if err != nil {
			panic(err)
		}
	}
}
