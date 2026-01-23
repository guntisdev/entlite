package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("hello world")

	if len(os.Args) < 2 {
		// TODO print usage with new and gen commands
		fmt.Println("Missing arguments")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "new":
		newCommand(os.Args[2:])
	case "gen":
		genCommand(os.Args[2:])
	default:
		// TODO print usage with new and gen commands
		fmt.Println("Unknow argument")
		os.Exit(1)
	}
}
