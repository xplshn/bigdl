// install.go // This file implements the install functionality //

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func installCommand(binaryName string, args []string) {
	installDir := os.Getenv("INSTALL_DIR")
	if len(args) > 0 && args[0] != "" {
		installDir = args[0]
	}

	if installDir == "" {
		installDir = filepath.Join(os.Getenv("HOME"), ".local", "bin")
	}

	if err := os.MkdirAll(installDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating installation directory: %v\n", err)
		os.Exit(1)
	}

	installPath := filepath.Join(installDir, binaryName)

	// Use ReturnCachedFile to check for a cached file
	cachedFile, errCode := ReturnCachedFile(binaryName)
	if errCode == 0 {
		// If the cached file exists, use it
		fmt.Printf("Using cached file: %s\n", cachedFile)
		// Copy the cached file to the install path
		if err := copyFile(cachedFile, installPath); err != nil {
			fmt.Printf("Error copying cached file: %v\n", err)
			os.Exit(1)
		}
	} else {
		// If the cached file does not exist, download the binary
		url, err := findURL(binaryName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if err := fetchBinaryFromURL(url, installPath); err != nil {
			fmt.Printf("Error installing binary: %v\n", err)
			os.Exit(1)
		}
	}

	installMessage := "Installation complete: %s at %s\n"
	if len(args) > 1 && args[1] != "" {
		installMessage = args[1]
	}

	// Use the escape sequence for newline directly
	fmt.Printf(installMessage, binaryName, installPath)
	fmt.Println() // Adding a newline for proper PS1 behavior
}
