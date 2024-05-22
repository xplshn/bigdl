// fsearch.go // this file implements the search option
package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// fSearch searches for binaries based on the given search term.
func fSearch(searchTerm string, limit int) {
	type tBinary struct {
		Architecture string `json:"architecture"`
		Name         string `json:"name"`
		Description  string `json:"description"`
	}

	// Fetch metadata
	var binaries []tBinary
	err := fetchJSON(RNMetadataURL, &binaries)
	if err != nil {
		fmt.Println("Failed to fetch and decode binary information:", err)
		return
	}

	// Filter binaries based on the search term and architecture
	searchResultsSet := make(map[string]struct{})
	for _, binary := range binaries {
		if strings.Contains(strings.ToLower(binary.Name), strings.ToLower(searchTerm)) || strings.Contains(strings.ToLower(binary.Description), strings.ToLower(searchTerm)) {
			ext := strings.ToLower(filepath.Ext(binary.Name))
			base := filepath.Base(binary.Name)
			if _, excluded := excludedFileTypes[ext]; excluded {
				continue // Skip this binary if its extension is excluded
			}
			if _, excludedName := excludedFileNames[base]; excludedName {
				continue // Skip this binary if its name is excluded
			}
			entry := fmt.Sprintf("%s - %s", binary.Name, binary.Description)
			searchResultsSet[entry] = struct{}{}
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
		} else {
			binaryPath, _ := exec.LookPath(name) // is it okay to ignore the err channel of LookPath?
			if binaryPath != "" {
				prefix = "[I]"
			} else if cachedLocation != "" && isExecutable(cachedLocation) {
				prefix = "[c]"
			} else {
				prefix = "[-]"
			}
		}

		truncatePrintf("%s %s - %s ", prefix, name, description)
		fmt.Printf("\n") // Escape sequences are truncated too...
	}
}
