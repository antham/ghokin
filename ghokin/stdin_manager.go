package ghokin

import (
	"io"

	"github.com/antham/ghokin/v3/ghokin/internal/transformer"
)

// StdinManager handles transformation from stdin
type StdinManager struct {
	indent  int
	aliases aliases
}

// NewStdinManager creates a brand new StdinManager, it requires indentation values and aliases defined
// as a shell commands in comments
func NewStdinManager(indent int, aliases map[string]string) StdinManager {
	return StdinManager{
		indent,
		aliases,
	}
}

// Transform formats and applies shell commands on stdin
func (s StdinManager) Transform(reader io.Reader) ([]byte, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return []byte{}, err
	}
	contentTransformer := &transformer.ContentTransformer{}
	contentTransformer.DetectSettings(content)
	content = contentTransformer.Prepare(content)
	section, err := extractSections(content)
	if err != nil {
		return []byte{}, err
	}
	content, err = transform(section, s.indent, s.aliases)
	if err != nil {
		return []byte{}, err
	}
	return contentTransformer.Restore(content), nil
}
