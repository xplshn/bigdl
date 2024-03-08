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

// TODO: padding should be used to align the messages
// TODO: Validation of programs could be done without a list (current bottleneck)

// update checks for updates to the valid programs and installs any that have changed.
func update(programsToUpdate []string) error {
	// 'Configure' external functions
	useProgressBar = false
	installUseCache = false

	// Initialize counters
	var (
		skipped, updated, errors, toBeChecked uint32
		checked                               uint32
		errorMessages                         string
		padding                               = " "
	)

	// Fetch the list of binaries from the remote source once
	remotePrograms, err := listBinaries()
	if err != nil {
		return fmt.Errorf("failed to list remote binaries: %w", err)
	}

	// If programsToUpdate is nil, list files from InstallDir and validate against remote
	if programsToUpdate == nil {
		files, err := listFilesInDir(InstallDir)
		if err != nil {
			return fmt.Errorf("failed to list files in InstallDir: %w", err)
		}

		programsToUpdate = make([]string, 0)
		for _, file := range files {
			// Extract the file name from the full path
			fileName := filepath.Base(file)
			if contains(remotePrograms, fileName) {
				programsToUpdate = append(programsToUpdate, fileName)
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

			installPath := filepath.Join(InstallDir, program)
			if !fileExists(installPath) {
				atomic.AddUint32(&skipped, 1)
				progressMutex.Lock()
				atomic.AddUint32(&checked, 1) // Increment the checked counter
				truncatePrintf("\033[2K\r<%d/%d> %s | Warning: Tried to update a non-existent program %s. Skipping.", atomic.LoadUint32(&checked), toBeChecked, padding, program)
				progressMutex.Unlock()
				return
			}
			localSHA256, err := getLocalSHA256(installPath)
			if err != nil {
				atomic.AddUint32(&skipped, 1)
				progressMutex.Lock()
				atomic.AddUint32(&checked, 1) // Increment the checked counter
				truncatePrintf("\033[2K\r<%d/%d> %s | Warning: Failed to get SHA256 for %s. Skipping.", atomic.LoadUint32(&checked), toBeChecked, padding, program)
				progressMutex.Unlock()
				return
			}

			binaryInfo, err := getBinaryInfo(program)
			if err != nil {
				atomic.AddUint32(&skipped, 1)
				progressMutex.Lock()
				atomic.AddUint32(&checked, 1) // Increment the checked counter
				truncatePrintf("\033[2K\r<%d/%d> %s | Warning: Failed to get metadata for %s. Skipping.", atomic.LoadUint32(&checked), toBeChecked, padding, program)
				progressMutex.Unlock()
				return
			}

			// Skip if the SHA field is null
			if binaryInfo.SHA256 == "" {
				atomic.AddUint32(&skipped, 1)
				progressMutex.Lock()
				atomic.AddUint32(&checked, 1) // Increment the checked counter
				truncatePrintf("\033[2K\r<%d/%d> %s | Skipping %s because the SHA256 field is null.", atomic.LoadUint32(&checked), toBeChecked, padding, program)
				progressMutex.Unlock()
				return
			}

			if checkDifferences(localSHA256, binaryInfo.SHA256) == 1 {
				truncatePrintf("\033[2K\r<%d/%d> %s | Detected a difference in %s. Updating...", atomic.LoadUint32(&checked), toBeChecked, padding, program)
				installMessage := truncateSprintf("\x1b[A\033[KUpdating %s", program)
				err := installCommand(program, installMessage)
				if err != nil {
					atomic.AddUint32(&errors, 1)
					progressMutex.Lock()
					errorMessages += sanitizeString(fmt.Sprintf("Failed to update '%s', please check this file's properties, etc\n", program))
					progressMutex.Unlock()
					return
				}
				progressMutex.Lock()
				atomic.AddUint32(&checked, 1) // Increment the checked counter
				truncatePrintf("\033[2K\r<%d/%d> %s | Successfully updated %s.", atomic.LoadUint32(&checked), toBeChecked, padding, program)
				progressMutex.Unlock()
				atomic.AddUint32(&updated, 1)
			} else {
				progressMutex.Lock()
				atomic.AddUint32(&checked, 1) // Increment the checked counter
				truncatePrintf("\033[2K\r<%d/%d> %s | No updates available for %s.", atomic.LoadUint32(&checked), toBeChecked, padding, program)
				progressMutex.Unlock()
			}
		}(program)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Prepare final counts
	finalCounts := fmt.Sprintf("\033[2K\rSkipped: %d\tUpdated: %d\tChecked: %d", atomic.LoadUint32(&skipped), atomic.LoadUint32(&updated), uint32(int(atomic.LoadUint32(&checked))-1))
	if errors > 0 {
		finalCounts += fmt.Sprintf("\tErrors: %d", atomic.LoadUint32(&errors))
	}
	// Print final counts
	fmt.Println(finalCounts)
	fmt.Printf(errorMessages)

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
