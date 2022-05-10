package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"os"
	"photoSorter/pkg/deduplicator"
	"photoSorter/pkg/metareader"
	"photoSorter/pkg/walker"
)

func main() {
	pflag.StringSliceP("sources", "s", nil, "source directories where to scan photos, can be repeated")
	pflag.StringP("dest", "d", "", "destination directory where to place sorted photos")
	pflag.StringP("filter-small", "F", "500K", "photos with size less than threshold in bytes  will be placed "+
		"into a separate directory. Available suffixes are K,M,G,T for kilobyte, megabyte, gigabyte, terabyte respectively")
	pflag.BoolP("move", "m", false, "if specified, photos will be moved, and originals will be removed")
	pflag.StringArray("exclude-dirs", nil, "exclude specified directories")
	pflag.StringArray("exclude-exts", []string{".gz", ".bz2", ".xz", ".tar", ".zip"}, "exclude files with given extensions")
	pflag.BoolP("skip-unknown", "U", false, "do not copy unknown files")
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatalln(err)
	}
	reader := metareader.NewDefaultExifReader()
	deduper := deduplicator.NewDeduplicator()
	var sources []string
	sources = viper.GetStringSlice("sources")
	if len(sources) == 0 {
		wd, wdErr := os.Getwd()
		if wdErr != nil {
			log.Fatalln("cannot use current working directory as a default source directory")
		}
		sources = append(sources, wd)
	}
	destination := viper.GetString("dest")
	if destination == "" {
		log.Fatalln("you must specify destination '-d' where to put sorted images")
	}
	sizeThreshold := viper.GetString("filter-small")
	threshold, err := walker.ConvertSize(sizeThreshold)
	if err != nil {
		log.Fatalf("wrong threshold size format %s", sizeThreshold)
	}
	registry, err := walker.NewFileRegistry(destination, threshold)
	if err != nil {
		log.Fatalf("failed to initialize registry %s", err)
	}
	defer registry.Close()
	w := walker.NewWalker(reader, deduper, registry)
	move := viper.GetBool("move")
	excludeDirs := viper.GetStringSlice("exclude-dirs")
	excludeExts := viper.GetStringSlice("exclude-exts")
	skipUnknown := viper.GetBool("skip-unknown")
	err = w.Walk(sources, destination, sizeThreshold, move, skipUnknown, excludeDirs, excludeExts)
	if err != nil {
		log.Fatalf("failed with an error %s", err)
	}
	log.Printf("Sucess! %d duplicates were found", w.DuplicatesCount())
}
