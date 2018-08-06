package ghokin

import (
	"os/exec"
	"testing"

	"github.com/cucumber/gherkin-go"
	"github.com/stretchr/testify/assert"
)

func TestIndentStrings(t *testing.T) {
	datas := []string{
		"hello",
		"world",
	}

	expected := []string{
		"    hello",
		"    world",
	}

	assert.Equal(t, expected, indentStrings(4, datas))
}

func TestExtractTokensText(t *testing.T) {
	tokens := []*gherkin.Token{
		&gherkin.Token{
			Text: "test1",
		},
		&gherkin.Token{
			Text: "test2",
		},
	}

	expected := []string{"test1", "test2"}

	assert.Equal(t, expected, extractTokensText(tokens))
}

func TestExtractTokensItemsText(t *testing.T) {
	tokens := []*gherkin.Token{
		&gherkin.Token{
			Items: []*gherkin.LineSpan{
				&gherkin.LineSpan{Text: "@test1"},
				&gherkin.LineSpan{Text: "@test2"},
			},
		},
		&gherkin.Token{
			Items: []*gherkin.LineSpan{
				&gherkin.LineSpan{Text: "@test3"},
				&gherkin.LineSpan{Text: "@test4"},
			},
		},
	}

	expected := []string{"@test1 @test2", "@test3 @test4"}

	assert.Equal(t, expected, extractTokensItemsText(tokens))
}

func TestExtractTokensKeywordAndText(t *testing.T) {
	tokens := []*gherkin.Token{
		&gherkin.Token{Keyword: "Then ", Text: "match some JSON properties"},
		&gherkin.Token{Keyword: "Then ", Text: "we do sometging"},
	}

	expected := []string{
		"Then match some JSON properties",
		"Then we do something",
	}

	assert.Equal(t, expected, extractTokensKeywordAndText(tokens))
}

func TestExtractKeywordAndTextSeparatedWithAColon(t *testing.T) {
	token := &gherkin.Token{Keyword: "Feature", Text: "Set api"}
	expected := []string{"Feature: Set api"}

	assert.Equal(t, expected, extractKeywordAndTextSeparatedWithAColon(token))
}

func TestExtractKeyword(t *testing.T) {
	token := &gherkin.Token{Keyword: `"""`}
	expected := []string{`"""`}

	assert.Equal(t, expected, extractKeyword(token))
}

func TestExtractTableRows(t *testing.T) {
	type scenario struct {
		tokens []*gherkin.Token
		test   func([]string)
	}

	scenarios := []scenario{
		{
			[]*gherkin.Token{
				&gherkin.Token{
					Items: []*gherkin.LineSpan{
						&gherkin.LineSpan{Text: "whatever"},
						&gherkin.LineSpan{Text: "whatever whatever"},
					},
				},
				&gherkin.Token{
					Items: []*gherkin.LineSpan{
						&gherkin.LineSpan{Text: "test"},
						&gherkin.LineSpan{Text: "test"},
					},
				},
				&gherkin.Token{
					Items: []*gherkin.LineSpan{
						&gherkin.LineSpan{Text: "t"},
						&gherkin.LineSpan{Text: "t"},
					},
				},
			},
			func(output []string) {
				expected := []string{
					"| whatever | whatever whatever |",
					"| test     | test              |",
					"| t        | t                 |",
				}
				assert.Equal(t, expected, output)
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.test(extractTableRows(scenario.tokens))
	}
}

func TestExtractCommand(t *testing.T) {
	type scenario struct {
		token *gherkin.Token
		test  func(*exec.Cmd)
	}

	commands := map[string]string{
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
		scenario.test(extractCommand(scenario.token, commands))
	}
}
