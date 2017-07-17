package main

import (
	"fmt"
	"os"

	"github.com/juntaki/toyotomimi/lib"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s outputDir", os.Args[0])
		os.Exit(1)
	}
	radiolib.RecordAll(os.Args[1])
}
