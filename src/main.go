package main

import (
	"flag"
	"os"
)
import "fmt"
import "github.com/fredex42/premconverter/reader"

func main() {
	inputFilePtr := flag.String("input","", "a single prproj file to process")
	outputFilePtr := flag.String("output","", "a single prproj file to output")
	flag.Parse()

	if(*inputFilePtr=="" || *outputFilePtr==""){
		fmt.Print("You must specify an input and an output file. Run with --help for more information\n")
		os.Exit(2)
	}

	fmt.Printf("Processing %s to %s\n", *inputFilePtr, *outputFilePtr)

	reader.GzipProcessor(*inputFilePtr, *outputFilePtr)
	fmt.Printf("Done.\n")
}