package rotate_writer

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

type RotatorFn = func(itemIdx int32, currentSize int32, addedSize int, startTime time.Time, endTime time.Time) io.WriteCloser

type RotateWriter struct {
	currentWriter io.WriteCloser

	currentSize      atomic.Int32
	currentStartTime atomic.Int64

	counter atomic.Int32

	rotateCondition RotatorFn
}

func (rw *RotateWriter) checkRotateCond(len_p int, newTime *time.Time) io.WriteCloser {
	startTimeInt := rw.currentStartTime.Load()
	startTime := time.Unix(startTimeInt, 0)

	return rw.rotateCondition(rw.counter.Load()+1, rw.currentSize.Load(), len_p, startTime, *newTime)
}

func (rw *RotateWriter) rotate(newWriter io.WriteCloser, newTime *time.Time) error {
	err := rw.currentWriter.Close()

	if err != nil {
		return err
	}

	rw.currentWriter = newWriter
	rw.currentStartTime.Store(newTime.Unix())
	rw.counter.Add(1)
	rw.currentSize.Store(0)

	return nil
}

func (rw *RotateWriter) Write(p []byte) (int, error) {
	newTime := time.Now()

	newWriter := rw.checkRotateCond(len(p), &newTime)

	if newWriter != nil {
		err := rw.rotate(newWriter, &newTime)
		if err != nil {
			return 0, fmt.Errorf("failed to rotate file: %s", err)
		}
	}

	if rw.currentWriter == nil {
		return 0, fmt.Errorf("failed to write to file: currentWriter is nil")
	}

	n, err := rw.currentWriter.Write(p)

	if err != nil {
		return 0, fmt.Errorf("failed to write to file: %s", err)
	}

	rw.currentSize.Add(int32(n))

	return n, nil
}

func (rw *RotateWriter) Reset(writerCloser io.WriteCloser) {
	rw.currentStartTime.Store(time.Now().Unix())
	rw.currentWriter = writerCloser
	rw.counter.Store(0)
	rw.currentSize.Store(0)
}

func NewRotateWriter(initialWriter io.WriteCloser, rotateCondition RotatorFn) (*RotateWriter, error) {

	if rotateCondition == nil {
		return nil, fmt.Errorf("rotateCondition cannot be nil")
	}

	if initialWriter == nil {
		return nil, fmt.Errorf("writerCloser cannot be nil")
	}

	rw := &RotateWriter{
		rotateCondition: rotateCondition,
	}

	rw.Reset(initialWriter)

	return rw, nil
}
