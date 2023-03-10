# goTx: Distributed Transaction Patterns for Go
goTx is a Go library that provides distributed transaction patterns to help you build reliable and scalable distributed systems. With goTx, you can leverage patterns such as Saga to coordinate transactions across multiple services or databases.

## Installation
To use goTx, you need to have Go 1.16 or higher installed on your system. Then, you can install the library using the following command:

> go get github.com/interwubs/goTx

## Usage
### Saga Pattern
The Saga pattern is a popular pattern for coordinating transactions across multiple services. In goTx, you can implement Sagas using the SagaTx struct:

```go
type SagaTx struct {
    txFuncs       []UpdateFunc
    rollbackFuncs []CompensateFunc
    async         bool
    //...
}
```

Each SagaTx consists of a sequence of txFuncs and rollbackFuncs, where each step represents an action to be taken to complete the transaction. You can define your own UpdateFunc and CompensateFunc types to perform the necessary actions:

```go
type UpdateFunc func() error
type CompensateFunc func() error
```

Here's an example of how you can define a Saga with two steps:
```go

func transferMoney(from, to string, amount float64) *SagaTx {
sagaTx := NewSagaTx(false)

	sagaTx.Append(func() error {
		// transfer money
		return nil
	}, func() error {
		// compensate by reversing transfer
		return nil
	})

	sagaTx.Append(func() error {
		// update account balance
		return nil
	}, func() error {
		// compensate by reverting account balance
		return nil
	})

	return sagaTx
}

```

In this example, we define a Saga that transfers money from one account to another and updates the account balance. The Append method adds a txFunc and a corresponding rollbackFunc to the SagaTx struct.

Once you've defined your Saga, you can execute it using the `ExecuteAll()` method:

```go
sagaTx := transferMoney("Alice", "Bob", 100.0)
err := sagaTx.ExecuteAll()
if err != nil {
// handle error
}
```

The `ExecuteAll()` method takes care of executing each step of the Saga in order and rolling back the transaction if any step fails.

#### Retries
You can also configure goTx to retry failed steps by setting the retries field to true and specifying the retry options:

```go
sagaTx := NewSagaTx(false)
sagaTx.retries = true
sagaTx.RetryOptions = RetryOptions{
                        MaxRetries: 3,
                        Backoff: &ExponentialBackoff{
                            InitialInterval: 1 * time.Second,
                            MaxInterval:     30 * time.Second,
                            Multiplier:      2,
                            RandomFactor:    0.2,
                        },
                    }
```

With retries enabled, goTx will automatically retry a failed step up to MaxRetries times with an exponential backoff delay between retries.

#### Asynchronous Execution
If you want to execute the steps of a Saga in parallel, you can set the async field to true:

> sagaTx := NewSagaTx(true)

With asynchronous execution, goTx will execute each step of the Saga in a separate Goroutine and return immediately, allowing you to perform other work while the transaction is being executed.


### Chain Operations
With goTx, you can implement chains of operations using the Chain struct:

```go
type Chain struct {
    RetryOptions

	ops     []*ChainOperation
	async   bool
	retries bool
	//...
}
```

Each Chain consists of a sequence of ChainOperations, where each operation represents an action to be taken to complete the chain. You can define your own ChainOperation types to perform the necessary actions:

```go
type ChainOperation struct {
    tryFunc     func() error
    secondaryOp *ChainOperation
}
```

Here's an example of how you can define a chain with two operations:

```go
func chainOperations() *Chain {
    chain := NewChain(false)

	chain.Append(NewOperation(func() error {
		// perform operation 1
		return nil
	}, nil))

	chain.Append(NewOperation(func() error {
		// perform operation 2
		return nil
	}, nil))

	return chain
}
```

In this example, we define a chain that performs two operations. The Append method adds a ChainOperation to the Chain struct.

Once you've defined your chain, you can execute it using the `ExecuteAll()` method:

```go
chain := chainOperations()
err := chain.ExecuteAll()
if err != nil {
// handle error
}
```

The `ExecuteAll()` method takes care of executing each operation in the chain in order and rolling back the transaction if any operation fails.

#### Retries
You can also configure goTx to retry failed operations by setting the retries field to true and specifying the retry options:

```go
chain := NewChain(false)
chain.retries = true
chain.RetryOptions = RetryOptions{
                        MaxRetries: 3,
                        Backoff: &ExponentialBackoff{
                        InitialInterval: 1 * time.Second,
                        MaxInterval:     30 * time.Second,
                        Multiplier:      2,
                        RandomFactor:    0.2,
						},
                    }
```

With retries enabled, goTx will automatically retry a failed operation up to MaxRetries times with an exponential backoff delay between retries.

#### Asynchronous Execution
If you want to execute the operations of a chain in parallel, you can set the async field to true:

> chain := NewChain(true)

With asynchronous execution, goTx will execute each operation of the chain in a separate Goroutine and return immediately, allowing you to perform other work while the transaction is being executed.

## Contributing
If you want to contribute to goTx, you can do so by submitting issues and pull requests.

Before submitting a pull request, please make sure your code follows the Go coding standards and includes tests.

## License
goTx is licensed under the MIT License. See the LICENSE file for more information.