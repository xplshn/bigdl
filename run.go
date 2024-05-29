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

var (
	verboseMode bool
	silentMode  bool
)

// ReturnCachedFile retrieves the cached file location. Returns an empty string and error code 1 if not found.
func ReturnCachedFile(binaryName string) (string, int) {
	cachedBinary := filepath.Join(TEMPDIR, binaryName)
	if fileExists(cachedBinary) {
		return cachedBinary, 0
	}
	// The file does not exist, return an error
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

	flagsAndBinaryName := append(strings.Fields(binaryName), args...)
	flag.CommandLine.Parse(flagsAndBinaryName)

	if *verbose && *silent {
		errorOut("error: --verbose and --silent are mutually exclusive\n")
	}

	if *verbose {
		verboseMode = true
		purifyVars()
	}

	if *silent {
		silentMode = true
		UseProgressBar = false
		purifyVars()
	}

	if *transparent {
		purifyVars()
		binaryPath, _ := exec.LookPath(binaryName) // is it okay to ignore the err channel of LookPath?

		if binaryPath != "" {
			if !silentMode {
				fmt.Printf("Running '%s' from PATH...\n", binaryName)
			}
			runBinary(binaryPath, args, verboseMode)
		}
	}

	if binaryName == "" {
		errorOut("error: Binary name not provided\n")
	}

	// Use the base name of binaryName to construc the cachedFile path. This way requests like toybox/wget are supported
	cachedFile := filepath.Join(TEMPDIR, filepath.Base(binaryName))

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
		InstallDir = TEMPDIR
		InstallMessage = ""
		if err := installCommand(silentMode, binaryName); err != nil {
			errorOut("%v\n", err)
		}
		cleanCache()
		runBinary(cachedFile, args, verboseMode)
	}
}

// runBinary executes the binary with the given arguments.
func runBinary(binaryPath string, args []string, verboseMode bool) {
	// Set the Controls for the Heart of the Sun
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	if err != nil && verboseMode {
		fmt.Printf("The program (%s) errored out with a non-zero exit code (%d).\n", binaryPath, exitCode)
	}

	os.Exit(exitCode)
}

// cleanCache removes the oldest binaries when the cache size exceeds MaxCacheSize.
func cleanCache() {
	// Get a list of all binaries in the cache directory
	files, err := os.ReadDir(TEMPDIR)
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
		err = syscall.Stat(filepath.Join(TEMPDIR, entry.Name()), &stat)
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

	// Delete the oldest binaries
	for i := 0; i < BinariesToDelete; i++ {
		err := os.Remove(filepath.Join(TEMPDIR, filesWithAtime[i].info.Name()))
		if err != nil {
			if !silentMode { // Check if not in silent mode before printing
				fmt.Printf("Error removing file: %v\n", err)
			}
		}
	}
}
