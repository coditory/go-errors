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
```

Output for `go run ./samples`

```
>>> Format: 0
foo failed

>>> Format: 1
foo failed
caused by: bar failed

>>> Format: 2
foo failed
	main.foo:32
	main.main:10
	runtime.main:250
	runtime.goexit:1598
caused by: bar failed
	main.bar:36
	main.foo:31
	main.main:10
	runtime.main:250
	runtime.goexit:1598

>>> Format: 3
foo failed
	./go:32
	./go:10
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/proc.go:250
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/asm_amd64.s:1598
caused by: bar failed
	./go:36
	./go:31
	./go:10
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/proc.go:250
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/asm_amd64.s:1598

>>> Format: 4
foo failed
	./go:32
		main.foo
	./go:10
		main.main
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/proc.go:250
		runtime.main
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/asm_amd64.s:1598
		runtime.goexit
caused by: bar failed
	./go:36
		main.bar
	./go:31
		main.foo
	./go:10
		main.main
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/proc.go:250
		runtime.main
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/asm_amd64.s:1598
		runtime.goexit

>>> Format: 5
foo failed
	/Users/mendlik/Development/go/go-errors/samples/samples.go:32
		main.foo
	/Users/mendlik/Development/go/go-errors/samples/samples.go:10
		main.main
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/proc.go:250
		runtime.main
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/asm_amd64.s:1598
		runtime.goexit
caused by: bar failed
	/Users/mendlik/Development/go/go-errors/samples/samples.go:36
		main.bar
	/Users/mendlik/Development/go/go-errors/samples/samples.go:31
		main.foo
	/Users/mendlik/Development/go/go-errors/samples/samples.go:10
		main.main
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/proc.go:250
		runtime.main
	/Users/mendlik/.sdkvm/sdk/go/1.20.2/src/runtime/asm_amd64.s:1598
		runtime.goexit

>>> Format go error:
go error
```
