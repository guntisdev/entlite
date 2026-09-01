package docs

import (
	"fmt"
	"strings"
)

const (
	markerStart = "<!-- %s:start -->"
	markerEnd   = "<!-- %s:end -->"
)

func region(content, marker string) (string, error) {
	start := strings.Index(content, fmt.Sprintf(markerStart, marker))
	end := strings.Index(content, fmt.Sprintf(markerEnd, marker))
	if start < 0 || end < 0 {
		return "", fmt.Errorf("missing %s markers", marker)
	}
	if end < start {
		return "", fmt.Errorf("%s markers are in the wrong order", marker)
	}

	start += len(fmt.Sprintf(markerStart, marker))

	return strings.TrimSpace(content[start:end]), nil
}

func afterRegion(content, marker string) (string, error) {
	end := strings.Index(content, fmt.Sprintf(markerEnd, marker))
	if end < 0 {
		return "", fmt.Errorf("missing %s markers", marker)
	}

	return strings.TrimSpace(content[end+len(fmt.Sprintf(markerEnd, marker)):]), nil
}

func replaceRegion(content, marker, body string) (string, error) {
	if _, err := region(content, marker); err != nil {
		return "", err
	}

	start := strings.Index(content, fmt.Sprintf(markerStart, marker)) + len(fmt.Sprintf(markerStart, marker))
	end := strings.Index(content, fmt.Sprintf(markerEnd, marker))

	return content[:start] + "\n" + strings.TrimSpace(body) + "\n" + content[end:], nil
}
