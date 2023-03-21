package errors_test

import (
	"github.com/coditory/go-errors"
)

func init() {
	errors.Config.BaseModule = "github.com/coditory/go-errors/"
	errors.Config.BasePath = "**/go-errors/"
	errors.Config.BaseCachePath = "**/mod/pkg/"
}
