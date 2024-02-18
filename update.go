// update.go // This file holds the implementation for the "update" functionality //>
// W.I.P // TODO: Handle signals like CTRL+C gracefully. // ENSURE THINGS ARE DONE RIGHT.

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// update checks for updates to the valid programs and installs any that have changed.
func update() error {
	validPrograms, err := listBinaries()
	if err != nil {
		return fmt.Errorf("failed to list binaries: %w", err)
	}

	installDir := os.Getenv("INSTALL_DIR")
	if installDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		installDir = filepath.Join(homeDir, ".local", "bin")
	}

	info, err := os.Stat(installDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("installation directory %s is not a directory", installDir)
	}

	files, err := os.ReadDir(installDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", installDir, err)
	}

	// Initialize counters
	var skipped, updated, checked int

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		binaryName := file.Name()
		if contains(validPrograms, binaryName) {
			checked++ // Increment the checked counter for every processed binary

			localSHA256, err := getLocalSHA256(filepath.Join(installDir, binaryName))
			if err != nil {
				fmt.Printf("\033[2K\rWarning: Failed to get SHA256 for %s. Skipping.   ", binaryName)
				skipped++
				continue
			}

			binaryInfo, err := getBinaryInfo(binaryName)
			if err != nil {
				fmt.Printf("\033[2K\rWarning: Failed to get metadata for %s. Skipping.   ", binaryName)
				skipped++
				continue
			}

			// Skip if the SHA field is null
			if binaryInfo.SHA256 == "" {
				fmt.Printf("\033[2K\rSkipping %s because the SHA256 field is null.   ", binaryName)
				skipped++
				continue
			}

			if checkDifferences(localSHA256, binaryInfo.SHA256) == 1 {
				fmt.Printf("\033[2K\rDetected a difference in %s. Updating...   ", binaryName)
				installMessage := fmt.Sprintf("Updating %s to version %s", binaryName, binaryInfo.SHA256)
				err := installCommand(binaryName, []string{installDir, installMessage})
				if err != nil {
					fmt.Printf("\033[2K\rError: Failed to update %s: %v   ", binaryName, err)
					continue
				}
				fmt.Printf("\033[2K\rSuccessfully updated %s.   ", binaryName)
				updated++
			} else {
				fmt.Printf("\033[2K\rNo updates available for %s.   ", binaryName)
			}
		}
	}

	// Print final counts
	fmt.Printf("\033[2K\rSkipped: %d\tUpdated: %d\tChecked: %d\n", skipped, updated, checked)

	return nil
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// getLocalSHA256 calculates the SHA256 checksum of the local file.
func getLocalSHA256(filePath string) (string, error) {
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Calculate SHA256 checksum
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate SHA256: %v", err)
	}
	sha256Checksum := hex.EncodeToString(hasher.Sum(nil))

	return sha256Checksum, nil
}

func checkDifferences(localSHA256, remoteSHA256 string) int {
	if localSHA256 != remoteSHA256 {
		return 1
	}
	return 0
}
