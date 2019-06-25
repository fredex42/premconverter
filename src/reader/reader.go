package reader

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
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

	reader, err := gzip.NewReader(file)

	//log.Printf("Buffering %s in memory...", filePathIn)
	//
	//buffer, bufErr := readToBuffer(reader)
	//if(bufErr!=nil){
	//	return 0, nil
	//}
	//
	//reader.Close()
	//file.Close()
	//
	//log.Printf("Buffering completed\n")

	defer reader.Close()
	defer file.Close()

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

func readToBuffer(reader io.Reader) (*bytes.Buffer, error) {
	var buffer bytes.Buffer

	bytesRead, err := buffer.ReadFrom(reader)

	if err != nil {
		log.Fatal("Could not buffer incoming file: ", err)
		return nil, err
	}
	log.Printf("Read %d bytes", bytesRead)
	return &buffer, nil
}

// Takes a reader and a writer, and applies the version change as a regex
// On error, returns an error; otherwise returns the number of lines processed.
func Scan(reader io.Reader, writer io.Writer) (int, error) {
	matcher, err := regexp.Compile(`<Project ObjectID="(\d)" ClassID="([\w\d\-]+)" Version="(\d+)">`)

	if err != nil {
		log.Fatal(err)
		return -1, err
	}

	lineCounter := 0

	//var zeroLengthReadsCount int = 0
	var foundIt = false

	//it seems that sometimes we get zero-length reads in the middle of the file.  Even 10 sometimes.
	//so, we must keep looping till we know that the whole stream is done.
	//if we have 1,000 zero-length reads, then we conclude that it's done.
	for true {
		if foundIt {
			log.Print("Version tag has been upgraded, performing binary copy of the rest of the file.")
			written, err := io.Copy(writer, reader)
			if err != nil {
				log.Fatalf("Could not copy remainder of file: %s", err)
				return lineCounter, err
			} else {
				if written == 0 {
					break
				}
				log.Printf("Copied %d (uncompressed) bytes", written)
			}
		} else {
			scanner := bufio.NewScanner(reader)
			scanner.Split(bufio.ScanLines)
			initialBuffer := make([]byte, 102400)
			scanner.Buffer(initialBuffer, 102400)

			for scanner.Scan() {
				lineCounter += 1
				//fmt.Printf("debug: got line %s\n", scanner.Text())

				matches := matcher.FindStringSubmatch(scanner.Text())
				if matches == nil {
					//log.Print("debug: got no matches\n")
					_, err := writer.Write(scanner.Bytes())
					if err != nil {
						log.Fatal(err)
						return -1, err
					}
					_, otherErr := writer.Write([]byte("\n"))
					if otherErr != nil {
						log.Fatal(err)
						return -1, err
					}
				} else {
					//log.Printf("debug: matches: %s", matches)
					version, err := strconv.ParseInt(matches[3], 10, 32)
					if err != nil {
						log.Fatalf("Detected version was not a number, got %s\n", matches[3])
						return lineCounter, err
					}
					log.Printf("ObjectID is %s, classID is %s, version is %d\n", matches[1], matches[2], version)
					if version == REPLACEMENT_VERSION {
						log.Printf("This file does not need updating.")
						//FIXME: should add custom error here.
						_, writeErr := writer.Write([]byte(scanner.Text()))
						if writeErr != nil {
							log.Fatal("Could not write output: ", writeErr)
						}
					} else if version > REPLACEMENT_VERSION {
						log.Printf("This file is at a higher version (%d) than the replacement (%d).", version, REPLACEMENT_VERSION)
					} else {
						replacementString := fmt.Sprintf(`<Project ObjectID="%s" ClassID="%s" Version="%d">`+"\n", matches[1], matches[2], REPLACEMENT_VERSION)
						replacementLine := matcher.ReplaceAllString(scanner.Text(), replacementString)
						_, writeErr := writer.Write([]byte(replacementLine))
						if writeErr != nil {
							log.Fatal("Could not write output: ", writeErr)
						}
						log.Printf("Version identifier tag updated to %d", REPLACEMENT_VERSION)
					}
					foundIt = true
				}
				if scanner.Text() == "<//PremiereData>" {
					break
				}
			}
		}
	}
	return lineCounter, nil
}
