// fsearch.go // this file implements the search option

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// fSearch searches for packages based on the given search term.
func fSearch(searchTerm string, desiredArch string) {
	// Fetch metadata
	response, err := http.Get(RMetadataURL)
	if err != nil {
		fmt.Println("Failed to fetch package information.")
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed to read metadata.")
		return
	}

	var metadata map[string][]PackageInfo
	if err := json.Unmarshal(body, &metadata); err != nil {
		fmt.Println("Failed to decode metadata.")
		return
	}

	// Filter packages based on the search term and architecture
	searchResultsSet := make(map[string]struct{}) // Use a set to keep track of unique entries
	for _, pkg := range metadata["packages"] {
		if strings.Contains(strings.ToLower(pkg.Name+pkg.Description), strings.ToLower(searchTerm)) {
			// Check if architecture matches the desired architecture
			if pkg.Architecture == desiredArch {
				entry := fmt.Sprintf("%s - %s", pkg.Name, pkg.Description)
				searchResultsSet[entry] = struct{}{} // Add the entry to the set
			}
		}
	}

	// Check if no matching packages found
	if len(searchResultsSet) == 0 {
		fmt.Printf("No matching packages found for '%s'.\n", searchTerm)
		return
	} else if len(searchResultsSet) > 90 {
		fmt.Printf("Too many matching packages found for '%s'.\n", searchTerm)
		return
	}

	// Convert set to slice for sorting
	var searchResults []string
	for entry := range searchResultsSet {
		searchResults = append(searchResults, entry)
	}

	// Sort the search results
	searchResults = sortBinaries(searchResults)

	// Determine the truncation length
	getTerminalWidth := func() int {
		cmd := exec.Command("tput", "cols")
		cmd.Stdin = os.Stdin
		out, err := cmd.Output()
		if err != nil {
			return 80 // Default to 80 columns if unable to get terminal width
		}
		width, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err != nil {
			return 80 // Default to 80 columns if unable to convert width to integer
		}
		return width
	}

	// Check if the package binary exists in the INSTALL_DIR and print results with installation state indicators
	for _, line := range searchResults {
		parts := strings.SplitN(line, " - ", 2)
		name := parts[0]
		description := parts[1]

		// Use INSTALL_DIR or fallback to $HOME/.local/bin
		installDir := os.Getenv("INSTALL_DIR")
		if installDir == "" {
			installDir = filepath.Join(os.Getenv("HOME"), ".local", "bin")
		}
		installDirLocation := filepath.Join(installDir, name)

		cachedLocation, _ := ReturnCachedFile(name)

		var prefix string
		if fileExists(installDirLocation) {
			prefix = "[i]"
		} else if cachedLocation != "" && isExecutable(cachedLocation) {
			prefix = "[c]"
		} else {
			prefix = "[-]"
		}

		// Calculate available space for description
		availableSpace := getTerminalWidth() - len(prefix) - len(name) - 4 // 4 accounts for space around ' - '

		// Truncate the description if it exceeds the available space
		if len(description) > availableSpace {
			description = fmt.Sprintf("%s...", description[:availableSpace-3]) // Shrink to the maximum line size
		}

		fmt.Printf("%s %s - %s\n", prefix, name, description)
	}
}
