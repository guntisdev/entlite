package docs

import (
	"fmt"
	"strings"
)

const (
	guideDir     = "guide"
	referenceDir = "reference"
	exampleDir   = "examples"
	docsIndex    = "README.md"
)

func injectDocsIndex(p *Pipeline) error {
	relPath := p.Out(docsIndex)

	content, err := p.Read(relPath)
	if err != nil {
		return err
	}

	body, err := indexBody(p)
	if err != nil {
		return err
	}

	updated, err := replaceRegion(string(content), pagesMarker, body)
	if err != nil {
		return fmt.Errorf("%s: %w", relPath, err)
	}
	p.Write(relPath, []byte(updated))

	return nil
}

func indexBody(p *Pipeline) (string, error) {
	guide, err := pageRows(p, guideDir)
	if err != nil {
		return "", err
	}
	reference, err := pageRows(p, referenceDir)
	if err != nil {
		return "", err
	}

	examples, err := scanExamples(p)
	if err != nil {
		return "", err
	}
	exampleRows := make([][]string, 0, len(examples))
	for _, ex := range examples {
		exampleRows = append(exampleRows, []string{
			fmt.Sprintf("[%s](%s/%s.md)", ex.Title, exampleDir, ex.Slug),
			escapePipe(ex.Intro),
		})
	}

	var body strings.Builder
	body.WriteString(fmt.Sprintf(generatedNotice, "it lists the pages of docs/"))
	body.WriteString("\n## Guide\n\n" + markdownTable([]string{"Page", "About"}, guide))
	body.WriteString("\n## Reference\n\n" + markdownTable([]string{"Page", "About"}, reference))
	body.WriteString("\n## Examples\n\n" + markdownTable([]string{"Page", "About"}, exampleRows))

	return body.String(), nil
}

func pageRows(p *Pipeline, dir string) ([][]string, error) {
	files, err := markdownFiles(p, p.Out(dir))
	if err != nil {
		return nil, err
	}

	rows := make([][]string, 0, len(files))
	for _, relPath := range files {
		content, err := p.Read(relPath)
		if err != nil {
			return nil, err
		}

		title, intro := titleAndIntro(string(content))
		name := strings.TrimPrefix(relPath, p.Out(dir)+"/")
		rows = append(rows, []string{
			fmt.Sprintf("[%s](%s/%s)", title, dir, name),
			escapePipe(firstSentence(intro)),
		})
	}

	return rows, nil
}

func firstSentence(text string) string {
	if end := strings.Index(text, ". "); end > 0 {
		return text[:end+1]
	}

	return text
}
