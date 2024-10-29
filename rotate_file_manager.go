package rotate_writer

import (
	"errors"
	"log"
	"os"
	"path"
	"slices"
	"strings"
)

type RotateFileManager struct {
	*RotateWriter
	maxFiles int32

	files []string

	Logger *log.Logger
}

func NewRotateFileManager(dir string, maxFiles int32, rotateCondition RotateConditionFn, onRotate OnRotateFn) *RotateFileManager {
	rfm := &RotateFileManager{
		maxFiles: maxFiles,
		files:    []string{},
	}

	onRotateFn := func(file *os.File, fileName string) {
		if onRotate != nil {
			onRotate(file, fileName)
		}
		rfm.files = append(rfm.files, fileName)
		_, err := rfm.PruneFiles()
		if err != nil {
			rfm.Logger.Printf("Failed to prune files: %s", err)
		}
	}

	rfm.RotateWriter = NewRotateWriter(dir, rotateCondition, onRotateFn)

	return rfm
}

func (rfm *RotateFileManager) PruneFiles() ([]string, error) {
	removedFiles := make([]string, 0)
	for rfm.GetNumFiles() > int(rfm.maxFiles) {
		fileToDelete := rfm.files[0]
		err := rfm.DeleteFileByIdx(0)
		if err != nil {
			return removedFiles, err
		}
		removedFiles = append(removedFiles, fileToDelete)
	}

	return removedFiles, nil
}

func (rfm *RotateFileManager) DiscoverDir(prefix string, postfix string) error {
	entriesInfo := make([]os.FileInfo, 0)

	entries, err := os.ReadDir(rfm.Dir())
	if err != nil {
		return err
	}
	for _, file := range entries {
		fileName := file.Name()
		if !file.IsDir() && strings.HasPrefix(fileName, prefix) && strings.HasSuffix(fileName, postfix) {
			fileInfo, err := file.Info()
			if err != nil {
				return err
			}
			// Sorting by modTime
			indexToEnter := 0
			for i, entryInfo := range entriesInfo {
				if fileInfo.ModTime().After(entryInfo.ModTime()) {
					continue
				} else {
					indexToEnter = i
					break
				}
			}

			entriesInfo = append(entriesInfo[:indexToEnter], append([]os.FileInfo{fileInfo}, entriesInfo[indexToEnter:]...)...)
		}
	}
	slices.Reverse(entriesInfo)

	// Prepend the prefix to the files to be rotated
	rfm.files = make([]string, 0)
	for _, entryInfo := range entriesInfo {
		rfm.files = append(rfm.files, entryInfo.Name())
	}

	_, err = rfm.PruneFiles()
	return err
}

func (rfm *RotateFileManager) DeleteFile(fileName string) error {
	fileIdx := slices.IndexFunc(rfm.files, func(f string) bool { return strings.EqualFold(f, fileName) })

	return rfm.DeleteFileByIdx(fileIdx)
}

func (rfm *RotateFileManager) DeleteFileByIdx(fileIdx int) error {
	fileName := rfm.files[fileIdx]
	filePath := path.Join(rfm.Dir(), fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	rfm.files = append(rfm.files[:fileIdx], rfm.files[fileIdx+1:]...)

	return nil
}

func (rfm *RotateFileManager) GetFiles() []string {
	return rfm.files
}

func (rfm *RotateFileManager) GetFilePaths() []string {
	filePaths := make([]string, 0)
	for _, fileName := range rfm.files {
		filePaths = append(filePaths, path.Join(rfm.Dir(), fileName))
	}
	return filePaths
}

func (rfm *RotateFileManager) GetNumFiles() int {
	return len(rfm.files)
}

func (rfm *RotateFileManager) ReadFileByIdx(fileIdx int) ([]byte, error) {
	if fileIdx < 0 || fileIdx >= len(rfm.files) {
		return nil, errors.New("invalid file index")
	}

	fileName := rfm.files[fileIdx]
	filePath := path.Join(rfm.Dir(), fileName)

	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("file not found")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (rfm *RotateFileManager) ReadFile(fileName string) ([]byte, error) {
	fileIdx := slices.IndexFunc(rfm.files, func(f string) bool { return strings.EqualFold(f, fileName) })

	return rfm.ReadFileByIdx(fileIdx)
}
