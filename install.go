// install.go // This file implements the install functionality //
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func installCommand(binaryName string, installMessage ...string) error {
	installPath := filepath.Join(InstallDir, binaryName)

	// Use ReturnCachedFile to check for a cached file
	if installUseCache {
		cachedFile, err := ReturnCachedFile(binaryName)
		if err == 0 {
			// If the cached file exists, use it
			fmt.Printf("\r\033[KUsing cached file: %s\n", cachedFile)
			// Copy the cached file to the install path
			if err := copyFile(cachedFile, installPath); err != nil {
				return fmt.Errorf("Error: Could not copy cached file: %v", err)
			}

			// Set executable bit immediately after copying
			if err := os.Chmod(installPath, 0755); err != nil {
				return fmt.Errorf("failed to set executable bit: %v", err)
			}

			return nil
		}
	}

	// If the cached file does not exist, download the binary
	url, err := findURL(binaryName)
	if err != nil {
		errorOut("%v\n", err)
	}
	if err := fetchBinaryFromURL(url, installPath); err != nil {
		return fmt.Errorf("Error: Could not install binary: %v", err)
	}

	// Check if the user provided a custom installMessage and If so, print it as per his requirements.
	if len(installMessage) != 0 {
		if installMessage[0] == "--fancy" {
			if installMessage[1] == "--truncate" {
				truncatePrintf(installMessage[2], binaryName)
			} else {
				fmt.Printf(installMessage[1], binaryName)
			}
			if len(installMessage) > 2 && installMessage[2] == "--newline" || len(installMessage) > 3 && installMessage[3] == "--newline" {
				fmt.Println()
			}
		} else {
			if installMessage[0] == "--truncate" {
				fmt.Println(truncateSprintf("%s", installMessage[1]))
			} else {
				fmt.Println(installMessage[0])
			}
		}
	} else {
		fmt.Printf("Installation complete: %s\n", installPath)
	}
	return nil
}
