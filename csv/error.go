package csv

import "errors"

var ErrWrongCellFormat = errors.New("wrong cell format")
var ErrUnreachableAddress = errors.New("address is unreachable")
var ErrTableFormat = errors.New("table is corrupted or it is not csv format")
