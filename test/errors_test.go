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
	BackupBasePath      string
	BackupBaseModule    string
	BackupBaseCachePath string
}

func TestErrorsSuite(t *testing.T) {
	suite.Run(t, new(ErrorsSuite))
}

func (suite *ErrorsSuite) SetupSuite() {
	suite.BackupBaseModule = errors.BaseModule
	suite.BackupBasePath = errors.BasePath
	suite.BackupBaseCachePath = errors.BaseCachePath
}

func (suite *ErrorsSuite) TearDownSuite() {
	errors.BaseModule = suite.BackupBaseModule
	errors.BasePath = suite.BackupBasePath
	errors.BaseCachePath = suite.BackupBaseCachePath
}

func (suite *ErrorsSuite) SetupTest() {
	errors.BaseModule = "github.com/coditory/go-errors"
	errors.BasePath = "**/go-errors"
	errors.BaseCachePath = "**/mod/pkg"
}

func (suite *ErrorsSuite) TestNewError() {
	err := errors.New("foo")
	suite.Equal("foo", err.Error())

	err = errors.New("")
	suite.Equal("", err.Error())

	err = errors.New("foo: %s", "bar")
	suite.Equal("foo: bar", err.Error())

	suite.Nil(err.Unwrap())

	suite.Equal("./test/errors_test.go", frameFile(err, 0))
	suite.Equal("./test_test.(*ErrorsSuite).TestNewError", frameFunc(err, 0))
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

func frameFile(err *errors.Error, idx int) string {
	frame := err.StackTrace()[idx]
	return fmt.Sprintf("%s", frame)
}

func frameFunc(err *errors.Error, idx int) string {
	frame := err.StackTrace()[idx]
	return fmt.Sprintf("%n", frame)
}

type errorString string

func (e errorString) Error() string {
	return string(e)
}
