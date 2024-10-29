package rotate_writer

import (
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type RotateStatus struct {
	ItemIdx     int32
	CurrentSize int32
	AddedSize   int
	StartTime   time.Time
	EndTime     time.Time
}

type RotatorFn = func(RotateStatus) io.WriteCloser

type RotateWriter struct {
	currentWriter io.WriteCloser

	currentSize      atomic.Int32
	currentStartTime atomic.Int64

	counter atomic.Int32

	rotatorFn RotatorFn

	mu sync.Mutex
}

func (rw *RotateWriter) Status() RotateStatus {
	return RotateStatus{
		ItemIdx:     rw.counter.Load(),
		CurrentSize: rw.currentSize.Load(),
		AddedSize:   0,
		StartTime:   time.Unix(rw.currentStartTime.Load(), 0),
		EndTime:     time.Now(),
	}
}

func (rw *RotateWriter) checkRotate(len_p int, newTime time.Time) io.WriteCloser {
	startTimeInt := rw.currentStartTime.Load()
	startTime := time.Unix(startTimeInt, 0)

	status := RotateStatus{
		ItemIdx:     rw.counter.Load() + 1,
		CurrentSize: rw.currentSize.Load(),
		AddedSize:   len_p,
		StartTime:   startTime,
		EndTime:     newTime,
	}
	return rw.rotatorFn(status)
}

func (rw *RotateWriter) Rotate(newWriter io.WriteCloser, newTime time.Time) error {
	rw.mu.Lock()
	err := rw.currentWriter.Close()
	if err != nil {
		rw.mu.Unlock()
		return err
	}
	rw.currentWriter = newWriter
	rw.mu.Unlock()

	rw.currentStartTime.Store(newTime.Unix())
	rw.counter.Add(1)
	rw.currentSize.Store(0)

	return nil
}

func (rw *RotateWriter) Write(p []byte) (int, error) {
	newTime := time.Now()

	newWriter := rw.checkRotate(len(p), newTime)

	if newWriter != nil {
		err := rw.Rotate(newWriter, newTime)
		if err != nil {
			return 0, &RotateError{err}
		}
	}

	if rw.currentWriter == nil {
		return 0, io.ErrClosedPipe
	}

	n, err := rw.currentWriter.Write(p)

	if err != nil {
		return 0, err
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

func (rw *RotateWriter) Close() error {
	if rw.currentWriter == nil {
		return nil
	}
	err := rw.currentWriter.Close()
	rw.currentWriter = nil
	return err
}

func NewRotateWriter(initialWriter io.WriteCloser, rotateCondition RotatorFn) *RotateWriter {
	if rotateCondition == nil || initialWriter == nil {
		return nil
	}

	rw := &RotateWriter{
		rotatorFn: rotateCondition,
	}

	rw.Reset(initialWriter)

	return rw
}
