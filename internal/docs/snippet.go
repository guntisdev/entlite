package docs

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
)

var snippetMarker = regexp.MustCompile(`<!-- snippet:([^\s:]+):([A-Za-z0-9_]+):start -->`)

func injectSnippets(p *Pipeline, relPath string) error {
	content, err := p.Read(relPath)
	if err != nil {
		return err
	}

	page := string(content)
	for _, match := range snippetMarker.FindAllStringSubmatch(page, -1) {
		marker, source, funcName := match[1]+":"+match[2], match[1], match[2]

		code, err := funcSource(p, source, funcName)
		if err != nil {
			return fmt.Errorf("%s: %w", relPath, err)
		}

		page, err = replaceRegion(page, "snippet:"+marker, "```go\n"+code+"\n```")
		if err != nil {
			return fmt.Errorf("%s: %w", relPath, err)
		}
	}
	p.Write(relPath, []byte(page))

	return nil
}

func funcSource(p *Pipeline, relPath, name string) (string, error) {
	content, err := p.Read(relPath)
	if err != nil {
		return "", err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, relPath, content, parser.SkipObjectResolution)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s: %w", relPath, err)
	}

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != name {
			continue
		}

		start := fset.Position(funcDecl.Pos()).Offset
		end := fset.Position(funcDecl.End()).Offset

		return string(content[start:end]), nil
	}

	return "", fmt.Errorf("%s has no func %s", relPath, name)
}
