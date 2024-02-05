// helperFunctions.go // This file contains commonly used functions //

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
)

// fetchBinaryFromURL fetches a binary from the given URL and saves it to the specified destination.
func fetchBinaryFromURL(url, destination string) error {
	// Use a wait group to wait for both the binary fetching and Spin to finish
	var wg sync.WaitGroup

	// Start Spin in a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		Spin()
	}()

	// Fetch the binary from the given URL
	resp, err := http.Get(url)
	if err != nil {
		StopSpinner() // Stop the spinner in case of an error
		return fmt.Errorf("Error fetching binary from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		StopSpinner() // Stop the spinner if fetching fails
		return fmt.Errorf("Failed to fetch binary from %s. HTTP status code: %d", url, resp.StatusCode)
	}

	// Create the file at the specified destination
	out, err := os.Create(destination)
	if err != nil {
		StopSpinner() // Stop the spinner in case of an error
		return fmt.Errorf("Failed to create file for binary: %v", err)
	}
	defer out.Close()

	// Write the binary to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		StopSpinner() // Stop the spinner in case of an error
		return fmt.Errorf("Failed to write binary to file: %v", err)
	}

	// Set executable bit
	if err := os.Chmod(destination, 0755); err != nil {
		StopSpinner() // Stop the spinner in case of an error
		return fmt.Errorf("Failed to set executable bit: %v", err)
	}

	StopSpinner() // Stop the spinner when binary fetching is successful
	wg.Wait()     // Wait for Spin to finish
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
