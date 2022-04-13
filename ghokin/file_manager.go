package ghokin

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	mpath "path"
	"path/filepath"
	"sync"
)

// ProcessFileError is emitted when processing a file trigger an error
type ProcessFileError struct {
	Message string
	File    string
}

// Error dumps a string error
func (p ProcessFileError) Error() string {
	return fmt.Sprintf(`an error occurred with file "%s" : %s`, p.File, p.Message)
}

type aliases map[string]string

type indent struct {
	featureDescription    int
	backgroundAndScenario int
	step                  int
	tableAndDocString     int
}

// FileManager handles transformation on feature files
type FileManager struct {
	indentConf indent
	aliases    aliases
}

// NewFileManager creates a brand new FileManager, it requires indentation values and aliases defined
// as a shell commands in comments
func NewFileManager(featureDescription int, backgroundAndScenarioIndent int, stepIndent int, tableAndDocStringIndent int, aliases map[string]string) FileManager {
	return FileManager{
		indent{
			featureDescription,
			backgroundAndScenarioIndent,
			stepIndent,
			tableAndDocStringIndent,
		},
		aliases,
	}
}

// Transform formats and applies shell commands on feature file
func (f FileManager) Transform(filename string) (bytes.Buffer, error) {
	/* #nosec */
	file, err := os.Open(filename)
	if err != nil {
		return bytes.Buffer{}, err
	}
	bom, err := skipBom(file)
	if err != nil {
		return bytes.Buffer{}, err
	}
	section, err := extractSections(file)
	if err != nil {
		return bytes.Buffer{}, err
	}

	buf, err := transform(section, f.indentConf, f.aliases)
	if err != nil {
		return bytes.Buffer{}, err
	}
	return *bytes.NewBuffer(append(bom, buf.Bytes()...)), nil
}

// TransformAndReplace formats and applies shell commands on file or folder
// and replace the content of files
func (f FileManager) TransformAndReplace(path string, extensions []string) []error {
	return f.process(path, extensions, replaceFileWithContent)
}

// Check ensures file or folder is well formatted
func (f FileManager) Check(path string, extensions []string) []error {
	return f.process(path, extensions, check)
}

func (f FileManager) process(path string, extensions []string, processFile func(file string, content []byte) error) []error {
	errors := []error{}
	fi, err := os.Stat(path)
	if err != nil {
		return append(errors, err)
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		errors = append(errors, f.processPath(path, extensions, processFile)...)
	case mode.IsRegular():
		b, err := f.Transform(path)
		if err != nil {
			return append(errors, err)
		}
		if err := processFile(path, b.Bytes()); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (f FileManager) processPath(path string, extensions []string, processFile func(file string, content []byte) error) []error {
	errors := []error{}
	fc := make(chan string)
	wg := sync.WaitGroup{}
	var mu sync.Mutex

	files, err := findFeatureFiles(path, extensions)
	if err != nil {
		return []error{err}
	}
	if len(files) == 0 {
		return []error{}
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			for file := range fc {
				b, err := f.Transform(file)
				if err != nil {
					mu.Lock()
					errors = append(errors, ProcessFileError{Message: err.Error(), File: file})
					mu.Unlock()
					continue
				}
				if err := processFile(file, b.Bytes()); err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
				}
			}
			wg.Done()
		}()
	}

	for _, file := range files {
		fc <- file
	}

	close(fc)
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()

	return errors
}

func replaceFileWithContent(file string, content []byte) error {
	f, err := os.Create(file)
	if err != nil {
		return ProcessFileError{Message: err.Error(), File: file}
	}
	_, err = f.Write(content)
	if err != nil {
		return ProcessFileError{Message: err.Error(), File: file}
	}
	return nil
}

func check(file string, content []byte) error {
	currentContent, err := ioutil.ReadFile(file) // #nosec

	if err != nil {
		return ProcessFileError{Message: err.Error(), File: file}
	}

	if !bytes.Equal(currentContent, content) {
		return ProcessFileError{Message: "file is not properly formatted", File: file}
	}

	return nil
}

func findFeatureFiles(rootPath string, extensions []string) ([]string, error) {
	files := []string{}

	if err := filepath.Walk(rootPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		for _, extension := range extensions {
			if !info.IsDir() && mpath.Ext(p) == "."+extension {
				files = append(files, p)
				break
			}
		}

		return nil
	}); err != nil {
		return []string{}, err
	}

	return files, nil
}

// skipBom moves file pointer after BOM if one is found and returns it
func skipBom(file *os.File) ([]byte, error) {
	bom := []byte{'\xef', '\xbb', '\xbf'}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return []byte{}, err
	}

	buffer := make([]byte, len(bom))

	if n, err := file.Read(buffer); err != nil || n < len(bom) {
		if _, serr := file.Seek(0, io.SeekStart); serr != nil {
			return []byte{}, serr
		}

		return []byte{}, err
	}

	if bytes.Equal(bom, buffer) {
		return bom, nil
	}

	_, err := file.Seek(0, io.SeekStart)

	return []byte{}, err
}
