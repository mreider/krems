package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

const defaultPort = "8080"
const outputDirName = ".tmp" // Changed from "tmp"

func printUsage() {
	fmt.Println("Usage: krems <command> [options]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  --build          Builds the static site into the ./.tmp directory.")
	fmt.Println("  --run [options]  Builds and serves the site locally from ./.tmp.")
	fmt.Println("    --port <number>  Port to run the local server on (default: 8080).")
	fmt.Println("  --clean          Removes the ./.tmp build directory.")
	fmt.Println("  --version        Displays the Krems version.")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Check for legacy 'markdown' directory and issue warning
	if _, err := os.Stat("markdown"); !os.IsNotExist(err) {
		fmt.Println("--------------------------------------------------------------------")
		fmt.Println("[KREMS WARNING] Breaking Change from v0.2.0 (or later):")
		fmt.Println("The 'markdown/' directory is no longer the primary source for markdown files or assets.")
		fmt.Println("Krems now processes markdown files and assets (css, js, images) from the project's root directory.")
		fmt.Println("\nPlease move all your content (markdown files, subdirectories, css, js, images)")
		fmt.Println("from the 'markdown/' directory to the root of your project.")
		fmt.Println("The 'markdown/' directory, if present, will be ignored for content processing.")
		fmt.Println("For more information, please see the README.md or project documentation.")
		fmt.Println("--------------------------------------------------------------------")
	}

	switch os.Args[1] {
	case "--build":
		handleBuild(false, outputDirName) // Use constant for output directory
	case "--run":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		portFlag := runCmd.String("port", defaultPort, "Port to run the local server on")
		
		// Parse flags specifically for the "run" command
		// os.Args[0] is program name, os.Args[1] is "--run"
		// so flags for "run" start from os.Args[2]
		err := runCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Printf("Error parsing --run flags: %v\n", err)
			runCmd.Usage() // Print usage for the run command
			os.Exit(1)
		}

		// Validate port
		if _, errConv := strconv.Atoi(*portFlag); errConv != nil {
			fmt.Printf("Error: Invalid port number '%s'. Port must be a number.\n", *portFlag)
			runCmd.Usage()
			os.Exit(1)
		}
		// handleRun signature will need to be updated to accept port string
		handleRun(*portFlag) 
	case "--clean":
		handleClean() // To be implemented
	case "--version":
		fmt.Println("Krems version 0.2.0-dev (simplified-workflow-changes)") // Placeholder
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}
