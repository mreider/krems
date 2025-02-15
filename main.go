package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("krems requires a command: --init, --build, or --run")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--init":
		handleInit()
	case "--build":
		handleBuild()
	case "--run":
		handleRun()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
