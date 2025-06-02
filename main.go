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

	// Check for legacy 'markdown' directory and issue warning
	// This check is primarily for local execution.
	// In a GitHub Action, this directory structure might not be relevant in the same way.
	if _, err := os.Stat("markdown"); !os.IsNotExist(err) {
		// The 'markdown' directory exists
		fmt.Println("--------------------------------------------------------------------")
		fmt.Println("[KREMS WARNING] Breaking Change from v0.2.0 (or later):")
		fmt.Println("The 'markdown/' directory is no longer the primary source for markdown files or assets.")
		fmt.Println("Krems now processes markdown files and assets (css, js, images) from the project's root directory.")
		fmt.Println("\nPlease move all your content (markdown files, subdirectories, css, js, images)")
		fmt.Println("from the 'markdown/' directory to the root of your project.")
		fmt.Println("The 'markdown/' directory, if present, will be ignored for content processing.")
		fmt.Println("For more information, please see the README.md or project documentation.")
		fmt.Println("--------------------------------------------------------------------")
		// Depending on desired strictness, you could os.Exit(1) here,
		// but for now, let's allow it to continue, as the new logic will ignore 'markdown/' anyway.
	}

	switch os.Args[1] {
	case "--init":
		handleInit()
	case "--build":
		handleBuild()
	case "--run":
		handleRun()
	case "--version":
		// Ideally, this version string is injected at build time.
		// For now, we'll use a placeholder.
		// You can replace "0.2.0-dev" with a more dynamic version later.
		fmt.Println("Krems version 0.2.0-dev (simplified-workflow-changes)")
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Available commands: --init, --build, --run, --version")
		os.Exit(1)
	}
}
