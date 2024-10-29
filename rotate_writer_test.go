package rotate_writer_test

import (
	"bytes"
	"io"
	"sync"
	"testing"

	"github.com/mohammadT77/rotate_writer"
	"github.com/stretchr/testify/require"
)

func newDummyBuffer() *rotate_writer.DummyWriteCloser {
	return rotate_writer.NewDummyWCloser(bytes.NewBuffer(nil))
}

func TestRotateWriterDummyBuffer(t *testing.T) {

	writers := []*rotate_writer.DummyWriteCloser{
		newDummyBuffer(),
	}

	rotatorFn := func(s rotate_writer.RotateStatus) io.WriteCloser {
		if s.CurrentSize >= 4 {
			newWriter := newDummyBuffer()
			writers = append(writers, newWriter)
			return newWriter
		} else {
			return nil
		}

	}

	rw := rotate_writer.NewRotateWriter(writers[0], rotatorFn)

	n, err := rw.Write([]byte("1234"))

	require.Nil(t, err)
	require.Equal(t, 4, n)

	require.Len(t, writers, 1)

	n, err = rw.Write([]byte("56789"))

	require.Nil(t, err)
	require.Equal(t, 5, n)

	require.Len(t, writers, 2)

	require.Equal(t, "1234", string(writers[0].Writer().(*bytes.Buffer).Bytes()))
	require.Equal(t, "56789", string(writers[1].Writer().(*bytes.Buffer).Bytes()))

	currStatus := rw.Status()
	require.Equal(t, int32(1), currStatus.ItemIdx)
	require.Equal(t, int32(5), currStatus.CurrentSize)
	require.Equal(t, 0, currStatus.AddedSize)
}

func TestRotateWriterDummyBufferParallel(t *testing.T) {

	writers := []*rotate_writer.DummyWriteCloser{
		newDummyBuffer(),
	}

	rotatorFn := func(s rotate_writer.RotateStatus) io.WriteCloser {
		if s.CurrentSize > 0 {
			newWriter := newDummyBuffer()
			writers = append(writers, newWriter)
			return newWriter
		} else {
			return nil
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(3)

	rw := rotate_writer.NewRotateWriter(writers[0], rotatorFn)

	printAll := func() {
		for _, w := range writers {
			t.Log(string(w.Writer().(*bytes.Buffer).Bytes()))
		}
	}

	go func() {
		defer wg.Done()
		n, err := rw.Write([]byte("1234"))
		require.Nil(t, err)
		require.Equal(t, 4, n)
		t.Log("first write done")
		printAll()
	}()

	go func() {
		defer wg.Done()
		n, err := rw.Write([]byte("56789"))
		require.Nil(t, err)
		require.Equal(t, 5, n)
		t.Log("second write done")
		printAll()
	}()

	go func() {
		defer wg.Done()
		n, err := rw.Write([]byte("abc"))
		require.Nil(t, err)
		require.Equal(t, 3, n)
		t.Log("third write done")
		printAll()
	}()

	wg.Wait()

}
