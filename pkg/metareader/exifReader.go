package metareader

import (
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"log"
	"os"
	"strings"
	"time"
)

type GPSCoordinates struct {
	Longitude float32
	Latitude  float32
	Accuracy  float32
}

type ExifMeta struct {
	Make        string
	Model       string
	Time        time.Time
	Coordinates *GPSCoordinates
	Records     map[string]string
	Unknown     bool
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
		log.Printf("EXIF decode error %s", err)
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
	cameraMake, err := x.Get(exif.Make)
	if err != nil {
		return nil, err
	}
	model, err := x.Get(exif.Model)
	if err != nil {
		return nil, err
	}
	tm, err := x.DateTime()
	if err != nil {
		return nil, err
	}
	var coordinates *GPSCoordinates
	lat, long, err := x.LatLong()
	if err != nil {
		log.Printf("position of %s is not available error %s", file.Name(), err)
	} else {
		coordinates = &GPSCoordinates{Latitude: float32(lat), Longitude: float32(long)}
	}
	makeStr := strings.TrimSpace(cameraMake.String())
	if makeStr == "" {
		makeStr = "no_make"
	}
	modelStr := strings.TrimSpace(model.String())
	if modelStr == "" {
		modelStr = "no_model"
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
	return ReadEXIFData(f)
}

func NewDefaultExifReader() ExifReader {
	return ExifReaderImpl{}
}
