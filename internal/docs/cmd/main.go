package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/guntisdev/entlite/internal/docs"
)

func main() {
	out := flag.String("out", "docs", "directory for the generated markdown")
	html := flag.String("html", "", "directory for the generated html, empty skips the site step")
	check := flag.Bool("check", false, "report stale docs instead of writing them")
	flag.Parse()

	result, err := docs.Run(docs.Config{Root: "", OutDir: *out, HTMLDir: *html, Check: *check})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building docs: %v\n", err)
		os.Exit(1)
	}

	if *check {
		if len(result.Stale) > 0 {
			fmt.Fprintf(os.Stderr, "%s\n%d file(s) out of date, run: make docs\n", result.Diff, len(result.Stale))
			os.Exit(1)
		}
		fmt.Println("docs are up to date")

		return
	}

	fmt.Printf("wrote %d file(s)\n", len(result.Written))
}
