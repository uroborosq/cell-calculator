package main

import (
	"cell-calculator/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func compareFile(path string, correct string) error {
	reader, err := csv.NewDiskParser(path, 100*1024*1024, 1024*1024)
	if err != nil {
		return err
	}
	defer reader.Close()
	var builder strings.Builder
	for {
		s, isNewLine, err := reader.GetNextCell()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		if isNewLine {
			builder.WriteString(s + "\n")
		} else {
			builder.WriteString(fmt.Sprintf("%s,", s))
		}
	}
	result := builder.String()
	if result != correct {
		return errors.New("output doesn't match correct answer")
	}
	return nil
}

func getErrorAndIndex(path string) (uint64, error) {
	index := uint64(0)
	reader, err := csv.NewDiskParser(path, 100*1024*1024, 1024*1024)
	if err != nil {
		return index, err
	}

	defer reader.Close()
	for {
		_, _, err := reader.GetNextCell()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return index, err
		}
		index++
	}
	return index, err
}
func TestAllOperators(t *testing.T) {
	err := compareFile("tests/valid/all-operators.csv", `,A,B,Cell
1,1,0,1
2,-97,3,100
30,10,2,5
4,1.5,3,2
`)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestTaskExample(t *testing.T) {
	err := compareFile("tests/valid/test.csv", `,A,B,Cell
1,1,0,1
2,2,6,0
30,0,1,5
`)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestFormulaReferences(t *testing.T) {
	err := compareFile("tests/valid/formula-reference-to-formula.csv", `,A,B,Cell
1,1,0,1
2,2,3,100
30,10,2,5
4,2,3,2
`)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestWrongNumberOfColumns(t *testing.T) {
	index, err := getErrorAndIndex("tests/error/different-number-of-columns.csv")
	if !errors.Is(err, csv.ErrTableFormat) && index != 12 {
		t.Fail()
	}
}

func TestWrongAddress(t *testing.T) {
	index, err := getErrorAndIndex("tests/error/wrong-address.csv")
	if !errors.Is(err, csv.ErrUnreachableAddress) && index != 14 {
		t.Fail()
	}
}

func TestWrongCellFormat(t *testing.T) {
	index, err := getErrorAndIndex("tests/error/wrong-cell-format.csv")
	if !errors.Is(err, csv.ErrWrongCellFormat) && index != 9 {
		t.Fail()
	}
}

func TestWrongColumnFormat(t *testing.T) {
	index, err := getErrorAndIndex("tests/error/wrong-column-format.csv")
	if !errors.Is(err, csv.ErrWrongCellFormat) && index != 4 {
		t.Fail()
	}
}

func TestWrongRowFormat(t *testing.T) {
	index, err := getErrorAndIndex("tests/error/wrong-row-format.csv")
	if !errors.Is(err, csv.ErrWrongCellFormat) && index != 5 {
		t.Fatalf("error: %s, index: %d", err.Error(), index)
	}
}

func TestMixAddressesAndNumbers(t *testing.T) {
	err := compareFile("tests/valid/formula-from-reference-and-number.csv", `,A,B,Cell
1,1,4,1
2,2,31,0
30,0,2,5
`)
	if err != nil {
		t.Fatal(err.Error())
	}
}
