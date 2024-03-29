package reader

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
)

const REPLACEMENT_VERSION = 39

func CheckFileExist(filePath string) (bool, error) {
	_, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return true, err
	} else {
		return true, nil
	}
}

// Opens an incoming and outgoing file and applies streaming gzip processing to them
// Then calls Scan() to process the results and returns the result from that.
func GzipProcessor(filePathIn string, filePathOut string, allowOverwrite bool) (int, int64, error) {

	file, err := os.Open(filePathIn)

	if err != nil {
		log.Print(err)
		return -1, -1, err
	}

	log.Printf("Opened %s", filePathIn)

	reader, err := gzip.NewReader(file)

	logtag := path.Base(filePathIn)

	if err != nil {
		log.Printf("[%s] Could not create gzip reader: %s", logtag, err)
		return -1, -1, err
	}

	doesExist, _ := CheckFileExist(filePathOut)
	if doesExist {
		if allowOverwrite {
			log.Printf("[%s] Warning: overwriting output file %s", logtag, filePathOut)
		} else {
			log.Printf("[%s] Not overwriting output file %s", logtag, filePathOut)
			return -1, -1, errors.New("Not overwriting output file")
		}
	}
	writeFile, writeErr := os.Create(filePathOut)

	if writeErr != nil {
		log.Printf("[%s] Could not open %s to write: %s", logtag, filePathOut, err)
		return -1, -1, err
	}

	log.Printf("[%s] Opened %s to write", logtag, filePathOut)
	writer := gzip.NewWriter(writeFile)

	defer func() {
		if reader != nil {
			reader.Close()
		}
		if file != nil {
			file.Close()
		}
		if writer != nil {
			writer.Close()
		}
		if writeFile != nil {
			writeFile.Close()
		}
	}()
	return Scan(reader, writer, logtag)
}

func UncompressedProcessor(filePathIn string, filePathOut string, allowOverwrite bool) (int, int64, error) {
	file, err := os.Open(filePathIn)

	if err != nil {
		log.Print(err)
		return -1, -1, err
	}

	log.Printf("Opened %s", filePathIn)

	logtag := path.Base(filePathIn)

	doesExist, _ := CheckFileExist(filePathOut)
	if doesExist {
		if allowOverwrite {
			log.Printf("Warning: overwriting output file %s", filePathOut)
		} else {
			log.Printf("Not overwriting output file %s", filePathOut)
			return -1, -1, errors.New("not overwriting output file")
		}
	}

	writeFile, writeErr := os.Create(filePathOut)

	if writeErr != nil {
		log.Printf("Could not open %s to write: %s", filePathOut, err)
		return -1, -1, err
	}

	log.Printf("Opened %s to write", filePathOut)

	defer func() {
		if writeFile != nil {
			writeFile.Close()
		}
		if file != nil {
			file.Close()
		}
	}()

	return Scan(file, writeFile, logtag)
}

func readToBuffer(reader io.Reader) (*bytes.Buffer, error) {
	var buffer bytes.Buffer

	bytesRead, err := buffer.ReadFrom(reader)

	if err != nil {
		log.Print("Could not buffer incoming file: ", err)
		return nil, err
	}
	log.Printf("Read %d bytes", bytesRead)
	return &buffer, nil
}

// Takes a reader and a writer, and applies the version change as a regex
// On error, returns an error; otherwise returns the number of lines processed and the number of uncompressed bytes processed
func Scan(reader io.Reader, writer io.Writer, logtag string) (int, int64, error) {
	matcher, err := regexp.Compile(`<Project ObjectID="(\d)" ClassID="([\w\d\-]+)" Version="(\d+)">`)

	if err != nil {
		log.Print(err)
		return -1, -1, err
	}

	lineCounter := 0
	lastLineCounter := 0
	//var zeroLengthReadsCount int = 0
	var foundIt = false
	var bytesWritten int64 = 0

	//it seems that sometimes we get zero-length reads in the middle of the file.  Even 10 sometimes.
	//so, we must keep looping till we know that the whole stream is done.
	//if we have 100,000 zero-length reads in a row, then we conclude that it must be finished
	for emptyReads := 0; emptyReads < 100000; {
		if foundIt {
			log.Printf("[%s] Version tag has been upgraded, performing binary copy of the rest of the file.", logtag)
			written, err := io.Copy(writer, reader)
			bytesWritten += written
			if err != nil {
				log.Printf("[%s] Could not copy remainder of file: %s", logtag, err)
				return lineCounter, written, err
			} else {
				if written == 0 {
					break
				}
				log.Printf("[%s] Copied %d (uncompressed) bytes", logtag, written)
			}
		} else {
			scanner := bufio.NewScanner(reader)
			scanner.Split(bufio.ScanLines)
			initialBuffer := make([]byte, 102400)
			scanner.Buffer(initialBuffer, 102400)

			//fmt.Printf("On line %d\n", lineCounter)
			for scanner.Scan() {
				lineCounter += 1
				//fmt.Printf("debug: got line %s\n", scanner.Text())

				matches := matcher.FindStringSubmatch(scanner.Text())
				if matches == nil {
					//log.Print("debug: got no matches\n")
					_, err := writer.Write(scanner.Bytes())
					if err != nil {
						log.Print(err)
						return -1, -1, err
					}
					_, otherErr := writer.Write([]byte("\n"))
					if otherErr != nil {
						log.Print(err)
						return -1, -1, err
					}
				} else {
					//log.Printf("debug: matches: %s", matches)
					version, err := strconv.ParseInt(matches[3], 10, 32)
					if err != nil {
						log.Printf("[%s] Detected version was not a number, got %s\n", logtag, matches[3])
						return lineCounter, -1, err
					}
					log.Printf("[%s] ObjectID is %s, classID is %s, version is %d\n", logtag, matches[1], matches[2], version)
					if version == REPLACEMENT_VERSION {
						log.Printf("This file does not need updating.")
						//FIXME: should add custom error here.
						_, writeErr := writer.Write([]byte(scanner.Text()))
						if writeErr != nil {
							log.Print("Could not write output: ", writeErr)
						}
					} else if version > REPLACEMENT_VERSION {
						log.Printf("[%s] This file is at a higher version (%d) than the replacement (%d).", logtag, version, REPLACEMENT_VERSION)
					} else {
						replacementString := fmt.Sprintf(`<Project ObjectID="%s" ClassID="%s" Version="%d">`+"\n", matches[1], matches[2], REPLACEMENT_VERSION)
						replacementLine := matcher.ReplaceAllString(scanner.Text(), replacementString)
						_, writeErr := writer.Write([]byte(replacementLine))
						if writeErr != nil {
							log.Printf("[%s] Could not write output: %s", logtag, writeErr)
						}
						log.Printf("[%s] Version identifier tag updated to %d", logtag, REPLACEMENT_VERSION)
					}
					foundIt = true
				}
				if scanner.Text() == "</PremiereData>" {
					foundIt = true
					break
				}
			}
		}
		if lineCounter == lastLineCounter {
			emptyReads += 1
			shouldPrintDiv := emptyReads % 1000
			if shouldPrintDiv == 0 { //only print the warning every 30 empty reads
				log.Printf("[%s] WARNING got %d empty reads in a row", logtag, emptyReads)
			}
		} else {
			emptyReads = 0
			lastLineCounter = lineCounter
		}
	}
	return lineCounter, bytesWritten, nil
}
