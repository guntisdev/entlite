package logic

import (
	"strings"

	"github.com/google/uuid"
)

// NewUUID generates a random UUID string, used as the Article primary key.
func NewUUID() string {
	return uuid.New().String()
}

// NotBlank rejects empty or whitespace-only strings.
func NotBlank(s string) bool {
	return strings.TrimSpace(s) != ""
}
