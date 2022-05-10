package walker

import (
	"errors"
	"fmt"
	"strconv"
)

func ConvertSize(size string) (int64, error) {
	if size == "" {
		return -1, errors.New("empty size")
	}
	if len(size) == 1 {
		s, err := strconv.Atoi(size)
		if err != nil {
			return -1, err
		}
		return int64(s), nil
	}
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
		if suffix >= '0' && suffix <= '9' {
			var s int
			s, err = strconv.Atoi(size)
			return int64(s), err
		}
		return -1, fmt.Errorf("wrong size %s", size)
	}
}
