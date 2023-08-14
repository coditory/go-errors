package errors_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/coditory/go-errors"
)

type FormatSuite struct {
	suite.Suite
}

func TestFormatSuite(t *testing.T) {
	suite.Run(t, new(FormatSuite))
}

func (suite *FormatSuite) TestFormatGoError() {
	err := fmt.Errorf("some err")
	text := errors.Format(err)
	suite.Equal("some err\n", text)
}

func (suite *FormatSuite) TestFormats() {
	tests := []struct {
		verbosity int
		chunks    []string
	}{
		{0, []string{"foo failed\n"}},
		{1, []string{"foo failed\ncaused by: bar failed\n"}},
		{2, []string{
			"foo failed\n\tgithub.com/coditory/go-errors/test_test.foo\\(\\):\\d+",
			"caused by: bar failed\n\tgithub.com/coditory/go-errors/test_test.bar\\(\\):\\d+",
		}},
		{3, []string{
			"foo failed\n\t\\./test/format_test.go:\\d+",
			"caused by: bar failed\n\t\\./test/format_test.go:\\d+",
		}},
		{4, []string{
			"foo failed\n\t\\./test/format_test.go:\\d+ foo\\(\\)",
			"caused by: bar failed\n\t\\./test/format_test.go:\\d+ bar\\(\\)",
		}},
		{5, []string{
			"foo failed\n\tgithub.com/coditory/go-errors/test_test.foo\\(\\)\n\t\t\\./test/format_test.go:\\d+\n",
			"\ncaused by: bar failed\n\tgithub.com/coditory/go-errors/test_test.bar\\(\\)\n\t\t\\./test/format_test.go:\\d+\n",
		}},
		{6, []string{
			"foo failed\n\t./test/format_test.go:\\d+\n\t\tfoo\\(\\)",
			"\ncaused by: bar failed\n\t./test/format_test.go:\\d+\n\t\tbar\\(\\)",
		}},
		{7, []string{
			"foo failed\n\t.+/test/format_test.go:\\d+\n\t\tgithub.com/coditory/go-errors/test_test.foo\\(\\)\n",
			"\ncaused by: bar failed\n\t.+/test/format_test.go:\\d+\n\t\tgithub.com/coditory/go-errors/test_test.bar\\(\\)\n",
		}},
	}
	tests = append(tests, tests[len(tests)-1])
	tests[len(tests)-1].verbosity = 100
	tests = append(tests, tests[0])
	tests[len(tests)-1].verbosity = -100
	for _, tt := range tests {
		name := fmt.Sprintf("Verbosity:%d", tt.verbosity)
		suite.Run(name, func() {
			err := foo()
			result := errors.Formatv(err, tt.verbosity)
			for _, c := range tt.chunks {
				re, _ := regexp.Compile(c)
				matches := re.MatchString(result)
				if !matches {
					suite.Failf("Invalid error format", "Expected:\n%s\n...to match:\n%q\n", result, c)
				}
			}
		})
	}
}

func (suite *FormatSuite) TestStackTraceSlice() {
	err := foo()
	got := errors.StackTraceSlice(err)
	suite.Assert().Len(got, 7, "StackTraceSlice has no expected len")
}

func (suite *FormatSuite) TestStackTraceSliceNoError() {
	err := fmt.Errorf("some err")
	got := errors.StackTraceSlice(err)
	suite.Assert().Len(got, 0, "StackTraceSlice has no expected len")
}

func foo() error {
	err := bar()
	return errors.Wrap(err, "foo failed")
}

func bar() error {
	return errors.New("bar failed")
}
