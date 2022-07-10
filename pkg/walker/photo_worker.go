package walker

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"photoSorter/pkg/metareader"
	"strconv"
)

//Available suffixes are K,M,G,T for kilobyte, megabyte, gigabyte, terabyte respectively
const SMALLISUNDER = "200k"
const NOTIME = "unknown_time"
const (
	TINYRES  = 384
	THUMBRES = 256
	SMALLRES = 768
	MEDIRES  = 2560
)

type PicSize int

const (
	Unknown           = -1
	Thumbnail PicSize = iota
	Tiny
	Small
	Medium
	Large
)

type PhotoWorker struct {
	mReader       metareader.ExifReader
	sizeThreshold int64
}

func NewDefaultPhotoWorker() (*PhotoWorker, error) {
	threshold, err := ConvertSize(SMALLISUNDER)
	if err != nil {
		log.Printf("cannot convert size %s %s", SMALLISUNDER, err)
		return nil, err
	}
	return NewPhotoWorker(threshold), nil
}

func NewPhotoWorker(sizeThreshold int64) *PhotoWorker {
	return &PhotoWorker{
		mReader:       metareader.NewDefaultExifReader(),
		sizeThreshold: sizeThreshold,
	}
}

func GetPhotoSizeName(ps PicSize) string {
	switch ps {
	case Thumbnail:
		return "thumbnail"
	case Tiny:
		return "tiny"
	case Small:
		return "small"
	case Medium:
		return "medium"
	case Large:
		return "large"
	default:
		return "unknown"
	}
}

func GetPicSize(filePath string) (PicSize, error) {
	pw, err := NewDefaultPhotoWorker()
	if err != nil {
		return -1, err
	}
	info, err := os.Stat(filePath)
	if err != nil {
		log.Printf("failed to get file stat %s", err)
		return -1, err
	}
	ps, _, err := pw.GetPicInfo(filePath, info)
	return ps, err
}

func (pw PhotoWorker) GetPicInfo(filePath string, fileInfo os.FileInfo) (PicSize, *metareader.ExifMeta, error) {
	x, err := pw.mReader.ReadEXIF(filePath)
	if err != nil {
		log.Printf("%s exif reading error %s", filePath, err)
		return 0, nil, err
	}
	if x == nil {
		return -1, nil, fmt.Errorf("no EXIF data for %s", filePath)
	}
	picSize := pw.whichSize(x, fileInfo.Size())
	return picSize, x, nil
}

func (pw PhotoWorker) whichSize(x *metareader.ExifMeta, size int64) PicSize {
	if x.Width > MEDIRES && x.Height > MEDIRES {
		return Large
	}
	if x.Width > SMALLRES && x.Height > SMALLRES {
		return Medium
	}
	if x.Height > TINYRES && x.Width > TINYRES {
		return Small
	}
	if x.Height <= TINYRES && x.Height == x.Width {
		return Thumbnail
	}
	if x.Height <= 0 || x.Width <= 0 {
		//If cannot read the picture data consider size of the file then
		if size < pw.sizeThreshold {
			return Small
		}
		return Unknown
	}
	return Tiny
}

func (pw PhotoWorker) GetPath(filePath string, fileInfo os.FileInfo) (string, PicSize, error) {
	pictures := filepath.Join("pictures")
	picSize, x, err := pw.GetPicInfo(filePath, fileInfo)
	if err != nil {
		log.Printf("failed to measure picture %s size %s", filePath, err)
		return "no-data", Unknown, err
	}
	if x.Unknown {
		unknownDir := filepath.Join(pictures, "unknown")
		ensureDir(unknownDir)
		return unknownDir, Unknown, nil
	}
	pictures = filepath.Join(pictures, GetPhotoSizeName(picSize))
	var finalDir string
	if x.Time == nil {
		finalDir = filepath.Join(pictures, NOTIME, x.Format, x.Make, x.Model)
	} else {
		finalDir = filepath.Join(pictures, strconv.Itoa(x.Time.Year()), x.Time.Month().String(), x.Format, x.Make, x.Model)
	}
	return finalDir, picSize, err
}
