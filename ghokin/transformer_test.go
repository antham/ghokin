package ghokin

import (
	"bytes"
	"io/ioutil"
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
		{
			Text: "test1",
		},
		{
			Text: "test2",
		},
	}

	expected := []string{"test1", "test2"}

	assert.Equal(t, expected, extractTokensText(tokens))
}

func TestExtractTokensItemsText(t *testing.T) {
	tokens := []*gherkin.Token{
		{
			Items: []*gherkin.LineSpan{
				{Text: "@test1"},
				{Text: "@test2"},
			},
		},
		{
			Items: []*gherkin.LineSpan{
				{Text: "@test3"},
				{Text: "@test4"},
			},
		},
	}

	expected := []string{"@test1 @test2", "@test3 @test4"}

	assert.Equal(t, expected, extractTokensItemsText(tokens))
}

func TestExtractTokensKeywordAndText(t *testing.T) {
	tokens := []*gherkin.Token{
		{Keyword: "Then ", Text: "match some JSON properties"},
		{Keyword: "Then ", Text: "we do something"},
	}

	expected := []string{
		"Then match some JSON properties",
		"Then we do something",
	}

	assert.Equal(t, expected, extractTokensKeywordAndText(tokens))
}

func TestExtractKeywordAndTextSeparatedWithAColon(t *testing.T) {
	tokens := []*gherkin.Token{{Keyword: "Feature", Text: "Set api"}}
	expected := []string{"Feature: Set api"}

	assert.Equal(t, expected, extractKeywordAndTextSeparatedWithAColon(tokens))
}

func TestExtractKeyword(t *testing.T) {
	tokens := []*gherkin.Token{{Keyword: `"""`}}
	expected := []string{`"""`}

	assert.Equal(t, expected, extractKeyword(tokens))
}

func TestExtractTableRows(t *testing.T) {
	type scenario struct {
		tokens []*gherkin.Token
		test   func([]string)
	}

	scenarios := []scenario{
		{
			[]*gherkin.Token{
				{
					Items: []*gherkin.LineSpan{
						{Text: "whatever"},
						{Text: "whatever whatever"},
					},
				},
				{
					Items: []*gherkin.LineSpan{
						{Text: "test"},
						{Text: "test"},
					},
				},
				{
					Items: []*gherkin.LineSpan{
						{Text: "t"},
						{Text: "t"},
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
		tokens []*gherkin.Token
		test   func(*exec.Cmd)
	}

	aliases := map[string]string{
		"cat": "cat",
		"jq":  "jq",
	}

	scenarios := []scenario{
		{
			[]*gherkin.Token{{
				Text: "",
			}},
			func(cmd *exec.Cmd) {
				assert.Nil(t, cmd)
			},
		},
		{
			[]*gherkin.Token{{
				Text: "# A comment",
			}},
			func(cmd *exec.Cmd) {
				assert.Nil(t, cmd)
			},
		},
		{
			[]*gherkin.Token{{
				Text: "# @jq",
			}},
			func(cmd *exec.Cmd) {
				expected := exec.Command("sh", "-c", "jq")

				assert.Equal(t, expected, cmd)
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.test(extractCommand(scenario.tokens, aliases))
	}
}

func TestTrimLinesSpace(t *testing.T) {
	datas := []string{
		"                        hello                          ",
		`		world


		`,
	}

	expected := []string{
		"hello",
		"world",
	}

	assert.Equal(t, expected, trimLinesSpace(datas))
}

func TestRunCommand(t *testing.T) {
	type scenario struct {
		cmd   *exec.Cmd
		lines []string
		test  func([]string, error)
	}

	scenarios := []scenario{
		{
			nil,
			[]string{},
			func(lines []string, err error) {
				assert.Empty(t, lines)
				assert.NoError(t, err)
			},
		},
		{
			exec.Command("sh", "-c", "cat"),
			[]string{"hello world !", "hello universe !"},
			func(lines []string, err error) {
				assert.Equal(t, []string{"hello world !", "hello universe !"}, lines)
				assert.NoError(t, err)
			},
		},
		{
			exec.Command("sh", "-c", "catttttt"),
			[]string{"hello world !", "hello universe !"},
			func(lines []string, err error) {
				assert.Equal(t, []string{}, lines)
				assert.Regexp(t, ".*catttttt.*not found.*", err.Error())
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.test(runCommand(scenario.cmd, scenario.lines))
	}
}

func TestExtractSections(t *testing.T) {
	type scenario struct {
		filename string
		test     func(*section, error)
	}

	scenarios := []scenario{
		{
			"",
			func(section *section, err error) {
				assert.EqualError(t, err, "open : no such file or directory")
			},
		},
		{
			"fixtures/file.txt",
			func(section *section, err error) {
				assert.EqualError(t, err, "Parser errors:\n(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'whatever'")
			},
		},
		{
			"fixtures/feature.feature",
			func(sec *section, err error) {
				type test struct {
					previousName string
					currentName  string
					nextName     string
					values       []map[string]string
				}

				assert.NoError(t, err)
				assert.Equal(t, sec.kind.Name(), "")

				ts := []test{
					{
						"",
						"FeatureLine",
						"Other",
						[]map[string]string{
							{
								"keyword": "Feature",
								"text":    "Test",
							},
						},
					},
					{
						"FeatureLine",
						"Other",
						"BackgroundLine",
						[]map[string]string{
							{
								"keyword": "",
								"text":    "  This is a description",
							},
							{
								"keyword": "",
								"text":    "",
							},
						},
					},
					{
						"Other",
						"BackgroundLine",
						"StepLine",
						[]map[string]string{
							{
								"keyword": "Background",
								"text":    "",
							},
						},
					},
					{
						"BackgroundLine",
						"StepLine",
						"ScenarioLine",
						[]map[string]string{
							{
								"keyword": "Given ",
								"text":    "something",
							},
						},
					},
					{
						"StepLine",
						"ScenarioLine",
						"StepLine",
						[]map[string]string{
							{
								"keyword": "Scenario",
								"text":    "A scenario to test",
							},
						},
					},
					{
						"ScenarioLine",
						"StepLine",
						"ScenarioLine",
						[]map[string]string{
							{
								"keyword": "Given ",
								"text":    "a thing",
							},
							{
								"keyword": "Given ",
								"text":    "something else",
							},
							{
								"keyword": "Then ",
								"text":    "something happened",
							},
						},
					},
					{
						"StepLine",
						"ScenarioLine",
						"StepLine",
						[]map[string]string{
							{
								"keyword": "Scenario",
								"text":    "Another scenario to test",
							},
						},
					},
					{
						"ScenarioLine",
						"StepLine",
						"",
						[]map[string]string{
							{
								"keyword": "Given ",
								"text":    "a second thing",
							},
							{
								"keyword": "Given ",
								"text":    "another second thing",
							},
							{
								"keyword": "Then ",
								"text":    "another thing happened",
							},
						},
					},
				}

				sec = sec.next([]gherkin.TokenType{gherkin.TokenType_Empty})

				for i := 0; i < len(ts); i++ {
					assert.Equal(t, sec.previous([]gherkin.TokenType{gherkin.TokenType_Empty}).kind.Name(), ts[i].previousName)
					assert.Equal(t, sec.kind.Name(), ts[i].currentName)

					if i == len(ts)-1 {
						assert.Equal(t, sec.next([]gherkin.TokenType{gherkin.TokenType_Empty}), (*section)(nil))
					} else {
						assert.Equal(t, sec.next([]gherkin.TokenType{gherkin.TokenType_Empty}).kind.Name(), ts[i].nextName)
					}

					for j, v := range sec.values {
						assert.Equal(t, ts[i].values[j]["keyword"], v.Keyword)
						assert.Equal(t, ts[i].values[j]["text"], v.Text)
					}

					sec = sec.next([]gherkin.TokenType{gherkin.TokenType_Empty})
				}
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.test(extractSections(scenario.filename))
	}
}

func TestTransform(t *testing.T) {
	type scenario struct {
		filename string
		test     func(bytes.Buffer, error)
	}

	scenarios := []scenario{
		{
			"fixtures/file1.feature",
			func(buf bytes.Buffer, err error) {
				assert.NoError(t, err)

				b, e := ioutil.ReadFile("fixtures/file1.feature")

				assert.NoError(t, e)
				assert.EqualValues(t, string(b), buf.String())
			},
		},
		{
			"fixtures/cmd.input.feature",
			func(buf bytes.Buffer, err error) {
				assert.NoError(t, err)

				b, e := ioutil.ReadFile("fixtures/cmd.expected.feature")

				assert.NoError(t, e)
				assert.EqualValues(t, string(b), buf.String())
			},
		},
		{
			"fixtures/multisize-table.input.feature",
			func(buf bytes.Buffer, err error) {
				assert.NoError(t, err)

				b, e := ioutil.ReadFile("fixtures/multisize-table.expected.feature")

				assert.NoError(t, e)
				assert.EqualValues(t, string(b), buf.String())
			},
		},
		{
			"fixtures/docstring-empty.input.feature",
			func(buf bytes.Buffer, err error) {
				assert.NoError(t, err)

				b, e := ioutil.ReadFile("fixtures/docstring-empty.expected.feature")

				assert.NoError(t, e)
				assert.EqualValues(t, string(b), buf.String())
			},
		},
	}

	for _, scenario := range scenarios {
		s, err := extractSections(scenario.filename)

		assert.NoError(t, err)

		aliases := map[string]string{
			"seq": "seq 1 3",
		}

		scenario.test(transform(s, indent{2, 4, 6}, aliases))
	}
}
