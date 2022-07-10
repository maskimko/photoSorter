package metareader

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

const (
	NOMAKE  = "unknown_make"
	NOMODEL = "unknown_model"
)

type GPSCoordinates struct {
	Longitude float32
	Latitude  float32
	Accuracy  float32
}

type ExifMeta struct {
	Make        string
	Model       string
	Time        *time.Time
	Coordinates *GPSCoordinates
	Records     map[string]string
	Unknown     bool
	Thumbnail   bool
	Height      int
	Width       int
	Format      string
}

type ExifReader interface {
	ReadEXIF(file string) (*ExifMeta, error)
}

type ExifAccumulator struct {
	Data map[exif.FieldName]*tiff.Tag
}

func NewExifAccumulator() ExifAccumulator {
	data := make(map[exif.FieldName]*tiff.Tag)
	return ExifAccumulator{data}
}

func (e ExifAccumulator) Walk(name exif.FieldName, tag *tiff.Tag) error {
	e.Data[name] = tag
	return nil
}

func (e ExifAccumulator) Meta() map[string]string {
	meta := make(map[string]string)
	for k, v := range e.Data {
		meta[string(k)] = v.String()
	}
	return meta
}

func ReadEXIFData(file *os.File) (*ExifMeta, error) {
	x, err := exif.Decode(file)
	if err != nil {
		if err.Error() != "EOF" {
			log.Printf("EXIF decode error %s %s", file.Name(), err)
		}
	}
	if x == nil {
		if err.Error() == "EOF" {
			return &ExifMeta{Unknown: true}, nil
		}
		return nil, fmt.Errorf("failed to decode EXIF data of %s", file.Name())
	}
	accum := NewExifAccumulator()
	err = x.Walk(accum)
	if err != nil {
		return nil, err
	}
	var cameraMake *tiff.Tag
	cameraMake, err = x.Get(exif.Make)
	if err != nil {
		if tErr, ok := err.(exif.TagNotPresentError); ok {
			log.Printf("cannot determine camera make of %s %s", file.Name(), tErr.Error())
			cameraMake = nil
		} else {
			return nil, err
		}
	}
	var model *tiff.Tag
	model, err = x.Get(exif.Model)
	if err != nil {
		if tErr, ok := err.(exif.TagNotPresentError); ok {
			log.Printf("cannot determine camera model of %s %s", file.Name(), tErr.Error())
			model = nil
		} else {
			return nil, err
		}
	}
	var makeStr string
	if cameraMake == nil {
		makeStr = NOMAKE
	} else {
		makeStr := strings.ReplaceAll(strings.Trim(strings.TrimSpace(cameraMake.String()), "\"'"), " ", "_")
		if makeStr == "" {
			makeStr = NOMAKE
		}
	}
	var modelStr string
	if model == nil {
		modelStr = NOMODEL
	} else {
		modelStr = strings.ReplaceAll(strings.Trim(strings.TrimSpace(model.String()), "\"'"), " ", "_")
		if modelStr == "" {
			modelStr = NOMODEL
		}
	}
	var tm *time.Time
	pictureTime, err := x.DateTime()
	tm = &pictureTime
	if err != nil {
		if tErr, ok := err.(exif.TagNotPresentError); ok {
			log.Printf("cannot determine creation time of %s %s", file.Name(), tErr.Error())
			tm = nil
		} else {
			return nil, err
		}
	}
	var coordinates *GPSCoordinates
	lat, long, err := x.LatLong()
	//Suppress the GPS EXIF errors
	//if err != nil {
	//	log.Printf("position of %s is not available error %s", file.Name(), err)
	//} else {
	//	coordinates = &GPSCoordinates{Latitude: float32(lat), Longitude: float32(long)}
	//}
	if err == nil {
		coordinates = &GPSCoordinates{Latitude: float32(lat), Longitude: float32(long)}
	}

	exifMeta := ExifMeta{
		Make:        makeStr,
		Model:       modelStr,
		Time:        tm,
		Coordinates: coordinates,
	}
	return &exifMeta, nil
}

type ExifReaderImpl struct {
}

func (e ExifReaderImpl) ReadEXIF(file string) (*ExifMeta, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	data, err := ReadEXIFData(f)
	if err != nil {
		return nil, err
	}
	h, w, format, err := e.decode(f)
	if err != nil {
		return data, err
	}
	data.Height = h
	data.Width = w
	data.Format = format
	return data, nil
}

func (e ExifReaderImpl) decode(file *os.File) (int, int, string, error) {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return -1, -1, "", err
	}
	ext := strings.ToLower(filepath.Ext(file.Name()))
	switch ext {
	case ".jpeg":
		fallthrough
	case ".jpg":
		img, err := jpeg.Decode(file)
		if err != nil {
			return -1, -1, "err", err
		}
		return img.Bounds().Dy(), img.Bounds().Dx(), "jpeg", nil
	case ".png":
		img, err := png.Decode(file)
		if err != nil {
			return -1, -1, "err", err
		}
		return img.Bounds().Dy(), img.Bounds().Dx(), "png", nil
	default:
		return -1, -1, "not_implemented" + ext, nil
	}

}

func NewDefaultExifReader() ExifReader {
	return ExifReaderImpl{}
}
