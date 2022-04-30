package ghokin

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type failingReader struct{}

func (f failingReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("an error occurred when reading data")
}

func TestStdinManagerTransform(t *testing.T) {
	type scenario struct {
		name  string
		setup func() (StdinManager, io.Reader)
		test  func([]byte, error)
	}

	scenarios := []scenario{
		{
			"Format the stream from stdin on stdout",
			func() (StdinManager, io.Reader) {
				stdinManager := NewStdinManager(2,
					map[string]string{
						"seq": "seq 1 3",
					},
				)
				content, err := os.ReadFile("fixtures/file1.feature")
				assert.NoError(t, err)
				return stdinManager, bytes.NewBuffer(content)
			},
			func(buf []byte, err error) {
				b, e := ioutil.ReadFile("fixtures/file1.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"Format a stream from stdin fails because reading the stdin stream fails",
			func() (StdinManager, io.Reader) {
				stdinManager := NewStdinManager(2,
					map[string]string{},
				)
				return stdinManager, failingReader{}
			},
			func(buf []byte, err error) {
				assert.Error(t, err)
			},
		},
		{
			"Format an invalid stream from stdin failed",
			func() (StdinManager, io.Reader) {
				stdinManager := NewStdinManager(2,
					map[string]string{
						"seq": "seq 1 3",
					},
				)
				content, err := os.ReadFile("fixtures/invalid.feature")
				assert.NoError(t, err)
				return stdinManager, bytes.NewBuffer(content)
			},
			func(buf []byte, err error) {
				assert.Error(t, err)
			},
		},
		{
			"Format a stream from stdin fails because of an invalid command",
			func() (StdinManager, io.Reader) {
				stdinManager := NewStdinManager(2,
					map[string]string{
						"abcdefg": "abcdefg",
					},
				)
				content, err := os.ReadFile("fixtures/invalid-cmd.feature")
				assert.NoError(t, err)
				return stdinManager, bytes.NewBuffer(content)
			},
			func(buf []byte, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			stdinManager, reader := scenario.setup()
			scenario.test(stdinManager.Transform(reader))
		})
	}
}
