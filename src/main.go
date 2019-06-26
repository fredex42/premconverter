package main

import (
	"flag"
	"fmt"
	"github.com/fredex42/premconverter/batcher"
	"github.com/fredex42/premconverter/reader"
	"log"
	"os"
	"strconv"
)

func singleFileMode(inputFilePtr *string, outputFilePtr *string) {
	if *inputFilePtr == "" || *outputFilePtr == "" {
		log.Print("You must specify an input and an output file. Run with --help for more information\n")
		os.Exit(2)
	}

	log.Printf("Processing %s to %s\n", *inputFilePtr, *outputFilePtr)

	lineCount, bytesCount, err := reader.GzipProcessor(*inputFilePtr, *outputFilePtr)
	if err != nil {
		log.Printf("Finished after %d lines with an error: %s\n", lineCount, err)
	} else {
		log.Printf("Done, processed %d lines and raw-copied %d bytes \n", lineCount, bytesCount)
	}
}

func listFileMode(listFilePtr *string, outputPathPtr *string, concurrency int) {
	if *outputPathPtr == "" {
		log.Print("You must specify an output path in the --output parameter")
		os.Exit(2)
	}

	file, err := os.Open(*outputPathPtr)
	if err != nil {
		log.Printf("Could not check output path '%s': %s", *outputPathPtr, err)
		os.Exit(3)
	}

	statInfo, statErr := file.Stat()
	if statErr != nil {
		log.Printf("Could not stat output path '%s': %s", *outputPathPtr, statErr)
		os.Exit(3)
	}

	if !statInfo.IsDir() {
		log.Printf("Output path '%s' is not a directory, can't continue", *outputPathPtr)
		os.Exit(3)
	}

	processor := &reader.RealReader{}

	filesList, listReadErr := batcher.ReadList(*listFilePtr)
	if listReadErr != nil {
		log.Printf("Could not read list '%s': %s", *listFilePtr, listReadErr)
		os.Exit(5)
	} else {
		log.Printf("List contains %d files to process", len(filesList))
	}

	done := make(chan bool)

	go batcher.ResultStats(done)
	wg := batcher.CreateWorkerPoolAndWait(concurrency, processor)
	batcher.Allocate(filesList, *outputPathPtr)

	wg.Wait()

	batcher.CloseResults()
	<-done
}

func main() {
	inputFilePtr := flag.String("input", "", "a single prproj file to process")
	outputFilePtr := flag.String("output", "", "a single prproj file to output, or a directory for output if using a batch list")
	listFilePtr := flag.String("list", "", "a newline-delimited list of input files to process")
	concurrencyStringPtr := flag.String("concurrency", "3", "how many projects to process at once when in batch mode")
	flag.Parse()

	if *listFilePtr == "" {
		singleFileMode(inputFilePtr, outputFilePtr)
	} else {
		concurrency, err := strconv.Atoi(*concurrencyStringPtr)
		if err != nil {
			fmt.Printf("Could not convert concurrency '%s' into an integer", *concurrencyStringPtr)
			os.Exit(4)
		}
		listFileMode(listFilePtr, outputFilePtr, concurrency)
	}

}
