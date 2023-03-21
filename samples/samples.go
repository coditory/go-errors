package main

import (
	"fmt"

	"github.com/coditory/go-errors"
)

func main() {
	err := foo()
	for i := 0; i < 8; i++ {
		fmt.Printf(">>> Format: %d\n", i)
		fmt.Println(errors.Formatv(err, i))
	}
}

func foo() error {
	err := bar()
	return errors.Wrap(err, "foo failed")
}

func bar() error {
	return errors.New("bar failed")
}
