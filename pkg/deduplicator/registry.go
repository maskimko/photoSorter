package deduplicator

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type fileInfo struct {
	Name           string
	Hash           []byte
	Size           int64
	Path           string
	DuplicatePaths []string
}

type DuplicateError struct {
	msg string
}

func (d DuplicateError) Error() string {
	return d.msg
}

type Deduplicator struct {
	file2hash       map[string]*fileInfo
	hash2file       map[string]*fileInfo
	duplicatesCount int
}

func (d Deduplicator) DuplicatesCount() int {
	return d.duplicatesCount
}

func (d Deduplicator) AddFile(file string) (*fileInfo, error) {
	fi, err := d.Contains(file)
	if err != nil {
		return nil, err
	}
	if fi != nil {
		absPath, absErr := filepath.Abs(file)
		if absErr == nil {
			if absPath != fi.Path {
				exists := false
				for _, p := range fi.DuplicatePaths {
					if p == absPath {
						exists = true
					}
				}
				if !exists {
					d.duplicatesCount++
					fi.DuplicatePaths = append(fi.DuplicatePaths, absPath)
				}
			}
		} else {
			log.Printf("failed to check %s file for duplicates", file)
		}
		return fi, &DuplicateError{fmt.Sprintf("duplicate %s", fi.Path)}
	}
	fi, err = d.getFileInfo(file)
	if err != nil {
		return nil, err
	}
	d.file2hash[fi.Name] = fi
	hash := base64.StdEncoding.EncodeToString(fi.Hash)
	d.hash2file[hash] = fi
	return fi, nil
}

func (d Deduplicator) Contains(file string) (*fileInfo, error) {
	info, err := d.getFileInfo(file)
	if err != nil {
		return nil, err
	}
	hash := base64.StdEncoding.EncodeToString(info.Hash)
	if fi, ok := d.hash2file[hash]; ok {
		return fi, nil
	}
	return nil, nil
}

func (d Deduplicator) getFileInfo(file string) (*fileInfo, error) {
	name := filepath.Base(file)
	if fi, ok := d.file2hash[name]; ok {
		return fi, nil
	}
	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	sum256 := sha256.Sum256(fileData)
	fileStats, err := os.Stat(file)
	if err != nil {
		return nil, err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	fi := fileInfo{Name: name, Path: path, Hash: sum256[:], Size: fileStats.Size()}
	return &fi, nil
}

func NewDeduplicator() *Deduplicator {
	f2h := make(map[string]*fileInfo)
	h2f := make(map[string]*fileInfo)
	return &Deduplicator{
		file2hash:       f2h,
		hash2file:       h2f,
		duplicatesCount: 0,
	}
}
