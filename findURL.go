// findURL.go // This file implements the findURL function //>
package main

import (
	"fmt"
	"net/http"
)

// findURLCommand returns the URL for the specified binary. We do not use info.go for this because unmarshalling such big files is slower than pinging to see which exists
func findURLCommand(binaryName string) {
	url, err := findURL(binaryName)
	if err != nil {
		errorOut("Error: %v\n", err)
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

	return "", fmt.Errorf("Error: Binary's SOURCE_URL was not found")
}
