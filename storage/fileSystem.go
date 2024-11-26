package storage

import (
	"io"
	"os"
	"sync"
)

var (
	fs      *fileSystem
	oneTime sync.Once
)

const (
	filePath = "./tmp/TSXhistory.data"
)

type fileSystem struct {
	filePtr *os.File
	path    string
}

var _ FileSystemInterface = &fileSystem{}

func GetFileSystem() (FileSystemInterface, error) {
	var err error
	oneTime.Do(func() {
		fs = &fileSystem{path: filePath}
		fs.filePtr, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			fs.filePtr = nil
		}
	})
	return fs, err
}

func (f *fileSystem) WriteToFile(data []byte) error {
	_, err := f.filePtr.Write(data)
	return err
}

func (f *fileSystem) ReadFromFile() ([]byte, error) {
	_, err := f.filePtr.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(f.filePtr)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *fileSystem) Close() error {
	return f.filePtr.Close()
}
