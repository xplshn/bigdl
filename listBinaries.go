// listBinaries.go // This file implements the listBinaries function //

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
func listBinaries(return_var bool) ([]string, error) {
	var allBinaries []string

	// Fetch binaries from each metadata URL
	for _, url := range MetadataURLs {
		// Fetch metadata from the given URL
		resp, err := http.Get(url)
		if err != nil {
			if return_var {
				return nil, fmt.Errorf("error fetching metadata from %s: %v", url, err)
			}
			fmt.Printf("Notice: Error fetching metadata from %s: %v\n", url, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if return_var {
				return nil, fmt.Errorf("failed to fetch metadata from %s. HTTP status code: %d", url, resp.StatusCode)
			}
			fmt.Printf("Notice: Failed to fetch metadata from %s. HTTP status code: %d\n", url, resp.StatusCode)
			continue
		}

		// Read response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			if return_var {
				return nil, fmt.Errorf("failed to read response body: %v", err)
			}
			fmt.Printf("Notice: Failed to read response body: %v\n", err)
			continue
		}

		// Unmarshal JSON
		var metadata []struct {
			Name    string `json:"name"`
			NameAlt string `json:"Name"` // Consider both "name" and "Name" fields
		}
		if err := json.Unmarshal(body, &metadata); err != nil {
			if return_var {
				return nil, fmt.Errorf("failed to unmarshal metadata JSON from %s: %v", url, err)
			}
			fmt.Printf("Notice: Failed to unmarshal metadata JSON from %s: %v\n", url, err)
			continue
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
		"_dir":  {},
	}

	excludedFileNames := map[string]struct{}{
		"robotstxt": {},
		"LICENSE":   {},
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

	// If return_var is true, return the list of binaries
	if return_var {
		return uniqueBinaries, nil
	}

	// Otherwise, print the binaries
	fmt.Println(strings.Join(uniqueBinaries, "\n"))
	return nil, nil
}
