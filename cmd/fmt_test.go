package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"

	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	msgHandler := messageHandler{
		func(exitCode int) {
			panic(exitCode)
		},
		&stdout,
		&stderr,
	}

	cmd := &cobra.Command{}
	args := []string{}

	format(msgHandler, cmd, args)

	assert.EqualValues(t, "", stdout.String())
	assert.EqualValues(t, "", stderr.String())
}
