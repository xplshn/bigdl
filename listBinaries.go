// listBinaries.go // This file implements the listBinaries function //>
package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// listBinaries fetches and lists binary names from the given URL.
func listBinaries() ([]string, error) {
	var allBinaries []struct {
		Name    string `json:"Name"`
		NameAlt string `json:"name"`
		SHA256  string `json:"sha256"`
		SHA     string `json:"sha"`
	}

	// Fetch binaries from each metadata URL
	for _, url := range MetadataURLs {
		var metadata []struct {
			Name    string `json:"Name"`
			NameAlt string `json:"name"`
			SHA256  string `json:"sha256"`
			SHA     string `json:"sha"`
		}

		// Use fetchJSON to fetch and unmarshal the JSON data
		if err := fetchJSON(url, &metadata); err != nil {
			return nil, fmt.Errorf("failed to fetch metadata from %s: %v", url, err)
		}

		// Extract binary names and SHA values
		allBinaries = append(allBinaries, metadata...)
	}

	// Filter out excluded file types and file names, and build a map of SHA256 and SHA values to binary names
	filteredBinaries := make(map[string]string)
	shaMap := make(map[string]bool)
	for _, item := range allBinaries {
		binary := item.Name
		if binary == "" {
			binary = item.NameAlt
		}

		if binary != "" {
			ext := strings.ToLower(filepath.Ext(binary))
			base := filepath.Base(binary)
			if _, excluded := excludedFileTypes[ext]; !excluded {
				if _, excludedName := excludedFileNames[base]; !excludedName {
					if _, exists := shaMap[item.SHA256]; !exists {
						shaMap[item.SHA256] = true
						filteredBinaries[binary] = item.SHA256
					}
					if _, exists := shaMap[item.SHA]; !exists {
						shaMap[item.SHA] = true
						filteredBinaries[binary] = item.SHA
					}
				}
			}
		}
	}

	// Define and fill uniqueBinaries
	var uniqueBinaries []string
	for binary := range filteredBinaries {
		uniqueBinaries = append(uniqueBinaries, binary)
	}
	// Remove duplicates (entries with same name)
	uniqueBinaries = removeDuplicates(uniqueBinaries)

	// Return the list of binaries
	return uniqueBinaries, nil
}
