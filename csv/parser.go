package csv

import (
	"bufio"
	"cell-calculator/buffer"
	"cell-calculator/calculator"
	"cell-calculator/validation"
	"errors"
	"io"
	"math/big"
	"os"
	"strings"
	"unicode"
)

type BufferPolitic int

const (
	SaveAll BufferPolitic = iota + 1
	SaveFormulasOnly
)

type Buffer interface {
	Get(string, string) (string, bool)
	Add(string, string, string)
	Update(string, string, string) bool
}

type DiskParser struct {
	path             string
	file             *os.File
	columnNumber     *big.Int
	scanner          *bufio.Scanner
	isFirstLineRead  bool
	isRowBegin       bool
	cellBuffer       Buffer
	bufferPolitic    BufferPolitic
	currentColumn    string
	currentRow       string
	columnNameReader *ColumnNameReader
}

func NewDiskParser(path string, byteThreshold int64, stringThreshold uint64) (*DiskParser, error) {
	stat, err := os.Stat(path)

	if err != nil {
		return nil, err
	}
	var politic BufferPolitic
	if stat.Size() > byteThreshold {
		politic = SaveFormulasOnly
	} else {
		politic = SaveAll
	}

	parser := DiskParser{
		path:          path,
		cellBuffer:    buffer.NewCellBuffer(stringThreshold),
		bufferPolitic: politic,
	}
	err = parser.init()
	if err != nil {
		return nil, err
	}
	return &parser, err
}

func (p *DiskParser) init() error {
	var err error
	p.file, err = os.Open(p.path)
	if err != nil {
		return err
	}
	p.columnNumber, err = GetColumnNumber(p.file)
	if err != nil {
		return err
	}
	p.columnNameReader, err = NewColumnReader(p.path)
	if err != nil {
		return err
	}
	p.file, err = os.Open(p.path)
	p.scanner = bufio.NewScanner(p.file)
	p.scanner.Split(csvSplit)

	if p.bufferPolitic == SaveAll {
		err = p.fillBuffer()
		if err != nil {
			return err
		}
	}

	return err
}

func (p *DiskParser) Close() error {
	return p.file.Close()
}

func (p *DiskParser) GetNextCell() (string, bool, error) {
	var cell string
	var err error
	p.currentColumn, err = p.columnNameReader.GetNext()
	if err != nil {
		return "", false, err
	}
	if p.scanner.Scan() {
		cell = p.scanner.Text()
	} else {
		return "", true, io.EOF
	}

	if p.isRowBegin && p.isFirstLineRead {
		p.isRowBegin = validation.IsNewLine(cell)
		cell = trimCell(cell)

		if validation.IsNumber(cell) {
			p.currentRow = cell
			return cell, p.isRowBegin, nil
		} else {
			return "", false, ErrWrongCellFormat
		}
	}
	p.isRowBegin = validation.IsNewLine(cell)
	cell = trimCell(cell)
	if validation.IsNumber(cell) && p.isFirstLineRead { // parse number cells
		return cell, p.isRowBegin, nil
	} else if cell == "" {
		return "", p.isRowBegin, nil
	} else if !p.isFirstLineRead { // parse column names
		if p.isRowBegin && !p.isFirstLineRead {
			p.isFirstLineRead = true
		}
		if validation.IsColumnName(cell) {
			return cell, p.isRowBegin, nil
		} else {
			return "", false, ErrWrongCellFormat
		}
	} else { // parse formulas
		if value, ok := p.cellBuffer.Get(p.currentColumn, p.currentRow); ok {
			cell = value
		}

		if validation.IsNumber(cell) {
			return cell, p.isRowBegin, nil
		}

		firstRow, firstColumn, secondRow, secondColumn, operator, err := p.parseArgs(cell)
		if err != nil {
			return "", p.isRowBegin, err
		}
		value, err := p.calculate(firstRow, firstColumn, secondRow, secondColumn, operator)
		if err != nil {
			return "", false, err
		}
		p.cellBuffer.Add(p.currentColumn, p.currentRow, value)
		return value, p.isRowBegin, nil
	}
}

func (p *DiskParser) parseArgs(cell string) (string, string, string, string, rune, error) {
	var (
		firstColumn, firstRow, secondColumn, secondRow strings.Builder
		operator                                       rune
	)

	if len(cell) > 1 && cell[0] != '=' {
		return "", "", "", "", ' ', ErrWrongCellFormat
	}
	currentIndex := 1
	for _, r := range cell[currentIndex:] {
		if unicode.IsLetter(r) {
			currentIndex++
			firstColumn.WriteRune(r)
		} else if r == ' ' {
			currentIndex++
		} else {
			break
		}
	}
	for _, r := range cell[currentIndex:] {
		if unicode.IsDigit(r) {
			currentIndex++
			firstRow.WriteRune(r)

		} else if r == ' ' {
			currentIndex++
		} else {
			break
		}
	}

	for _, r := range cell[currentIndex:] {
		if validation.IsOperator(r) {
			currentIndex++
			operator = r
			break
		} else if r == ' ' {
			currentIndex++
		} else {
			return "", "", "", "", 0, ErrWrongCellFormat
		}
	}

	for _, r := range cell[currentIndex:] {
		if unicode.IsLetter(r) {
			currentIndex++
			secondColumn.WriteRune(r)
		} else if r == ' ' {
			currentIndex++
			continue
		} else {
			break
		}
	}

	for _, r := range cell[currentIndex:] {
		if unicode.IsDigit(r) {
			currentIndex++
			secondRow.WriteRune(r)
		} else if r == ' ' {
			currentIndex++
			continue
		} else {
			break
		}
	}

	if currentIndex != len(cell) {
		return "", "", "", "", 0, ErrWrongCellFormat
	}

	return firstRow.String(), firstColumn.String(), secondRow.String(), secondColumn.String(), operator, nil
}

func (p *DiskParser) calculate(firstRow string, firstColumn string, secondRow string, secondColumn string, operator rune) (string, error) {
	var (
		values            [2]string
		firstOk, secondOk bool
		err               error
	)

	if firstColumn == "" && firstRow != "" {
		values[0] = firstRow
	} else {
		values[0], firstOk = p.cellBuffer.Get(firstColumn, firstRow)
	}
	if secondColumn == "" && secondRow != "" {
		values[1] = secondRow
	} else {
		values[1], secondOk = p.cellBuffer.Get(secondColumn, secondRow)
	}

	if !firstOk && values[0] == "" && values[1] == "" && !secondOk {
		valuesSlice, err := p.getCellsValue([]string{firstRow, secondRow}, []string{firstColumn, secondColumn})
		if err != nil {
			return "", err
		}
		values[0] = valuesSlice[0]
		values[1] = valuesSlice[1]
	} else if !firstOk && values[0] == "" {
		valuesSlice, err := p.getCellsValue([]string{firstRow}, []string{firstColumn})
		if err != nil {
			return "", err
		}
		values[0] = valuesSlice[0]
	} else if !secondOk && values[1] == "" {
		valuesSlice, err := p.getCellsValue([]string{secondRow}, []string{secondColumn})
		if err != nil {
			return "", err
		}
		values[1] = valuesSlice[0]
	}

	if err != nil {
		return "", err
	}
	if !validation.IsNumber(values[0]) {
		firstLocalRow, firstLocalColumn, secondLocalRow, secondLocalColumn, operator, err := p.parseArgs(values[0])
		if err != nil {
			return "", err
		}
		values[0], err = p.calculate(firstLocalRow, firstLocalColumn, secondLocalRow, secondLocalColumn, operator)
		if err != nil {
			return "", err
		}
		p.cellBuffer.Add(firstColumn, firstRow, values[0])
	}
	if !validation.IsNumber(values[1]) {
		firstLocalRow, firstLocalColumn, secondLocalRow, secondLocalColumn, operator, err := p.parseArgs(values[1])
		if err != nil {
			return "", err
		}
		values[1], err = p.calculate(firstLocalRow, firstLocalColumn, secondLocalRow, secondLocalColumn, operator)
		if err != nil {
			return "", err
		}
		p.cellBuffer.Add(secondColumn, secondRow, values[1])

	}
	return calculator.ProcessWithOperator(values[0], values[1], operator)
}

func (p *DiskParser) getCellsValue(rows []string, columns []string) ([]string, error) {
	file, err := os.Open(p.path)
	if err != nil {
		return nil, err
	}
	if len(rows) != len(columns) {
		return nil, errors.New("getCellsValue: rows length doesn't fit columns")
	}
	size := len(rows)
	values := make([]string, size)
	indexes := make([]*big.Int, size)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(csvSplit)
	var (
		cell         string
		currentIndex *big.Int
		valuesFound  int
	)
	bigOne := new(big.Int).SetInt64(1)
	currentIndex = big.NewInt(0)
	for scanner.Scan() {
		cell = scanner.Text()
		rawCell := trimCell(cell)
		for i := 0; i < size; i++ {
			if rawCell == columns[i] {
				if indexes[i] != nil {
					return nil, ErrTableFormat
				}
				indexes[i] = new(big.Int).Set(currentIndex)
			}
		}
		if cell[len(cell)-1] == '\n' {
			break
		}
		currentIndex.Add(currentIndex, bigOne)
	}
	for scanner.Scan() {
		currentRow := scanner.Text()
		currentRow = currentRow[:len(currentRow)-1]
		currentIndex.Set(bigOne)
		for scanner.Scan() {
			cell = scanner.Text()
			for i := 0; i < size; i++ {
				if indexes[i].Cmp(currentIndex) == 0 && currentRow == rows[i] {
					values[i] = cell
					valuesFound++
				}
			}
			currentIndex = currentIndex.Add(currentIndex, bigOne)
			if cell[len(cell)-1] == '\n' {
				break
			}
		}
		if currentIndex.Cmp(p.columnNumber) != 0 {
			return nil, ErrTableFormat
		}
		if valuesFound == size {
			break
		}
	}
	if valuesFound != size {
		return nil, ErrUnreachableAddress
	}
	for i := 0; i < size; i++ {
		values[i] = trimCell(values[i])
	}

	return values, nil
}

func (p *DiskParser) fillBuffer() error {
	file, err := os.Open(p.path)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(csvSplit)
	cells := make(map[string]map[string]struct{})
	for scanner.Scan() {
		cell := scanner.Text()
		if cell[len(cell)-1] == '\n' {
			break
		}
	}
	for scanner.Scan() {
		cell := scanner.Text()
		cell = trimCell(cell)
		if !validation.IsNumber(cell) {
			firstRow, firstColumn, secondRow, secondColumn, _, err := p.parseArgs(cell)
			if firstColumn == "" && secondColumn == "" {
				continue
			}
			if err != nil {
				return err
			}
			if cells[firstRow] == nil {
				cells[firstRow] = make(map[string]struct{}, 1)
			}
			if cells[secondRow] == nil {
				cells[secondRow] = make(map[string]struct{}, 1)
			}
			cells[firstRow][firstColumn] = struct{}{}
			cells[secondRow][secondColumn] = struct{}{}
		}
	}

	file, err = os.Open(p.path)
	if err != nil {
		return err
	}

	scanner = bufio.NewScanner(file)
	scanner.Split(csvSplit)

	var currentRow string
	var currentColumn string
	columnReader, err := NewColumnReader(p.path)
	if err != nil {
		return err
	}
	if scanner.Scan() {
		currentRow = scanner.Text()
	}
	for scanner.Scan() {
		cell := scanner.Text()
		if cell[len(cell)-1] == '\n' {
			break
		}
	}
	if scanner.Scan() {
		currentRow = scanner.Text()
		currentColumn, err = columnReader.GetNext()
		if err != nil {
			return err
		}
	}
	for scanner.Scan() {
		cell := scanner.Text()
		currentColumn, err = columnReader.GetNext()
		if err != nil {
			return err
		}

		if m, ok := cells[currentRow]; ok {
			if _, ok := m[currentColumn]; ok {
				p.cellBuffer.Add(currentColumn, currentRow, trimCell(cell))
			}
		}

		if cell[len(cell)-1] == '\n' {
			if scanner.Scan() {
				currentRow = trimCell(scanner.Text())
				currentColumn, err = columnReader.GetNext()
				if err != nil {
					return err
				}
				currentColumn = trimCell(currentColumn)
			} else {
				break
			}
		}
	}
	return nil
}

func trimCell(s string) string {
	s = strings.TrimSuffix(s, ",")
	s = strings.TrimSuffix(s, "\n")
	s = strings.TrimSuffix(s, "\r")
	return s
}
