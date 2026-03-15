package logic

import (
	"crypto/rand"
	"fmt"
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

func GenerateAPIKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(fmt.Sprintf("failed to generate secure random bytes: %v", err))
	}
	return key
}
