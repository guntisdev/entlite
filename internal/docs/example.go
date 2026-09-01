package docs

import (
	"fmt"
	"os"
	"strings"
)

const examplesDir = "examples"
const teachesMarker = "teaches"

type example struct {
	Slug    string   // directory name, e.g. 01-basic-entity
	Path    string   // path from the repo root, e.g. examples/01-basic-entity
	Title   string   // the h1 of its readme
	Intro   string   // the paragraph under the h1
	Teaches []string // the bullets of the teaches block
	Readme  string   // the whole readme
}

func scanExamples(p *Pipeline) ([]example, error) {
	entries, err := os.ReadDir(p.Abs(examplesDir))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", examplesDir, err)
	}

	var examples []example
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := examplesDir + "/" + entry.Name()
		readme, err := p.Read(path + "/README.md")
		if err != nil {
			return nil, err
		}

		parsed, err := newExample(entry.Name(), path, string(readme))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		examples = append(examples, parsed)
	}

	return examples, nil
}

func newExample(slug, path, readme string) (example, error) {
	teaches, err := region(readme, teachesMarker)
	if err != nil {
		return example{}, err
	}

	var bullets []string
	for _, line := range strings.Split(teaches, "\n") {
		line = strings.TrimSpace(line)
		if after, found := strings.CutPrefix(line, "- "); found {
			bullets = append(bullets, after)
		}
	}

	title, intro := titleAndIntro(readme)

	return example{
		Slug:    slug,
		Path:    path,
		Title:   title,
		Intro:   intro,
		Teaches: bullets,
		Readme:  readme,
	}, nil
}

func titleAndIntro(readme string) (string, string) {
	title := ""
	var intro []string

	for _, line := range strings.Split(readme, "\n") {
		line = strings.TrimSpace(line)

		if title == "" {
			if after, found := strings.CutPrefix(line, "# "); found {
				title = after
			}
			continue
		}

		if line == "" {
			if len(intro) > 0 {
				break // the intro paragraph ended
			}
			continue
		}
		if strings.HasPrefix(line, "<!--") || strings.HasPrefix(line, "#") {
			break // a marker or the next heading, there is no intro
		}
		intro = append(intro, line)
	}

	return title, strings.Join(intro, " ")
}
