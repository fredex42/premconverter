package reader

import (
	"bytes"
	"strings"
	"testing"
)

// Scan should read in a bunch of XML and convert just the "version" entry
func TestScanOK(t *testing.T) {
	test_data := `<?xml version="1.0" encoding="UTF-8" ?>
<PremiereData Version="3">
        <Project ObjectRef="1"/>
        <Project ObjectID="1" ClassID="62ad66dd-0dcd-42da-a660-6d8fbde94876" Version="32">
                <Node Version="1">
                        <Properties Version="1">
                                <ProjectViewState.List ObjectID="2" ClassID="aab0946f-7a21-4425-8908-fafa2119e30e" Version="3">
                                        <ProjectViewStates Version="1">
                                                <ProjectViewState Version="1" Index="0">
                                                        <First>8fd5ff01-787e-41bc-9302-193332660c4c</First>
                                                        <Second ObjectRef="1"/>
                                                </ProjectViewState>
                                                <ProjectViewState Version="1" Index="1">
                                                        <First>8a4a4716-6f1b-46fc-8b9c-e8c012ee89d4</First>
                                                        <Second ObjectRef="2"/>
                                                </ProjectViewState>
                                        </ProjectViewStates>`

	expected := `<?xml version="1.0" encoding="UTF-8" ?>
<PremiereData Version="3">
        <Project ObjectRef="1"/>
        <Project ObjectID="1" ClassID="62ad66dd-0dcd-42da-a660-6d8fbde94876" Version="35">
                <Node Version="1">
                        <Properties Version="1">
                                <ProjectViewState.List ObjectID="2" ClassID="aab0946f-7a21-4425-8908-fafa2119e30e" Version="3">
                                        <ProjectViewStates Version="1">
                                                <ProjectViewState Version="1" Index="0">
                                                        <First>8fd5ff01-787e-41bc-9302-193332660c4c</First>
                                                        <Second ObjectRef="1"/>
                                                </ProjectViewState>
                                                <ProjectViewState Version="1" Index="1">
                                                        <First>8a4a4716-6f1b-46fc-8b9c-e8c012ee89d4</First>
                                                        <Second ObjectRef="2"/>
                                                </ProjectViewState>
                                        </ProjectViewStates>
`

	reader := strings.NewReader(test_data)
	writer := bytes.NewBufferString("")

	lineCount, _, err := Scan(reader, writer, "test")

	if err != nil {
		t.Errorf("Scan returned an error: %s", err)
	}

	t.Logf("Processed %d lines", lineCount)
	output := writer.String()

	if output != expected {
		t.Errorf("Scan did not process the string properly, got %s\n", output)
		t.Errorf("Expected %s\n", expected)
	}
}
