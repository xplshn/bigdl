// remove.go // This file implements the functionality of "remove" //>
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func remove(binariesToRemove []string) {
	for _, binaryName := range binariesToRemove {
		installPath := filepath.Join(InstallDir, binaryName)
		err := os.Remove(installPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: Binary %s does not exist in %s\n", binaryName, InstallDir)
			} else {
				fmt.Fprintf(os.Stderr, "Error: Failed to remove binary %s from %s. %v\n", binaryName, InstallDir, err)
			}
			continue
		}
		fmt.Printf("Binary %s removed from %s\n", binaryName, InstallDir)
	}
}
