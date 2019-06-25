package reader

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"log"
	"regexp"
	"strconv"
)

const REPLACEMENT_VERSION = 35

// Opens an incoming and outgoing file and applies streaming gzip processing to them
// Then calls Scan() to process the results and returns the result from that.
func GzipProcessor(filePathIn string, filePathOut string) (int, error) {
	file, err := os.Open(filePathIn)

	if err != nil {
		log.Fatal(err)
		return -1, err
	}

	log.Printf("Opened %s", filePathIn)
	defer file.Close()

	reader, err := gzip.NewReader(file)

	if err != nil {
		log.Fatal("Could not create gzip reader: ", err)
		return -1, err
	}

	writeFile, writeErr := os.Create(filePathOut)

	if writeErr != nil {
		log.Fatalf("Could not open %s to write: %s", filePathOut, err)
		return -1, err
	}



	log.Printf("Opened %s to write", filePathOut)
	writer := gzip.NewWriter(writeFile)

	defer writer.Close()
	return Scan(reader, writer)
}

// Takes a reader and a writer, and applies the version change as a regex
// On error, returns an error; otherwise returns the number of lines processed.
func Scan(reader io.Reader, writer io.Writer) (int, error) {
	matcher, err := regexp.Compile(`<Project ObjectID="(\d)" ClassID="([\w\d\-]+)" Version="(\d+)">`)

	if(err!=nil){
		log.Fatal(err)
		return -1, err
	}

	lineCounter := 0

	var zeroLengthReadsCount int = 0

	//it seems that sometimes we get zero-length reads in the middle of the file.
	//so, we must keep looping till we know that the whole stream is done.
	//if we have 1,000 zero-length reads, then we conclude that it's done.
	for zeroLengthReadsCount<1000 {
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanLines)
		if len(scanner.Bytes())==0 {
			zeroLengthReadsCount+=1
		}

		for scanner.Scan() {
			lineCounter += 1
			//fmt.Printf("debug: got line %s\n", scanner.Text())

			matches := matcher.FindStringSubmatch(scanner.Text())
			if (matches == nil) {
				//log.Print("debug: got no matches\n")
				_, err := writer.Write(scanner.Bytes())
				if (err != nil) {
					log.Fatal(err)
					return -1, err
				}
				_, otherErr := writer.Write([]byte("\n"))
				if (otherErr != nil) {
					log.Fatal(err)
					return -1, err
				}
			} else {
				//log.Printf("debug: matches: %s", matches)
				version, err := strconv.ParseInt(matches[3], 10, 32)
				if (err != nil) {
					log.Fatalf("Detected version was not a number, got %s\n", matches[3])
					return lineCounter, err
				}
				log.Printf("ObjectID is %s, classID is %s, version is %d\n", matches[1], matches[2], version)
				if(version==REPLACEMENT_VERSION){
					log.Printf("This file does not need updating.")
				}
				replacementString := fmt.Sprintf(`<Project ObjectID="%s" ClassID="%s" Version="%d">`+"\n", matches[1], matches[2], REPLACEMENT_VERSION)
				replacementLine := matcher.ReplaceAllString(scanner.Text(), replacementString)
				_, writeErr := writer.Write([]byte(replacementLine))
				if (writeErr != nil) {
					log.Fatal("Could not write output: ", writeErr)
				}
			}
		}
	}
	return lineCounter, nil
}