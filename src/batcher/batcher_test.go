package batcher

import (
	"strings"
	"testing"
)

func TestProcessListOK(t *testing.T) {
	testData := `/path/to/file1
/path/to/file2
/path/to/file3
/path/to/file4
`
	reader := strings.NewReader(testData)

	result, err := processList(reader)

	if err != nil {
		t.Fatalf("processList returned an error %s", err)
	}

	t.Logf("Got results %s", result)
	if result[0] != "/path/to/file1" {
		t.Errorf("Path 1 did not match")
	}
	if result[1] != "/path/to/file2" {
		t.Errorf("Path 2 did not match")
	}
	if result[2] != "/path/to/file3" {
		t.Errorf("Path 3 did not match")
	}
	if result[3] != "/path/to/file4" {
		t.Errorf("Path 4 did not match")
	}
	if len(result) != 4 {
		t.Errorf("Array length was incorrect, got %d expected 4", len(result))
	}
}
