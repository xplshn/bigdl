// run.go // This file implements the "run" functionality //>
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

var verboseMode bool
var silentMode bool
var transparentMode bool

// ReturnCachedFile retrieves the cached file location. Returns an empty string and error code 1 if not found.
func ReturnCachedFile(binaryName string) (string, int) {
	// Construct the expected cached file pattern
	expectedCachedFile := filepath.Join(TEMP_DIR, fmt.Sprintf("%s.bin", binaryName))

	// Check if the file exists using the fileExists function
	if fileExists(expectedCachedFile) {
		return expectedCachedFile, 0
	}

	// If the file does not exist, return an error
	return "", 1
}

// RunFromCache runs the binary from cache or fetches it if not found.
func RunFromCache(binaryName string, args []string) {

	// purifyVars is a function to purify binaryName and args.
	purifyVars := func() {
		if len(args) > 0 {
			binaryName = (args)[0] // Purify binaryName
			args = (args)[1:]      // Appropriately set args to exclude any of the flags
		} else {
			errorOut("Error: Binary name not provided after flag.\n")
		}
	}

	// Process flags
	verbose := flag.Bool("verbose", false, "Enable verbose mode")
	silent := flag.Bool("silent", false, "Enable silent mode")
	transparent := flag.Bool("transparent", false, "Enable transparent mode")

	flags_AndBinaryName := append(strings.Fields(binaryName), args...)
	flag.CommandLine.Parse(flags_AndBinaryName)

	if *verbose && *silent {
		errorOut("Error: --verbose and --silent are mutually exclusive\n")
	}

	if *verbose {
		verboseMode = true
		purifyVars()
	}

	if *silent {
		silentMode = true
		purifyVars()
	}

	if *transparent {
		transparentMode = true

		purifyVars()
		isInPath, err := isBinaryInPath(binaryName)
		if err != nil {
			errorOut("Error checking if binary is in PATH: %s\n", err)
		}

		if isInPath {
			if !silentMode {
				fmt.Printf("Running '%s' from PATH...\n", binaryName)
			}
			runBinary(binaryName, args, verboseMode)
		}
	}

	if binaryName == "" {
		errorOut("Error: Binary name not provided\n")
	}

	cachedFile := filepath.Join(TEMP_DIR, binaryName+".bin")
	if fileExists(cachedFile) && isExecutable(cachedFile) {
		if !silentMode {
			fmt.Printf("Running '%s' from cache...\n", binaryName)
		}
		runBinary(cachedFile, args, verboseMode)
		cleanCache()
	} else {
		if verboseMode {
			fmt.Printf("Couldn't find '%s' in the cache. Fetching a new one...\n", binaryName)
		}
		err := fetchBinary(binaryName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching binary for '%s'\n", binaryName)
			errorOut("Error: %s\n", err)
		}
		cleanCache()
		runBinary(cachedFile, args, verboseMode)
	}
}

// runBinary executes the binary with the given arguments, handling .bin files as needed.
func runBinary(binaryPath string, args []string, verboseMode bool) {
	var programExitCode int = 1
	executeBinary := func(rbinaryPath string, args []string, verboseMode bool) {
		cmd := exec.Command(rbinaryPath, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					if verboseMode {
						fmt.Printf("The program (%s) errored out with a non-zero exit code (%d).\n", binaryPath, status.ExitStatus())
					}
					programExitCode = status.ExitStatus()
				}
			} else {
				programExitCode = 1
			}
		} else {
			programExitCode = 0
		}
	}
	// If the file ends with .bin, remove the suffix and proceed to run it, afterwards, set the same suffix again.
	if strings.HasSuffix(binaryPath, ".bin") {
		tempFile := filepath.Join(filepath.Dir(binaryPath), strings.TrimSuffix(filepath.Base(binaryPath), ".bin"))
		if err := copyFile(binaryPath, tempFile); err != nil {
			fmt.Printf("failed to move binary to temporary location: %v\n", err)
			return
		}
		if err := os.Chmod(tempFile, 0755); err != nil {
			fmt.Printf("failed to set executable bit: %v\n", err)
			return
		}

		executeBinary(tempFile, args, verboseMode)
		if err := os.Rename(tempFile, binaryPath); err != nil {
			fmt.Printf("failed to move temporary binary back to original name: %v\n", err)
			return
		}
	}

	// Exit the program with the exit code from the executed binary or 1 if we couldn't even get to the execution
	os.Exit(programExitCode)
}

// fetchBinary downloads the binary and caches it.
func fetchBinary(binaryName string) error {
	if silentMode {
		useProgressBar = false
	}

	url, err := findURL(binaryName)
	if err != nil {
		return err
	}

	cachedFile := filepath.Join(TEMP_DIR, binaryName+".bin")

	// Fetch the binary from the internet and save it to the cache
	err = fetchBinaryFromURL(url, cachedFile)
	if err != nil {
		return fmt.Errorf("error fetching binary for %s: %v", binaryName, err)
	}

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
			if !silentMode { // Check if not in silent mode before printing
				fmt.Printf("Error removing file: %v\n", err)
			}
		}
	}
}
