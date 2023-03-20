package errors_test

import (
	"github.com/coditory/go-errors"
)

func init() {
	errors.BaseModule = "github.com/coditory/go-errors"
	errors.BasePath = "**/go-errors"
	errors.BaseCachePath = "**/mod/pkg"
}
