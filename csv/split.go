package csv

import "bytes"

func csvSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	i := bytes.IndexByte(data, ',')
	j := bytes.IndexByte(data, '\n')

	if i == -1 && j != -1 {
		return j + 1, dropCR(data[0 : j+1]), nil
	} else if j == -1 && i != -1 {
		return i + 1, dropCR(data[0 : i+1]), nil
	} else if j != -1 && i != -1 && i <= j {
		return i + 1, dropCR(data[0 : i+1]), nil
	} else if j != -1 && i != -1 && j < i {
		return j + 1, dropCR(data[0 : j+1]), nil
	}

	if atEOF {
		return len(data), dropCR(data), nil
	}

	return 0, nil, nil
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
