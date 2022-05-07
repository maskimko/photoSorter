package deduplicator

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type fileInfo struct {
	name string
	hash []byte
	size int64
	path string
}

type DuplicateError struct {
	msg string
}

func (d DuplicateError) Error() string {
	return d.msg
}

type deduplicator struct {
	file2hash map[string]*fileInfo
	hash2file map[string]*fileInfo
}

func (d deduplicator) addFile(file string) (*fileInfo, error) {
	fi, err := d.contains(file)
	if err != nil {
		return nil, err
	}
	if fi != nil {
		return fi, &DuplicateError{fmt.Sprintf("duplicate %s", fi.path)}
	}

}

func (d deduplicator) Contains(file string) (*fileInfo, error) {
	name := filepath.Base(file)
	if fi, ok := d.file2hash[name]; ok {
		return fi, nil
	}
	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	sum256 := sha256.Sum256(fileData)
	hash := base64.StdEncoding.EncodeToString(sum256[:])
	if fi, ok := d.hash2file[hash]; ok {
		return fi, nil
	}
	return nil, nil
}

func getFileInfo(file string) (*fileInfo, error) {
	name := filepath.Base(file)
	if fi, ok := d.file2hash[name]; ok {
		return fi, nil
	}
	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	sum256 := sha256.Sum256(fileData)
	hash := base64.StdEncoding.EncodeToString(sum256[:])

}
