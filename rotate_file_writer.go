package rotate_writer

import (
	"io"
	"os"
	"path"
)

type FileRotatorFn = func(RotateStatus) (rotate bool, fileName string)

func getDirName(filePath string) string {
	return path.Dir(filePath)
}

func getRotatorFn(dir string, fileRotatorFn FileRotatorFn) RotatorFn {
	return func(status RotateStatus) io.WriteCloser {
		if rotate, fileName := fileRotatorFn(status); rotate {
			newFile, err := os.OpenFile(path.Join(dir, fileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				return newFile
			}
		}

		return nil
	}
}

type RotateFileWriter struct {
	*RotateWriter

	dir string
}

func NewRotateFileWriter(initFilePath string, fileRotatorFn FileRotatorFn) *RotateFileWriter {
	if fileRotatorFn == nil {
		return nil
	}

	dir := getDirName(initFilePath)
	rotatorFn := getRotatorFn(dir, fileRotatorFn)
	initFile, err := os.Open(initFilePath)
	if err != nil {
		return nil
	}

	return &RotateFileWriter{
		RotateWriter: NewRotateWriter(initFile, rotatorFn),
		dir:          dir,
	}
}
