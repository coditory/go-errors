package errors_test

import (
	goerrors "errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/coditory/go-errors"
)

type ErrorsSuite struct {
	suite.Suite
}

func TestErrorsSuite(t *testing.T) {
	suite.Run(t, new(ErrorsSuite))
}

func (suite *ErrorsSuite) TestNewError() {
	err := errors.New("foo")
	suite.Equal("foo", err.Error())

	err = errors.New("")
	suite.Equal("", err.Error())

	err = errors.New("foo: %s", "bar")
	suite.Equal("foo: bar", err.Error())

	suite.Nil(err.Unwrap())

	suite.Equal("./test/errors_test.go", frameRelFile(err, 0))
	suite.Equal("./test_test.(*ErrorsSuite).TestNewError", frameRelFunc(err, 0))
}

func (suite *ErrorsSuite) TestWrapError() {
	werr := errors.New("bar")

	err := errors.Wrap(werr, "foo")
	suite.Equal("foo", err.Error())

	err = errors.Wrap(werr, "foo %s", "baz")
	suite.Equal("foo baz", err.Error())

	suite.Equal(werr, err.Unwrap())

	suite.Equal("./test/errors_test.go", frameRelFile(err, 0))
	suite.Equal("./test_test.(*ErrorsSuite).TestWrapError", frameRelFunc(err, 0))
}

func (suite *ErrorsSuite) TestPrefixError() {
	werr := errors.New("bar")

	err := errors.Prefix(werr, "foo")
	suite.Equal("foo: bar", err.Error())

	err = errors.Prefix(werr, "foo %s", "baz")
	suite.Equal("foo baz: bar", err.Error())

	suite.Equal(werr, err.Unwrap())

	suite.Equal("./test/errors_test.go", frameRelFile(err, 0))
	suite.Equal("./test_test.(*ErrorsSuite).TestPrefixError", frameRelFunc(err, 0))
}

func (suite *ErrorsSuite) TestRecoverPanic() {
	tests := []struct {
		err      any
		expected string
	}{
		{"some text", "caught panic: some text"},
		{fmt.Errorf("go err"), "caught panic: go err"},
		{io.EOF, "caught panic: EOF"},
		{errors.New("some err"), "caught panic: some err"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v", tt.err)
		suite.Run(name, func() {
			do := func() (err error) {
				defer func() {
					errors.RecoverPanic(recover(), &err)
				}()
				panic(tt.err)
			}
			err := do().(*errors.Error)
			suite.Equal(tt.expected, err.Error())
			if serr, ok := tt.err.(string); ok {
				suite.Equal(fmt.Errorf(serr), err.Unwrap())
			} else {
				suite.Equal(tt.err, err.Unwrap())
			}
			suite.Equal("./test/errors_test.go", frameRelFile(err, 0))
			suite.Equal("./test_test.(*ErrorsSuite).TestRecoverPanic.func1.1", frameRelFunc(err, 0))
			suite.Equal("./test_test.(*ErrorsSuite).TestRecoverPanic.func1", frameRelFunc(err, 1))
		})
	}
}

func (suite *ErrorsSuite) TestRecover() {
	tests := []struct {
		err      any
		expected string
	}{
		{"some text", "caught panic: some text"},
		{fmt.Errorf("go err"), "caught panic: go err"},
		{io.EOF, "caught panic: EOF"},
		{errors.New("some err"), "caught panic: some err"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v", tt.err)
		suite.Run(name, func() {
			do := func() {
				panic(tt.err)
			}
			err := errors.Recover(do).(*errors.Error)
			suite.Equal(tt.expected, err.Error())
			if serr, ok := tt.err.(string); ok {
				suite.Equal(fmt.Errorf(serr), err.Unwrap())
			} else {
				suite.Equal(tt.err, err.Unwrap())
			}
			suite.Equal("./test/errors_test.go", frameRelFile(err, 0))
			suite.Equal("./test_test.(*ErrorsSuite).TestRecover.func1", frameRelFunc(err, 0))
		})
	}
}

func (suite *ErrorsSuite) TestIs() {
	err := errors.New("err")
	suite.True(errors.Is(err, err),
		"err is not err")
	suite.True(!errors.Is(goerrors.New("xxx"), errors.New("xxx")),
		"New(\"xxx\") is not New(\"xxx\")")
	suite.True(!errors.Is(nil, io.EOF),
		"nil is io.EOF")
	suite.True(errors.Is(io.EOF, io.EOF),
		"io.EOF is not io.EOF")
	suite.True(errors.Is(io.EOF, errors.Wrap(io.EOF)),
		"io.EOF is not Trace(io.EOF)")
	suite.True(errors.Is(errors.Wrap(io.EOF), errors.Wrap(io.EOF)),
		"Trace(io.EOF) is not Trace(io.EOF)")
	suite.True(!errors.Is(io.EOF, fmt.Errorf("io.EOF")),
		"io.EOF is fmt.Errorf")
}

func (suite *ErrorsSuite) TestAs() {
	var errStrIn errorString = "TestForFun"
	var errStrOut errorString

	if errors.As(errStrIn, &errStrOut) {
		suite.Equal(errStrIn, errStrOut)
	} else {
		suite.FailNow("direct errStr is not returned")
	}

	errStrOut = ""
	err := errors.Wrap(errStrIn)
	if errors.As(err, &errStrOut) {
		suite.Equal(errStrIn, errStrOut)
	} else {
		suite.FailNow("wrapped errStr is not returned")
	}
}

func frameRelFile(err *errors.Error, idx int) string {
	frame := err.StackTrace()[idx]
	return frame.RelFile()
}

func frameRelFunc(err *errors.Error, idx int) string {
	frame := err.StackTrace()[idx]
	return frame.RelFuncName()
}

type errorString string

func (e errorString) Error() string {
	return string(e)
}
