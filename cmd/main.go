package main

import (
	"log"
	"photoSorter/pkg/deduplicator"
	"photoSorter/pkg/metareader"
	"photoSorter/pkg/walker"
)

func main() {
	reader := metareader.NewDefaultExifReader()
	deduper := deduplicator.NewDeduplicator()
	w := walker.NewWalker(reader, deduper)
	dst := "/home/maskimko/Sorted/Pictures/automation"
	src := "/home/maskimko/evacuation/maskimko/priority/Pictures"
	excludExt := []string{".gz", ".bz2", ".xz", ".tar"}
	err := w.Walk(src, dst, "500K", false, nil, excludExt)
	if err != nil {
		log.Fatalf("failed with an error %s", err)
	}
	log.Printf("Sucess! %d duplicates were found", w.DuplicatesCount())
}
