package logic

import (
	"unicode"

	"github.com/google/uuid"
)

func GetUuidStr() string {
	return uuid.New().String()
}

func StartsWithCapital(s string) bool {
	if len(s) == 0 {
		return false
	}
	return unicode.IsUpper(rune(s[0]))
}
