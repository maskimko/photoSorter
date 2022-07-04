package main

import (
	"log"
	"os"
	"photoSorter/pkg/deduplicator"
	"photoSorter/pkg/metareader"
	"photoSorter/pkg/walker"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const SMALLISUNDER = "200k"

func main() {
	pflag.StringSliceP("sources", "s", nil, "source directories where to scan photos, can be repeated")
	pflag.StringP("dest", "d", "", "destination directory where to place sorted photos")
	pflag.StringP("filter-small", "F", "500K", "photos with size less than threshold in bytes  will be placed "+
		"into a separate directory. Available suffixes are K,M,G,T for kilobyte, megabyte, gigabyte, terabyte respectively")
	pflag.BoolP("move", "m", false, "if specified, photos will be moved, and originals will be removed")
	pflag.StringArray("exclude-dirs", nil, "exclude specified directories")
	pflag.StringArray("exclude-exts", []string{".gz", ".bz2", ".xz", ".tar", ".zip"}, "exclude files with given extensions")
	pflag.BoolP("skip-unsupported", "U", false, "do not copy unsupported files")
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

	registry, err := walker.NewFileRegistry(destination)
	if err != nil {
		log.Fatalf("failed to initialize registry %s", err)
	}
	defer registry.Close()
	w := walker.NewWalker(reader, deduper, registry, SMALLISUNDER)
	move := viper.GetBool("move")
	excludeDirs := viper.GetStringSlice("exclude-dirs")
	excludeExts := viper.GetStringSlice("exclude-exts")
	skipUnsupported := viper.GetBool("skip-unsupported")
	err = w.Walk(sources, destination, move, skipUnsupported, excludeDirs, excludeExts)
	if err != nil {
		log.Fatalf("failed with an error %s", err)
	}
	log.Printf("Sucess! %d duplicates were found", w.DuplicatesCount())
}
