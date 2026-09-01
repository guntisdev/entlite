package docs

import (
	"go/ast"
	"go/token"
	"strings"
)

const cliDir = "cmd/entlite"

type cliCommand struct {
	Name  string // as typed, e.g. sqlc-wrap
	Usage string
	Doc   string
	Flags []cliFlag
}

type cliFlag struct {
	Name    string
	Default string
	Usage   string
}

var flagKinds = map[string]bool{
	"String": true, "Bool": true, "Int": true, "Int64": true,
	"Float64": true, "Uint": true, "Duration": true,
}

func cliPage(p *Pipeline) ([]byte, error) {
	commands, err := parseCLI(p.Abs(cliDir))
	if err != nil {
		return nil, err
	}

	page := newPage("CLI", "The `entlite` binary. In a generated project every command is wired into a "+
		"`go:generate` line, so `go generate ./...` inside `ent/` runs the whole pipeline.")

	rows := make([][]string, 0, len(commands))
	for _, cmd := range commands {
		rows = append(rows, []string{code(cmd.Name), cmd.Doc})
	}
	page.Table([]string{"Command", "Description"}, rows)

	for _, cmd := range commands {
		page.Heading(2, cmd.Name)
		page.Text(cmd.Doc)
		if cmd.Usage != "" {
			page.Code("bash", cmd.Usage)
		}

		flagRows := make([][]string, 0, len(cmd.Flags))
		for _, flag := range cmd.Flags {
			flagRows = append(flagRows, []string{code("-" + flag.Name), code(flag.Default), flag.Usage})
		}
		page.Table([]string{"Flag", "Default", "Description"}, flagRows)
	}

	return page.Bytes(), nil
}

func parseCLI(dir string) ([]cliCommand, error) {
	files, fset, err := parseGoFiles(dir)
	if err != nil {
		return nil, err
	}

	funcs := map[string]*ast.FuncDecl{}
	for _, file := range files {
		for _, decl := range file.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Recv == nil {
				funcs[funcDecl.Name.Name] = funcDecl
			}
		}
	}

	main, ok := funcs["main"]
	if !ok {
		return nil, nil
	}

	var commands []cliCommand
	for _, cmd := range commandSwitch(fset, main) {
		funcDecl, ok := funcs[cmd.handler]
		if !ok {
			continue
		}

		doc, usage := splitUsage(funcDecl.Name.Name, funcDecl.Doc.Text())
		commands = append(commands, cliCommand{
			Name:  cmd.name,
			Usage: usage,
			Doc:   doc,
			Flags: commandFlags(fset, funcDecl),
		})
	}

	return commands, nil
}

type switchCase struct {
	name    string
	handler string
}

func commandSwitch(fset *token.FileSet, main *ast.FuncDecl) []switchCase {
	var commands []switchCase

	ast.Inspect(main, func(node ast.Node) bool {
		clause, ok := node.(*ast.CaseClause)
		if !ok || len(clause.List) == 0 {
			return true
		}

		name := strings.Trim(exprString(fset, clause.List[0]), `"`)
		for _, stmt := range clause.Body {
			exprStmt, ok := stmt.(*ast.ExprStmt)
			if !ok {
				continue
			}
			call, ok := exprStmt.X.(*ast.CallExpr)
			if !ok {
				continue
			}
			if ident, ok := call.Fun.(*ast.Ident); ok {
				commands = append(commands, switchCase{name: name, handler: ident.Name})
				break
			}
		}

		return true
	})

	return commands
}

func commandFlags(fset *token.FileSet, funcDecl *ast.FuncDecl) []cliFlag {
	var flags []cliFlag

	ast.Inspect(funcDecl, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) != 3 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !flagKinds[selector.Sel.Name] {
			return true
		}

		name, nameOK := stringLit(fset, call.Args[0])
		usage, usageOK := stringLit(fset, call.Args[2])
		if !nameOK || !usageOK {
			return true
		}

		flags = append(flags, cliFlag{
			Name:    name,
			Default: strings.Trim(exprString(fset, call.Args[1]), `"`),
			Usage:   usage,
		})

		return true
	})

	return flags
}

func stringLit(fset *token.FileSet, expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok {
		return "", false
	}

	return strings.Trim(exprString(fset, lit), `"`), true
}

func splitUsage(name, doc string) (string, string) {
	var description, usage []string

	for _, line := range strings.Split(strings.TrimSpace(doc), "\n") {
		line = strings.TrimSpace(line)
		if after, found := strings.CutPrefix(line, "usage:"); found {
			usage = append(usage, strings.TrimSpace(after))
			continue
		}
		description = append(description, line)
	}

	return docLine(name, strings.Join(description, " ")), strings.Join(usage, "\n")
}
