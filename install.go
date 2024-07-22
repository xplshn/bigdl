package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	installSuccesses uint32
	errorCount       uint32
	errorMessages    []string   // Slice to collect error messages
	progressMutex    sync.Mutex // Mutex to synchronize progress printing
	nextPrintIndex   int
	interrupted      bool
)

// installBinary handles the installation of a single binary
func installBinary(ctx context.Context, binaryName string, index int, totalCount int, wg *sync.WaitGroup, interruptChecker func() bool) {
	defer wg.Done()

	// Helper function to check context cancellation and log an error if interrupted
	checkContext := func(action string) bool {
		if ctx.Err() != nil || interruptChecker() {
			appendErrorMessage(fmt.Sprintf("%s interrupted for %s", action, binaryName))
			atomic.AddUint32(&errorCount, 1)
			return true
		}
		return false
	}

	// Check initial context
	if checkContext("initial") {
		return
	}

	// Extract the filename from the binaryName
	fileName := filepath.Base(binaryName)

	// Construct the installation path
	installPath := filepath.Join(InstallDir, fileName)

	if InstallUseCache {
		cachedFilePath, err := ReturnCachedFile(binaryName)
		if err == 0 {
			if checkContext("before copying cached file") {
				return
			}
			if err := copyFile(cachedFilePath, installPath); err != nil {
				appendErrorMessage(fmt.Sprintf("error: Could not copy cached file for %s: %v", binaryName, err))
				atomic.AddUint32(&errorCount, 1)
				return
			}
			if err := os.Chmod(installPath, 0o755); err != nil {
				appendErrorMessage(fmt.Sprintf("failed to set executable bit for %s: %v", binaryName, err))
				atomic.AddUint32(&errorCount, 1)
				return
			}
			atomic.AddUint32(&installSuccesses, 1)
			printProgress(index, totalCount, installPath)
			return
		}
	}

	// If the cached file does not exist, download the binary
	downloadURL, err := findURL(binaryName)
	if err != nil {
		appendErrorMessage(fmt.Sprintf("failed to find URL for %s: %v", binaryName, err))
		atomic.AddUint32(&errorCount, 1)
		return
	}

	if checkContext("before fetching binary") {
		return
	}

	if err := fetchBinaryFromURL(downloadURL, installPath); err != nil {
		appendErrorMessage(fmt.Sprintf("failed to fetch binary from URL for %s: %v", binaryName, err))
		atomic.AddUint32(&errorCount, 1)
		return
	}

	if checkContext("after fetching binary") {
		return
	}

	if TrackFiles {
		if err := addToTrackerFile(binaryName); err != nil {
			appendErrorMessage(fmt.Sprintf("failed to update tracker file for %s: %v", binaryName, err))
			atomic.AddUint32(&errorCount, 1)
			return
		}
	}

	atomic.AddUint32(&installSuccesses, 1)
	printProgress(index, totalCount, installPath)
}

// appendErrorMessage appends an error message to the global slice in a thread-safe manner
func appendErrorMessage(message string) {
	progressMutex.Lock()
	defer progressMutex.Unlock()
	errorMessages = append(errorMessages, message)
}

// printProgress ensures progress messages are printed in the correct order
func printProgress(index int, totalCount int, installPath string) {
	progressMutex.Lock()
	defer progressMutex.Unlock()

	for nextPrintIndex != index {
		progressMutex.Unlock()
		progressMutex.Lock()
	}

	truncatePrintf("\033[2K\r<%d/%d> | Successfully created %s", index+1, totalCount, filepath.Base(installPath))
	nextPrintIndex++
}

// installCommand handles the overall installation process
func installCommand(silent bool, binaryNames string) error {
	if silent {
		UseProgressBar = false
	}

	binaryList := strings.Fields(binaryNames)
	totalBinaries := len(binaryList)

	var wg sync.WaitGroup
	nextPrintIndex = 0 // Initialize the next print index

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the cancel function is called when the function returns

	// Set up signal handling
	interruptedFunc, err := signalHandler(ctx, cancel)
	if err != nil {
		return fmt.Errorf("failed to set up signal handler: %v", err)
	}

	for i, binaryName := range binaryList {
		wg.Add(1)
		go installBinary(ctx, binaryName, i, totalBinaries, &wg, interruptedFunc)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	if interrupted {
		fmt.Println("\033[2K\rinstallCommand: Quitting, user bailed out. Cleaning up...")
		return nil
	}

	// Print final counts
	finalCounts := fmt.Sprintf("\033[2K\rInstalled: %d", atomic.LoadUint32(&installSuccesses))
	if atomic.LoadUint32(&errorCount) > 0 {
		finalCounts += fmt.Sprintf("\tErrors: %d", atomic.LoadUint32(&errorCount))
	}
	fmt.Println(finalCounts)

	// Print collected error messages
	for _, errMsg := range errorMessages {
		fmt.Println(errMsg)
	}

	return nil
}
