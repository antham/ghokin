package cmd

import (
	"bytes"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

func TestFormatOnStdout(t *testing.T) {
	var code int
	var w sync.WaitGroup
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	viper.Set("indent", 2)

	msgHandler := messageHandler{
		func(exitCode int) {
			panic(exitCode)
		},
		&stdout,
		&stderr,
	}

	w.Add(1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				code = r.(int)
			}

			w.Done()
		}()

		cmd := &cobra.Command{}
		args := []string{"fixtures/feature.feature"}

		formatOnStdout(msgHandler, cmd, args)
	}()

	w.Wait()

	b, err := ioutil.ReadFile("fixtures/feature.feature")

	assert.NoError(t, err)

	assert.EqualValues(t, 0, code, "Must exit with no errors (exit 0)")
	assert.EqualValues(t, string(b), stdout.String())
}

func TestFormatOnStdoutWithErrors(t *testing.T) {
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
			"you must provide a filename as argument\n",
		},
		{
			[]string{"fixtures/featurefeature.feature"},
			"open fixtures/featurefeature.feature: no such file or directory\n",
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

			formatOnStdout(msgHandler, cmd, s.args)
		}()

		w.Wait()

		assert.EqualValues(t, 1, code, "Must exit with errors (exit 1)")
		assert.EqualValues(t, s.errMsg, stderr.String())

		stderr.Reset()
		stdout.Reset()
	}
}
