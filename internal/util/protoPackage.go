package util

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var nonIdentChars = regexp.MustCompile(`[^a-z0-9]+`)
var leadingDigits = regexp.MustCompile(`^[0-9_]+`)

// generates proto package name (e.g. "examples.basic_entity.v1")
// from entDir's path relative to its Go module root when there is no entlite.yaml
func AutoProtoPackage(entDir string) (string, error) {
	absDir, err := filepath.Abs(entDir)
	if err != nil {
		return "", fmt.Errorf("getting absolute path: %w", err)
	}

	_, moduleRoot, err := FindModuleInfo(absDir)
	if err != nil {
		return "", fmt.Errorf("finding module info: %w", err)
	}

	relPath, err := filepath.Rel(moduleRoot, absDir)
	if err != nil {
		return "", fmt.Errorf("getting relative path: %w", err)
	}

	segments := strings.Split(filepath.ToSlash(relPath), "/")
	sanitized := make([]string, 0, len(segments)+1)
	for _, seg := range segments {
		if s := sanitizeProtoSegment(seg); s != "" {
			sanitized = append(sanitized, s)
		}
	}
	sanitized = append(sanitized, "v1")

	return strings.Join(sanitized, "."), nil
}

// make valid proto package name:
// lowercase letters, digits and underscores, not starting with a digit.
func sanitizeProtoSegment(seg string) string {
	seg = strings.ToLower(seg)
	seg = nonIdentChars.ReplaceAllString(seg, "_")
	seg = strings.Trim(seg, "_")
	seg = leadingDigits.ReplaceAllString(seg, "")
	seg = strings.Trim(seg, "_")
	return seg
}
