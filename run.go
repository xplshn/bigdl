// run.go // This file implements functions related to the Run options //

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"syscall"
	"time"
)

var verboseMode bool

// ReturnCachedFile retrieves the cached file location.
// Returns an empty string and error code  1 if not found.
func ReturnCachedFile(packageName string) (string, int) {
	// Construct the expected cached file pattern
	expectedCachedFile := filepath.Join(TEMP_DIR, fmt.Sprintf("%s.bin", packageName))

	// Check if the file exists using the fileExists function
	if fileExists(expectedCachedFile) {
		return expectedCachedFile, 0
	}

	// If the file does not exist, return an error
	return "", 1
}

// RunFromCache runs the binary from cache or fetches it if not found.
func RunFromCache(packageName string, args []string) {
	// Check for verbose mode flag
	if packageName == "--verbose" {
		// In this case, we should set packageName to the next argument if available
		if len(args) > 0 {
			packageName = args[0]
			args = args[1:] // Remove the flag from the arguments
			verboseMode = true
		} else {
			fmt.Println("Error: Package name not provided after --verbose flag.")
			os.Exit(1)
		}
	}

	cachedFile := filepath.Join(TEMP_DIR, packageName+".bin")
	if fileExists(cachedFile) && isExecutable(cachedFile) {
		fmt.Printf("Running '%s' from cache...\n", packageName)
		runBinary(cachedFile, args, verboseMode)
		cleanCache()
	} else {
		fmt.Printf("Error: cached binary for '%s' not found. Fetching a new one...\n", packageName)
		err := fetchBinary(packageName)
		if err != nil {
			fmt.Printf("Error fetching binary for '%s': %v\n", packageName, err)
			os.Exit(1)
		}
		cleanCache()
		runBinary(cachedFile, args, verboseMode)
	}
}

// runBinary executes the binary with the given arguments.
func runBinary(binaryPath string, args []string, verboseMode bool) {
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		// Check if the error is an exit error and if the exit status is non-zero
		if exitError, ok := err.(*exec.ExitError); ok {
			// The program has exited with a non-zero exit status
			// This is an error from the binary itself
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				// In verbose mode, print the error message with the program name and exit code
				if verboseMode {
					fmt.Printf("The program (%s) errored out with a non-zero exit code (%d).\n", binaryPath, status.ExitStatus())
				}
				// Exit with the same exit code as the binary
				os.Exit(status.ExitStatus())
			}
		}
		// If we can't determine the exit code, exit with a default code
		os.Exit(1)
	}

	// If the command executed successfully, exit with its exit code
	if status, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
		os.Exit(status.ExitStatus())
	}
}

// fetchBinary downloads the binary and caches it.
func fetchBinary(packageName string) error {
	url, err := findURL(packageName)
	if err != nil {
		return err
	}

	cachedFile := filepath.Join(TEMP_DIR, packageName+".bin")

	// Fetch the binary from the internet and save it to the cache
	err = fetchBinaryFromURL(url, cachedFile)
	if err != nil {
		return fmt.Errorf("error fetching binary for %s: %v", packageName, err)
	}

	// Ensure the cache size does not exceed the limit
	cleanCache()

	return nil
}

// cleanCache removes the oldest binaries when the cache size exceeds MaxCacheSize.
func cleanCache() {
	// Get a list of all binaries in the cache directory
	files, err := os.ReadDir(TEMP_DIR)
	if err != nil {
		fmt.Printf("Error reading cache directory: %v\n", err)
		return
	}

	// Check if the cache size has exceeded the limit
	if len(files) <= MaxCacheSize {
		return
	}

	// Use a custom struct to hold file info and atime
	type fileWithAtime struct {
		info  os.FileInfo
		atime time.Time
	}

	// Convert os.DirEntry to fileWithAtime and track the last accessed time
	var filesWithAtime []fileWithAtime
	for _, entry := range files {
		fileInfo, err := entry.Info()
		if err != nil {
			fmt.Printf("Error getting file info: %v\n", err)
			continue
		}

		// Use syscall to get atime
		var stat syscall.Stat_t
		err = syscall.Stat(filepath.Join(TEMP_DIR, entry.Name()), &stat)
		if err != nil {
			fmt.Printf("Error getting file stat: %v\n", err)
			continue
		}

		atime := time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
		filesWithAtime = append(filesWithAtime, fileWithAtime{info: fileInfo, atime: atime})
	}

	// Sort files by last accessed time
	sort.Slice(filesWithAtime, func(i, j int) bool {
		return filesWithAtime[i].atime.Before(filesWithAtime[j].atime)
	})

	// Delete the oldest BinariesToDelete
	for i := 0; i < BinariesToDelete; i++ {
		err := os.Remove(filepath.Join(TEMP_DIR, filesWithAtime[i].info.Name()))
		if err != nil {
			fmt.Printf("Error removing file: %v\n", err)
		}
	}
}
