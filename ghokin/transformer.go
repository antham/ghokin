package ghokin

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cucumber/gherkin/go/v26"
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

func extractSections(content []byte) (*section, error) {
	section := &section{}
	builder := &tokenGenerator{section: section}
	matcher := gherkin.NewMatcher(gherkin.DialectsBuiltin())
	scanner := gherkin.NewScanner(bytes.NewBuffer(content))
	parser := gherkin.NewParser(builder)
	parser.StopAtFirstError(true)
	return section, parser.Parse(scanner, matcher)
}

func transform(section *section, indent int, aliases aliases) ([]byte, error) {
	paddings := map[gherkin.TokenType]int{
		gherkin.TokenTypeFeatureLine:        0,
		gherkin.TokenTypeBackgroundLine:     indent,
		gherkin.TokenTypeScenarioLine:       indent,
		gherkin.TokenTypeDocStringSeparator: 3 * indent,
		gherkin.TokenTypeStepLine:           2 * indent,
		gherkin.TokenTypeExamplesLine:       2 * indent,
		gherkin.TokenTypeOther:              3 * indent,
		gherkin.TokenTypeTableRow:           3 * indent,
	}

	formats := map[gherkin.TokenType](func(values []*gherkin.Token) []string){
		gherkin.TokenTypeFeatureLine:        extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenTypeBackgroundLine:     extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenTypeScenarioLine:       extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenTypeExamplesLine:       extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenTypeComment:            extractTokensText,
		gherkin.TokenTypeTagLine:            extractTokensItemsText,
		gherkin.TokenTypeDocStringSeparator: extractKeyword,
		gherkin.TokenTypeRuleLine:           extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenTypeOther:              extractTokensText,
		gherkin.TokenTypeStepLine:           extractTokensKeywordAndText,
		gherkin.TokenTypeTableRow:           extractTableRows,
		gherkin.TokenTypeEmpty:              extractTokensItemsText,
		gherkin.TokenTypeLanguage:           extractLanguage,
	}

	var cmd *exec.Cmd
	document := []string{}
	optionalRulePadding := 0

	for sec := section; sec != nil; sec = sec.nex {
		if sec.kind == 0 {
			continue
		}

		padding := paddings[sec.kind] + optionalRulePadding
		lines := formats[sec.kind](sec.values)
		switch sec.kind {
		case gherkin.TokenTypeRuleLine:
			optionalRulePadding = indent
			padding = indent
		case gherkin.TokenTypeComment, gherkin.TokenTypeLanguage:
			cmd = extractCommand(sec.values, aliases)
			padding = getTagOrCommentPadding(paddings, indent, sec)
			lines = trimLinesSpace(lines)
		case gherkin.TokenTypeTagLine:
			padding = getTagOrCommentPadding(paddings, indent, sec)
		case gherkin.TokenTypeDocStringSeparator:
			lines = extractKeyword(sec.values)
		case gherkin.TokenTypeOther:
			if isDescriptionFeature(sec) {
				lines = trimLinesSpace(lines)
				padding = indent
			}
		}

		computed, lines, err := computeCommand(cmd, lines, sec)
		if err != nil {
			return []byte{}, err
		}
		if computed {
			cmd = nil
		}
		document = append(document, trimExtraTrailingSpace(indentStrings(padding, lines))...)
	}
	return []byte(strings.Join(document, "\n") + "\n"), nil
}

func getTagOrCommentPadding(paddings map[gherkin.TokenType]int, indent int, sec *section) int {
	var kind gherkin.TokenType
	excluded := []gherkin.TokenType{gherkin.TokenTypeTagLine, gherkin.TokenTypeComment}
	if sec.next(excluded) != nil {
		kind = sec.next(excluded).kind
	}
	if kind == 0 && sec.previous(excluded) != nil {
		kind = sec.previous(excluded).kind
	}
	// indent the last comment line at the same level than scenario and background
	if sec.next([]gherkin.TokenType{gherkin.TokenTypeEmpty}) == nil {
		return indent
	}
	return paddings[kind]
}

func computeCommand(cmd *exec.Cmd, lines []string, sec *section) (bool, []string, error) {
	if sec.kind == gherkin.TokenTypeComment || sec.kind == gherkin.TokenTypeDocStringSeparator || cmd == nil {
		return false, lines, nil
	}
	l, err := runCommand(cmd, lines)
	if err != nil {
		return true, []string{}, err
	}
	return true, l, err
}

func isDescriptionFeature(sec *section) bool {
	excluded := []gherkin.TokenType{gherkin.TokenTypeEmpty}
	if sec.previous(excluded) != nil && sec.previous(excluded).kind == gherkin.TokenTypeFeatureLine {
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
	content := []string{}
	for _, token := range tokens {
		content = append(content, fmt.Sprintf("%s: %s", token.Keyword, token.Text))
	}
	return content
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
			// A remaining pipe means it was escaped before to not be messed with pipe column delimiter
			// so here we introduce the escaping sequence back
			text := data.Text
			text = strings.ReplaceAll(text, "|", "\\|")
			text = strings.ReplaceAll(text, "\n", "\\n")
			row = append(row, text)
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
				lengths = append(lengths, utf8.RuneCountInString(str))
			case i != 0 && len(lengths) > j && lengths[j] < utf8.RuneCountInString(str):
				lengths[j] = utf8.RuneCountInString(str)
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

	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n"))
	o, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, CmdErr{strings.TrimRight(string(o), "\n")}
	}
	return strings.Split(strings.TrimRight(string(o), "\n"), "\n"), nil
}
