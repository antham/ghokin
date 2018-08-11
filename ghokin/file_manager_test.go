package ghokin

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileManagerTransform(t *testing.T) {
	type scenario struct {
		filename string
		test     func(bytes.Buffer, error)
	}

	scenarios := []scenario{
		{
			"fixtures/file1.feature",
			func(buf bytes.Buffer, err error) {
				b, e := ioutil.ReadFile("fixtures/file1.feature")

				assert.NoError(t, e)
				assert.EqualValues(t, string(b[:len(b)-1]), buf.String())
			},
		},
		{
			"fixtures/",
			func(buf bytes.Buffer, err error) {
				assert.EqualError(t, err, "Parser errors:\nread fixtures/: is a directory")
			},
		},
	}

	for _, scenario := range scenarios {
		f := NewFileManager(2, 4, 6,
			map[string]string{
				"seq": "seq 1 3",
			},
		)

		scenario.test(f.Transform(scenario.filename))
	}
}
