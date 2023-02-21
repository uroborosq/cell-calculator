package main

import (
	"fmt"
	"log"
	"os"
)


func main() {
	if len(os.Args) > 1 {
		filePath := os.Args[1]
		fmt.Println(filePath)
	} else {
		log.Fatal("File path wasn't specified! Exiting...")
	}

}