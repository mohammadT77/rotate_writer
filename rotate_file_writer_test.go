package rotate_writer_test

import (
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/mohammadT77/rotate_writer"
	"github.com/stretchr/testify/require"
)

func TestRotateFileWriter(t *testing.T) {
	os.RemoveAll("testdata")
	rotator := func(status rotate_writer.RotateStatus) (rotate bool, fileName string) {
		if status.CurrentSize > 0 {
			return true, fmt.Sprint("TestRotateFileWriter_", status.ItemIdx, ".txt")
		}

		return false, ""
	}

	rfw, err := rotate_writer.NewRotateFileWriter("testdata/TestRotateFileWriter_0.txt", rotator)

	require.Nil(t, err)
	require.NotNil(t, rfw)

	require.True(t, rfw.IsOpen())

	n, err := rfw.Write([]byte("1234"))

	require.Nil(t, err)
	require.Equal(t, 4, n)

	n, err = rfw.Write([]byte("5678"))

	require.Nil(t, err)
	require.Equal(t, 4, n)

	require.True(t, rfw.IsOpen())

	rfw.Close()

	require.False(t, rfw.IsOpen())

}

func TestRotateFileWriterOpenClose(t *testing.T) {
	os.RemoveAll("testdata")
	rotator := func(status rotate_writer.RotateStatus) (rotate bool, fileName string) {
		if status.CurrentSize > 10 {
			return true, fmt.Sprint("TestRotateFileWriterOpenClose_", status.ItemIdx, ".txt")
		}

		return false, ""
	}

	rfw, err := rotate_writer.NewRotateFileWriter("testdata/TestRotateFileWriterOpenClose_0.txt", rotator)

	require.Nil(t, err)
	require.NotNil(t, rfw)

	require.True(t, rfw.IsOpen())

	n, err := rfw.Write([]byte("1234"))

	require.Nil(t, err)
	require.Equal(t, 4, n)

	require.True(t, rfw.IsOpen())

	rfw.Close()

	require.False(t, rfw.IsOpen())

	n, err = rfw.Write([]byte("x"))

	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Equal(t, 0, n)

	err = rfw.Open()

	require.Nil(t, err)
	n, err = rfw.Write([]byte("x"))

	require.Nil(t, err)
	require.Equal(t, 1, n)

	require.True(t, rfw.IsOpen())

}

func TestRotateFileWriterRotate(t *testing.T) {
	os.RemoveAll("testdata")
	rotator := func(status rotate_writer.RotateStatus) (rotate bool, fileName string) {
		if status.CurrentSize > 10 {
			return true, fmt.Sprint("TestRotateFileWriterRotate", status.ItemIdx, ".txt")
		}

		return false, ""
	}

	rfw, err := rotate_writer.NewRotateFileWriter("testdata/TestRotateFileWriterRotate_0.txt", rotator)

	require.Nil(t, err)
	require.NotNil(t, rfw)

	require.True(t, rfw.IsOpen())

	n, err := rfw.Write([]byte("1234"))

	require.Nil(t, err)
	require.Equal(t, 4, n)

	require.True(t, rfw.IsOpen())

	err = rfw.Rotate("TestRotateFileWriterRotate_1.txt", time.Now())

	require.Nil(t, err)

	require.True(t, rfw.IsOpen())

	n, err = rfw.Write([]byte("56789abc"))

	require.Nil(t, err)
	require.Equal(t, 8, n)

	rfw.Close()

	require.False(t, rfw.IsOpen())

}

func TestRotateFileWriterParallel(t *testing.T) {
	os.RemoveAll("testdata")
	rotator := func(status rotate_writer.RotateStatus) (rotate bool, fileName string) {
		if status.CurrentSize+int32(status.AddedSize) > 15 {
			return true, fmt.Sprint("TestRotateFileWriterParallel", status.ItemIdx, ".txt")
		}

		return false, ""
	}

	rfw, err := rotate_writer.NewRotateFileWriter("testdata/TestRotateFileWriterParallel_0.txt", rotator)

	require.Nil(t, err)
	require.NotNil(t, rfw)

	require.True(t, rfw.IsOpen())

	wg := &sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			s := ""
			for j := 0; j < 6; j++ {
				s += string(byte((i+1)*(j+1)) + 'a' - 1)
			}
			t.Log(s)
			n, err := rfw.Write([]byte(s))

			require.Nil(t, err)
			require.Equal(t, len(s), n)
		}()
	}

	wg.Wait()

	rfw.Close()
	t.Error()
}
