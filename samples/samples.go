package main

import (
	"fmt"

	"github.com/coditory/go-errors"
)

func main() {
	err := foo()
	fmt.Printf("\n>>> Format %%s:\n%s", err)
	fmt.Printf("\n>>> Format %%v:\n%v", err)
	fmt.Printf("\n>>> Format %%+v:\n%+v", err)
	fmt.Printf("\n>>> Format %%#v:\n%#v", err)
}

func foo() error {
	err := bar()
	return errors.Wrap(err, "foo failed")
}

func bar() error {
	return errors.New("bar failed")
}
