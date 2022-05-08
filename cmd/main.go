package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"photoSorter/pkg/deduplicator"
	"photoSorter/pkg/metareader"
	"photoSorter/pkg/walker"
)

func main() {
	pflag.StringP("source", "s", ".", "source directory where to scan photos")
	pflag.StringP("dest", "d", "", "destination directory where to place sorted photos")
	pflag.StringP("small", "S", "500K", "photos with size less than threshold in bytes  will be placed "+
		"into a separate directory. Available suffixes are K,M,G,T for kilobyte, megabyte, gigabyte, terabyte respectively")
	pflag.BoolP("move", "m", false, "if specified, photos will be moved, and originals will be removed")
	pflag.StringArray("exclude-dirs", nil, "exclude specified directories")
	pflag.StringArray("exclude-exts", []string{".gz", ".bz2", ".xz", ".tar", ".zip"}, "exclude files with given extensions")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	reader := metareader.NewDefaultExifReader()
	deduper := deduplicator.NewDeduplicator()
	w := walker.NewWalker(reader, deduper)
	source := viper.GetString("source")
	destination := viper.GetString("dest")
	if destination == "" {
		log.Fatalln("you must specify destination '-d' where to put sorted images")
	}
	move := viper.GetBool("move")
	excludeDirs := viper.GetStringSlice("exclude-dirs")
	excludeExts := viper.GetStringSlice("exclude-exts")
	sizeThreshold := viper.GetString("small")
	err := w.Walk(source, destination, sizeThreshold, move, excludeDirs, excludeExts)
	if err != nil {
		log.Fatalf("failed with an error %s", err)
	}
	log.Printf("Sucess! %d duplicates were found", w.DuplicatesCount())
}
