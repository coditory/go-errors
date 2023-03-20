# Coditory - Go Errors
[![GitHub release](https://img.shields.io/github/v/release/coditory/go-errors.svg)](https://github.com/coditory/go-errors/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/coditory/go-errors.svg)](https://pkg.go.dev/github.com/coditory/go-errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/coditory/go-errors)](https://goreportcard.com/report/github.com/coditory/go-errors)
[![Build Status](https://github.com/coditory/go-errors/workflows/Build/badge.svg?branch=main)](https://github.com/coditory/go-errors/actions?query=workflow%3ABuild+branch%3Amain)
[![Coverage](https://codecov.io/gh/coditory/go-errors/branch/main/graph/badge.svg?token=EPRs5LiPje)](https://codecov.io/gh/coditory/go-errors)

**🚧 This library as under heavy development until release of version `1.x.x` 🚧**

> Wrapper for Go errors that prints error causes with theis stack traces.

- Prints stacks traces from all of the causes
- Shortens file paths and function names for readability
- Supports and exports `errors.Is`, `errors.As`, `errors.Unwrap`

# Getting started

## Installation
Get the dependency with:
```sh
go get github.com/coditory/go-errors
```

and import it in the project:
```go
import "github.com/coditory/go-errors"
```

The exported package is `errors`, basic usage:
```go
import "github.com/coditory/go-errors"

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
```

Output for `go run ./samples`

```
>>> Format %s:
foo failed

>>> Format %v:
foo failed
	main.foo:19
	main.main:10
	runtime.main:250
	runtime.goexit:1598
caused by: bar failed
	main.bar:23
	main.foo:18
	main.main:10
	runtime.main:250
	runtime.goexit:1598

>>> Format %+v:
foo failed
	./samples.go:19
		main.foo
	./samples.go:10
		main.main
	<GO_SRC_DIR>/runtime/proc.go:250
		runtime.main
	<GO_SRC_DIR>/runtime/asm_amd64.s:1598
		runtime.goexit
caused by: bar failed
	./samples.go:23
		main.bar
	./samples.go:18
		main.foo
	./samples.go:10
		main.main
	<GO_SRC_DIR>/runtime/proc.go:250
		runtime.main
	<GO_SRC_DIR>/runtime/asm_amd64.s:1598
		runtime.goexit

>>> Format %#v:
foo failed
	<PROJECT_DIR>/samples/samples.go:19
		main.foo
	<PROJECT_DIR>/samples/samples.go:10
		main.main
	<GO_SRC_DIR>/runtime/proc.go:250
		runtime.main
	<GO_SRC_DIR>/runtime/asm_amd64.s:1598
		runtime.goexit
caused by: bar failed
	<PROJECT_DIR>/samples/samples.go:23
		main.bar
	<PROJECT_DIR>/samples/samples.go:18
		main.foo
	/Users/mendlik/Development/go/go-errors/samples/samples.go:10
		main.main
	<GO_SRC_DIR>/runtime/proc.go:250
		runtime.main
	<GO_SRC_DIR>/runtime/asm_amd64.s:1598
		runtime.goexit
```
