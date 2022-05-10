photoSorter
======


## Description

This is a simple tool to organise photos and sort them by date, camera make and model


## How to use
```
Usage of ./photo-sorter:
  -d, --dest string                destination directory where to place sorted photos
      --exclude-dirs stringArray   exclude specified directories
      --exclude-exts stringArray   exclude files with given extensions (default [.gz,.bz2,.xz,.tar,.zip])
  -F, --filter-small string        photos with size less than threshold in bytes  will be placed into a separate directory. Available suffixes are K,M,G,T for kilobyte, megabyte, gigabyte, terabyte respectively (default "500K")
  -m, --move                       if specified, photos will be moved, and originals will be removed
  -s, --sources strings            source directories where to scan photos, can be repeated
```
