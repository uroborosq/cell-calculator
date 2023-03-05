package csv_test

import (
	"bytes"
	"cell-calculator/csv"
	"math/big"
	"testing"
)

func TestColumnNumber(t *testing.T) {
	data := []byte(`,A,B,C,Cell
	1,2,4,5`)
	reader := bytes.NewReader(data)
	columnNumber, err := csv.GetColumnNumber(reader)
	if columnNumber.Cmp(new(big.Int).SetInt64(5)) != 0 {
		t.Fatalf("Wrong number of columns, must be 5, received %d", columnNumber)
	} else if err != nil {
		t.Fatalf("Error occurred, description: %s", err.Error())
	}
}
