// install.go // This file implements the install functionality //
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func installCommand(binaryName string, args []string, messages ...string) error {
	installDir := os.Getenv("INSTALL_DIR")
	if len(args) > 0 && args[0] != "" {
		installDir = args[0]
	}

	if installDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("Error: Failed to get user's Home directory: %w", err)
		}
		installDir = filepath.Join(homeDir, ".local", "bin")
	}

	if err := os.MkdirAll(installDir, os.ModePerm); err != nil {
		return fmt.Errorf("Error: Could not create installation directory: %v", err)
	}

	installPath := filepath.Join(installDir, binaryName)

	// Use ReturnCachedFile to check for a cached file
	cachedFile, errCode := ReturnCachedFile(binaryName)
	if errCode == 0 {
		// If the cached file exists, use it
		fmt.Printf("Using cached file: %s\n", cachedFile)
		// Copy the cached file to the install path
		if err := copyFile(cachedFile, installPath); err != nil {
			return fmt.Errorf("Error: Could not copy cached file: %v", err)
		}
	} else {
		// If the cached file does not exist, download the binary
		url, err := findURL(binaryName)
		if err != nil {
			return fmt.Errorf("Error: %v", err)
		}
		if err := fetchBinaryFromURL(url, installPath); err != nil {
			return fmt.Errorf("Error: Could not install binary: %v", err)
		}
	}

	// Check if any messages are provided and print them
	if len(messages) > 0 && messages[0] != "" {
		for _, message := range messages {
			fmt.Printf(message)
		}
	} else {
		// If no message provided, print default installation complete message
		fmt.Printf("Installation complete: %s \n", installPath)
	}
	return nil
}
