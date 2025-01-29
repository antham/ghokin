package ghokin

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cucumber/gherkin/go/v28"
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
		gherkin.TokenTypeTableRow:           extractTableRowsAndComments,
		gherkin.TokenTypeEmpty:              extractTokensItemsText,
		gherkin.TokenTypeLanguage:           extractLanguage,
	}

	var cmd *exec.Cmd
	document := []string{}
	optionalRulePadding := 0
	accumulator := []*gherkin.Token{}

	for sec := section; sec != nil; sec = sec.nex {
		values := sec.values
		if len(accumulator) > 0 &&
			sec.kind == gherkin.TokenTypeTableRow &&
			(sec.nex != nil && sec.nex.kind != gherkin.TokenTypeComment) || sec.nex == nil {
			values = append(accumulator, sec.values...)
			accumulator = []*gherkin.Token{}
		}
		if sec.kind == gherkin.TokenTypeTableRow &&
			sec.nex != nil &&
			sec.nex.kind == gherkin.TokenTypeComment &&
			sec.nex.nex != nil &&
			sec.nex.nex.kind == gherkin.TokenTypeTableRow ||
			len(accumulator) > 0 && sec.kind == gherkin.TokenTypeComment ||
			len(accumulator) > 0 && sec.kind == gherkin.TokenTypeTableRow {
			accumulator = append(accumulator, sec.values...)
			continue
		}

		if sec.kind == 0 {
			continue
		}
		padding := paddings[sec.kind] + optionalRulePadding
		lines := formats[sec.kind](values)
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
			} else if isDescriptionScenario(sec) {
				lines = trimLinesSpace(lines)
				padding = paddings[gherkin.TokenTypeScenarioLine]
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
		if s := sec.next(excluded); s != nil {
			kind = s.kind
		}
	}
	if kind == 0 && sec.previous(excluded) != nil {
		if s := sec.previous(excluded); s != nil {
			kind = s.kind
		}
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
	if sec.previous(excluded) != nil {
		if s := sec.previous(excluded); s != nil && s.kind == gherkin.TokenTypeFeatureLine {
			return true
		}
	}
	return false
}

func isDescriptionScenario(sec *section) bool {
	excluded := []gherkin.TokenType{gherkin.TokenTypeEmpty}
	if sec.previous(excluded) != nil {
		if s := sec.previous(excluded); s != nil && s.kind == gherkin.TokenTypeScenarioLine {
			return true
		}
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

func extractTableRowsAndComments(tokens []*gherkin.Token) []string {
	type tableElement struct {
		content []string
		kind    gherkin.TokenType
	}
	rows := [][]string{}
	tableElements := []tableElement{}
	for _, token := range tokens {
		element := tableElement{}
		if token.Type == gherkin.TokenTypeComment {
			element.kind = token.Type
			element.content = []string{token.Text}
		} else {
			row := []string{}
			for _, data := range token.Items {
				// A remaining pipe means it was escaped before to not be messed with pipe column delimiter
				// so here we introduce the escaping sequence back
				text := data.Text
				text = strings.ReplaceAll(text, "\\", "\\\\")
				text = strings.ReplaceAll(text, "\n", "\\n")
				text = strings.ReplaceAll(text, "|", "\\|")
				row = append(row, text)
			}
			element.kind = token.Type
			element.content = row
			rows = append(rows, row)
		}
		tableElements = append(tableElements, element)
	}

	var tableRows []string
	lengths := calculateLonguestLineLengthPerColumn(rows)
	for _, tableElement := range tableElements {
		inputs := []interface{}{}
		fmtDirective := ""
		if tableElement.kind == gherkin.TokenTypeComment {
			inputs = append(inputs, trimLinesSpace(tableElement.content)[0])
			fmtDirective = "%s"
		} else {
			for i, str := range tableElement.content {
				inputs = append(inputs, str)
				fmtDirective += "| %-" + strconv.Itoa(lengths[i]) + "s "
			}
			fmtDirective += "|"
		}
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
