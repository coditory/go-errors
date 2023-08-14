package errors

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Export a number of functions or variables from pkg/errors. We want people to be able to
// use them, if only via the entrypoints we've vetted in this file.
var (
	As     = errors.As
	Unwrap = errors.Unwrap
)

func Formatv(err error, verbosity int) string {
	if serr, ok := err.(*Error); ok {
		return serr.StackTraceString(verbosity)
	}
	return err.Error() + "\n"
}

func Format(err error) string {
	return Formatv(err, Config.Verbosity)
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
	if Config.StackTraceFormatter != nil {
		return Config.StackTraceFormatter(e, verbosity)
	}
	if verbosity <= 0 {
		return e.Error() + "\n"
	}
	var err error
	err = e
	buf := bytes.NewBufferString("")
	causes := 0
	for err != nil && causes < Config.MaxPrintCauses {
		msg := err.Error()
		if causes == 0 {
			fmt.Fprintf(buf, "%s\n", msg)
		} else {
			fmt.Fprintf(buf, "caused by: %s\n", msg)
		}
		serr, ok := err.(stackTracer)
		if ok && verbosity > 1 {
			stacktrace := serr.StackTrace()
			n := Config.MaxPrintStackFrames
			if n > len(stacktrace) {
				n = len(stacktrace)
			}
			for i := 0; i < n; i++ {
				frame := stacktrace[i]
				if Config.FrameFormatter != nil {
					Config.FrameFormatter(buf, &frame, verbosity)
				} else {
					switch verbosity {
					case 2:
						fmt.Fprintf(buf, "\t%s():%d\n",
							frame.FuncName(), frame.FileLine())
					case 3:
						fmt.Fprintf(buf, "\t%s:%d\n",
							frame.RelFileName(), frame.FileLine())
					case 4:
						fmt.Fprintf(buf, "\t%s:%d %s()\n",
							frame.RelFileName(), frame.FileLine(), frame.ShortFuncName())
					case 5:
						fmt.Fprintf(buf, "\t%s()\n", frame.FuncName())
						fmt.Fprintf(buf, "\t\t%s:%d\n", frame.RelFileName(), frame.FileLine())
					case 6:
						fmt.Fprintf(buf, "\t%s:%d\n", frame.RelFileName(), frame.FileLine())
						fmt.Fprintf(buf, "\t\t%s()\n", frame.ShortFuncName())
					default:
						fmt.Fprintf(buf, "\t%s:%d\n", frame.File(), frame.FileLine())
						fmt.Fprintf(buf, "\t\t%s()\n", frame.FuncName())
					}
				}
			}
			if n < len(stacktrace) {
				if len(stacktrace) >= Config.MaxStackDepth {
					fmt.Fprintf(buf, "\t...skipped\n")
				} else {
					fmt.Fprintf(buf, "\t...skipped: %d\n", len(stacktrace)-Config.MaxPrintStackFrames)
				}
			}
		}
		if werr, ok := err.(wrapper); ok {
			err = werr.Unwrap()
		} else {
			err = nil
		}
		causes++
		if causes >= Config.MaxPrintCauses {
			fmt.Fprint(buf, "...skipped\n")
		}
	}
	return buf.String()
}

func StackTraceSlice(err error) []string {
	var stack []string
	stacked, ok := err.(*Error)
	if !ok {
		return stack
	}
	for _, d := range stacked.StackTrace() {
		stack = append(stack, fmt.Sprintf("%s:%d", d.RelFileName(), d.FileLine()))
	}
	return stack
}

// Creates a new error with a stack trace.
// Supports interpolating of message parameters.
//
// Example:
// err := err.New("error %d", 42)
// err.Error() == "error 42"
func New(msg string, args ...interface{}) *Error {
	return &Error{
		message: fmt.Sprintf(msg, args...),
		stack:   callers(1),
	}
}

// Creates a new error with a cause and a stack trace.
// Supports interpolating of message parameters.
//
// Example:
// wrapped := fmt.Errorf("wrapped")
// err := errors.wrap(wrapped, "errored happened")
// err.Error() == "error happened"
// err.Unwrap() == wrapped
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

// Similiar to wrap. Creates a new error with message that prefixes wrapped errro.
// Supports interpolating of message parameters.
//
// Example:
// wrapped := fmt.Errorf("wrapped")
// err := errors.Prefix(wrapped, "errored happened")
// err.Error() == "error happened: wrapped"
// err.Unwrap() == wrapped
func Prefix(err error, msg string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		message: fmt.Sprintf(msg, args...) + ": " + err.Error(),
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
			message: "caught panic: " + err.Error(),
			err:     err,
			stack:   callers(3),
		}
	}
}

// Recover executes a function and turns a panic into an error.
//
// Example:
//
//	err := errors.Recover(func() {
//	  somePanicingLogic()
//	})
func Recover(action func()) error {
	err := func() (err error) {
		defer func() {
			RecoverPanic(recover(), &err)
			if err != nil {
				e := err.(*Error)
				e.stack = callers(5)
			}
		}()
		action()
		return nil
	}()
	return err
}

func callers(skip int) []uintptr {
	pc := make([]uintptr, Config.MaxStackDepth)
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

// FileLine returns the FileLine number of source code of the
// function for this Frame's pc.
func (f Frame) FileLine() int {
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
	if Config.BaseModule != "" && strings.HasPrefix(name, Config.BaseModule) {
		name = "./" + name[len(Config.BaseModule):]
	}
	return name
}

// function name relative to main package
func (f Frame) ShortFuncName() string {
	name := f.FuncName()
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}

// file name relateive to BasePath or BaseCachePath
func (f Frame) RelFileName() string {
	name := f.File()
	relPath, ok := trimBasePath(Config.BasePath, name)
	if ok {
		return "./" + relPath
	}
	relPath, ok = trimBasePath(Config.BaseCachePath, name)
	if ok {
		return relPath
	}
	baseGoSrcPath := Config.BaseGoSrcPath
	if baseGoSrcPath != "" {
		if strings.HasPrefix(name, baseGoSrcPath) {
			return fmt.Sprintf("%s/%s", Config.BaseGoSrcToken, name[len(baseGoSrcPath)+1:])
		}
	}
	return name
}

func trimBasePath(basePath string, path string) (string, bool) {
	if basePath == "" {
		return "", false
	}
	if strings.HasPrefix(basePath, "**/") {
		i := strings.LastIndex(path, basePath[3:])
		if i > 0 {
			return path[i+len(basePath)-3:], true
		}
	} else if strings.HasPrefix(path, basePath) {
		return path[len(basePath):], true
	}
	return "", false
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
