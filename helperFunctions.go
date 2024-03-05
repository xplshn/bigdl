// helperFunctions.go // This file contains commonly used functions //>
// TODO: Add *PROPER* error handling in the truncate functions. Ensure escape sequences are correctly handled?
package main

import (
	"context"
	"fmt"
	"github.com/schollz/progressbar/v3"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the cancel function is called when the function returns

	// Set up signal handling
	interrupted, err := signalHandler(ctx, cancel)
	if err != nil {
		return fmt.Errorf("failed to set up signal handler: %v", err)
	}

	// Create a temporary directory if it doesn't exist
	if err := os.MkdirAll(TEMP_DIR, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}

	// Create a temporary file to download the binary
	tempFile := filepath.Join(TEMP_DIR, filepath.Base(destination)+".tmp")
	out, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer out.Close()

	// Schedule the deletion of the temporary file
	defer func() {
		if err := os.Remove(tempFile); err != nil && !os.IsNotExist(err) {
			fmt.Printf("\r\033[Kfailed to remove temporary file: %v\n", err)
		}
	}()

	// Fetch the binary from the given URL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Ensure that redirects are followed
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching binary from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch binary from %s. HTTP status code: %d", url, resp.StatusCode)
	}

	// Create a progress bar
	var bar *progressbar.ProgressBar
	if useProgressBar {
		bar = progressbar.NewOptions(int(resp.ContentLength),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionFullWidth(),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerHead:    ">",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)
	} else {
		bar = progressbar.NewOptions(-1,
			progressbar.OptionSetVisibility(false), // Couldn't make update.go work well enough with it.
			progressbar.OptionSpinnerType(9),       // Type 9 spinner (Classic BSD styled spinner; "|/-\").
		)
	}

	// Write the binary to the temporary file with progress bar
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to temporary file: %v", err)
	}

	// Close the file before setting executable bit
	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Use copyFile to move the binary to its destination
	if err := copyFile(tempFile, destination); err != nil {
		return fmt.Errorf("failed to move binary to destination: %v", err)
	}

	// Set executable bit immediately after copying
	if err := os.Chmod(destination, 0755); err != nil {
		return fmt.Errorf("failed to set executable bit: %v", err)
	}

	// Check if the operation was interrupted
	if interrupted() {
		fmt.Println("\r\033[KDownload interrupted. Cleaning up...")
		// Ensure the temporary file is removed if the operation was interrupted
		if err := os.Remove(tempFile); err != nil && !os.IsNotExist(err) {
			fmt.Printf("failed to remove temporary file: %v\n", err)
		}
	}
	fmt.Println("\r\033[K")
	return nil
}

// copyFile copies(removes original after copy) a file from src to dst
func copyFile(src, dst string) error {
	// Check if the destination file already exists
	if fileExists(dst) {
		if err := os.Remove(dst); err != nil {
			return fmt.Errorf("%v", err)
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

// listFilesInDir lists all files in a directory
func listFilesInDir(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// sanitizeString removes certain punctuation from the end of the string and converts it to lower case.
func sanitizeString(s string) string {
	// Define the punctuation to remove
	punctuation := []string{".", " ", ",", "!", "?"}

	// Convert string to lower case
	s = strings.ToLower(s)

	// Remove specified punctuation from the end of the string
	for _, p := range punctuation {
		if strings.HasSuffix(s, p) {
			s = s[:len(s)-len(p)]
		}
	}

	return s
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
	availableSpace := getTerminalWidth() - len(indicator)
	if len(formatted) > availableSpace {
		formatted = fmt.Sprintf("%s", formatted[:availableSpace])
		for strings.HasSuffix(formatted, ",") || strings.HasSuffix(formatted, ".") || strings.HasSuffix(formatted, " ") {
			formatted = formatted[:len(formatted)-1]
		}
		formatted = fmt.Sprintf("%s%s", formatted, indicator) // Add the dots.
	}

	return formatted
}

// truncatePrintf is a drop-in replacement for fmt.Printf that truncates the input string if it exceeds a certain length.
func truncatePrintf(format string, a ...interface{}) (n int, err error) {
	formatted := truncateSprintf(format, a...)
	return fmt.Print(formatted)
}

// NOTE: Both truncate functions will remove the escape sequences of truncated lines, and sometimes break them in half because of the truncation. Avoid using escape sequences with truncate functions, as it is UNSAFE.
