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

var videoExts []string = []string{".mp4", ".avi", ".mkv"}
var picExts []string = []string{".jpg", ".jpeg", ".png", ".heic", ".bmp", ".tiff", ".gif"}

func convertSize(size string) (int64, error) {
	r := []rune(size)
	suffix := r[len(r)-1:][0]
	baseSizeStr := string(r[0 : len(r)-1])
	bs, err := strconv.Atoi(baseSizeStr)
	if err != nil {
		return -1, err
	}
	baseSize := int64(bs)
	switch suffix {
	case 'K':
		fallthrough
	case 'k':
		return baseSize * (1 << 10), nil
	case 'M':
		fallthrough
	case 'm':
		return baseSize * (2 << 10), nil
	case 'G':
		fallthrough
	case 'g':
		return baseSize * (3 << 10), nil
	case 'T':
		fallthrough
	case 't':
		return baseSize * (4 << 10), nil
	default:
		var s int
		s, err = strconv.Atoi(size)
		return int64(s), err
	}
}

type Walker struct {
	MetaReader   metareader.ExifReader
	deduplicator *deduplicator.Deduplicator
}

func (w Walker) Walk(source, dest, sizeThreshold string, move bool, excludeDir, excludeExt []string) error {
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

	threshold, err := convertSize(sizeThreshold)
	if err != nil {
		return err
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		var small bool
		if err != nil {
			log.Printf("%s walking error %s", path, err.Error())
		}
		if isExcluded(path, excludeDir, excludeExt) {
			log.Printf("excluded skipping %s", path)
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".thumbnails" {
				log.Printf("skipping %s thumbnails directory", path)
				return filepath.SkipDir
			}
			return nil
		}
		if isThumbnail(path) {
			log.Printf("skipping thumbnail %s", path)
			return nil
		}
		if info.Size() < threshold {
			small = true
		}
		x, err := w.MetaReader.ReadEXIF(path)
		if err != nil {
			log.Printf("%s exif reading error %s", path, err)
		}
		finalDir := getDestDir(x, path, dest, small)
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
				}
				log.Printf("failed to check duplicates of %s", path)
			}
		}
		err = w.processFile(path, finalDest, move)
		if err != nil {
			log.Printf("failed to process file %s error %s", path, err)
		}
		return nil
	})
	if err != nil {
		return err
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

func getDestDir(x *metareader.ExifMeta, file, dest string, small bool) string {
	destRoot := dest
	isPic := isPicture(file)
	isVid := isVideo(file)
	switch {
	case isPic:
		destRoot = filepath.Join(destRoot, "pictures")
	case isVid:
		destRoot = filepath.Join(destRoot, "video")
	default:
		destRoot = filepath.Join(destRoot, "others")
	}

	if x == nil {
		noDataDir := filepath.Join(dest, "no-data")
		ensureDir(noDataDir)
		return noDataDir
	}
	if isTrash(file) {
		trashDir := filepath.Join(dest, "trash")
		ensureDir(trashDir)
		return trashDir
	}
	if small {
		destRoot = filepath.Join(dest, "small")
	}
	if x.Unknown {
		unknownDir := filepath.Join(destRoot, "unknown")
		ensureDir(unknownDir)
		return unknownDir
	}
	if x == nil {
		log.Printf("no exif data for %s", file)
		return ""
	}
	finalDir := filepath.Join(destRoot, strconv.Itoa(x.Time.Year()), x.Time.Month().String(), x.Make, x.Model)
	ensureDir(finalDir)
	return finalDir
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
	return strings.Contains(path, ".dtrash")
}

func NewWalker(reader metareader.ExifReader, deduplicator *deduplicator.Deduplicator) Walker {
	return Walker{reader, deduplicator}
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

func isVideo(path string) bool {
	extension := filepath.Ext(path)
	for _, ext := range videoExts {
		if extension == strings.ToLower(ext) {
			return true
		}
	}
	return false
}

func isPicture(path string) bool {
	extension := filepath.Ext(path)
	for _, ext := range picExts {
		if extension == strings.ToLower(ext) {
			return true
		}
	}
	return false
}

func (w Walker) DuplicatesCount() int {
	return w.deduplicator.DuplicatesCount()
}
