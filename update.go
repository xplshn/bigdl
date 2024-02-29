// update.go // This file holds the implementation for the "update" functionality - (parallel) //>
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// update checks for updates to the valid programs and installs any that have changed.
func update(programsToUpdate []string) error {
	// Initialize counters
	var skipped, updated, toBeChecked uint32
	var checked uint32 = 1

	// Define installDir at the beginning of the function
	installDir := os.Getenv("INSTALL_DIR")
	if installDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		installDir = filepath.Join(homeDir, ".local", "bin")
	}

	// Fetch the list of binaries from the remote source once
	remotePrograms, err := listBinaries()
	if err != nil {
		return fmt.Errorf("failed to list remote binaries: %w", err)
	}

	// If programsToUpdate is nil, list files from InstallDir and validate against remote
	if programsToUpdate == nil {
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
			if !file.IsDir() && contains(remotePrograms, file.Name()) {
				programsToUpdate = append(programsToUpdate, file.Name())
			}
		}
	}

	// Calculate toBeChecked
	toBeChecked = uint32(len(programsToUpdate))

	// Use a mutex for thread-safe updates to the progress
	var progressMutex sync.Mutex

	// Use a wait group to wait for all programs to finish updating
	var wg sync.WaitGroup

	// Iterate over programsToUpdate and download/update each one concurrently
	for _, program := range programsToUpdate {
		// Increment the WaitGroup counter
		wg.Add(1)

		// Launch a goroutine to update the program
		go func(program string) {
			defer wg.Done()
			localFilePath := filepath.Join(installDir, program)
			_, err := os.Stat(localFilePath)
			if os.IsNotExist(err) {
				progressMutex.Lock()
				truncatePrintf("\033[2K\rWarning: Tried to update a non-existent program %s. <%d/%d>", program, atomic.LoadUint32(&checked), toBeChecked)
				fmt.Printf("\n")
				progressMutex.Unlock()
				return
			}
			localSHA256, err := getLocalSHA256(localFilePath)
			if err != nil {
				atomic.AddUint32(&skipped, 1)
				progressMutex.Lock()
				truncatePrintf("\033[2K\rWarning: Failed to get SHA256 for %s. Skipping. <%d/%d>", program, atomic.LoadUint32(&checked), toBeChecked)
				progressMutex.Unlock()
				return
			}

			binaryInfo, err := getBinaryInfo(program)
			if err != nil {
				atomic.AddUint32(&skipped, 1)
				progressMutex.Lock()
				truncatePrintf("\033[2K\rWarning: Failed to get metadata for %s. Skipping. <%d/%d>", program, atomic.LoadUint32(&checked), toBeChecked)
				progressMutex.Unlock()
				return
			}

			// Skip if the SHA field is null
			if binaryInfo.SHA256 == "" {
				atomic.AddUint32(&skipped, 1)
				progressMutex.Lock()
				truncatePrintf("\033[2K\rSkipping %s because the SHA256 field is null. <%d/%d>", program, atomic.LoadUint32(&checked), toBeChecked)
				progressMutex.Unlock()
				return
			}

			if checkDifferences(localSHA256, binaryInfo.SHA256) == 1 {
				truncatePrintf("\033[2K\rDetected a difference in %s. Updating...", program)
				installMessage := truncateSprintf("\x1b[A\033[KUpdating %s to version %s", program, binaryInfo.SHA256)
				installUseCache = false //I hate myself, this is DISGUSTING.
				useProgressBar = false  // I hate myself, this is AWFUL.
				err := installCommand(program, []string{installDir}, installMessage)
				if err != nil {
					progressMutex.Lock()
					truncatePrintf("\033[2K\rFailed to update %s: %s <%d/%d>", program, err.Error(), atomic.LoadUint32(&checked), toBeChecked)
					progressMutex.Unlock()
					return
				}
				installUseCache = true //I hate myself, this is DISGUSTING.
				useProgressBar = true  // I hate myself, this is AWFUL.
				progressMutex.Lock()
				truncatePrintf("\033[2K\rSuccessfully updated %s. <%d/%d>", program, atomic.LoadUint32(&checked), toBeChecked)
				progressMutex.Unlock()
				atomic.AddUint32(&updated, 1)
			} else {
				progressMutex.Lock()
				truncatePrintf("\033[2K\rNo updates available for %s. <%d/%d>", program, atomic.LoadUint32(&checked), toBeChecked)
				progressMutex.Unlock()
			}
			atomic.AddUint32(&checked, 1)
		}(program)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Print final counts
	fmt.Printf("\033[2K\rSkipped: %d\tUpdated: %d\tChecked: %d\n", atomic.LoadUint32(&skipped), atomic.LoadUint32(&updated), uint32(int(atomic.LoadUint32(&checked))-1))

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
