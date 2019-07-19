package batcher

import (
	"log"
	"strings"
	"sync"
	"os"
)
import "github.com/fredex42/premconverter/reader"
import "path"

type Job struct {
	inputFileName  string
	outputFileName string
}

type Result struct {
	outputFileName string
	linesProcessed int
	bytesProcessed int64
	success        bool
	err            error
}

type ResultList []Result

//implement sort.Interface for results - https://gobyexample.com/sorting-by-functions
func (r ResultList) Len() int {
	return len(r)
}

func (r ResultList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ResultList) Less(i, j int) bool {
	return strings.Compare(r[i].outputFileName, r[j].outputFileName) < 0
}

var jobsChan = make(chan Job, 10)
var resultsChan = make(chan Result, 10)

func Exists(name string) bool {
    if _, err := os.Stat(name); err != nil {
        if os.IsNotExist(err) {
            return false
        }
    }
    return true
}

// body of the worker thread that reads a job description and calls out to reader to process it
// need to pass a pointer to a WaitGroup that will be notified of termination, and an interface pointer
// to the implmentation of Reader to use
func worker(wg *sync.WaitGroup, reader2 reader.Reader, allowOverwrite bool) {
	for job := range jobsChan {
		lineCount, bytesWritten, err := reader2.GzipProcessor(job.inputFileName, job.outputFileName, allowOverwrite)

		if err != nil && strings.Contains(err.Error(), "gzip: invalid header") {
			log.Printf("[%s] File does not appear to be gzipped.  Attempting to run without gzip....", job.inputFileName)
			lineCount, bytesWritten, err = reader.UncompressedProcessor(job.inputFileName, job.outputFileName, allowOverwrite)
		}

		result := Result{
			job.outputFileName,
			lineCount,
			bytesWritten,
			err == nil,
			err,
		}

		errStr := "no error"
		if err != nil {
			errStr = "error: " + err.Error()
		}

		if err!=nil {
			if Exists(job.outputFileName){
				log.Printf("[%s] - output file %s exists but an error occurred. Removing corrupt output file.", job.inputFileName, job.outputFileName)
				os.Remove(job.outputFileName)
			}
		}
		log.Printf("[%s] Completed processing, %s", job.inputFileName, errStr)
		resultsChan <- result
	}

	wg.Done()
}

// creates a worker pool and runs it. Only returns when all jobs have been processed.
// You should run Allocate() to put jobs onto the queue before running this
func CreateWorkerPoolAndWait(workerCount int, reader2 reader.Reader, allowOverwrite bool) *sync.WaitGroup {
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(&wg, reader2, allowOverwrite)
	}

	return &wg
}

// puts jobs onto the queue.  Pass in an array of the input file paths, and the single output
// path where the updated files should go.
func Allocate(fileList []string, outputPath string) {
	for i := range fileList {
		inputFilePath := fileList[i]
		outputFilePath := path.Join(outputPath, path.Base(inputFilePath))
		if inputFilePath == outputFilePath {
			log.Printf("Both input and output files are the same, refusing to trash existing files")
			panic("Refusing to trash files")
		}
		job := Job{
			inputFilePath,
			outputFilePath,
		}
		jobsChan <- job
	}
	close(jobsChan)
}

// call this to close the results channel, once the WaitGroup returned by `CreatePoolAndWait` returns
func CloseResults() {
	close(resultsChan)
}

//Gather results from the output channel and return them
func CollectResults() []Result {
	resultList := make([]Result, len(resultsChan))
	var i = 0

	for result := range resultsChan {
		resultList[i] = result
		i++
	}

	return resultList
}

// reads the results channel and outputs stats to the console
func ResultStats(done chan bool) {
	var successCount int = 0
	var errorCount int = 0

	for result := range resultsChan {
		if result.success {
			successCount++
		} else {
			errorCount++
			log.Printf("%s failed: %s", result.outputFileName, result.err)
		}
	}

	log.Printf("%d conversions succeeded, %d failed", successCount, errorCount)
	done <- true
}
