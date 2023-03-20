package errors

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
)

// Export a number of functions or variables from pkg/errors. We want people to be able to
// use them, if only via the entrypoints we've vetted in this file.
var (
	As                  = errors.As
	Unwrap              = errors.Unwrap
	Verbosity           = 2
	BasePath            = ""
	BaseCachePath       = ""
	BaseModule          = ""
	BaseGoSrcPath       = ""
	BaseGoSrcToken      = ""
	MaxStackDepth       = 32
	MaxPrintStackFrames = 5
	MaxPrintCauses      = 5
)

func Formatv(err error, verbosity int) string {
	if serr, ok := err.(*Error); ok {
		return serr.StackTraceString(verbosity)
	}
	return err.Error() + "\n"
}

func Format(err error) string {
	return Formatv(err, Verbosity)
}

func init() {
	bi, ok := debug.ReadBuildInfo()
	if ok && bi.Path != "" {
		BaseModule = bi.Path
		BasePath = "**/" + path.Base(BaseModule)
	}
	BaseCachePath = "**/pkg/mod"
	BaseGoSrcPath = runtime.GOROOT()
	BaseGoSrcToken = runtime.Version()
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

func (e *Error) StackTraceString(verbosity int) string {
	if verbosity <= 0 {
		return e.Error() + "\n"
	}
	var err error
	err = e
	buf := bytes.NewBufferString("")
	causes := 0
	for err != nil && causes < MaxPrintCauses {
		msg := err.Error()
		if causes == 0 {
			fmt.Fprintf(buf, "%s\n", msg)
		} else {
			fmt.Fprintf(buf, "caused by: %s\n", msg)
		}
		serr, ok := err.(stackTracer)
		if ok && verbosity > 1 {
			stacktrace := serr.StackTrace()
			n := MaxPrintStackFrames
			if n > len(stacktrace) {
				n = len(stacktrace)
			}
			for i := 0; i < n; i++ {
				frame := stacktrace[i]
				switch verbosity {
				case 2:
					fmt.Fprintf(buf, "\t%s:%d\n", frame.RelFuncName(), frame.Line())
				case 3:
					fmt.Fprintf(buf, "\t%s:%d\n", frame.RelFile(), frame.Line())
				case 4:
					fmt.Fprintf(buf, "\t%s:%d\n", frame.RelFile(), frame.Line())
					fmt.Fprintf(buf, "\t\t%s\n", frame.RelFuncName())
				default:
					fmt.Fprintf(buf, "\t%s:%d\n", frame.File(), frame.Line())
					fmt.Fprintf(buf, "\t\t%s\n", frame.FuncName())
				}
			}
			if n < len(stacktrace) {
				if len(stacktrace) >= MaxStackDepth {
					fmt.Fprintf(buf, "\t...skipped")
				} else {
					fmt.Fprintf(buf, "\t...skipped: %d\n", len(stacktrace)-MaxPrintStackFrames)
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
			fmt.Fprint(buf, "...skipped\n")
		}
	}
	return buf.String()
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

// Pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f Frame) Pc() uintptr { return uintptr(f) - 1 }

// File returns the full path to the File that contains the
// function for this Frame's pc.
func (f Frame) File() string {
	fn := runtime.FuncForPC(f.Pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.Pc())
	return file
}

// Line returns the Line number of source code of the
// function for this Frame's pc.
func (f Frame) Line() int {
	fn := runtime.FuncForPC(f.Pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.Pc())
	return line
}

// FuncName returns the FuncName of this function, if known.
func (f Frame) FuncName() string {
	fn := runtime.FuncForPC(f.Pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// function name relative to main package
func (f Frame) RelFuncName() string {
	name := f.FuncName()
	if BaseModule != "" && strings.HasPrefix(name, BaseModule) {
		name = "./" + name[len(BaseModule)+1:]
	}
	return name
}

// file name relateive to BasePath or BaseCachePath
func (f Frame) RelFile() string {
	name := f.File()
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
	if BaseGoSrcPath != "" {
		if strings.HasPrefix(name, BaseGoSrcPath) {
			return fmt.Sprintf("%s/%s", BaseGoSrcToken, name[len(BaseGoSrcPath)+1:])
		}
	}
	return name
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
