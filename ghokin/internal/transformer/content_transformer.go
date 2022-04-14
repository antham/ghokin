package transformer

import "bytes"

// eolType represents a new line according to the underlying OS system (Linux, Windows, MacOSX)
type eolType string

const (
	noEol eolType = ""
	lf    eolType = "\n"
	cr    eolType = "\r"
	crlf  eolType = "\r\n"
)

// ContentTransformer adapts a content for the gherkin parser
// to fix everything that is not taking into account in the parser.
// The transformer records all settings, remove/replace everything
// that could triggers errors then it restores all settings that need
// to be restored in the original content
type ContentTransformer struct {
	eol eolType
	bom []byte
}

// DetectSettings stores all settings specifics to the content being processed
func (c *ContentTransformer) DetectSettings(content []byte) {
	c.detectEOL(content)
	c.detectBom(content)
}

// Prepare removes/replaces all settings int the content
// that would not be handle properly by the gherkin parser
func (c *ContentTransformer) Prepare(content []byte) []byte {
	return c.replaceEOLWithLF(c.removeBom(content))
}

// Restore used recorded settings to restore all settings from the original content
// that need to be preserved
func (c *ContentTransformer) Restore(content []byte) []byte {
	return c.addBom(c.replaceLFWithEOl(content))
}

// detectBom checks if a content contains a BOM (https://en.wikipedia.org/wiki/Byte_order_mark)
func (c *ContentTransformer) detectBom(content []byte) {
	bom := []byte{'\xef', '\xbb', '\xbf'}
	if len(content) >= len(bom) && bytes.Equal(bom, content[:3]) {
		c.bom = bom
	}
}

// removeBom removes BOM if one was detected and returns the content without it
func (c *ContentTransformer) removeBom(content []byte) []byte {
	if len(c.bom) > 0 {
		return content[3:]
	}
	return content
}

// addBom adds back a previously detected BOM if any to a content
func (c *ContentTransformer) addBom(content []byte) []byte {
	return append(c.bom, content...)
}

// detectEOL checks a content to find out what is the line separator in the content
func (c *ContentTransformer) detectEOL(content []byte) {
	c.eol = noEol
	var previousChar byte
	for _, char := range content {
		switch {
		case previousChar == '\r' && char == '\n':
			c.eol = crlf
			return
		case previousChar == '\r' && char != '\n':
			c.eol = cr
			return
		case char == '\n':
			c.eol = lf
			return
		}
		previousChar = char
	}
}

// replaceEOLWithLF replaces the detected EOL with the linux standard line separator
func (c *ContentTransformer) replaceEOLWithLF(content []byte) []byte {
	return bytes.ReplaceAll(content, []byte(c.eol), []byte(lf))
}

// replaceEOLWithLF replaces the linux standard line separator with the one detected in content
func (c *ContentTransformer) replaceLFWithEOl(content []byte) []byte {
	return bytes.ReplaceAll(content, []byte(lf), []byte(c.eol))
}
