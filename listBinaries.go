// listBinaries.go // This file implements the listBinaries function //>
package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// listBinaries fetches and lists binary names from the given URLs.
func listBinaries() ([]string, error) {
	var metadata, allBinaries []struct {
		Name    string `json:"Name"`
		NameAlt string `json:"name"`
		SHA256  string `json:"sha256,omitempty"`
		SHA     string `json:"sha,omitempty"`
	}

	for _, url := range MetadataURLs {
		// Use fetchJSON to fetch and unmarshal the JSON data
		if err := fetchJSON(url, &metadata); err != nil {
			return nil, fmt.Errorf("failed to fetch metadata from %s: %v", url, err)
		}

		allBinaries = append(allBinaries, metadata...)
	}

	filteredBinaries := make(map[string]string)
	excludedFileTypes := map[string]bool{}

	for _, item := range allBinaries {
		binary := item.Name
		if binary == "" {
			binary = item.NameAlt
		}

		ext := strings.ToLower(filepath.Ext(binary))
		if _, excluded := excludedFileTypes[ext]; !excluded {
			filteredBinaries[binary] = binary
			/* // BROKEN. Will filter out binaries WITHOUT a SHA, and those with an EMPTY SHA.
			if item.SHA256 != "" {
							filteredBinaries[item.SHA256] = binary
						}
						if item.SHA != "" {
							filteredBinaries[item.SHA] = binary
						}
			*/
		}
	}

	uniqueBinaries := make([]string, 0, len(filteredBinaries))
	for binary := range filteredBinaries {
		uniqueBinaries = append(uniqueBinaries, binary)
	}

	return uniqueBinaries, nil
}
