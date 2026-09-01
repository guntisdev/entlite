package docs

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

var variantOrder = []string{"sqlite", "postgresql", "mysql"}
var contractFiles = []struct{ Path, Lang string }{
	{Path: "ent/contract/sqlc/schema.sql", Lang: "sql"},
	{Path: "ent/contract/sqlc/queries.sql", Lang: "sql"},
	{Path: "ent/contract/proto/schema.proto", Lang: "proto"},
}

var mdLink = regexp.MustCompile(`\]\(([^)]+)\)`)

// examples writes one page per example, built from its readme and its ent files.
func examples(p *Pipeline) error {
	list, err := scanExamples(p)
	if err != nil {
		return err
	}

	for _, ex := range list {
		content, err := examplePage(p, ex)
		if err != nil {
			return fmt.Errorf("%s: %w", ex.Path, err)
		}
		p.Write(p.Out("examples", ex.Slug+".md"), content)
	}

	return nil
}

func examplePage(p *Pipeline, ex example) ([]byte, error) {
	page := newPage(ex.Title, ex.Intro)

	variants, err := exampleVariants(p, ex)
	if err != nil {
		return nil, err
	}
	page.Text(fmt.Sprintf("Source: [%s](%s)", ex.Path, docsToRepo(ex.Path)))

	page.Heading(2, "What it teaches")
	page.Text("- " + strings.Join(ex.Teaches, "\n- "))

	body, err := afterRegion(ex.Readme, teachesMarker)
	if err != nil {
		return nil, err
	}
	page.Text(rewriteLinks(body, ex.Path))

	if len(variants) == 0 {
		return page.Bytes(), nil
	}
	variant := variants[0]

	page.Heading(2, "Schema")
	page.Text(fmt.Sprintf("The dsl input, from the `%s` variant.", variant))
	if err := writeSchemaFiles(p, page, ex, variant); err != nil {
		return nil, err
	}

	page.Heading(2, "Generated contract")
	page.Text(fmt.Sprintf("What `entlite gen` writes from the schema above. See %s for the other variants.",
		variantLinks(ex, variants)))
	if err := writeContractFiles(p, page, ex, variant); err != nil {
		return nil, err
	}

	return page.Bytes(), nil
}

func writeSchemaFiles(p *Pipeline, page *page, ex example, variant string) error {
	dir := path.Join(ex.Path, variant, "ent", "schema")

	entries, err := os.ReadDir(p.Abs(dir))
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", dir, err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || name == "generate.go" {
			continue
		}

		content, err := p.Read(path.Join(dir, name))
		if err != nil {
			return err
		}
		page.Heading(3, fmt.Sprintf("[%s](%s)", name, docsToRepo(path.Join(dir, name))))
		page.Code("go", string(content))
	}

	return nil
}

func writeContractFiles(p *Pipeline, page *page, ex example, variant string) error {
	for _, file := range contractFiles {
		relPath := path.Join(ex.Path, variant, file.Path)

		content, err := p.Read(relPath)
		if os.IsNotExist(err) {
			continue // an example without that contract
		}
		if err != nil {
			return err
		}
		page.Details(file.Path, file.Lang, string(content))
	}

	return nil
}

func exampleVariants(p *Pipeline, ex example) ([]string, error) {
	entries, err := os.ReadDir(p.Abs(ex.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", ex.Path, err)
	}

	found := map[string]bool{}
	var rest []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(p.Abs(path.Join(ex.Path, entry.Name(), "ent"))); err != nil {
			continue // not a dialect folder
		}
		found[entry.Name()] = true
		rest = append(rest, entry.Name())
	}

	var variants []string
	for _, name := range variantOrder {
		if found[name] {
			variants = append(variants, name)
			delete(found, name)
		}
	}
	for _, name := range rest {
		if found[name] {
			variants = append(variants, name)
		}
	}

	return variants, nil
}

func variantLinks(ex example, variants []string) string {
	links := make([]string, 0, len(variants))
	for _, variant := range variants {
		links = append(links, fmt.Sprintf("[`%s`](%s)", variant, docsToRepo(path.Join(ex.Path, variant))))
	}

	return strings.Join(links, ", ")
}

func rewriteLinks(body, examplePath string) string {
	return mdLink.ReplaceAllStringFunc(body, func(match string) string {
		target := strings.TrimSuffix(strings.TrimPrefix(match, "]("), ")")
		if isAbsoluteLink(target) {
			return match
		}

		link, fragment, hasFragment := strings.Cut(target, "#")
		rewritten := docsToRepo(path.Join(examplePath, link))
		if hasFragment {
			rewritten += "#" + fragment
		}

		return "](" + rewritten + ")"
	})
}

func isAbsoluteLink(target string) bool {
	return target == "" ||
		strings.HasPrefix(target, "#") ||
		strings.HasPrefix(target, "/") ||
		strings.Contains(target, "://") ||
		strings.HasPrefix(target, "mailto:")
}

func docsToRepo(repoPath string) string {
	return "../../" + repoPath
}
