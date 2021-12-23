package cmd

import (
	"bytes"
	"io/ioutil"
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
				assert.EqualValues(t, 2, viper.GetInt("indent.featureDescription"))
				assert.EqualValues(t, 2, viper.GetInt("indent.backgroundAndScenario"))
				assert.EqualValues(t, 4, viper.GetInt("indent.step"))
				assert.EqualValues(t, 6, viper.GetInt("indent.tableAndDocString"))
				assert.EqualValues(t, map[string]string{}, viper.GetStringMapString("aliases"))
			},
			func() {},
		},
		{
			func() {
				os.Setenv("GHOKIN_INDENT_FEATUREDESCRIPTION", "1")
				os.Setenv("GHOKIN_INDENT_BACKGROUNDANDSCENARIO", "4")
				os.Setenv("GHOKIN_INDENT_STEP", "6")
				os.Setenv("GHOKIN_INDENT_TABLEANDDOCSTRING", "8")
				os.Setenv("GHOKIN_ALIASES", `{"json":"jq"}`)
			},
			func(exitCode int, stdin string, stderr string) {
				assert.EqualValues(t, 1, viper.GetInt("indent.featureDescription"))
				assert.EqualValues(t, 4, viper.GetInt("indent.backgroundAndScenario"))
				assert.EqualValues(t, 6, viper.GetInt("indent.step"))
				assert.EqualValues(t, 8, viper.GetInt("indent.tableAndDocString"))
				assert.EqualValues(t, map[string]string{"json": "jq"}, viper.GetStringMapString("aliases"))
			},
			func() {
				os.Unsetenv("GHOKIN_INDENT_FEATUREDESCRIPTION")
				os.Unsetenv("GHOKIN_INDENT_BACKGROUNDANDSCENARIO")
				os.Unsetenv("GHOKIN_INDENT_STEP")
				os.Unsetenv("GHOKIN_INDENT_TABLEANDDOCSTRING")
				os.Unsetenv("GHOKIN_ALIASES")
			},
		},
		{
			func() {
				os.Setenv("GHOKIN_ALIASES", `{"json":"jq"`)
			},
			func(exitCode int, stdin string, stderr string) {
				assert.EqualValues(t, 1, exitCode)
				assert.EqualValues(t, "check aliases is a well-formed JSON : unexpected end of JSON input\n", stderr)
			},
			func() {
				os.Unsetenv("GHOKIN_ALIASES")
			},
		},
		{
			func() {
				data := `indent:
  featureDescription: 6
  backgroundAndScenario: 8
  step: 10
  tableAndDocString: 12
aliases:
  cat: cat
`
				assert.NoError(t, ioutil.WriteFile(".ghokin.yml", []byte(data), 0777))
			},
			func(exitCode int, stdin string, stderr string) {
				assert.EqualValues(t, 6, viper.GetInt("indent.featureDescription"))
				assert.EqualValues(t, 8, viper.GetInt("indent.backgroundAndScenario"))
				assert.EqualValues(t, 10, viper.GetInt("indent.step"))
				assert.EqualValues(t, 12, viper.GetInt("indent.tableAndDocString"))
				assert.EqualValues(t, map[string]string{"cat": "cat"}, viper.GetStringMapString("aliases"))
			},
			func() {
				assert.NoError(t, os.Remove(".ghokin.yml"))
			},
		},
		{
			func() {
				data := `indent:
  featureDescription: 8
  backgroundAndScenario: 10
  step: 12
  tableAndDocString: 14
aliases:
  seq: seq
`
				cfgFile = ".test.yml"
				assert.NoError(t, ioutil.WriteFile(".test.yml", []byte(data), 0777))
			},
			func(exitCode int, stdin string, stderr string) {
				assert.EqualValues(t, 8, viper.GetInt("indent.featureDescription"))
				assert.EqualValues(t, 10, viper.GetInt("indent.backgroundAndScenario"))
				assert.EqualValues(t, 12, viper.GetInt("indent.step"))
				assert.EqualValues(t, 14, viper.GetInt("indent.tableAndDocString"))
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
				assert.NoError(t, ioutil.WriteFile(".ghokin.yml", []byte(data), 0777))
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
