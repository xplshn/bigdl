// findURL.go // This file implements the findURL function //>
package main

import (
	"fmt"
	"net/http"
	"os"
)

// findURLCommand returns the URL for the specified binary.
func findURLCommand(binaryName string) {
	url, err := findURL(binaryName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(url)
}

// findURL fetches the URL for the specified binary.
func findURL(binaryName string) (string, error) {
	for _, Repository := range Repositories {
		url := fmt.Sprintf("%s%s", Repository, binaryName)
		resp, err := http.Head(url)
		if err != nil {
			return "", err
		}

		if resp.StatusCode == http.StatusOK {
			return url, nil
		}
	}

	return "", fmt.Errorf("Binary's SOURCE_URL was not found")
}
