package runtime

import (
	"os"
	"sync"
)

var fileMap = make(map[string]*File)

type File struct {
	ID      string
	File    *os.File
	mux     sync.Mutex
	ActiveN int
}

func (f *File) Add() {
	f.ActiveN++
}

func (f *File) Done() {
	f.mux.Lock()
	defer f.mux.Unlock()
	f.ActiveN--
	if f.ActiveN <= 0 {
		f.File.Close()
		fileMap[f.ID] = nil
	}
}

func LookupFile(key string) (*File, bool) {
	f, ok := fileMap[key]
	return f, ok && f != nil && f.File != nil
}

func RegisterFile(key string, f *os.File) *File {
	file := &File{
		ID:      key,
		File:    f,
		ActiveN: 1,
	}
	fileMap[key] = file
	return file
}
