package rotate_writer

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type RotateFileWriter struct {
	dir string

	currentFilePath  string
	currentSize      atomic.Int32
	currentStartTime time.Time

	counter atomic.Int32

	rotateCondition FileRotateConditionFn
	OnRotate        OnRotateFn
}

func (rw *RotateFileWriter) rotateFile(newFileName string, newTime *time.Time) error {
	rw.currentFilePath = filepath.Join(rw.dir, newFileName)
	rw.currentStartTime = *newTime

	file, err := os.OpenFile(rw.currentFilePath, os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	rw.counter.Add(1)

	if rw.OnRotate != nil {
		rw.OnRotate(file, newFileName)
	}

	rw.currentSize.Store(0)

	return file.Close()
}

func (rw *RotateFileWriter) Dir() string {
	return rw.dir
}

func (rw *RotateFileWriter) RotateNow() (string, error) {
	now := time.Now()
	_, fileName := rw.rotateCondition(int(rw.counter.Load()), int(rw.currentSize.Load()), 0, now, now)
	return fileName, rw.rotateFile(fileName, &now)
}

func (rw *RotateFileWriter) OpenAndWriteToCurrent(p []byte) error {

	file, err := os.OpenFile(rw.currentFilePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	n, err := file.Write(p)

	if err != nil {
		return err
	}
	rw.currentSize.Add(int32(n))

	return file.Close()
}

func (rw *RotateFileWriter) checkRotateCond(len_p int, newTime *time.Time) (bool, string) {
	currentSize := rw.currentSize.Load()
	item := rw.counter.Load()

	return rw.rotateCondition(int(item)+1, int(currentSize), len_p, rw.currentStartTime, *newTime)
}

func (rw *RotateFileWriter) Write(p []byte) (int, error) {
	newTime := time.Now()
	rotate, newFileName := rw.checkRotateCond(len(p), &newTime)

	if rotate {
		if err := rw.rotateFile(newFileName, &newTime); err != nil {
			return 0, fmt.Errorf("failed to rotate file: %s", err)
		}
	}

	err := rw.OpenAndWriteToCurrent(p)
	if err != nil {
		return 0, fmt.Errorf("failed to open and write to file: %s", err)
	}

	return len(p), nil
}

func (rw *RotateFileWriter) Reset() error {
	rw.currentSize.Store(0)
	_, err := rw.RotateNow()

	return err
}

func NewRotateFileWriter(dir string, rotateCondition FileRotateConditionFn, onRotate OnRotateFn) *RotateFileWriter {
	if rotateCondition == nil {
		rotateCondition = DefaultRotateCondition
	}

	rw := &RotateFileWriter{
		dir:             dir,
		OnRotate:        onRotate,
		rotateCondition: rotateCondition,
	}

	rw.Reset()

	return rw
}
