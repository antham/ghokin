package ghokin

import (
	"bytes"
)

type commands map[string]string

type indent struct {
	backgroundAndScenario int
	step                  int
	tableAndDocString     int
}

// FileManager handles transformation on feature files
type FileManager struct {
	indentConf indent
	commands   commands
}

// NewFileManager creates a brand new FileManager, it requires indentation values and commands defined
// as a shell commands in comments
func NewFileManager(backgroundAndScenarioIndent int, stepIndent int, tableAndDocStringIndent int, commands map[string]string) FileManager {
	return FileManager{
		indent{
			backgroundAndScenarioIndent,
			stepIndent,
			tableAndDocStringIndent,
		},
		commands,
	}
}

// Transform formats and applies shell commands on feature file
func (f FileManager) Transform(filename string) (bytes.Buffer, error) {
	section, err := extractSections(filename)

	if err != nil {
		return bytes.Buffer{}, err
	}

	return transform(section, f.indentConf, f.commands)
}
