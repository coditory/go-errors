package errors

import (
	"path"
	"runtime"
	"runtime/debug"
)

var Config struct {
	StackTraceFormatter func(err *Error, verbosity int) string
	Verbosity           int
	BasePath            string
	BaseCachePath       string
	BaseModule          string
	BaseGoSrcPath       string
	BaseGoSrcToken      string
	MaxStackDepth       int
	MaxPrintStackFrames int
	MaxPrintCauses      int
}

func init() {
	Config.Verbosity = 4
	Config.MaxStackDepth = 32
	Config.MaxPrintCauses = 5
	Config.MaxPrintStackFrames = 5
	Config.BaseCachePath = "**/pkg/mod/"
	Config.BaseGoSrcPath = runtime.GOROOT() + "/"
	Config.BaseGoSrcToken = runtime.Version()
	bi, ok := debug.ReadBuildInfo()
	if ok && bi.Path != "" {
		Config.BaseModule = bi.Path + "/"
		Config.BasePath = "**/" + path.Base(Config.BaseModule) + "/"
	}
}
