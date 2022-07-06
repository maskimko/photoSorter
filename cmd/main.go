package main

import (
	"log"
	"os"
	"photoSorter/pkg/deduplicator"
	"photoSorter/pkg/walker"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	pflag.StringSliceP("sources", "s", nil, "source directories where to scan photos, can be repeated")
	pflag.StringP("dest", "d", "", "destination directory where to place sorted photos")
	pflag.BoolP("move", "m", false, "if specified, photos will be moved, and originals will be removed")
	pflag.StringArray("exclude-dirs", nil, "exclude specified directories")
	pflag.StringArray("exclude-exts", []string{".gz", ".bz2", ".xz", ".tar", ".zip"}, "exclude files with given extensions")
	pflag.BoolP("skip-unsupported", "U", false, "do not copy unsupported files")
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatalln(err)
	}
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
	pw, err := walker.NewDefaultPhotoWorker()
	w := walker.NewWalker(pw, deduper, registry)
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
