// helperFunctions.go // This file contains commonly used functions //>
package main

import (
	"context"
	"encoding/json"
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

	"github.com/schollz/progressbar/v3"
)

// TODO: Add *PROPER* error handling in the truncate functions. Ensure escape sequences are correctly handled?

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
	if err := os.MkdirAll(TEMPDIR, 0o755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}

	// Create a temporary file to download the binary
	tempFile := filepath.Join(TEMPDIR, filepath.Base(destination)+".tmp")
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
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch binary from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch binary from %s. HTTP status code: %d", url, resp.StatusCode)
	}

	bar := spawnProgressBar(resp.ContentLength)

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
	if err := os.Chmod(destination, 0o755); err != nil {
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
	fmt.Print("\033[2K\r") // Clean the line
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

func fetchJSON(url string, v interface{}) error {
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error fetching from %s: %v", url, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading from %s: %v", url, err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("error decoding from %s: %v", url, err)
	}

	return nil
}

// removeDuplicates removes duplicate elements from the input slice.
func removeDuplicates(input []string) []string {
	seen := make(map[string]bool)
	unique := []string{}
	if input != nil {
		for _, entry := range input {
			if _, value := seen[entry]; !value {
				seen[entry] = true
				unique = append(unique, entry)
			}
		}
	} else {
		unique = input
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

// isExecutable checks if the file at the specified path is executable.
func isExecutable(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular() && (info.Mode().Perm()&0o111) != 0
}

// listFilesInDir lists all files in a directory
func listFilesInDir(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, dir+"/"+entry.Name())
		}
	}
	return files, nil
}

func spawnProgressBar(contentLength int64) *progressbar.ProgressBar {
	if useProgressBar {
		return progressbar.NewOptions(int(contentLength),
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
	}
	return progressbar.NewOptions(0, progressbar.OptionClearOnFinish()) // A dummy
}

// sanitizeString removes certain punctuation from the end of the string and converts it to lower case.
func sanitizeString(s string) string {
	// Define the punctuation to remove
	punctuation := []string{".", " ", ",", "!", "?"}

	// Convert string to lower case
	s = strings.ToLower(s)

	// Remove specified punctuation from the end of the string
	for _, p := range punctuation {
		s = s[:len(s)-len(p)]
	}

	return s
}

// contanins will return true if the provided slice of []strings contains the word str
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// errorEncoder generates a unique error code based on the sum of ASCII values of the error message.
func errorEncoder(format string, args ...interface{}) int {
	formattedErrorMessage := fmt.Sprintf(format, args...)

	var sum int
	for _, char := range formattedErrorMessage {
		sum += int(char)
	}
	errorCode := sum % 256
	fmt.Fprint(os.Stderr, formattedErrorMessage)
	return errorCode
}

// errorOut prints the error message to stderr and exits the program with the error code generated by errorEncoder.
func errorOut(format string, args ...interface{}) {
	os.Exit(errorEncoder(format, args...))
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
			width, _ := strconv.Atoi(parts[1])
			return width
		}
	}

	// Fallback to tput cols
	cmd = exec.Command("tput", "cols")
	cmd.Stdin = os.Stdin
	out, err = cmd.Output()
	if err == nil {
		width, _ := strconv.Atoi(strings.TrimSpace(string(out)))
		return width
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
	if disableTruncation {
		return fmt.Print(fmt.Sprintf(format, a...))
	}
	return fmt.Print(truncateSprintf(format, a...))
} // NOTE: Both truncate functions will remove the escape sequences of truncated lines, and sometimes break them in half because of the truncation. Avoid using escape sequences with truncate functions, as it is UNSAFE.

// validateProgramsFrom validates programs against the files in the specified directory against the remote binaries.
// It returns the validated programs based on the last element of the received list of programs.
func validateProgramsFrom(InstallDir string, programsToValidate []string) ([]string, error) {
	// Fetch the list of binaries from the remote source once
	remotePrograms, err := listBinaries()
	if err != nil {
		return nil, fmt.Errorf("failed to list remote binaries: %w", err)
	}

	// List files from the specified directory
	files, err := listFilesInDir(InstallDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in %s: %w", InstallDir, err)
	}

	validPrograms := make([]string, 0)
	invalidPrograms := make([]string, 0)

	programsToValidate = removeDuplicates(programsToValidate)

	// If programsToValidate is nil, validate all programs in the install directory
	if programsToValidate == nil {
		for _, file := range files {
			// Extract the file name from the full path
			fileName := filepath.Base(file)
			if contains(remotePrograms, fileName) {
				validPrograms = append(validPrograms, fileName)
			} else {
				invalidPrograms = append(invalidPrograms, fileName)
			}
		}
	} else {
		// Only check the ones specified in programsToValidate
		for _, program := range programsToValidate {
			if contains(remotePrograms, program) {
				validPrograms = append(validPrograms, program)
			} else {
				invalidPrograms = append(invalidPrograms, program)
			}
		}
	}

	// Handle the list of programs received based on the last element
	// If programsToValidate is not nil, handle based on the last element
	if len(programsToValidate) != 0 {
		lastElement := programsToValidate[len(programsToValidate)-1]
		switch lastElement {
		case "_2_":
			return invalidPrograms, nil
		case "_3_":
			return append(validPrograms, invalidPrograms...), nil
		default:
			return validPrograms, nil
		}
	}
	return validPrograms, nil
}
