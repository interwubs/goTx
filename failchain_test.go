package goTx

import (
	"fmt"
	"testing"
)

var f = 1

func TestOps(t *testing.T) {
	ch := NewChain(false)
	ch.Append(
		NewOperation(
			func() error {
				return doSomething()
			}, NewOperation(
				func() error {
					return doSomethingElse()
				}, NewOperation(
					func() error {
						if err := undoEverything(); err != nil {
							fmt.Println(err)
							return err
						}
						return nil
					}, nil,
				),
			),
		),
	)
	err := ch.ExecuteAll()
	if err != nil {
		fmt.Println(err)
	}
}

func doSomething() error {
	f += 1
	fmt.Println(f)
	return fmt.Errorf("something")
}

func doSomethingElse() error {
	f += 2
	fmt.Println(f)
	return fmt.Errorf("something else")
}

func undoEverything() error {
	f -= 3
	fmt.Println(f)
	return nil
}
