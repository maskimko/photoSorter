package main

import (
	"photoSorter/pkg/metareader"
	"photoSorter/pkg/walker"
)

func main() {
	reader := metareader.NewDefaultExifReader()
	w := walker.NewWalker(reader)
	dst := "/home/maskimko/Sorted/Pictures/automation"
	src := "/home/maskimko/evacuation/maskimko/priority/Pictures"
	excludExt := []string{".gz", ".bz2", ".xz", ".tar"}
	w.Walk(src, dst, "500K", false, nil, excludExt)
}
