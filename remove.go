// remove.go // This file implements the functionality of "remove" //>
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func remove(binariesToRemove []string) {
	for _, binaryName := range binariesToRemove {
		// Use the base name of binaryName for constructing the cachedFile path
		baseName := filepath.Base(binaryName)
		installPath := filepath.Join(InstallDir, baseName)
		err := os.Remove(installPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: '%s' does not exist in %s\n", baseName, InstallDir)
			} else {
				fmt.Fprintf(os.Stderr, "Error: Failed to remove '%s' from %s. %v\n", baseName, InstallDir, err)
			}
			continue
		}
		fmt.Printf("'%s' removed from %s\n", baseName, InstallDir)
	}
}
