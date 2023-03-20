package main

import (
	"fmt"

	"github.com/coditory/go-errors"
)

func main() {
	err := foo()
	// 0: error
	fmt.Printf("\n>>> Format: 0\n%s", errors.Formatv(err, 0))
	// 1: error + causes
	fmt.Printf("\n>>> Format: 1\n%s", errors.Formatv(err, 1))
	// 2: error + causes with stack traces of relative func names and lines
	fmt.Printf("\n>>> Format: 2\n%s", errors.Formatv(err, 2))
	// 3: error + causes with stack traces of relative file names and lines
	fmt.Printf("\n>>> Format: 3\n%s", errors.Formatv(err, 3))
	// 4: error + causes with stack traces of relative file names and lines
	//                        ...and relative func names
	fmt.Printf("\n>>> Format: 4\n%s", errors.Formatv(err, 4))
	// 5: like 4 but uses absolute file names and func names
	fmt.Printf("\n>>> Format: 5\n%s", errors.Formatv(err, 5))

	// standard errors are formatted with err.Error()
	goerr := fmt.Errorf("go error")
	fmt.Printf("\n>>> Format go error:\n%s", errors.Format(goerr))
}

func foo() error {
	err := bar()
	return errors.Wrap(err, "foo failed")
}

func bar() error {
	return errors.New("bar failed")
}
