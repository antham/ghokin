package ghokin

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/cucumber/gherkin-go"
)

var commandMatcher map[string]string

var tableIndent = 6
var stepIndent = 4
func formatDocStringContent(tokens []*gherkin.Token) string {
	content := []string{}

	for _, token := range tokens {
		content = append(content, strings.Repeat(" ", tableIndent)+token.Text)
	}

	return strings.Join(content, "\n") + "\n"
}
func formatTags(tokens []*gherkin.Token) string {
	tags := []string{}

	for _, token := range tokens {
		for _, data := range token.Items {
			tags = append(tags, data.Text)
		}
	}

	return fmt.Sprintf("%s\n", strings.Join(tags, " "))
}

func formatComments(tokens []*gherkin.Token) string {
	comments := []string{}

	for _, token := range tokens {
		comments = append(comments, token.Text)
	}

	return strings.Join(comments, "\n")
}

func formatStepOrExampleLine(token *gherkin.Token) string {
	return fmt.Sprintf("%s%s%s\n", strings.Repeat(" ", stepIndent), token.Keyword, token.Text)
}

func formatFeatureOrBackgroundLine(token *gherkin.Token) string {
	return fmt.Sprintf("%s: %s\n", token.Keyword, token.Text)
}

func formatDocStringOrRuleLine(token *gherkin.Token) string {
	return fmt.Sprintf("%s%s\n", strings.Repeat(" ", tableIndent), token.Keyword)
}

func formatTableRows(tokens []*gherkin.Token) string {
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
		fmtDirective := strings.Repeat(" ", tableIndent)

		for i, str := range row {
			inputs = append(inputs, str)
			fmtDirective += "| %-" + strconv.Itoa(lengths[i]) + "s "
		}

		fmtDirective += "|"

		tableRows = append(tableRows, fmt.Sprintf(fmtDirective, inputs...))
	}

	return strings.Join(tableRows, "\n") + "\n"
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

func extractCommand(comment *gherkin.Token) *exec.Cmd {
	re := regexp.MustCompile("(\\@[a-zA-Z0-9]+)")
	matches := re.FindStringSubmatch(comment.Text)

	if len(matches) == 0 {
		return nil
	}

	if cmd, ok := commandMatcher[matches[0][1:]]; ok {
		return exec.Command("sh", "-c", cmd)
	}

	return nil
}
