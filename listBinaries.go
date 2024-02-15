// listBinaries.go // This file implements the listBinaries function //

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// listBinariesCommand fetches and lists binary names from the given URL.
func listBinaries() {
	var allBinaries []string

	// Fetch binaries from each metadata URL
	for _, url := range MetadataURLs {
		// Fetch metadata from the given URL
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error fetching metadata from %s: %v\n", url, err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Failed to fetch metadata from %s. HTTP status code: %d\n", url, resp.StatusCode)
			os.Exit(1)
		}

		// Read response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response body: %v\n", err)
			os.Exit(1)
		}

		// Unmarshal JSON
		var metadata []struct {
			Name    string `json:"name"`
			NameAlt string `json:"Name"` // Consider both "name" and "Name" fields
		}
		if err := json.Unmarshal(body, &metadata); err != nil {
			fmt.Printf("Failed to unmarshal metadata JSON from %s: %v\n", url, err)
			os.Exit(1)
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
		"_dir":  {},
		".zip":  {},
	}

	excludedFileNames := map[string]struct{}{
		"robotstxt": {},
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

	// Print binaries
	fmt.Println(strings.Join(uniqueBinaries, "\n"))
}
