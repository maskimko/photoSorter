package walker

import "strings"

type MediaType uint8

const (
	Unsupported = iota
	Photo
	Video
	Audio
)

var Ext2MediaType map[string]MediaType = map[string]MediaType{
	".jpg":  Photo,
	".jpeg": Photo,
	".png":  Photo,
	".heic": Photo,
	".bmp":  Photo,
	".tiff": Photo,
	".gif":  Photo,
	".mp4":  Video,
	".avi":  Video,
	".mkv":  Video,
	".mp3":  Audio,
	".aac":  Audio,
	".ogg":  Audio,
	".flac": Audio,
	".dsf":  Audio,
	".alac": Audio,
	".m4a":  Audio,
	".m4b":  Audio,
	".m4p":  Audio}

func (m MediaType) IsPhoto() bool {
	return m == Photo
}

func (m MediaType) IsVideo() bool {
	return m == Video
}

func (m MediaType) IsAudio() bool {
	return m == Audio
}

func (m MediaType) IsUnknown() bool {
	return m == Unsupported
}

func WhichMediaType(ext string) MediaType {
	if mt, ok := Ext2MediaType[strings.ToLower(ext)]; ok {
		return mt
	}
	return Unsupported
}
