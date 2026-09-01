package docs

import (
	"fmt"
	"strings"

	"github.com/guntisdev/entlite/internal/util"
)

type page struct {
	buf strings.Builder
}

func newPage(title, intro string) *page {
	p := &page{buf: strings.Builder{}}
	p.buf.WriteString(util.GeneratedMD)
	p.buf.WriteString("# " + title + "\n")
	if intro != "" {
		p.buf.WriteString("\n" + intro + "\n")
	}

	return p
}

func (p *page) Heading(level int, text string) {
	p.buf.WriteString("\n" + strings.Repeat("#", level) + " " + text + "\n")
}

func (p *page) Text(text string) {
	if text == "" {
		return
	}
	p.buf.WriteString("\n" + text + "\n")
}

func (p *page) Code(lang, content string) {
	p.buf.WriteString(fmt.Sprintf("\n```%s\n%s\n```\n", lang, strings.TrimRight(content, "\n")))
}

func (p *page) Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	p.buf.WriteString("\n| " + strings.Join(headers, " | ") + " |\n")
	p.buf.WriteString("|" + strings.Repeat(" --- |", len(headers)) + "\n")
	for _, row := range rows {
		p.buf.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}
}

func (p *page) MethodTable(methods []apiMethod) {
	rows := make([][]string, 0, len(methods))
	for _, m := range methods {
		rows = append(rows, []string{code(m.Signature), m.Doc})
	}
	p.Table([]string{"Method", "Description"}, rows)
}

func (p *page) FuncTable(header string, pkgName string, funcs []apiFunc) {
	rows := make([][]string, 0, len(funcs))
	for _, f := range funcs {
		rows = append(rows, []string{code(pkgName + "." + f.Signature), f.Doc})
	}
	p.Table([]string{header, "Description"}, rows)
}

func (p *page) ConstTable(pkgName string, consts []apiConst) {
	rows := make([][]string, 0, len(consts))
	for _, c := range consts {
		rows = append(rows, []string{code(pkgName + "." + c.Name), code(c.Value), c.Doc})
	}
	p.Table([]string{"Constant", "Value", "Description"}, rows)
}

func (p *page) Bytes() []byte {
	return []byte(p.buf.String())
}

func code(text string) string {
	if text == "" {
		return ""
	}

	return "`" + text + "`"
}
