// helperFunctions.go // This file contains commonly used functions //>

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// signalHandler sets up a channel to listen for interrupt signals and returns a function
// that can be called to check if an interrupt has been received.
func signalHandler(ctx context.Context, cancel context.CancelFunc) (func() bool, error) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel() // Call the cancel function when an interrupt is received
	}()

	return func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}, nil
}

// fetchBinaryFromURL fetches a binary from the given URL and saves it to the specified destination.
func fetchBinaryFromURL(url, destination string) error {
	// Set up the signal handler at the start of the function.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the cancel function is called when the function returns

	// Create a temporary directory if it doesn't exist
	if err := os.MkdirAll(TEMP_DIR, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}

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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching binary from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch binary from %s. HTTP status code: %d", url, resp.StatusCode)
	}

	// Write the binary to the temporary file
	buf := make([]byte, 32*1024) //  32KB buffer
	for {
		n, err := resp.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				// End of file, break the loop
				break
			}
			// Delete the line before printing the error
			fmt.Printf("\r\033[K") // Overwrite the current line with spaces
			return fmt.Errorf("failed to read from response body: %v", err)
		}

		// Write to the temporary file
		_, err = out.Write(buf[:n])
		if err != nil {
			// Delete the line before printing the error
			fmt.Printf("\r\033[K") // Overwrite the current line with spaces
			return fmt.Errorf("failed to write to temporary file: %v", err)
		}
	}

	// Close the file before setting executable bit
	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Stop the spinner
	StopSpinner()

	// Move the binary to its destination
	if err := copyFile(tempFile, destination); err != nil {
		return fmt.Errorf("failed to move binary to destination: %v", err)
	}

	// Set executable bit immediately after copying
	if err := os.Chmod(destination, 0755); err != nil {
		return fmt.Errorf("failed to set executable bit: %v", err)
	}

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

// GetTerminalWidth attempts to determine the width of the terminal.
// It first tries using "stty size", then "tput cols", and finally falls back to  80 columns.
func getTerminalWidth() int {
	// Try using stty size
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err == nil {
		// stty size returns rows and columns
		parts := strings.Split(strings.TrimSpace(string(out)), " ")
		if len(parts) == 2 {
			width, err := strconv.Atoi(parts[1])
			if err == nil {
				return width
			}
		}
	}

	// Fallback to tput cols
	cmd = exec.Command("tput", "cols")
	cmd.Stdin = os.Stdin
	out, err = cmd.Output()
	if err == nil {
		width, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err == nil {
			return width
		}
	}

	// Fallback to  80 columns
	return 80
}

// truncateSprintf formats the string and truncates it if it exceeds the terminal width.
func truncateSprintf(format string, a ...interface{}) string {
	// Format the string first
	formatted := fmt.Sprintf(format, a...)

	// Determine the truncation length & truncate the formatted string if it exceeds the available space
	availableSpace := getTerminalWidth()
	if len(formatted) > availableSpace {
		formatted = fmt.Sprintf("%s", formatted[:availableSpace-4]) // Shrink to the maximum line size, accounting for the dots to be added.
		for strings.HasSuffix(formatted, ",") || strings.HasSuffix(formatted, ".") || strings.HasSuffix(formatted, " ") {
			formatted = formatted[:len(formatted)-1]
		}
		formatted = fmt.Sprintf("%s...>", formatted) // Add the dots
	}

	return formatted
}

// truncatePrintf is a drop-in replacement for fmt.Printf that truncates the input string if it exceeds a certain length.
func truncatePrintf(format string, a ...interface{}) (n int, err error) {
	// Call truncateSprintf to get the formatted and truncated string
	formatted := truncateSprintf(format, a...)

	// Print the possibly truncated string
	return fmt.Print(formatted)
}
