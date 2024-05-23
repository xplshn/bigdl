// install.go // This file implements the install functionality //
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func installCommand(silent bool, binaryNames string) error {
	// Disable the progressbar if the installation is to be performed silently
	if silent {
		UseProgressBar = false
	}
	binaries := strings.Fields(binaryNames)
	for _, binaryName := range binaries {
		// Extract the last part of the binaryName to use as the filename
		fileName := filepath.Base(binaryName)

		// Construct the installPath using the extracted filename
		installPath := filepath.Join(InstallDir, fileName)

		// Use ReturnCachedFile to check for a cached file
		if InstallUseCache {
			cachedFile, err := ReturnCachedFile(binaryName)
			if err == 0 {
				// If the cached file exists, use it
				if !silent {
					fmt.Printf("Using cached file: %s\n", cachedFile)
				}
				// Copy the cached file to the install path
				if err := copyFile(cachedFile, installPath); err != nil {
					return fmt.Errorf("error: Could not copy cached file: %v", err)
				}

				// Set executable bit immediately after copying
				if err := os.Chmod(installPath, 0o755); err != nil {
					return fmt.Errorf("failed to set executable bit: %v", err)
				}

				continue
			}
		}

		// If the cached file does not exist, download the binary
		url, err := findURL(binaryName)
		if err != nil {
			errorOut("%v\n", err)
		}

		if err := fetchBinaryFromURL(url, installPath); err != nil {
			return fmt.Errorf("error: Could not install binary: %v", err)
		}

		if !silent {
			fmt.Printf("Installed: %s\n", installPath)
		}
	}
	return nil
}
