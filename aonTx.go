package aonTx

import (
	"sync"
)

type TxFunc func() error
type RollbackFunc func() error

type Tx struct {
	txFuncs        []TxFunc
	rollbackFuncs  []RollbackFunc
	async          bool
	lock           sync.Mutex
	completedCount int
	completedErr   error
}

func NewTx(async bool) *Tx {
	return &Tx{
		txFuncs:       make([]TxFunc, 0),
		rollbackFuncs: make([]RollbackFunc, 0),
		async:         async,
	}
}

func (t *Tx) AddFunc(txFunc TxFunc, rollbackFunc RollbackFunc) {
	t.txFuncs = append(t.txFuncs, txFunc)
	t.rollbackFuncs = append(t.rollbackFuncs, rollbackFunc)
}

func (t *Tx) ExecuteFunc(txFunc TxFunc, rollbackFunc RollbackFunc) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.completedErr = nil

	err := txFunc()
	if err != nil {
		t.AddFunc(func() error { return nil }, rollbackFunc)
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
				err := txF()
				t.handleCompletion(err)
			}(txF)
		} else {
			err := txF()
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
