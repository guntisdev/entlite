package docs

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
)

const treeMarker = "tree"

var entTree = []struct{ Path, Note string }{
	{Path: "schema/", Note: "you write: one file per entity"},
	{Path: "logic/", Note: "you write: Go funcs for DefaultFunc and Validate"},
	{Path: "contract/", Note: "the input for sqlc and buf"},
	{Path: "contract/proto/", Note: "generated: schema.proto, plus your custom .proto"},
	{Path: "contract/sqlc/", Note: "generated: schema.sql, queries.sql, plus your custom .sql"},
	{Path: "gen/", Note: "the output, rewritten on every run"},
	{Path: "gen/db/", Note: "generated: typed wrapper, pointers instead of null types"},
	{Path: "gen/db/internal/", Note: "generated: raw sqlc output, never import it"},
	{Path: "gen/pb/", Note: "generated: proto messages and the connect service"},
	{Path: "gen/ts/", Note: "generated: TypeScript client"},
	{Path: "buf.yaml", Note: ""},
	{Path: "buf.gen.yaml", Note: ""},
	{Path: "buf.lock", Note: ""},
	{Path: "sqlc.yaml", Note: "dialect and sqlc settings"},
	{Path: "generate.go", Note: "the go:generate lines that run the pipeline"},
}

func injectEntTree(p *Pipeline, pages []string) error {
	dir, err := sampleEntDir(p)
	if err != nil {
		return err
	}
	if err := checkEntTree(p, dir); err != nil {
		return err
	}

	block := "```\n" + renderTree() + "```"
	for _, relPath := range pages {
		content, err := p.Read(relPath)
		if err != nil {
			return err
		}

		if !hasRegion(string(content), treeMarker) {
			continue
		}

		updated, err := replaceRegion(string(content), treeMarker, block)
		if err != nil {
			return fmt.Errorf("%s: %w", relPath, err)
		}
		p.Write(relPath, []byte(updated))
	}

	return nil
}

func renderTree() string {
	lines := make([]string, 0, len(entTree))
	notes := make([]string, 0, len(entTree))
	width := 0

	for i, entry := range entTree {
		depth := strings.Count(strings.TrimSuffix(entry.Path, "/"), "/")

		connector := "├── "
		if isLast(i, depth) {
			connector = "└── "
		}

		name := path.Base(entry.Path)
		if strings.HasSuffix(entry.Path, "/") {
			name += "/"
		}

		line := treePrefix(i, depth) + connector + name
		if entry.Note != "" && len([]rune(line)) > width {
			width = len([]rune(line))
		}
		lines = append(lines, line)
		notes = append(notes, entry.Note)
	}

	var out strings.Builder
	out.WriteString("ent/\n")
	for i, line := range lines {
		if notes[i] == "" {
			out.WriteString(line + "\n")
			continue
		}
		padding := strings.Repeat(" ", width-len([]rune(line)))
		out.WriteString(line + padding + "  # " + notes[i] + "\n")
	}

	return out.String()
}

func treePrefix(index, depth int) string {
	var prefix strings.Builder
	for level := 0; level < depth; level++ {
		if isLast(ancestor(index, level), level) {
			prefix.WriteString("    ")
		} else {
			prefix.WriteString("│   ")
		}
	}

	return prefix.String()
}

func ancestor(index, level int) int {
	for i := index; i >= 0; i-- {
		if strings.Count(strings.TrimSuffix(entTree[i].Path, "/"), "/") == level {
			return i
		}
	}

	return index
}

func isLast(index, depth int) bool {
	for i := index + 1; i < len(entTree); i++ {
		next := strings.Count(strings.TrimSuffix(entTree[i].Path, "/"), "/")
		if next < depth {
			return true
		}
		if next == depth {
			return false
		}
	}

	return true
}

func sampleEntDir(p *Pipeline) (string, error) {
	examples, err := scanExamples(p)
	if err != nil {
		return "", err
	}
	if len(examples) == 0 {
		return "", fmt.Errorf("no examples to check the ent tree against")
	}

	variants, err := exampleVariants(p, examples[0])
	if err != nil {
		return "", err
	}
	if len(variants) == 0 {
		return "", fmt.Errorf("%s has no dialect folder", examples[0].Path)
	}

	return path.Join(examples[0].Path, variants[0], "ent"), nil
}

// checkEntTree fails when the tree and a real generated project disagree.
func checkEntTree(p *Pipeline, entDir string) error {
	listed := map[string]bool{}
	for _, entry := range entTree {
		name := strings.TrimSuffix(entry.Path, "/")
		listed[name] = true

		if _, err := os.Stat(p.Abs(path.Join(entDir, name))); err != nil {
			return fmt.Errorf("ent tree lists %s, %s has no such path", entry.Path, entDir)
		}
	}

	missing, err := unlistedPaths(p, entDir, listed)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s has %s, add them to entTree in internal/docs/tree.go",
			entDir, strings.Join(missing, ", "))
	}

	return nil
}

func unlistedPaths(p *Pipeline, entDir string, listed map[string]bool) ([]string, error) {
	var missing []string

	for _, dir := range []string{"", "contract", "gen"} {
		entries, err := os.ReadDir(p.Abs(path.Join(entDir, dir)))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path.Join(entDir, dir), err)
		}

		for _, entry := range entries {
			name := path.Join(dir, entry.Name())
			if dir == "" && !entry.IsDir() && !listed[name] {
				missing = append(missing, name) // a new config file at the root
			}
			if entry.IsDir() && !listed[name] {
				missing = append(missing, name+"/")
			}
		}
	}
	sort.Strings(missing)

	return missing, nil
}
