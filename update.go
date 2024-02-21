// update.go // This file holds the implementation for the "update" functionality //>
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
func update(programsToUpdate []string) error {
	// If programsToUpdate is nil, list files from InstallDir
	if programsToUpdate == nil {
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

		programsToUpdate = make([]string, 0)
		for _, file := range files {
			if !file.IsDir() && contains(validPrograms, file.Name()) {
				programsToUpdate = append(programsToUpdate, file.Name())
			}
		}
	}

	totalPrograms := len(programsToUpdate)

	// Initialize counters
	var skipped, updated, checked int

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

	for _, program := range programsToUpdate {
		checked++ // Increment the checked counter for every processed program
		leftToGoStr := fmt.Sprintf("(%d/%d)", checked, totalPrograms)
		localFilePath := filepath.Join(installDir, program)
		_, err := os.Stat(localFilePath)
		if os.IsNotExist(err) {
			truncatePrintf("\033[2K\rWarning: Tried to update a non-existent program %s. %s\n", program, leftToGoStr)
			skipped++
			continue
		} else if err != nil {
			truncatePrintf("\033[2K\rWarning: Failed to access program %s. Skipping. %s", program, leftToGoStr)
			skipped++
			continue
		}

		localSHA256, err := getLocalSHA256(localFilePath)
		if err != nil {
			truncatePrintf("\033[2K\rWarning: Failed to get SHA256 for %s. Skipping. %s", program, leftToGoStr)
			skipped++
			continue
		}

		binaryInfo, err := getBinaryInfo(program)
		if err != nil {
			truncatePrintf("\033[2K\rWarning: Failed to get metadata for %s. Skipping. %s", program, leftToGoStr)
			skipped++
			continue
		}

		// Skip if the SHA field is null
		if binaryInfo.SHA256 == "" {
			truncatePrintf("\033[2K\rSkipping %s because the SHA256 field is null. %s", program, leftToGoStr)
			skipped++
			continue
		}

		if checkDifferences(localSHA256, binaryInfo.SHA256) == 1 {
			truncatePrintf("\033[2K\rDetected a difference in %s. Updating... %s", program, leftToGoStr)
			installMessage := truncateSprintf("\033[2K\rUpdating %s to version %s", program, binaryInfo.SHA256)
			err := installCommand(program, []string{installDir}, installMessage)
			if err != nil {
				truncatePrintf("\033[2K\rError: Failed to update %s: %v", program, err)
				continue
			}
			truncatePrintf("\033[2K\rSuccessfully updated %s. %s", program, leftToGoStr)
			updated++
		} else {
			truncatePrintf("\033[2K\rNo updates available for %s. %s", program, leftToGoStr)
		}
	}

	// Print final counts
	truncatePrintf("\033[2K\rSkipped: %d\tUpdated: %d\tChecked: %d\n", skipped, updated, checked)

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
