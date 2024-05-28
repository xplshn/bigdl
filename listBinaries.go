// listBinaries.go // This file implements the listBinaries function //>
package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// listBinaries fetches and lists binary names from the given URL.
func listBinaries() ([]string, error) {
	var allBinaries []string

	// Fetch binaries from each metadata URL
	for _, url := range MetadataURLs {
		var metadata []struct {
			Name    string `json:"Name"` // Consider both "name" and "Name" fields
			NameAlt string `json:"name"`
		}

		// Use fetchJSON to fetch and unmarshal the JSON data
		if err := fetchJSON(url, &metadata); err != nil {
			return nil, fmt.Errorf("failed to fetch metadata from %s: %v", url, err)
		}

		// Extract binary names
		for _, item := range metadata {
			if item.Name != "" {
				allBinaries = append(allBinaries, item.Name)
			}
			if item.NameAlt != "" {
				allBinaries = append(allBinaries, item.NameAlt)
			}
		}
	}

	// Filter out excluded file types and file names
	var filteredBinaries []string
	for _, binary := range allBinaries {
		ext := strings.ToLower(filepath.Ext(binary))
		base := filepath.Base(binary)
		if _, excluded := excludedFileTypes[ext]; !excluded {
			if _, excludedName := excludedFileNames[base]; !excludedName {
				filteredBinaries = append(filteredBinaries, binary)
			}
		}
	}

	// Remove duplicates
	uniqueBinaries := removeDuplicates(filteredBinaries)

	// Sort binaries alphabetically
	sort.Strings(uniqueBinaries)

	// Return the list of binaries
	return uniqueBinaries, nil
}
