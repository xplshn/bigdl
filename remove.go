// remove.go // This file implements the functionality of "remove" //>
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

	removeFromDir(InstallDir)
}
