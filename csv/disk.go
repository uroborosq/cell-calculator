package csv

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type CsvParser struct {
	path            string
	file            *os.File
	columnNumber    uint64
	bufferSize      uint64
	reader          *bufio.Reader
	buffer          rune
	stringBuilder   strings.Builder
	stringBuffer    string
}

func NewCsvParserWithBuffer(path string, bufferSize uint64) (*CsvParser, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return &CsvParser{
		path:       path,
		bufferSize: bufferSize,
	}, nil
}

func NewCsvParser(path string) (*CsvParser, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return &CsvParser{
		path:       path,
		bufferSize: 1024,
	}, nil
}

func (p *CsvParser) Init() error {
	var err error
	p.file, err = os.Open(p.path)
	if err != nil {
		return err
	}
	p.columnNumber, err = GetColumnNumber(p.reader)
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

func (p *CsvParser) GetNextCell() (string, error) {
	var err error
	for {
		p.buffer, _, err = p.reader.ReadRune()
		if err != nil {
			return "", err
		}
		if p.buffer == ',' || p.buffer == '\n' || p.buffer == '\r' {
			break
		} else {
			p.stringBuilder.WriteRune(p.buffer)
		}
	}
	p.stringBuffer = p.stringBuilder.String()
	if IsNumber(p.stringBuffer) {
		return p.stringBuffer, nil
	} else {
		var firstColumn strings.Builder
		var firstRow strings.Builder
		var secondColumn strings.Builder
		var secondRow strings.Builder
		var operator rune
		var isFirstAddressParsed bool
		for _, r := range p.stringBuffer {
			if unicode.IsDigit(r) {
				if isFirstAddressParsed {
					firstRow.WriteRune(r)
				} else {
					secondRow.WriteRune(r)
				}
				firstColumn.WriteRune(r)
			} else if r == ' ' {
				continue
			} else if IsOperator(r) {
				operator = r
				isFirstAddressParsed = true
			} else {
				if isFirstAddressParsed {
					firstColumn.WriteRune(r)
				} else {
					secondColumn.WriteRune(r)
				}
			}
		}
		return p.Calculate(firstRow.String(), firstColumn.String(), secondRow.String(), secondColumn.String(), operator)
	}
}

// must be faster than strconv.Atoi
func IsNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func IsOperator(r rune) bool {
	if r == '+' || r == '-' || r == '*' || r == '/' {
		return true
	}
	return false
}

func ParseArgs(cell string) (string, string, string, string, rune) {
	var firstColumn strings.Builder
	var firstRow strings.Builder
	var secondColumn strings.Builder
	var secondRow strings.Builder
	var operator rune
	var isFirstAddressParsed bool
	for _, r := range cell {
		if unicode.IsDigit(r) {
			if isFirstAddressParsed {
				firstRow.WriteRune(r)
			} else {
				secondRow.WriteRune(r)
			}
			firstColumn.WriteRune(r)
		} else if r == ' ' {
			continue
		} else if IsOperator(r) {
			operator = r
			isFirstAddressParsed = true
		} else {
			if isFirstAddressParsed {
				firstColumn.WriteRune(r)
			} else {
				secondColumn.WriteRune(r)
			}
		}
	}
	return firstRow.String(), firstColumn.String(), secondRow.String(), secondColumn.String(), operator
}

func (p *CsvParser) Calculate(firstRow string, firstColunm string, secondRow string, secondColunm string, operator rune) (string, error) {
	firstStr, secondStr, err := p.GetCellsValue(firstRow, firstColunm, secondRow, secondColunm)
	if err != nil {
		return "", err
	}
	if !IsNumber(firstStr) {
		firstRow, firstColumn, secondRow, secondColumn, operator := ParseArgs(firstStr)
		firstStr, err = p.Calculate(firstRow, firstColumn, secondRow, secondColumn, operator)
		if err != nil {
			return "", err
		}
	} else if !IsNumber(secondStr) {
		firstRow, firstColumn, secondRow, secondColumn, operator := ParseArgs(firstStr)
		firstStr, err = p.Calculate(firstRow, firstColumn, secondRow, secondColumn, operator)
		if err != nil {
			return "", err
		}
	}
	first, err := strconv.ParseFloat(firstStr, 64)
	if err != nil {
		return "", err
	}
	second, err := strconv.ParseFloat(secondStr, 64)
	if err != nil {
		return "", err
	}
	switch operator {
	case '+':
		return strconv.FormatFloat(first+second, 'g', 1, 64), nil
	case '-':
		return strconv.FormatFloat(first-second, 'g', 1, 64), nil
	case '*':
		return strconv.FormatFloat(first*second, 'g', 1, 64), nil
	case '/':
		return strconv.FormatFloat(first/second, 'g', 1, 64), nil
	}

	return "", errors.New("operator is not supported")
}

func (p *CsvParser) GetCellsValue(firstRow string, firstColunm string, secondRow string, secondColunm string) (string, string, error) {
	file, err := os.Open(p.path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	var buffer rune
	var builder strings.Builder
	var firstColumnIndex uint64
	var secondColumnIndex uint64
	var currentIndex uint64
	var firstValue string
	var secondValue string
	for {
		buffer, _, err = reader.ReadRune()
		if err != nil {
			return "", "", err
		}
		if buffer == '\r' {
			_, _, err = reader.ReadRune()
			if err != nil {
				return "", "", err
			}
			break
		} else if buffer == '\n' {
			break
		} else if buffer == ',' {
			name := builder.String()
			if name == firstColunm {
				firstColumnIndex = currentIndex
			} else if name == secondColunm {
				secondColumnIndex = currentIndex
			}
			builder.Reset()
		} else {
			builder.WriteRune(buffer)
		}
		currentIndex++
	}
	builder.Reset()
	var valuesFound int
	for {
		var currentRow string
		currentIndex = 0
		for {
			buffer, _, err = reader.ReadRune()
			if err != nil {
				return "", "", err
			}
			if buffer == ',' {
				break
			} else {
				builder.WriteRune(buffer)
			}
		}
		currentRow = builder.String()
		currentIndex++
		builder.Reset()
		if currentRow == firstRow {
			for {
				if currentIndex >= firstColumnIndex {
					break
				}
				buffer, _, err = reader.ReadRune()
				if err != nil {
					return "", "", err
				}
				if buffer == ',' {
					currentIndex++
				}
			}
			for {
				buffer, _, err = reader.ReadRune()
				if err != nil {
					return "", "", err
				}
				if buffer == ',' {
					break
				} else {
					builder.WriteRune(buffer)
				}
			}
			firstValue = builder.String()
			builder.Reset()
			valuesFound++
		} else if currentRow == secondRow {
			for {
				if currentIndex >= secondColumnIndex {
					break
				}
				buffer, _, err = reader.ReadRune()
				if err != nil {
					return "", "", err
				}
				if buffer == ',' {
					currentIndex++
				}
			}
			for {
				buffer, _, err = reader.ReadRune()
				if err != nil {
					return "", "", err
				}
				if buffer == ',' {
					break
				} else {
					builder.WriteRune(buffer)
				}
			}
			secondValue = builder.String()
			builder.Reset()
			valuesFound++
		} else {
			for {
				if currentIndex >= p.columnNumber {
					break
				}
				buffer, _, err = reader.ReadRune()
				if err != nil {
					return "", "", err
				}
				if buffer == ',' {
					currentIndex++
				}
			}
		}
		if valuesFound == 2 {
			break
		}
	}
	return firstValue, secondValue, nil
}
