package walker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	b "github.com/cenkalti/backoff/v4"
)

type Registry interface {
	io.Closer
	Add(path string, picSize PicSize) error
}

type FileRegistry struct {
	files   map[PicSize]*os.File
	backoff b.BackOff
}

func getRegistryFilename(p PicSize) string {
	size := fmt.Sprintf("%s.txt", GetPhotoSizeName(p))
	return size
}

func (f FileRegistry) Close() error {
	for _, v := range f.files {
		if v != nil {
			err := v.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f FileRegistry) Add(path string, picSize PicSize) error {

	var fi os.FileInfo
	var err error
	rErr := b.Retry(func() error {
		fi, err = os.Stat(path)
		if err != nil {
			return err
		}
		return nil
	}, f.backoff)
	if rErr != nil {
		return err
	}
	if !fi.IsDir() {
		fName := f.files[picSize].Name()
		if _, pErr := os.Stat(fName); os.IsNotExist(pErr) {
			dErr := ensureDir(filepath.Dir(fName))
			if dErr != nil {
				return fmt.Errorf("failed to create directory for a registry file %s %s", fName, dErr)
			}
		}
		_, err = fmt.Fprintln(f.files[picSize], path)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewFileRegistry(dir string) (Registry, error) {
	fr := FileRegistry{}
	fr.files = make(map[PicSize]*os.File)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0775)
		if err != nil {
			return nil, err
		}
	}
	for i := -1; i <= 5; i++ {
		if i == 0 {
			continue
		}
		p := PicSize(i)
		f, err := os.OpenFile(filepath.Join(dir, getRegistryFilename(p)), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
		if err != nil {
			return nil, err
		}
		fr.files[p] = f
	}
	b := b.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(60) * time.Second // 60 seconds
	b.InitialInterval = time.Duration(3) * time.Second //3 seconds
	fr.backoff = b
	return &fr, nil
}
