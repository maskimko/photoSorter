package walker

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"photoSorter/pkg/deduplicator"
	"photoSorter/pkg/metareader"
	"strconv"
	"strings"
)

type Walker struct {
	MetaReader   metareader.ExifReader
	deduplicator *deduplicator.Deduplicator
	registry     Registry
	sizeTreshold int64
}

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

func GetSizeName(ps PicSize) string {
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

func (w Walker) whichSize(x *metareader.ExifMeta, size int64) PicSize {
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
		if size < w.sizeTreshold {
			return Small
		}
		return Unknown
	}
	return Tiny
}

func (w Walker) Walk(sources []string, dest string, move,
	skipUnsupported bool, excludeDir, excludeExt []string) error {
	destStat, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dest, 0755)
			if err != nil {
				return fmt.Errorf("failed to create destination dir %w", err)
			}
		} else {
			return err
		}
	} else {
		if !destStat.IsDir() {
			return fmt.Errorf("destination path %s must be a directory", dest)
		}
	}

	for _, source := range sources {
		log.Printf("processing source directory %s", source)
		err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("%s walking error %s", path, err.Error())
			}
			if isExcluded(path, excludeDir, excludeExt) {
				log.Printf("excluded skipping %s", path)
				return nil
			}
			if info.IsDir() {
				if info.Name() == ".thumbnails" || info.Name() == ".videoThumbnails" {
					log.Printf("skipping %s thumbnails directory", path)
					return filepath.SkipDir
				}
				return nil
			}
			if isThumbnail(path) {
				log.Printf("skipping thumbnail %s", path)
				return nil
			}
			x, err := w.MetaReader.ReadEXIF(path)
			if err != nil {
				log.Printf("%s exif reading error %s", path, err)
			}
			picSize := w.whichSize(x, info.Size())
			finalDir, skip := w.getDestDir(x, path, dest, picSize, skipUnsupported)
			if skip {
				return nil
			}
			if finalDir == "" {
				log.Printf("no exif data for %s, skipping...", path)
				return nil
			}
			finalDest := filepath.Join(finalDir, info.Name())
			if !isTrash(path) {
				fileInfo, err := w.deduplicator.AddFile(path)
				if err != nil {
					if _, ok := err.(deduplicator.DuplicateError); ok {
						log.Printf("file %s is a duplicate of %s, which has already been processed", path, fileInfo.Path)
						return nil
					} else {
						log.Printf("failed to check duplicates of %s", path)
					}
				}
			}
			err = w.processFile(path, finalDest, move)

			if err != nil {
				log.Printf("failed to process file %s error %s", path, err)
			}
			go func() {
				err = w.registry.Add(finalDest, picSize)
				if err != nil {
					log.Printf("failed to register file %s error %s", finalDest, err)
				}
			}()
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (w Walker) processFile(src, dst string, move bool) error {
	sFile, err := os.Open(src)
	if err != nil {
		log.Printf("failed to copy file %s %s", src, err)
		return err
	}
	defer sFile.Close()
	dFile, err := os.Create(dst)
	if err != nil {
		log.Printf("failed to copy file to %s %s", dst, err)
		return err
	}
	defer dFile.Close()
	_, err = io.Copy(dFile, sFile)
	if err != nil {
		log.Printf("failed to copy file from %s to %s  %s", src, dst, err)
		return err
	}
	if move {
		err = os.Remove(src)
		if err != nil {
			log.Printf("failed to remove source file %s %s", src, err)
			return err
		}
	}
	return nil
}

func (w Walker) getDestDir(x *metareader.ExifMeta, file, dest string, picSize PicSize, skipUnsupported bool) (string, bool) {
	destRoot := dest
	switch WhichMediaType(filepath.Ext(file)) {
	case Photo:
		destRoot = filepath.Join(destRoot, "pictures")
	case Audio:
		log.Printf("audio processing is not implemented")
		destRoot = filepath.Join(destRoot, "audio")
	case Video:
		log.Printf("audio processing is not implemented")
		destRoot = filepath.Join(destRoot, "video")
	case Unsupported:
		if skipUnsupported {
			return filepath.Join(destRoot, "others"), true
		}
		fallthrough
	default:
		destRoot = filepath.Join(destRoot, "others")
	}

	if x == nil {
		noDataDir := filepath.Join(dest, "no-data")
		ensureDir(noDataDir)
		return noDataDir, false
	}
	if isTrash(file) {
		trashDir := filepath.Join(dest, "trash")
		ensureDir(trashDir)
		return trashDir, false
	}
	destRoot = filepath.Join(GetSizeName(picSize))

	if x.Unknown {
		unknownDir := filepath.Join(destRoot, "unknown")
		ensureDir(unknownDir)
		return unknownDir, false
	}
	if x == nil {
		log.Printf("no exif data for %s", file)
		return "", skipUnsupported
	}
	//TODO: for audio and video here should be a different paths
	finalDir := filepath.Join(destRoot, strconv.Itoa(x.Time.Year()), x.Time.Month().String(), x.Format, x.Make, x.Model)
	ensureDir(finalDir)
	return finalDir, false
}

func ensureDir(finalDir string) {
	if _, err := os.Stat(finalDir); os.IsNotExist(err) {
		err = os.MkdirAll(finalDir, 0755)
		if err != nil {
			log.Printf("failed to create directory %s", finalDir)
		}
	}
}

func isThumbnail(path string) bool {
	return strings.Contains(path, ".thumbnail") || strings.Contains(path, ".videoThumbnail")
}

func isTrash(path string) bool {
	return strings.Contains(path, ".dtrash") || strings.Contains(path, ".trash")
}

func NewWalker(reader metareader.ExifReader, deduplicator *deduplicator.Deduplicator, registry Registry, sizeThreshold string) Walker {
	threshold, err := ConvertSize(sizeThreshold)
	if err != nil {
		log.Fatalf("cannot convert size %s %s", sizeThreshold, err)
	}
	return Walker{reader, deduplicator, registry, threshold}
}

func isExcluded(path string, dirs, extensions []string) bool {
	for _, d := range dirs {
		if path == d || filepath.Dir(path) == d {
			return true
		}
	}
	for _, ext := range extensions {
		if filepath.Ext(path) == ext {
			return true
		}
	}
	return false
}

func (w Walker) DuplicatesCount() int {
	return w.deduplicator.DuplicatesCount()
}
