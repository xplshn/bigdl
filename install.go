package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// installBinary handles the installation of a single binary
func installBinary(binaryName string, silent bool, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

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
				errCh <- fmt.Errorf("error: Could not copy cached file: %v", err)
				return
			}

			// Set executable bit immediately after copying
			if err := os.Chmod(installPath, 0o755); err != nil {
				errCh <- fmt.Errorf("failed to set executable bit: %v", err)
				return
			}

			return
		}
	}

	// If the cached file does not exist, download the binary
	url, err := findURL(binaryName)
	if err != nil {
		errCh <- fmt.Errorf("%v", err)
		return
	}

	if err := fetchBinaryFromURL(url, installPath); err != nil {
		errCh <- fmt.Errorf("%v", err)
		return
	}

	if TrackFiles {
		if err := addToTrackerFile(binaryName); err != nil {
			errCh <- fmt.Errorf("failed to update tracker file: %v", err)
			return
		}
	}

	if !silent {
		if InstallMessage != "disabled" {
			fmt.Print(InstallMessage)
		} else {
			fmt.Printf("Successfully created %s\n", installPath)
		}
	}
}

func installCommand(silent bool, binaryNames string) error {
	// Disable the progressbar if the installation is to be performed silently
	if silent {
		UseProgressBar = false
	}
	binaries := strings.Fields(binaryNames)

	var wg sync.WaitGroup
	errCh := make(chan error, len(binaries)) // Buffered channel to collect errors

	for _, binaryName := range binaries {
		wg.Add(1)
		go installBinary(binaryName, silent, &wg, errCh)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errCh)

	// Collect and return the first error encountered
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
