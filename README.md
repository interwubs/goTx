# All-or-Nothing Transactions
### All-or-nothing transactions (aonTx) is a Go package that provides a simple way to manage custom transactions. It allows you to create and append related functions and their arguments, with corresponding rollback functions to be performed upon error. The package can handle synchronous or asynchronous execution of the functions, and if an error is encountered, the package calls the corresponding rollback functions for all completed functions in reverse order to return to the previous state.

### aonTx can be used to manage more complex transactions involving multiple functions with different signatures, and it provides a convenient way to ensure that transactions are atomic and all-or-nothing, even in the face of errors.

## Installation
### To install the aonTx package, use the following command:

> go get github.com/interwubs/aonTx


## Usage
### Here's an example usage of the aonTx package:

```go
package main

import (
	"fmt"

	"github.com/interwubs/aonTx"
)

func main() {
	tx := aonTx.NewSagaTx(false)

	tx.Append(
		func() error {
			fmt.Println("Executing function 1")
			return nil
		}, func() error {
			fmt.Println("Rolling back function 1")
			return nil
		},
	)

	tx.Append(
		func() error {
			fmt.Println("Executing function 2")
			return fmt.Errorf("error in function 2")
		}, func() error {
			fmt.Println("Rolling back function 2")
			return nil
		},
	)

	tx.Append(
		func() error {
			fmt.Println("Executing function 3")
			return nil
		}, func() error {
			fmt.Println("Rolling back function 3")
			return nil
		},
	)

	if err := tx.ExecuteAll(); err != nil {
		fmt.Printf("Transaction failed: %v\n", err)
	} else {
		fmt.Println("Transaction succeeded")
	}
}

```


#### In this example, we create a new transaction using aonTx.NewTx, and then add three functions to the transaction using tx.AddFunc. The first and third functions simply print a message, while the second function returns an error to simulate a failure.

#### We then call tx.Execute to run the transaction. Since the second function returns an error, the transaction fails and the rollback functions for the first and third functions are called to return the system to its previous state.

#### If all functions complete successfully, the transaction succeeds and the tx.Execute method returns nil.

### Contributing
#### Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

#### Please make sure to update tests as appropriate.

### License
MIT




