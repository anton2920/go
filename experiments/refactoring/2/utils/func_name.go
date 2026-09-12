package utils

import (
	"path"
	"reflect"
	"runtime"
)

func FunctionName(fn interface{}) string {
	return path.Base(runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name())
}
