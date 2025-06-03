package main

import (
	"fmt"
	"os"
)

// handleClean removes the ./.tmp output directory.
func handleClean() {
	// outputDirName is defined in main.go as ".tmp"
	fmt.Printf("Attempting to remove output directory: %s\n", outputDirName)
	err := os.RemoveAll(outputDirName)
	if err != nil {
		fmt.Printf("Error removing directory %s: %v\n", outputDirName, err)
		// Optionally, os.Exit(1) if clean must succeed
		return
	}
	fmt.Printf("Successfully removed directory: %s\n", outputDirName)
}
