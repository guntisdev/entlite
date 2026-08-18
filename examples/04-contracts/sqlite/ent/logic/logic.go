package logic

import "strings"

var results = []string{"1-0", "0-1", "1/2-1/2"}

func NotBlank(s string) bool {
	return strings.TrimSpace(s) != ""
}

func IsKnownResult(s string) bool {
	for _, result := range results {
		if s == result {
			return true
		}
	}
	return false
}
