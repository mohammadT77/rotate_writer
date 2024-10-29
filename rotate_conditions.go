package rotate_writer

import (
	"fmt"
	"os"
	"time"
)

type FileNameFn = func(itemIdx int, currentSize int, addedSize int, startTime time.Time, endTime time.Time) string

type RotateConditionFn = func(itemIdx int, currentSize int, addedSize int, startTime time.Time, endTime time.Time) (rotate bool, fileName string)

type OnRotateFn = func(file *os.File, fileName string)

type OnPruneFn = func(removedFiles []string, err error)

var DefaultRotateCondition = func(itemIdx int, currentSize int, addedSize int, startTime time.Time, endTime time.Time) (rotate bool, fileName string) {
	return false, fmt.Sprint(itemIdx, "-", startTime.Format(time.RFC3339))
}
