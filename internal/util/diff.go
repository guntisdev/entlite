package testutil

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var splitLines = cmpopts.AcyclicTransformer("SplitLines", func(s string) []string {
	return strings.Split(strings.TrimSpace(s), "\n")
})

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
)

// Diff styles cmp.Diff in readable format
func Diff(expected, actual string) string {
	d := cmp.Diff(expected, actual, splitLines)
	if d == "" {
		return ""
	}

	// Remove the wrapper lines that cmp.Diff adds
	lines := strings.Split(d, "\n")
	if len(lines) > 2 {
		lines = lines[1 : len(lines)-1]
	}

	for i, line := range lines {
		if strings.HasPrefix(line, "-") {
			lines[i] = colorRed + line + colorReset
		} else if strings.HasPrefix(line, "+") {
			lines[i] = colorGreen + line + colorReset
		}
	}

	return strings.Join(lines, "\n")
}
