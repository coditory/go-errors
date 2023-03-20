package errors

import (
	"errors"
	"fmt"
	"io"
	"path"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

// Export a number of functions or variables from pkg/errors. We want people to be able to
// use them, if only via the entrypoints we've vetted in this file.
var (
	As                  = errors.As
	Unwrap              = errors.Unwrap
	BasePath            = ""
	BaseCachePath       = ""
	BaseModule          = ""
	MaxStackDepth       = 32
	MaxPrintStackFrames = 5
	MaxPrintCauses      = 5
)

func init() {
	bi, ok := debug.ReadBuildInfo()
	if ok && bi.Path != "" {
		BaseModule = bi.Path
		BasePath = "**/" + path.Base(BaseModule)
	}
	BaseCachePath = "**/pkg/mod"
}

type stackTracer interface {
	StackTrace() StackTrace
}

type wrapper interface {
	Unwrap() error
}

type Error struct {
	message string
	err     error
	stack   []uintptr
}

func (e *Error) Error() string {
	if e.message != "" {
		return e.message
	}
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) StackTrace() StackTrace {
	return newStackTrace(e.stack)
}

func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		e.formatCauses(s)
	case 's':
		write(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

func (e *Error) formatCauses(s fmt.State) {
	var err error
	err = e
	causes := 0
	for err != nil && causes < MaxPrintCauses {
		msg := err.Error()
		if causes == 0 {
			fmt.Fprintf(s, "%s\n", msg)
		} else {
			fmt.Fprintf(s, "caused by: %s\n", msg)
		}
		serr, ok := err.(stackTracer)
		if ok {
			stacktrace := serr.StackTrace()
			n := MaxPrintStackFrames
			if n > len(stacktrace) {
				n = len(stacktrace)
			}
			for i := 0; i < n; i++ {
				frame := stacktrace[i]
				if s.Flag('+') {
					fmt.Fprintf(s, "\t%s:%d\n", frame, frame)
					fmt.Fprintf(s, "\t\t%n\n", frame)
				} else if s.Flag('#') {
					fmt.Fprintf(s, "\t%+s:%d\n", frame, frame)
					fmt.Fprintf(s, "\t\t%+n\n", frame)
				} else {
					fmt.Fprintf(s, "\t%n:%d\n", frame, frame)
				}
			}
			if n < len(stacktrace) {
				if len(stacktrace) >= MaxStackDepth {
					fmt.Fprintf(s, "\t...skipped")
				} else {
					fmt.Fprintf(s, "\t...skipped: %d\n", len(stacktrace)-MaxPrintStackFrames)
				}
			}
		}
		if werr, ok := err.(wrapper); ok {
			err = werr.Unwrap()
		} else {
			err = nil
		}
		causes++
		if causes >= MaxPrintCauses {
			write(s, "...skipped\n")
		}
	}
}

// Creates a new error with a stack trace.
// Supports interpolating of message parameters.
func New(msg string, args ...interface{}) *Error {
	return &Error{
		message: fmt.Sprintf(msg, args...),
		stack:   callers(1),
	}
}

// Creates a new error with a cause and a stack trace.
func Wrap(err error, msgAndArgs ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		message: messageFromMsgAndArgs(msgAndArgs...),
		err:     err,
		stack:   callers(1),
	}
}

// RecoverPanic turns a panic into an error.
//
// Example:
//
//	func Do() (err error) {
//	  defer func() {
//	    errors.RecoverPanic(recover(), &err)
//	  }()
//	}
func RecoverPanic(r interface{}, errPtr *error) {
	var err error
	if r != nil {
		if panicErr, ok := r.(error); ok {
			err = panicErr
		} else {
			err = fmt.Errorf("%v", r)
		}
	}
	if err != nil {
		// 2 skips: errors.go and defer
		*errPtr = &Error{
			message: "caught panic",
			err:     err,
			stack:   callers(2),
		}
	}
}

func callers(skip int) []uintptr {
	pc := make([]uintptr, MaxStackDepth)
	n := runtime.Callers(skip+2, pc)
	return pc[:n]
}

// StackTrace is stack of Frames from innermost (newest) to outermost (oldest).
type StackTrace []Frame

// StackTrace is stack of Frames from innermost (newest) to outermost (oldest).
func newStackTrace(stack []uintptr) StackTrace {
	f := make([]Frame, len(stack))
	for i := 0; i < len(f); i++ {
		f[i] = Frame(stack[i])
	}
	return f
}

// Detects whether the error is equal to a given error. Errors
// are considered equal by this function if they are matched by errors.Is
// or if their contained errors are matched through errors.Is
func Is(e error, original error) bool {
	if errors.Is(e, original) {
		return true
	}
	if e, ok := e.(*Error); ok {
		return Is(e.err, original)
	}
	if original, ok := original.(*Error); ok {
		return Is(e, original.err)
	}
	return false
}

// Frame represents a program counter inside a stack frame.
// For historical reasons if Frame is interpreted as a uintptr
// its value represents the program counter + 1.
type Frame uintptr

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f Frame) pc() uintptr { return uintptr(f) - 1 }

// file returns the full path to the file that contains the
// function for this Frame's pc.
func (f Frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// line returns the line number of source code of the
// function for this Frame's pc.
func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// name returns the name of this function, if known.
func (f Frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// Format formats the frame according to the fmt.Formatter interface.
//
//	%s    source file relative to the compile time GOPATH
//	%d    source line
//	%n    function name without package
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//	%+s   full source file path
//	%+n   function name with package
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			write(s, f.file())
		default:
			write(s, relName(f.file()))
		}
	case 'd':
		write(s, strconv.Itoa(f.line()))
	case 'n':
		switch {
		case s.Flag('+'):
			write(s, f.name())
		default:
			write(s, relFuncname(f.name()))
		}
	}
}

// function name relative to main package
func relFuncname(name string) string {
	if BaseModule != "" && strings.HasPrefix(name, BaseModule) {
		name = "./" + name[len(BaseModule)+1:]
	}
	return name
}

// file name relateive to BasePath or BaseCachePath
func relName(name string) string {
	if BasePath != "" {
		if strings.HasPrefix(BasePath, "**/") {
			i := strings.Index(name, BasePath[3:])
			if i > 0 {
				name = name[i+len(BasePath)-2:]
				for strings.HasPrefix(name, BasePath[3:]) {
					name = name[len(BasePath)-2:]
				}
				return "./" + name
			}
		} else if strings.HasPrefix(name, BasePath) {
			return name[len(BasePath)+1:]
		}
	}
	if BaseCachePath != "" {
		if strings.HasPrefix(BaseCachePath, "**/") {
			i := strings.Index(name, BaseCachePath[3:])
			if i > 0 {
				return name[i+len(BaseCachePath)-2:]
			}
		} else if strings.HasPrefix(name, BaseCachePath) {
			return name[len(BaseCachePath)+1:]
		}
	}
	return name
}

func write(state fmt.State, text string) {
	_, _ = io.WriteString(state, text)
}

func messageFromMsgAndArgs(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return ""
	}
	if len(msgAndArgs) == 1 {
		msg := msgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", msg)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}
