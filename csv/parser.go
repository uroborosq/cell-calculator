package csv

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

type CsvParser struct {
	path string
	file *os.File
	rowsNumber int
	bufferSize uint64
	reader *bufio.Reader
	isFirstLineRead bool
}

func NewCsvParserWithBuffer(path string, bufferSize uint64) (*CsvParser, error) {

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return &CsvParser{
		path: path,
		bufferSize: bufferSize,
	}, nil
}

func NewCsvParser(path string) (*CsvParser, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return &CsvParser{
		path: path,
		bufferSize: 1024,
	}, nil
}

func (p *CsvParser) Init() error {
	var err error
	p.file, err = os.Open(p.path)
	if err != nil {
		return err
	}
	p.reader = bufio.NewReader(p.file)
	buffer := make([]byte, p.bufferSize)
	for i := uint64(0); i < p.bufferSize; i++ {
		s := strings.Builder{}
		if buffer[i] == ',' {
			p.rowsNumber++
		} else if buffer[i] == ' ' {
			continue
		} else {
			s.WriteByte(buffer[i])
		}
	}
	
	return err
}

func (p *CsvParser) GetRowPart(row string) (map[string]string, bool, error) {
	buffer := make([]byte, p.bufferSize)
	readBytes, err := p.reader.Read(buffer)
	if err != nil {
		return nil, false, err
	}
	buffer = buffer[:readBytes]
	
	for i := 0; i < readBytes; i++ {
		s := strings.Builder{}
		if buffer[i] == ',' {

		} else {
			s.WriteByte(buffer[i])
		}
	}

	return nil, false, nil
}

func (p *CsvParser) Close() error {
	return p.file.Close()
}


