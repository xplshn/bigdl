// findURL.go // This file implements the findURL function //>
package main

import (
	"fmt"
	"net/http"
	"path/filepath"
)

// findURLCommand returns the URL for the specified binary. We do not use info.go for this because unmarshalling such big files is slower than pinging to see which exists
func findURLCommand(binaryName string) {
	url, err := findURL(binaryName)
	if err != nil {
		errorOut("error: %v\n", err)
	}

	fmt.Println(url)
}

// findURL fetches the URL for the specified binary.
func findURL(binaryName string) (string, error) {
	// Check the tracker file first
	realBinaryName, err := getBinaryNameFromTrackerFile(filepath.Base(binaryName))
	if err == nil {
		binaryName = realBinaryName
	}

	iterations := 0
	for _, Repository := range Repositories {
		iterations++
		url := fmt.Sprintf("%s%s", Repository, binaryName)
		fmt.Printf("\033[2K\r<%d/%d> | Working: Checking if \"%s\" is in the repos.", iterations, len(Repositories), binaryName)
		resp, err := http.Head(url)
		if err != nil {
			return "", err
		}

		if resp.StatusCode == http.StatusOK {
			fmt.Printf("\033[2K\r<%d/%d> | Found \"%s\" at %s", iterations, len(Repositories), binaryName, Repository)
			return url, nil
		}
	}

	fmt.Printf("\033[2K\r")
	return "", fmt.Errorf("Didn't find the SOURCE_URL for [%s]", binaryName)
}
