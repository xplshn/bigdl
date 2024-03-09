// listBinaries.go // This file implements the listBinaries function //>
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
)

// listBinariesCommand fetches and lists binary names from the given URL.
func listBinaries() ([]string, error) {
	var allBinaries []string

	// Fetch binaries from each metadata URL
	for _, url := range MetadataURLs {
		// Fetch metadata from the given URL
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("error fetching metadata from %s: %v", url, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch metadata from %s. HTTP status code: %d", url, resp.StatusCode)
		}

		// Read response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}

		// Unmarshal JSON
		var metadata []struct {
			Name    string `json:"name"`
			NameAlt string `json:"Name"` // Consider both "name" and "Name" fields
		}
		if err := json.Unmarshal(body, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata JSON from %s: %v", url, err)
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

	// Exclude specified file types and file names
	excludedFileTypes := map[string]struct{}{
		".7z":   {},
		".bz2":  {},
		".json": {},
		".gz":   {},
		".md":   {},
		".txt":  {},
		".tar":  {},
		".zip":  {},
	}

	excludedFileNames := map[string]struct{}{
		"robotstxt":                {},
		"LICENSE":                  {},
		"experimentalBinaries_dir": {},
		"bdl.sh":                   {},
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
