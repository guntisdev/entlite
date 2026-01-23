package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func newCommand(entityNames []string) {
	if len(entityNames) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify at least one entity name")
		os.Exit(1)
	}

	entDir := "ent"
	schemaDir := filepath.Join(entDir, "schema")
	contractDir := filepath.Join(entDir, "contract")
	genDir := filepath.Join(entDir, "gen")

	for _, dir := range []string{entDir, schemaDir, contractDir, genDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s directory: %v\n", dir, err)
			os.Exit(1)
		}
	}
}
