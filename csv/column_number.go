package csv

import (
	"bufio"
	"io"
	"math/big"
	"os"
)

func GetColumnNumber(file io.Reader) (*big.Int, error) {
	reader := bufio.NewReader(file)
	var columnNumber big.Int
	for {
		chr, _, err := reader.ReadRune()
		if err != nil {
			return nil, err
		}
		if chr == ',' {
			columnNumber = *columnNumber.Add(&columnNumber, big.NewInt(1))
		} else if chr == '\n' || chr == '\r' {
			columnNumber = *columnNumber.Add(&columnNumber, big.NewInt(1))
			break
		}
	}
	return &columnNumber, nil
}

type ColumnNameReader struct {
	path   string
	reader *bufio.Scanner
}

func NewColumnReader(path string) (*ColumnNameReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewScanner(file)
	reader.Split(csvSplit)

	return &ColumnNameReader{path: path, reader: reader}, nil
}

func (r *ColumnNameReader) GetNext() (string, error) {
	if r.reader.Scan() {
		column := r.reader.Text()
		if column[len(column)-1] == '\n' {
			file, err := os.Open(r.path)
			if err != nil {
				return "", err
			}
			r.reader = bufio.NewScanner(file)
			r.reader.Split(csvSplit)
		}
		return trimCell(column), nil
	}
	return "", io.EOF
}
