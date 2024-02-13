// update.go // This file contains the complete implementation of the updateBulk function //

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
)

// Define custom error for files with the same size
var ErrSameSize = fmt.Errorf("source and destination files have the same size")

func updateBinary(binaryName, installDir string) error {
	url, err := findURL(binaryName)
	if err != nil {
		return fmt.Errorf("Error finding URL for binary %s: %v", binaryName, err)
	}

	// Generate a random string for the temporary file name
	randomStr := randomString(9)

	// Download the binary to a temporary directory with a random file extension
	tempFilePath := filepath.Join(TEMP_DIR, fmt.Sprintf("%s_%s", binaryName, randomStr))

	if err := fetchBinaryFromURL(url, tempFilePath); err != nil {
		return fmt.Errorf("Error downloading binary %s: %v", binaryName, err)
	}

	// Construct the final temporary file path with the same random string
	finalTempFilePath := filepath.Join(TEMP_DIR, fmt.Sprintf("%s_%s", binaryName, randomStr))

	// Move the temporary file to the installation directory
	if err := moveFile(finalTempFilePath, filepath.Join(installDir, binaryName)); err != nil {
		if err == ErrSameSize {
			fmt.Printf("Skipped updating binary: '%s' as they are the same version\n", binaryName)
			return nil
		}
		return fmt.Errorf("Error replacing binary %s: %v", binaryName, err)
	}

	if err == nil {
		fmt.Printf("Updated binary: '%s'\n", binaryName)
	}

	return nil
}

func updateBulk(args []string) error {
	var installDir = args[0]

	// Call listBinaries to print the list of all available binaries
	binaries, err := listBinaries(true)
	if err != nil {
		fmt.Printf("Error fetching list of binaries: %v\n", err)
		return err
	}

	// Get the list of binaries in the installation directory
	installedBinaries, err := listBinariesInDir(installDir)
	if err != nil {
		return fmt.Errorf("Error listing binaries in installation directory: %v", err)
	}

	for _, binaryName := range installedBinaries {
		if existsInList(binaryName, binaries) {
			if err := updateBinary(binaryName, installDir); err != nil {
				fmt.Printf("Error updating binary %s: %v\n", binaryName, err)
			}
		} else {
			fmt.Printf("Binary %s does not seem to come from our repos.\n", binaryName)
		}
	}

	return nil
}

func listBinariesInDir(directory string) ([]string, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var binaries []string
	for _, file := range files {
		if !file.IsDir() {
			binaries = append(binaries, file.Name())
		}
	}

	return binaries, nil
}

func existsInList(binaryName string, list []string) bool {
	for _, item := range list {
		if binaryName == item {
			return true
		}
	}
	return false
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func moveFile(src, dest string) error {
	// Get sizes of source and destination files
	srcSize := fileSize(src)
	destSize := fileSize(dest)

	// If source and destination files have the same size, no need to move
	if srcSize == destSize {
		return ErrSameSize
	}

	// Open source file for reading
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy content from source to destination
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	// Close destination file
	err = destFile.Close()
	if err != nil {
		return err
	}

	// Remove the source file
	err = os.Remove(src)
	if err != nil {
		return err
	}

	fmt.Printf("Moved file %s to %s\n", src, dest)

	return nil
}
