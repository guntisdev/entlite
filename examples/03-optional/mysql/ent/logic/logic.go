package logic

import (
	"strings"

	"github.com/google/uuid"
)

func NewUUID() string {
	return uuid.New().String()
}

func NotBlank(s string) bool {
	return strings.TrimSpace(s) != ""
}
