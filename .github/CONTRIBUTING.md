# Contributing

## Build
Before pushing any changes make sure project builds without errors with:
```sh
make
# Tuns subcommands:
# clean, lint, test, build
```

## Unit tests

This project uses [github.com/stretchr/testify](https://github.com/stretchr/testify) for testing.
Run tests with:
```sh
make test
```

Pull requests that lower test coverage will not be merged.
Test coverage metric will be visible in GitHub Pull requests and can be generated locally with:
```sh
make coverage
# Coverage report:
# out/report/test/coverage.html
```

Wherever possible prefer [table tests](https://dev.to/boncheff/table-driven-unit-tests-in-go-407b)
and [test suites](https://pkg.go.dev/github.com/stretchr/testify/suite)

## Formatting
Codestyle is enforced by [gofumpt](https://github.com/mvdan/gofumpt) and [golangci-lint](https://github.com/golangci/golangci-lint).

```sh
# Format code
make format

# Check linting errors
make lint
```

## Commit messages
Before writing a commit message read [this article](https://chris.beams.io/posts/git-commit/).
