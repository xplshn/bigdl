// install.go // This file implements the install functionality //
package main

import (
	"fmt"
	"path/filepath"
)

func installCommand(binaryName string, args []string, messages ...string) error {
	if len(args) > 0 && args[0] != "" {
		InstallDir = args[0]
	}

	installPath := filepath.Join(InstallDir, binaryName)

	// Use ReturnCachedFile to check for a cached file
	if installUseCache {
		cachedFile, errCode := ReturnCachedFile(binaryName)
		if errCode == 0 {
			// If the cached file exists, use it
			fmt.Printf("\r\033[KUsing cached file: %s\n", cachedFile)
			// Copy the cached file to the install path
			if err := copyFile(cachedFile, installPath); err != nil {
				return fmt.Errorf("Error: Could not copy cached file: %v", err)
			}
			return nil
		}
	}

	// If the cached file does not exist, download the binary
	url, err := findURL(binaryName)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	if err := fetchBinaryFromURL(url, installPath); err != nil {
		return fmt.Errorf("Error: Could not install binary: %v", err)
	}

	// Check if any messages are provided and print them
	if len(messages) > 0 && messages[0] != "" {
		for _, message := range messages {
			fmt.Printf(message)
		}
	} else {
		// If no message provided, print default installation complete message
		fmt.Printf("\x1b[A\033[KInstallation complete: %s \n", installPath)
	}
	return nil
}
