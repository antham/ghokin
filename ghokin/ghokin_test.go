package ghokin

import (
	"os/exec"
	"testing"

	"github.com/cucumber/gherkin-go"
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

func TestExtractCommand(t *testing.T) {
	type scenario struct {
		token *gherkin.Token
		test  func(*exec.Cmd)
	}

	commandMatcher = map[string]string{
		"cat": "cat",
		"jq":  "jq",
	}

	scenarios := []scenario{
		{
			&gherkin.Token{
				Text: "",
			},
			func(cmd *exec.Cmd) {
				assert.Nil(t, cmd)
			},
		},
		{
			&gherkin.Token{
				Text: "# A comment",
			},
			func(cmd *exec.Cmd) {
				assert.Nil(t, cmd)
			},
		},
		{
			&gherkin.Token{
				Text: "# @jq",
			},
			func(cmd *exec.Cmd) {
				expected := exec.Command("sh", "-c", "jq")

				assert.Equal(t, expected, cmd)
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.test(extractCommand(scenario.token))
	}
}
