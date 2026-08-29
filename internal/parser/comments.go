package parser

import (
	"go/ast"
	"go/token"
	"strings"
)

// commentLookup resolves the code comments written above declarations.
type commentLookup struct {
	fset   *token.FileSet
	groups []*ast.CommentGroup
}

func newCommentLookup(fset *token.FileSet, file *ast.File) commentLookup {
	return commentLookup{fset: fset, groups: file.Comments}
}

// docAbove returns the comment block written directly above pos.
func (c commentLookup) docAbove(pos, prevEnd token.Pos) string {
	if c.fset == nil {
		return ""
	}

	line := c.fset.Position(pos).Line

	for _, group := range c.groups {
		if c.fset.Position(group.End()).Line != line-1 {
			continue
		}
		if prevEnd.IsValid() && c.fset.Position(group.Pos()).Line <= c.fset.Position(prevEnd).Line {
			return ""
		}

		// Text() strips the markers and drops directives like //go:generate
		return strings.TrimSpace(group.Text())
	}

	return ""
}
