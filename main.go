package main

import (
	"fmt"
	"os"

	"github.com/juntaki/toyotomimi/lib"
)

const version = "1.3"

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s outputDir\n", os.Args[0])
		fmt.Printf("Version: %s\n", version)
		os.Exit(1)
	}

	radiolib.RecordAll(os.Args[1])
}
