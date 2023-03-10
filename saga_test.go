package goTx

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

var (
	a = 4
	b = 5
	c = 2
)

func TestTx_Execute(t1 *testing.T) {
	wg := sync.WaitGroup{}
	type fields struct {
		txFuncs        []UpdateFunc
		rollbackFuncs  []CompensateFunc
		async          bool
		completedCount int
		completedErr   error
	}
	tests := []struct {
		name       string
		fields     fields
		wantErr    bool
		assertions func(t *testing.T, a, b interface{}, c ...interface{}) bool
	}{
		{
			name: "happypath",
			fields: fields{
				txFuncs: []UpdateFunc{
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
				rollbackFuncs: []CompensateFunc{
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
			wantErr: false,
			assertions: func(ts *testing.T, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])
				return a == 6 && b == 6 && c[0] == 6
			},
		},
		{
			name: "error-mid",
			fields: fields{
				txFuncs: []UpdateFunc{
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
				rollbackFuncs: []CompensateFunc{
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
			wantErr: true,
			assertions: func(ts *testing.T, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])

				return a == 4 && b == 5 && c[0] == 2
			},
		},
		{
			name: "happypath-async",
			fields: fields{
				txFuncs: []UpdateFunc{
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
				rollbackFuncs: []CompensateFunc{
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
			wantErr: false,
			assertions: func(ts *testing.T, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])

				return a == 6 && b == 6 && c[0] == 6
			},
		},
	}
	for _, tt := range tests {
		t1.Run(
			tt.name, func(t1 *testing.T) {
				t := &SagaTx{
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
					t.AppendFunc(txFunc, tt.fields.rollbackFuncs[i])
				}

				err := t.ExecuteFuncs()
				if (err != nil) != tt.wantErr {
					t1.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				}
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
		txFuncs        []UpdateFunc
		rollbackFuncs  []CompensateFunc
		async          bool
		completedCount int
		completedErr   error
	}
	tests := []struct {
		name       string
		fields     fields
		appendTx   UpdateFunc
		appendRB   CompensateFunc
		wantErr    bool
		assertions func(t *testing.T, a, b interface{}, c ...interface{}) bool
	}{
		{
			name: "happypath",
			fields: fields{
				txFuncs: []UpdateFunc{
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
				rollbackFuncs: []CompensateFunc{
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
			wantErr: false,
			assertions: func(ts *testing.T, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])
				return a == 6 && b == 6 && c[0] == 6
			},
		},
		{
			name: "error-mid",
			fields: fields{
				txFuncs: []UpdateFunc{
					func() error {
						a += 2
						return nil
					},
					func() error {
						b += 1
						return nil
					},
				},
				rollbackFuncs: []CompensateFunc{
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
			wantErr: true,
			assertions: func(ts *testing.T, a interface{}, b interface{}, c ...interface{}) bool {
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
				t := &SagaTx{
					async:          tt.fields.async,
					lock:           sync.Mutex{},
					completedCount: tt.fields.completedCount,
					completedErr:   tt.fields.completedErr,
				}

				defer t.rollback()
				for i, txFunc := range tt.fields.txFuncs {
					t.AppendFunc(txFunc, tt.fields.rollbackFuncs[i])
				}

				if err := t.ExecuteFuncs(); err != nil {
					t1.Error(err)
					t1.Fail()
				}

				err := t.ExecuteFunc(tt.appendTx, tt.appendRB)
				if (err != nil) != tt.wantErr {
					t1.Errorf("ExecuteFunc() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.assertions(t1, a, b, c) {
					t1.Fail()
				}
				t.AppendFunc(func() error { return nil }, tt.appendRB)
			},
		)
	}
}

func TestTx_ExecuteRetry(t1 *testing.T) {
	wg := sync.WaitGroup{}
	type fields struct {
		txFuncs        []UpdateFunc
		rollbackFuncs  []CompensateFunc
		async          bool
		completedCount int
		completedErr   error
	}
	tests := []struct {
		name       string
		fields     fields
		wantErr    bool
		assertions func(t *testing.T, a, b interface{}, c ...interface{}) bool
	}{
		{
			name: "happypath",
			fields: fields{
				txFuncs: []UpdateFunc{
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
				rollbackFuncs: []CompensateFunc{
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
			wantErr: false,
			assertions: func(ts *testing.T, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])
				return a == 6 && b == 6 && c[0] == 6
			},
		},
		{
			name: "error-mid",
			fields: fields{
				txFuncs: []UpdateFunc{
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
				rollbackFuncs: []CompensateFunc{
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
			wantErr: true,
			assertions: func(ts *testing.T, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])

				return a == 4 && b == 5 && c[0] == 2
			},
		},
		{
			name: "happypath-async",
			fields: fields{
				txFuncs: []UpdateFunc{
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
				rollbackFuncs: []CompensateFunc{
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
			wantErr: false,
			assertions: func(ts *testing.T, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])

				return a == 6 && b == 6 && c[0] == 6
			},
		},
	}
	for _, tt := range tests {
		t1.Run(
			tt.name, func(t1 *testing.T) {
				t := &SagaTx{
					async:          tt.fields.async,
					lock:           sync.Mutex{},
					completedCount: tt.fields.completedCount,
					completedErr:   tt.fields.completedErr,
					retries:        true,
					RetryOptions: RetryOptions{
						MaxRetries: 3,
						Backoff: &ExponentialBackoff{
							InitialInterval: 500 * time.Millisecond,
							MaxInterval:     30 * time.Second,
							Multiplier:      2,
							RandomFactor:    0.2,
						},
					},
				}
				defer t.rollback()
				defer func() { wg = sync.WaitGroup{} }()
				for i, txFunc := range tt.fields.txFuncs {
					if tt.fields.async {
						wg.Add(1)
					}
					t.AppendFunc(txFunc, tt.fields.rollbackFuncs[i])
				}

				err := t.ExecuteFuncs()
				if (err != nil) != tt.wantErr {
					t1.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				}
				wg.Wait()
				if !tt.assertions(t1, a, b, c) {
					t1.Fail()
				}
			},
		)
	}
}

func TestTx_ExecuteFuncRetry(t1 *testing.T) {
	type fields struct {
		txFuncs        []UpdateFunc
		rollbackFuncs  []CompensateFunc
		async          bool
		completedCount int
		completedErr   error
	}
	tests := []struct {
		name       string
		fields     fields
		appendTx   UpdateFunc
		appendRB   CompensateFunc
		wantErr    bool
		assertions func(t *testing.T, a, b interface{}, c ...interface{}) bool
	}{
		{
			name: "happypath",
			fields: fields{
				txFuncs: []UpdateFunc{
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
				rollbackFuncs: []CompensateFunc{
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
			wantErr: false,
			assertions: func(ts *testing.T, a interface{}, b interface{}, c ...interface{}) bool {
				fmt.Println(a, b, c[0])
				return a == 6 && b == 6 && c[0] == 6
			},
		},
		{
			name: "error-mid",
			fields: fields{
				txFuncs: []UpdateFunc{
					func() error {
						a += 2
						return nil
					},
					func() error {
						b += 1
						return nil
					},
				},
				rollbackFuncs: []CompensateFunc{
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
			wantErr: true,
			assertions: func(ts *testing.T, a interface{}, b interface{}, c ...interface{}) bool {
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
				t := &SagaTx{
					async:          tt.fields.async,
					lock:           sync.Mutex{},
					completedCount: tt.fields.completedCount,
					completedErr:   tt.fields.completedErr,
					retries:        true,
					RetryOptions: RetryOptions{
						MaxRetries: 3,
						Backoff: &ExponentialBackoff{
							InitialInterval: 500 * time.Millisecond,
							MaxInterval:     30 * time.Second,
							Multiplier:      2,
							RandomFactor:    0.2,
						},
					},
				}

				defer t.rollback()
				for i, txFunc := range tt.fields.txFuncs {
					t.AppendFunc(txFunc, tt.fields.rollbackFuncs[i])
				}

				if err := t.ExecuteFuncs(); err != nil {
					t1.Error(err)
					t1.Fail()
				}

				err := t.ExecuteFunc(tt.appendTx, tt.appendRB)
				if (err != nil) != tt.wantErr {
					t1.Errorf("ExecuteFunc() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.assertions(t1, a, b, c) {
					t1.Fail()
				}
				t.AppendFunc(func() error { return nil }, tt.appendRB)
			},
		)
	}
}
