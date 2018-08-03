package ghokin

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

var tableIndent = 6

func formatTableRows(rows [][]string) (bytes.Buffer, error) {
	var output bytes.Buffer

	lengths := calculateLonguestLineLengthPerRow(rows)

	for _, row := range rows {
		inputs := []interface{}{}
		fmtDirective := strings.Repeat(" ", tableIndent)

		for i, str := range row {
			inputs = append(inputs, str)
			fmtDirective += "| %-" + strconv.Itoa(lengths[i]) + "s "
		}

		fmtDirective += "|\n"

		if _, err := output.WriteString(fmt.Sprintf(fmtDirective, inputs...)); err != nil {
			return bytes.Buffer{}, err
		}
	}

	return output, nil
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
