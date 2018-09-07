package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	var code int
	var w sync.WaitGroup
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	viper.Set("indent.backgroundAndScenario", 2)
	viper.Set("indent.step", 4)
	viper.Set("indent.tableAndDocString", 6)

	msgHandler := messageHandler{
		func(exitCode int) {
			panic(exitCode)
		},
		&stdout,
		&stderr,
	}

	assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
	assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0777))
	assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/file1.feature", []byte("Feature: Test\n      Test\n  Scenario: Scenario1\n    Given a test\n"), 0755))
	assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/file2.feature", []byte("Feature: Test\n      Test\n  Scenario: Scenario2\n    Given a test\n"), 0755))

	w.Add(1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				code = r.(int)
			}

			w.Done()
		}()

		cmd := &cobra.Command{}

		check(msgHandler, cmd, []string{"/tmp/ghokin"})
	}()

	w.Wait()

	assert.EqualValues(t, 0, code, "Must exit with errors (exit 0)")
	assert.EqualValues(t, `"/tmp/ghokin" is well formatted`+"\n", stdout.String())
}

func TestCheckErrors(t *testing.T) {
	var code int
	var w sync.WaitGroup
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	msgHandler := messageHandler{
		func(exitCode int) {
			panic(exitCode)
		},
		&stdout,
		&stderr,
	}

	type scenario struct {
		args   []string
		errMsg string
	}

	scenarios := []scenario{
		{
			[]string{},
			"you must provide a filename or a folder as argument\n",
		},
		{
			[]string{"fixtures/whatever.feature"},
			"stat fixtures/whatever.feature: no such file or directory\n",
		},
		{
			[]string{"fixtures/file.txt"},
			"Parser errors:\n(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'Whatever'\n",
		},
	}

	for _, s := range scenarios {
		w.Add(1)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					code = r.(int)
				}

				w.Done()
			}()

			cmd := &cobra.Command{}

			check(msgHandler, cmd, s.args)
		}()

		w.Wait()

		assert.EqualValues(t, 1, code, "Must exit with errors (exit 1)")
		assert.EqualValues(t, s.errMsg, stderr.String())

		stderr.Reset()
		stdout.Reset()
	}
}
