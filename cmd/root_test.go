package cmd

import (
	"bytes"
	"os"
	"sync"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var code int
	var w sync.WaitGroup

	msgHandler := messageHandler{
		func(exitCode int) {
			panic(exitCode)
		},
		&stdout,
		&stderr,
	}

	type scenario struct {
		setup    func()
		test     func(exitCode int, stdin string, stderr string)
		teardown func()
	}

	scenarios := []scenario{
		{
			func() {},
			func(exitCode int, stdin string, stderr string) {
				assert.EqualValues(t, 2, viper.GetInt("indent"))
				assert.EqualValues(t, map[string]string{}, viper.GetStringMapString("aliases"))
			},
			func() {},
		},
		{
			func() {
				assert.NoError(t, os.Setenv("GHOKIN_INDENT", "1"))
				assert.NoError(t, os.Setenv("GHOKIN_ALIASES", `{"json":"jq"}`))
			},
			func(exitCode int, stdin string, stderr string) {
				assert.EqualValues(t, 1, viper.GetInt("indent"))
				assert.EqualValues(t, map[string]string{"json": "jq"}, viper.GetStringMapString("aliases"))
			},
			func() {
				assert.NoError(t, os.Unsetenv("GHOKIN_INDENT"))
				assert.NoError(t, os.Unsetenv("GHOKIN_ALIASES"))
			},
		},
		{
			func() {
				assert.NoError(t, os.Setenv("GHOKIN_ALIASES", `{"json":"jq"`))
			},
			func(exitCode int, stdin string, stderr string) {
				assert.EqualValues(t, 1, exitCode)
				assert.EqualValues(t, "check aliases is a well-formed JSON : unexpected end of JSON input\n", stderr)
			},
			func() {
				assert.NoError(t, os.Unsetenv("GHOKIN_ALIASES"))
			},
		},
		{
			func() {
				data := `indent: 12
aliases:
  cat: cat
`
				assert.NoError(t, os.WriteFile(".ghokin.yml", []byte(data), 0o777))
			},
			func(exitCode int, stdin string, stderr string) {
				assert.EqualValues(t, 12, viper.GetInt("indent"))
				assert.EqualValues(t, map[string]string{"cat": "cat"}, viper.GetStringMapString("aliases"))
			},
			func() {
				assert.NoError(t, os.Remove(".ghokin.yml"))
			},
		},
		{
			func() {
				data := `indent: 14
aliases:
  seq: seq
`
				cfgFile = ".test.yml"
				assert.NoError(t, os.WriteFile(".test.yml", []byte(data), 0o777))
			},
			func(exitCode int, stdin string, stderr string) {
				assert.EqualValues(t, 14, viper.GetInt("indent"))
				assert.EqualValues(t, map[string]string{"seq": "seq"}, viper.GetStringMapString("aliases"))
			},
			func() {
				assert.NoError(t, os.Remove(".test.yml"))
				cfgFile = ""
			},
		},
		{
			func() {
				data := `indent`
				assert.NoError(t, os.WriteFile(".ghokin.yml", []byte(data), 0o777))
			},
			func(exitCode int, stdin string, stderr string) {
				assert.EqualValues(t, 1, exitCode)
				assert.EqualValues(t, "check your yaml config file is well-formed : While parsing config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `indent` into map[string]interface {}\n", stderr)
			},
			func() {
				assert.NoError(t, os.Remove(".ghokin.yml"))
			},
		},
	}

	for _, s := range scenarios {
		s.setup()

		w.Add(1)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					code = r.(int)
				}

				w.Done()
			}()

			initConfig(msgHandler)()
		}()

		w.Wait()

		s.test(code, stdout.String(), stderr.String())
		s.teardown()
		viper.Reset()
		stderr.Reset()
		stdout.Reset()
	}
}
