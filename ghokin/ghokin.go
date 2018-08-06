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


var featureDescIndent = 4
var tableIndent = 6
var stepIndent = 4


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
	re := regexp.MustCompile("(\\@[a-zA-Z0-9]+)")
	matches := re.FindStringSubmatch(token.Text)

	if len(matches) == 0 {
		return nil
	}

	if cmd, ok := commands[matches[0][1:]]; ok {
		return exec.Command("sh", "-c", cmd)
	}

	return nil
}
