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
	/* #nosec */
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

func transform(section *section, indentConf indent, aliases aliases) (bytes.Buffer, error) {
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

	formats := map[gherkin.TokenType](func(values []*gherkin.Token) []string){
		gherkin.TokenType_FeatureLine:        extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenType_BackgroundLine:     extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenType_ScenarioLine:       extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenType_ExamplesLine:       extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenType_Comment:            extractTokensText,
		gherkin.TokenType_TagLine:            extractTokensItemsText,
		gherkin.TokenType_DocStringSeparator: extractKeyword,
		gherkin.TokenType_RuleLine:           extractKeyword,
		gherkin.TokenType_Other:              extractTokensText,
		gherkin.TokenType_StepLine:           extractTokensKeywordAndText,
		gherkin.TokenType_TableRow:           extractTableRows,
		gherkin.TokenType_Empty:              extractTokensItemsText,
		gherkin.TokenType_Language:           extractLanguage,
	}

	var cmd *exec.Cmd
	document := []string{}

	for sec := section; sec != nil; sec = sec.nex {
		if sec.kind == 0 {
			continue
		}

		var err error
		padding := paddings[sec.kind]
		lines := formats[sec.kind](sec.values)

		switch sec.kind {
		case gherkin.TokenType_Comment, gherkin.TokenType_Language:
			cmd = extractCommand(sec.values, aliases)
			padding = getTagOrCommentPadding(paddings, sec)
			lines = trimLinesSpace(lines)
		case gherkin.TokenType_TagLine:
			padding = getTagOrCommentPadding(paddings, sec)
		case gherkin.TokenType_DocStringSeparator, gherkin.TokenType_RuleLine:
			lines = extractKeyword(sec.values)
		case gherkin.TokenType_Other:
			if isDescriptionFeature(sec) {
				lines = trimLinesSpace(lines)
			}
		}

		computed, lines, err := computeCommand(cmd, lines, sec)

		if err != nil {
			return bytes.Buffer{}, err
		}

		if computed {
			cmd = nil
		}

		document = append(document, trimExtraTrailingSpace(indentStrings(padding, lines))...)
	}

	return buildBuffer(document)
}

func buildBuffer(document []string) (bytes.Buffer, error) {
	var buf bytes.Buffer

	if _, err := buf.WriteString(strings.Join(document, "\n") + "\n"); err != nil {
		return bytes.Buffer{}, err
	}

	return buf, nil
}

func getTagOrCommentPadding(paddings map[gherkin.TokenType]int, sec *section) int {
	var kind gherkin.TokenType
	excluded := []gherkin.TokenType{gherkin.TokenType_Empty, gherkin.TokenType_TagLine, gherkin.TokenType_Comment}

	if sec.next(excluded) != nil {
		kind = sec.next(excluded).kind
	}

	if kind == 0 && sec.previous(excluded) != nil {
		kind = sec.previous(excluded).kind
	}

	return paddings[kind]
}

func computeCommand(cmd *exec.Cmd, lines []string, sec *section) (bool, []string, error) {
	if sec.kind == gherkin.TokenType_Comment || sec.kind == gherkin.TokenType_DocStringSeparator || cmd == nil {
		return false, lines, nil
	}

	l, err := runCommand(cmd, lines)

	if err != nil {
		return true, []string{}, err
	}

	return true, l, err
}

func isDescriptionFeature(sec *section) bool {
	excluded := []gherkin.TokenType{gherkin.TokenType_Empty}

	if sec.previous(excluded) != nil && sec.previous(excluded).kind == gherkin.TokenType_FeatureLine {
		return true
	}

	return false
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

func extractLanguage(tokens []*gherkin.Token) []string {
	return []string{fmt.Sprintf("# language: %s", tokens[0].Text)}
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

func extractKeywordAndTextSeparatedWithAColon(tokens []*gherkin.Token) []string {
	return []string{fmt.Sprintf("%s: %s", tokens[0].Keyword, tokens[0].Text)}
}

func extractKeyword(tokens []*gherkin.Token) []string {
	content := []string{}

	for _, t := range tokens {
		content = append(content, t.Keyword)
	}

	return content
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

	lengths := calculateLonguestLineLengthPerColumn(rows)

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

func calculateLonguestLineLengthPerColumn(rows [][]string) []int {
	lengths := []int{}

	for i, row := range rows {
		for j, str := range row {
			switch true {
			case i == 0:
				lengths = append(lengths, len(str))
			case i != 0 && len(lengths) > j && lengths[j] < len(str):
				lengths[j] = len(str)
			default:
				lengths = append(lengths, 0)
			}
		}
	}

	return lengths
}

func extractCommand(tokens []*gherkin.Token, aliases map[string]string) *exec.Cmd {
	re := regexp.MustCompile(`(\@[a-zA-Z0-9]+)`)
	matches := re.FindStringSubmatch(tokens[0].Text)

	if len(matches) == 0 {
		return nil
	}

	/* #nosec */
	if cmd, ok := aliases[matches[0][1:]]; ok {
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
