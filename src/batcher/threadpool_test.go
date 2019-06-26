package batcher

import (
	"fmt"
	"github.com/fredex42/premconverter/reader"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestThreadPoolNormal(t *testing.T) {
	testData := make([]string, 5)
	testData[0] = "/path/to/file1"
	testData[1] = "/path/to/file2"
	testData[2] = "/path/to/file3"
	testData[3] = "/path/to/file4"
	testData[4] = "/path/to/file5"

	Allocate(testData, "/output/path/")

	testSpy := &reader.SpyReader{}
	testSpy.Initialise()

	wg := CreateWorkerPoolAndWait(2, testSpy)

	wg.Wait()
	var finalResults ResultList = CollectResults()
	//threads mean they can come through in any order; we sort here to ensure that our assertions
	//work properly below
	sort.Sort(finalResults)

	if testSpy.Calls != len(testData) {
		t.Errorf("Got %d calls, expected %d", testSpy.Calls, len(testData))
	}

	for i := 0; i < len(finalResults); i++ {
		if finalResults[i].success != true {
			t.Errorf("Result %d should have been successful", i)
		}
		if finalResults[i].outputFileName != fmt.Sprintf("/output/path/file%d", i+1) {
			t.Errorf("Output filename %s is incorrect, wanted %s", finalResults[i].outputFileName, fmt.Sprintf("/output/path/file%d", i+1))
		}
	}
	if len(finalResults) != len(testData) {
		t.Errorf("Got %d results, expected %d", len(finalResults), len(testData))
	}
}

// Allocate should not allow jobs to be queued if the input and output files would be the same
func TestThreadPoolBadPath(t *testing.T) {
	testData := make([]string, 5)
	testData[0] = "/path/to/file1"
	testData[1] = "/path/to/file2"
	testData[2] = "/path/to/file3"
	testData[3] = "/path/to/file4"
	testData[4] = "/path/to/file5"

	assert.Panics(t, func() { Allocate(testData, "/path/to/") }, "Got no panic")

}
