// run.go // This file implements functions related to the Run options //

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ReturnCachedFile retrieves the cached file location.
// Returns an empty string and error code 1 if not found.
func ReturnCachedFile(packageName string) (string, int) {
	if fileExists(CACHE_FILE) {
		lines, err := readLines(CACHE_FILE)
		if err != nil {
			return "", 1
		}

		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[0] == packageName {
				cachedFile := filepath.Join(TEMP_DIR, fields[1])
				if fileExists(cachedFile) {
					return cachedFile, 0
				}
			}
		}
	}

	return "", 1
}

// CleanCache removes duplicate entries and non-existent files from the cache.
func CleanCache() {
	if fileExists(CACHE_FILE) {
		// Read all lines from the cache file
		lines, err := readLines(CACHE_FILE)
		if err != nil {
			fmt.Printf("Error reading cache file: %v\n", err)
			return
		}

		// Track unique entries to remove duplicates
		uniqueLines := make(map[string]struct{})
		var resultLines []string

		// Remove non-existent files and duplicates from the cache
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				packageName := fields[0]
				cachedFile := filepath.Join(TEMP_DIR, fields[1])

				// Check if the program is in the cache
				if fileExists(cachedFile) && packageName == getPackageName(fields[1]) {
					// Remove the non-existent file only if not duplicated
					key := packageName + " " + cachedFile
					if _, exists := uniqueLines[key]; !exists {
						uniqueLines[key] = struct{}{}
						resultLines = append(resultLines, line)
					}
				}
			}
		}

		// Write the unique lines back to the cache file
		writeLines(CACHE_FILE, resultLines)
	}
}

// fetchBinary downloads the binary and returns the cache entry.
func fetchBinary(packageName string, tempDir, cachedLocation string) error {
	url, err := findURL(packageName)
	if err != nil {
		return err
	}

	// Use TEMP_DIR for caching
	cacheDir := TEMP_DIR

	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return fmt.Errorf("Error creating cache directory: %v", err)
	}

	installPath := cachedLocation

	// Use fetchBinaryFromURL to download the binary
	if err := fetchBinaryFromURL(url, installPath); err != nil {
		return fmt.Errorf("Error installing binary: %v", err)
	}

	installMessage := "Cached %s to %s\n"
	fmt.Printf(installMessage, packageName, installPath)

	// Update the cache file with the fetched binary using the original cached location
	if err := appendLineToFile(CACHE_FILE, fmt.Sprintf("%s %s", packageName, cachedLocation)); err != nil {
		return fmt.Errorf("Error updating cache file: %v", err)
	}

	return nil
}

// RunFromCache runs the binary from cache or fetches it if not found.
func RunFromCache(packageName string, args []string) {
	cachedLocation, errCode := ReturnCachedFile(packageName)

	if errCode == 0 && isExecutable(cachedLocation) {
		fmt.Printf("Running '%s' from cache...\n", packageName)
		CleanCache()

		cmd := exec.Command(cachedLocation, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running '%s' from cache: %v\n", packageName, err)
			os.Exit(1)
		}
	} else {
		newCachedLocation := fmt.Sprintf("bigdl_%s-%d", packageName, getCurrentTimestamp())
		if errCode == 1 {
			fmt.Printf("Error: cached binary for '%s' not found. Fetching a new one...\n", packageName)
		}
		if err := fetchBinary(packageName, TEMP_DIR, filepath.Join(TEMP_DIR, newCachedLocation)); err == nil {
			writeLineToFile(CACHE_FILE, fmt.Sprintf("%s %s", packageName, newCachedLocation))
			CleanCache()

			// Refresh the cachedLocation after fetching
			cachedLocation, _ = ReturnCachedFile(packageName)

			cmd := exec.Command(cachedLocation, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			if err := cmd.Run(); err != nil {
				fmt.Printf("Error running fetched binary '%s': %v\n", packageName, err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Error fetching binary for '%s': %v\n", packageName, err)
			os.Exit(1)
		}
	}
}

// readLines reads all lines from a file and returns them as a slice of strings.
func readLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// writeLines writes the given lines to a file, overwriting its previous content.
func writeLines(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range lines {
		_, err := fmt.Fprintln(file, line)
		if err != nil {
			return err
		}
	}

	return nil
}

// removeLineFromFile removes the specified line from a file.
func removeLineFromFile(filePath string, lineToRemove string) error {
	lines, err := readLines(filePath)
	if err != nil {
		return err
	}

	var updatedLines []string
	for _, line := range lines {
		if line != lineToRemove {
			updatedLines = append(updatedLines, line)
		}
	}

	return writeLines(filePath, updatedLines)
}

// getPackageName extracts the package name from a cached file name.
func getPackageName(cachedFileName string) string {
	// Assuming the format is "bigdl_PackageName-Timestamp"
	parts := strings.Split(cachedFileName, "-")
	if len(parts) >= 2 {
		return strings.TrimPrefix(parts[0], "bigdl_")
	}
	return ""
}

// getCurrentTimestamp returns the current Unix timestamp.
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// isExecutable checks if the file at the specified path is executable.
func isExecutable(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.Mode().Perm()&0100 != 0
}

// writeLineToFile appends a line to the end of a file.
func writeLineToFile(filePath, line string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, line)
	return err
}
