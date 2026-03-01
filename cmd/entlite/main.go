package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Missing arguments")
		// TODO print usage with new and gen commands
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "new":
		newCommand(os.Args[2:])
	case "gen":
		genCommand(os.Args[2:])
	case "sqlc-wrap":
		sqlcWrapCommand(os.Args[2:])
	case "proto-validate":
		protoValidate(os.Args[2:])
	case "convert":
		convertCommand(os.Args[2:])
	default:
		// TODO print usage with new and gen commands
		fmt.Println("Unknow argument")
		os.Exit(1)
	}
}
