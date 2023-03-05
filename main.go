package main

import (
	"cell-calculator/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

type CsvReader interface {
	Close() error
	GetNextCell() (string, bool, error)
}

func main() {
	var filePath string
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	} else {
		log.Fatal("File path wasn't specified! Exiting...")
	}

	var reader CsvReader
	var err error
	reader, err = csv.NewDiskParser(filePath, 100*1024*1024, 1024*1024)
	defer reader.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		value, isNewLine, err := reader.GetNextCell()
		if err != nil && errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			log.Fatalf("Error occured while proccesing: %s", err.Error())
		}
		if isNewLine {
			fmt.Println(value)
		} else {
			fmt.Printf("%s,", value)
		}
	}
}
