package walker

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"photoSorter/pkg/deduplicator"
	"strings"
)

type Walker struct {
	pw           *PhotoWorker
	deduplicator *deduplicator.Deduplicator
	registry     Registry
}

func (w Walker) Walk(sources []string, dest string, move, skipUnsupported bool, excludeDir, excludeExt []string) error {
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
				if os.IsNotExist(err) {
					return err
				}
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

			finalDir, skip := w.getDestDir(path, dest, info)
			if skip {
				log.Printf("skip %s", path)
				return nil
			}
			if finalDir == "" {
				log.Printf("no exif data for %s, skipping...", path)
				return nil
			}
			err = ensureDir(finalDir)
			if err != nil {
				log.Printf("failed to create a destination directory %s", err)
				return err
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

func (w Walker) getDestDir(file, dest string, info os.FileInfo) (string, bool) {
	destRoot := dest
	if isTrash(file) {
		trashDir := filepath.Join(dest, "trash")
		ensureDir(trashDir)
		return trashDir, false
	}

	switch WhichMediaType(filepath.Ext(file)) {
	case Photo:
		picPath, picSize, err := w.pw.GetPath(file, info)
		if err != nil {
			log.Printf("skipping %s error: %s", file, err)
			return filepath.Join(destRoot, "others"), true
		}
		finalDest := filepath.Join(dest, picPath)
		go func() {
			err = w.registry.Add(finalDest, picSize)
			if err != nil {
				log.Printf("failed to register file %s error %s", finalDest, err)
			}
		}()
		return finalDest, false
	case Audio:
		log.Printf("audio processing is not implemented")
		destRoot = filepath.Join(destRoot, "audio")
	case Video:
		log.Printf("audio processing is not implemented")
		destRoot = filepath.Join(destRoot, "video")
	case Unsupported:
		log.Printf("unsupported file %s goes to others", file)
		fallthrough
	default:
		destRoot = filepath.Join(destRoot, "others")
	}

	//TODO: for audio and video here should be a different paths
	return destRoot, false
}

func ensureDir(finalDir string) error {
	if _, err := os.Stat(finalDir); os.IsNotExist(err) {
		err = os.MkdirAll(finalDir, 0755)
		if err != nil {
			log.Printf("failed to create directory %s", finalDir)
			return err
		}
	} else {
		if err != nil {
			return err
		}
	}
	return nil
}

func isThumbnail(path string) bool {
	return strings.Contains(path, ".thumbnail") || strings.Contains(path, ".videoThumbnail")
}

func isTrash(path string) bool {
	return strings.Contains(path, ".dtrash") || strings.Contains(path, ".trash")
}

func NewWalker(pw *PhotoWorker, deduplicator *deduplicator.Deduplicator, registry Registry) Walker {

	return Walker{pw, deduplicator, registry}
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
