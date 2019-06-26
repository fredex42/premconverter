package batcher

import (
	"bufio"
	"io"
	"log"
	"os"
)

const INITIAL_LIST_SIZE = 50

func ReadList(listFilePath string) ([]string, error) {
	file, err := os.Open(listFilePath)
	if err != nil {
		log.Fatalf("Could not open %s: %s", listFilePath, err)
		return nil, err
	}
	return processList(file)
}

// Reads lines from a Reader which are assumed to be filenames
// outputs these as a slice of strings
func processList(file io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var lineCounter int = 0

	var rtnList = make([]string, INITIAL_LIST_SIZE)

	for scanner.Scan() {
		if len(rtnList) == cap(rtnList) {
			newSlice := make([]string, len(rtnList), len(rtnList)+INITIAL_LIST_SIZE)
			copy(newSlice, rtnList)
			rtnList = newSlice
		}
		rtnList[lineCounter] = scanner.Text()

		lineCounter += 1
	}

	finalSlice := make([]string, lineCounter)
	copy(finalSlice, rtnList)
	return finalSlice, nil
}
