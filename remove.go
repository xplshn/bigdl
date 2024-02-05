package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func remove(binaryToRemove string) {
	removeFromDir := func(dirPath string) {
		binaryPath := filepath.Join(dirPath, binaryToRemove)
		err := os.Remove(binaryPath)
		if err != nil {
			fmt.Printf("Binary %s not found in %s\n", binaryToRemove, dirPath)
			return
		}
		fmt.Printf("Binary %s removed from %s\n", binaryToRemove, dirPath)
	}

	// Check if INSTALL_DIR environment variable is set
	if installDir := os.Getenv("INSTALL_DIR"); installDir != "" {
		removeFromDir(installDir)
		return
	}

	// If INSTALL_DIR is not set, check %HOME/.local/bin
	if homeDir, err := os.UserHomeDir(); err == nil {
		localBinDir := filepath.Join(homeDir, ".local", "bin")
		removeFromDir(localBinDir)
	} else {
		fmt.Println("Error getting user home directory:", err)
		os.Exit(1)
	}
}
