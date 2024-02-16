// helperFunctions.go // This file contains commonly used functions //

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
)

// signalHandler sets up a channel to listen for interrupt signals and returns a function
// that can be called to check if an interrupt has been received.
func signalHandler() (func() bool, error) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	interrupted := false
	go func() {
		<-sigChan
		interrupted = true
	}()

	return func() bool {
		return interrupted
	}, nil
}

// fetchBinaryFromURL fetches a binary from the given URL and saves it to the specified destination.
// fetchBinaryFromURL fetches a binary from the given URL and saves it to the specified destination.
func fetchBinaryFromURL(url, destination string) error {
	// Set up the signal handler at the start of the function.
	isInterrupted, err := signalHandler()
	if err != nil {
		return fmt.Errorf("error setting up signal handler: %v", err)
	}

	// Create a temporary directory if it doesn't exist
	if err := os.MkdirAll(TEMP_DIR, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}

	// Create a channel to handle interruption with CTRL+C or other signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Start spinner
	Spin("")

	// Create a temporary file to download the binary
	tempFile := filepath.Join(TEMP_DIR, filepath.Base(destination)+".tmp")
	out, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer out.Close()

	// Fetch the binary from the given URL
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error fetching binary from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch binary from %s. HTTP status code: %d", url, resp.StatusCode)
	}

	// Write the binary to the temporary file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write binary to file: %v", err)
	}

	// Close the file before setting executable bit
	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Check for interruption after the binary has been downloaded.
	if isInterrupted() {
		// If an interrupt was received, clean up and exit.
		StopSpinner()
		return fmt.Errorf("installation interrupted")
	}

	// Move the binary to its destination
	if err := copyFile(tempFile, destination); err != nil {
		// If copying fails, remove the temporary file
		os.Remove(tempFile)
		return fmt.Errorf("failed to move binary to destination: %v", err)
	}

	// Set executable bit immediately after copying
	if err := os.Chmod(destination, 0755); err != nil {
		return fmt.Errorf("failed to set executable bit: %v", err)
	}

	// Stop spinner
	StopSpinner()

	return nil
}

// copyFile copies(removes original after copy) a file from src to dst
func copyFile(src, dst string) error {
	// Check if the destination file already exists
	if fileExists(dst) {
		// File exists, handle accordingly (e.g., overwrite or skip)
		// For this example, we'll overwrite the file
		if err := os.Remove(dst); err != nil {
			return fmt.Errorf("failed to remove existing destination file: %v", err)
		}
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		destFile.Close() // Ensure the destination file is closed
		return fmt.Errorf("failed to copy file: %v", err)
	}

	if err := destFile.Close(); err != nil {
		return fmt.Errorf("failed to close destination file: %v", err)
	}

	// Remove the temporary file after copying
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove source file: %v", err)
	}

	return nil
}

// removeDuplicates removes duplicate elements from the input slice.
func removeDuplicates(input []string) []string {
	seen := make(map[string]struct{})
	var unique []string
	for _, s := range input {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			unique = append(unique, s)
		}
	}
	return unique
}

// sortBinaries sorts the input slice of binaries.
func sortBinaries(binaries []string) []string {
	sort.Strings(binaries)
	return binaries
}

// fileExists checks if a file exists.
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// appendLineToFile appends a line to the end of a file.
func appendLineToFile(filePath, line string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = fmt.Fprintln(file, line)
	return err
}

// fileSize returns the size of the file at the specified path.
func fileSize(filePath string) int64 {
	file, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return 0
	}

	return stat.Size()
}

// isExecutable checks if the file at the specified path is executable.
func isExecutable(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular() && (info.Mode().Perm()&0111) != 0
}
