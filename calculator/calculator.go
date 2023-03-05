package calculator

import (
	"errors"
	"strconv"
)

var ErrOperatorIsNotSupported = errors.New("operator is not supported")
var ErrNotNumber = errors.New("")
var ErrZeroDivision = errors.New("zero division")

func ProcessWithOperator(firstStr string, secondStr string, operator rune) (string, error) {
	first, err := strconv.ParseFloat(firstStr, 64)
	if err != nil {
		return "", ErrNotNumber
	}
	second, err := strconv.ParseFloat(secondStr, 64)
	if err != nil {
		return "", ErrNotNumber
	}
	switch operator {
	case '+':
		return strconv.FormatFloat(first+second, 'f', -1, 64), nil
	case '-':
		return strconv.FormatFloat(first-second, 'f', -1, 64), nil
	case '*':
		return strconv.FormatFloat(first*second, 'f', -1, 64), nil
	case '/':
		if second == 0 {
			return "", ErrZeroDivision
		}
		return strconv.FormatFloat(first/second, 'f', -1, 64), nil
	}
	return "", ErrOperatorIsNotSupported
}
