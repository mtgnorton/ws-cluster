package kit

import (
	"fmt"
	"runtime"
	"strings"
)

// GetCallerPosition 返回文件名和行号
func GetCallerPosition[T int | int8 | int16 | int32 | int64](skip T) string {

	callerFilePathFunc := func(filePath string) string {
		// To make sure we trim the path correctly on Windows too, we
		// counter-intuitively need to use '/' and *not* os.PathSeparator here,
		// because the path given originates from Go stdlib, specifically
		// runtime.Caller() which (as of Mar/17) returns forward slashes even on
		// Windows.
		//
		// See https://github.com/golang/go/issues/3335
		// and https://github.com/golang/go/issues/18151
		//
		// for discussion on the issue on Go side.
		idx := strings.LastIndexByte(filePath, '/')
		if idx == -1 {
			return filePath
		}
		idx = strings.LastIndexByte(filePath[:idx], '/')
		if idx == -1 {
			return filePath
		}
		return filePath[idx+1:]
	}

	if _, file, line, ok := runtime.Caller(int(skip)); ok {
		return fmt.Sprintf("%s:%d", callerFilePathFunc(file), line)
	}

	return ""

}
