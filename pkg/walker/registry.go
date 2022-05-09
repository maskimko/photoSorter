package walker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"photoSorter/pkg/metareader"
)

type Registry interface {
	io.Closer
	Add(path string, info *metareader.ExifMeta) error
}

type FileRegistry struct {
	threshold  int64
	smallFile  *os.File
	noDataFile *os.File
}

func (f FileRegistry) Close() error {
	if f.smallFile != nil {
		err := f.smallFile.Close()
		if err != nil {
			return err
		}
	}
	if f.noDataFile != nil {
		err := f.noDataFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f FileRegistry) Add(path string, info *metareader.ExifMeta) error {
	if info == nil || info.Unknown {
		_, err := fmt.Fprintln(f.noDataFile, path)
		if err != nil {
			return err
		}
	}
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() && fi.Size() < f.threshold {
		_, err = fmt.Fprintln(f.smallFile, path)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewFileRegistry(dir string, threshold int64) (Registry, error) {
	fr := FileRegistry{threshold: threshold}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0775)
		if err != nil {
			return nil, err
		}
	}
	noDataFile, err := os.OpenFile(filepath.Join(dir, "no_data_file.txt"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}
	smallFile, err := os.OpenFile(filepath.Join(dir, "small_file.txt"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}
	fr.noDataFile = noDataFile
	fr.smallFile = smallFile
	return &fr, nil
}
