package ghokin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatTable(t *testing.T) {
	input := [][]string{
		[]string{
			"whatever",
			"whatever whatever",
		},
		[]string{
			"test",
			"test",
		},
		[]string{
			"t",
			"t",
		},
	}

	expected := `      | whatever | whatever whatever |
      | test     | test              |
      | t        | t                 |
`

	actual, err := formatTableRows(input)

	assert.NoError(t, err)
	assert.Equal(t, expected, string(actual.String()))
}
