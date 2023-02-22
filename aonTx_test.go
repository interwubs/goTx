package aonTx

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	a = 4
	b = 5
	c = 2
)

func TestTx_Execute(t1 *testing.T) {
	wg := sync.WaitGroup{}
	type fields struct {
		txFuncs        []TxFunc
		rollbackFuncs  []RollbackFunc
		async          bool
		completedCount int
		completedErr   error
	}
	tests := []struct {
		name       string
		fields     fields
		wantErr    assert.ErrorAssertionFunc
		assertions assert.ComparisonAssertionFunc
	}{
		{
			name: "happypath",
			fields: fields{
				txFuncs: []TxFunc{
					func() error {
						a += 2
						fmt.Println("a =", a)
						return nil
					},
					func() error {
						b += 1
						fmt.Println("b =", b)
						return nil
					},
					func() error {
						c += 4
						fmt.Println("c =", c)
						return nil
					},
				},
				rollbackFuncs: []RollbackFunc{
					func() error {
						a = 4
						fmt.Println("ROLLBACK A")
						return nil
					},
					func() error {
						b = 5
						fmt.Println("ROLLBACK B")
						return nil
					},
					func() error {
						c = 2
						fmt.Println("ROLLBACK C")
						return nil
					},
				},
				async:          false,
				completedCount: 0,
				completedErr:   nil,
			},
			wantErr: assert.NoError,
			assertions: func(ts assert.TestingT, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])
				return a == 6 && b == 6 && c[0] == 6
			},
		},
		{
			name: "error-mid",
			fields: fields{
				txFuncs: []TxFunc{
					func() error {
						a += 2
						return nil
					},
					func() error {
						b += 1
						return errors.New("error")
					},

					func() error {
						c += 4
						return nil
					},
				},
				rollbackFuncs: []RollbackFunc{
					func() error {
						a = 4
						return nil
					},
					func() error {
						b = 5
						return nil
					},
					func() error {
						c = 2
						return nil
					},
				},
				async:          false,
				completedCount: 0,
				completedErr:   nil,
			},
			wantErr: assert.Error,
			assertions: func(t assert.TestingT, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])

				return a == 4 && b == 5 && c[0] == 2
			},
		},
		{
			name: "happypath-async",
			fields: fields{
				txFuncs: []TxFunc{
					func() error {
						a += 2
						fmt.Println("a =", a)
						wg.Done()
						return nil
					},
					func() error {
						b += 1
						fmt.Println("b =", b)
						wg.Done()
						return nil
					},
					func() error {
						c += 4
						fmt.Println("c =", c)
						wg.Done()
						return nil
					},
				},
				rollbackFuncs: []RollbackFunc{
					func() error {
						a = 4
						fmt.Println("ROLLBACK A")
						return nil
					},
					func() error {
						b = 5
						fmt.Println("ROLLBACK B")
						return nil
					},
					func() error {
						c = 2
						fmt.Println("ROLLBACK C")
						return nil
					},
				},
				async:          true,
				completedCount: 0,
				completedErr:   nil,
			},
			wantErr: assert.NoError,
			assertions: func(t assert.TestingT, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])

				return a == 6 && b == 6 && c[0] == 6
			},
		},
	}
	for _, tt := range tests {
		t1.Run(
			tt.name, func(t1 *testing.T) {
				t := &Tx{
					async:          tt.fields.async,
					lock:           sync.Mutex{},
					completedCount: tt.fields.completedCount,
					completedErr:   tt.fields.completedErr,
				}
				defer t.rollback()
				defer func() { wg = sync.WaitGroup{} }()
				for i, txFunc := range tt.fields.txFuncs {
					if tt.fields.async {
						wg.Add(1)
					}
					t.AddFunc(txFunc, tt.fields.rollbackFuncs[i])
				}

				tt.wantErr(t1, t.Execute(), fmt.Sprintf("Execute()"))
				wg.Wait()
				if !tt.assertions(t1, a, b, c) {
					t1.Fail()
				}
			},
		)
	}
}
func TestTx_ExecuteFunc(t1 *testing.T) {
	type fields struct {
		txFuncs        []TxFunc
		rollbackFuncs  []RollbackFunc
		async          bool
		completedCount int
		completedErr   error
	}
	tests := []struct {
		name       string
		fields     fields
		appendTx   TxFunc
		appendRB   RollbackFunc
		wantErr    assert.ErrorAssertionFunc
		assertions assert.ComparisonAssertionFunc
	}{
		{
			name: "happypath",
			fields: fields{
				txFuncs: []TxFunc{
					func() error {
						a += 2
						fmt.Println("a =", a)
						return nil
					},
					func() error {
						b += 1
						fmt.Println("b =", b)
						return nil
					},
				},
				rollbackFuncs: []RollbackFunc{
					func() error {
						a = 4
						fmt.Println("ROLLBACK A")
						return nil
					},
					func() error {
						b = 5
						fmt.Println("ROLLBACK B")
						return nil
					},
				},
				async:          false,
				completedCount: 0,
				completedErr:   nil,
			},
			appendTx: func() error {
				c += 4
				fmt.Println("c =", c)
				return nil
			},
			appendRB: func() error {
				c = 2
				fmt.Println("ROLLBACK C")
				return nil
			},
			wantErr: assert.NoError,
			assertions: func(ts assert.TestingT, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])
				return a == 6 && b == 6 && c[0] == 6
			},
		},
		{
			name: "error-mid",
			fields: fields{
				txFuncs: []TxFunc{
					func() error {
						a += 2
						return nil
					},
					func() error {
						b += 1
						return nil
					},
				},
				rollbackFuncs: []RollbackFunc{
					func() error {
						a = 4
						return nil
					},
					func() error {
						b = 5
						return nil
					},
				},
				async:          false,
				completedCount: 0,
				completedErr:   nil,
			},
			wantErr: assert.Error,
			assertions: func(t assert.TestingT, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])

				return a == 4 && b == 5 && c[0] == 2
			},
			appendRB: func() error {
				c = 2
				return nil
			},
			appendTx: func() error {
				c += 4
				return errors.New("error")
			},
		},
	}
	for _, tt := range tests {
		t1.Run(
			tt.name, func(t1 *testing.T) {
				t := &Tx{
					async:          tt.fields.async,
					lock:           sync.Mutex{},
					completedCount: tt.fields.completedCount,
					completedErr:   tt.fields.completedErr,
				}
				defer t.rollback()
				for i, txFunc := range tt.fields.txFuncs {
					t.AddFunc(txFunc, tt.fields.rollbackFuncs[i])
				}
				if err := t.Execute(); err != nil {
					t1.Error(err)
					t1.Fail()
				}
				tt.wantErr(t1, t.ExecuteFunc(tt.appendTx, tt.appendRB), fmt.Sprintf("ExecuteFunc()"))
				if !tt.assertions(t1, a, b, c) {
					t1.Fail()
				}
				t.AddFunc(func() error { return nil }, tt.appendRB)
			},
		)
	}
}
