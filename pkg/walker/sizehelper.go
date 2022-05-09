package walker

import "strconv"

func ConvertSize(size string) (int64, error) {
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
		var s int
		s, err = strconv.Atoi(size)
		return int64(s), err
	}
}
