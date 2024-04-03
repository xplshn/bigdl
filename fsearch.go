// fsearch.go // this file implements the search option
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

// fSearch searches for binaries based on the given search term.
func fSearch(searchTerm string, limit int) {
	// Fetch metadata
	response, err := http.Get(RMetadataURL)
	if err != nil {
		fmt.Println("Failed to fetch binary information.")
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed to read metadata.")
		return
	}

	// Define a struct to match the JSON structure from RMetadataURL
	type Binary struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		// Include other fields if needed
	}

	type RMetadata struct {
		Binaries []Binary `json:"packages"`
	}

	// Unmarshal the description as an RMetadata object
	var rMetadata RMetadata
	if err := json.Unmarshal(body, &rMetadata); err != nil {
		errorOut("Error while decoding metadata from %s: %v\n", RMetadataURL, err)
	}

	// Filter binaries based on the search term and architecture
	searchResultsSet := make(map[string]struct{}) // Use a set to keep track of unique entries
	for _, binary := range rMetadata.Binaries {
		if strings.Contains(strings.ToLower(binary.Name+binary.Description), strings.ToLower(searchTerm)) {
			// Filter out excluded file types and file names
			ext := strings.ToLower(filepath.Ext(binary.Name))
			base := filepath.Base(binary.Name)
			if _, excluded := excludedFileTypes[ext]; !excluded {
				if _, excludedName := excludedFileNames[base]; !excludedName {
					entry := fmt.Sprintf("%s - %s", binary.Name, binary.Description)
					searchResultsSet[entry] = struct{}{} // Add the entry to the set
				}
			}
		}
	}

	// Check if no matching binaries found
	if len(searchResultsSet) == 0 {
		fmt.Printf("No matching binaries found for '%s'.\n", searchTerm)
		return
	} else if len(searchResultsSet) > limit {
		fmt.Printf("Too many matching binaries (+%d. [Use --limit before your query]) found for '%s'.\n", limit, searchTerm)
		return
	}

	// Convert set to slice for sorting
	var searchResults []string
	for entry := range searchResultsSet {
		searchResults = append(searchResults, entry)
	}

	// Sort the search results
	searchResults = sortBinaries(searchResults)

	// Check if the binary exists in the INSTALL_DIR and print results with installation state indicators
	for _, line := range searchResults {
		parts := strings.SplitN(line, " - ", 2)
		name := parts[0]
		description := parts[1]

		installPath := filepath.Join(InstallDir, name)

		cachedLocation, _ := ReturnCachedFile(name)

		var prefix string
		if fileExists(installPath) {
			prefix = "[i]"
		} else if cachedLocation != "" && isExecutable(cachedLocation) {
			prefix = "[c]"
		} else {
			prefix = "[-]"
		}

		truncatePrintf("%s %s - %s ", prefix, name, description)
		fmt.Printf("\n") // Escape sequences are truncated too...
	}
}
