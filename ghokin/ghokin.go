package ghokin

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/cucumber/gherkin-go"
)

// CmdErr is thrown when an error occurred when calling
// a command on an input, both stdout and stderr are stored
type CmdErr struct {
	output string
}

// Error outputs both stdout and stderr
func (e CmdErr) Error() string {
	return e.output
}

func extractSections(filename string) (*section, error) {
	file, err := os.Open(filename)

	if err != nil {
		return &section{}, err
	}

	section := &section{}
	builder := &tokenGenerator{section: section}

	matcher := gherkin.NewMatcher(gherkin.GherkinDialectsBuildin())
	scanner := gherkin.NewScanner(file)
	parser := gherkin.NewParser(builder)

	parser.StopAtFirstError(true)

	return section, parser.Parse(scanner, matcher)
}

func transform(section *section, indentConf indent, commands commands) (bytes.Buffer, error) {
	paddings := map[gherkin.TokenType]int{
		gherkin.TokenType_FeatureLine:        0,
		gherkin.TokenType_BackgroundLine:     indentConf.backgroundAndScenario,
		gherkin.TokenType_ScenarioLine:       indentConf.backgroundAndScenario,
		gherkin.TokenType_DocStringSeparator: indentConf.tableAndDocString,
		gherkin.TokenType_RuleLine:           indentConf.tableAndDocString,
		gherkin.TokenType_StepLine:           indentConf.step,
		gherkin.TokenType_ExamplesLine:       indentConf.step,
		gherkin.TokenType_Other:              indentConf.tableAndDocString,
		gherkin.TokenType_TableRow:           indentConf.tableAndDocString,
	}

	var padding int
	var cmd *exec.Cmd
	document := []string{}

	for sec := section; sec != nil; sec = sec.nex {
		var lines []string

		switch sec.kind {
		case gherkin.TokenType_FeatureLine:
			lines = extractKeywordAndTextSeparatedWithAColon(sec.values[0])
			padding = paddings[sec.kind]
		case gherkin.TokenType_BackgroundLine, gherkin.TokenType_ScenarioLine:
			lines = extractKeywordAndTextSeparatedWithAColon(sec.values[0])
			padding = paddings[sec.kind]
		case gherkin.TokenType_Comment:
			var nt gherkin.TokenType
			excluded := []gherkin.TokenType{gherkin.TokenType_Empty, gherkin.TokenType_TagLine, gherkin.TokenType_Comment}

			if sec.next(excluded) != nil {
				nt = sec.next(excluded).kind
			}

			if nt == 0 && sec.previous(excluded) != nil {
				nt = sec.previous(excluded).kind
			}

			cmd = extractCommand(sec.values[0], commands)
			lines = trimLinesSpace(extractTokensText(sec.values))
			padding = paddings[nt]
		case gherkin.TokenType_TagLine:
			var nt gherkin.TokenType
			excluded := []gherkin.TokenType{gherkin.TokenType_Empty, gherkin.TokenType_TagLine, gherkin.TokenType_Comment}

			if sec.next(excluded) != nil {
				nt = sec.next(excluded).kind
			}

			if nt == 0 && sec.previous(excluded) != nil {
				nt = sec.previous(excluded).kind
			}

			lines = extractTokensItemsText(sec.values)
			padding = paddings[nt]
		case gherkin.TokenType_DocStringSeparator, gherkin.TokenType_RuleLine:
			lines = extractKeyword(sec.values[0])
			padding = paddings[sec.kind]
		case gherkin.TokenType_Other:
			lines = extractTokensText(sec.values)
			excluded := []gherkin.TokenType{gherkin.TokenType_Empty}

			if sec.previous(excluded) != nil && sec.previous(excluded).kind == gherkin.TokenType_FeatureLine {
				lines = trimLinesSpace(lines)
			}

			padding = paddings[sec.kind]
		case gherkin.TokenType_StepLine:
			lines = extractTokensKeywordAndText(sec.values)
			padding = paddings[sec.kind]
		case gherkin.TokenType_ExamplesLine:
			lines = extractKeywordAndTextSeparatedWithAColon(sec.values[0])
			padding = paddings[sec.kind]
		case gherkin.TokenType_TableRow:
			lines = extractTableRows(sec.values)
			padding = paddings[sec.kind]
		case gherkin.TokenType_Empty:
			lines = extractTokensItemsText(sec.values)
		}

		if sec.kind != gherkin.TokenType_Comment && sec.kind != gherkin.TokenType_DocStringSeparator && cmd != nil {
			l, err := runCommand(cmd, lines)

			if err != nil {
				return bytes.Buffer{}, err
			}

			lines = l
			cmd = nil
		}

		document = append(document, trimExtraTrailingSpace(indentStrings(padding, lines))...)
	}

	var buf bytes.Buffer

	if _, err := buf.WriteString(strings.Join(document, "\n")); err != nil {
		return bytes.Buffer{}, err
	}

	return buf, nil
}

func trimLinesSpace(lines []string) []string {
	content := []string{}

	for _, line := range lines {
		content = append(content, strings.TrimSpace(line))
	}

	return content
}

func trimExtraTrailingSpace(lines []string) []string {
	content := []string{}

	for _, line := range lines {
		content = append(content, strings.TrimRight(line, " \t"))
	}

	return content
}

func indentStrings(padding int, lines []string) []string {
	content := []string{}

	for _, line := range lines {
		content = append(content, strings.Repeat(" ", padding)+line)
	}

	return content
}

func extractTokensText(tokens []*gherkin.Token) []string {
	content := []string{}

	for _, token := range tokens {
		content = append(content, token.Text)
	}

	return content
}

func extractTokensItemsText(tokens []*gherkin.Token) []string {
	content := []string{}

	for _, token := range tokens {
		t := []string{}

		for _, item := range token.Items {
			t = append(t, item.Text)
		}

		content = append(content, strings.Join(t, " "))
	}

	return content
}

func extractTokensKeywordAndText(tokens []*gherkin.Token) []string {
	content := []string{}

	for _, token := range tokens {
		content = append(content, fmt.Sprintf("%s%s", token.Keyword, token.Text))
	}

	return content
}

func extractKeywordAndTextSeparatedWithAColon(token *gherkin.Token) []string {
	return []string{fmt.Sprintf("%s: %s", token.Keyword, token.Text)}
}

func extractKeyword(token *gherkin.Token) []string {
	return []string{token.Keyword}
}

func extractTableRows(tokens []*gherkin.Token) []string {
	rows := [][]string{}

	for _, tab := range tokens {
		row := []string{}

		for _, data := range tab.Items {
			row = append(row, data.Text)
		}

		rows = append(rows, row)
	}

	var tableRows []string

	lengths := calculateLonguestLineLengthPerRow(rows)

	for _, row := range rows {
		inputs := []interface{}{}
		fmtDirective := ""

		for i, str := range row {
			inputs = append(inputs, str)
			fmtDirective += "| %-" + strconv.Itoa(lengths[i]) + "s "
		}

		fmtDirective += "|"

		tableRows = append(tableRows, fmt.Sprintf(fmtDirective, inputs...))
	}

	return tableRows
}

func calculateLonguestLineLengthPerRow(rows [][]string) []int {
	lengths := []int{}

	for i, row := range rows {
		for j, str := range row {
			if i == 0 {
				lengths = append(lengths, len(str))
			}

			if i != 0 && lengths[j] < len(str) {
				lengths[j] = len(str)
			}
		}
	}

	return lengths
}

func extractCommand(token *gherkin.Token, commands map[string]string) *exec.Cmd {
	re := regexp.MustCompile(`(\@[a-zA-Z0-9]+)`)
	matches := re.FindStringSubmatch(token.Text)

	if len(matches) == 0 {
		return nil
	}

	if cmd, ok := commands[matches[0][1:]]; ok {
		return exec.Command("sh", "-c", cmd)
	}

	return nil
}

func runCommand(cmd *exec.Cmd, lines []string) ([]string, error) {
	if len(lines) == 0 {
		return lines, nil
	}

	datas := strings.Join(lines, "\n")
	cmd.Stdin = strings.NewReader(datas)

	o, err := cmd.CombinedOutput()

	if err != nil {
		return []string{}, CmdErr{strings.TrimRight(string(o), "\n")}
	}

	return strings.Split(strings.TrimRight(string(o), "\n"), "\n"), nil
}
