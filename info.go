// info.go // this file implements the functionality of 'info'

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// PackageInfo represents the structure of package information.
type PackageInfo struct {
	Description  string `json:"description"`
	Name         string `json:"name"`
	Architecture string `json:"architecture"`
	Version      string `json:"version"`
	Updated      string `json:"updated"`
	Size         string `json:"size"`
	SHA          string `json:"sha"`
	Source       string `json:"source"`
}

// showPackageInfo fetches package information from RMetadataURL and prints it.
func showPackageInfo(packageName string, validatedArch string) {
	response, err := http.Get(RMetadataURL)
	if err != nil {
		fmt.Printf("Error fetching metadata: %v\n", err)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading metadata: %v\n", err)
		return
	}

	var metadata map[string][]PackageInfo
	if err := json.Unmarshal(body, &metadata); err != nil {
		fmt.Printf("Error decoding metadata: %v\n", err)
		return
	}

	packages, exists := metadata["packages"]
	if !exists {
		fmt.Println("Invalid metadata format. 'packages' field not found.")
		return
	}

	var found bool
	var nonMatchingArchCount int

	for _, pkg := range packages {
		if pkg.Name == packageName && pkg.Architecture == validatedArch {
			// Print the package information
			fmt.Printf("Package Name: %s\n", pkg.Name)
			fmt.Printf("Description: %s\n", pkg.Description)
			if pkg.Version != "" {
				fmt.Printf("Version: %s\n", pkg.Version)
			}
			if pkg.Updated != "" {
				fmt.Printf("Updated: %s\n", pkg.Updated)
			}
			if pkg.Size != "" {
				fmt.Printf("Size: %s\n", pkg.Size)
			}
			if pkg.SHA != "" {
				fmt.Printf("SHA: %s\n", pkg.SHA)
			}
			fmt.Printf("Source: %s\n", pkg.Source)
			found = true
			break
		} else if pkg.Name == packageName {
			// Package found but not for the system's architecture
			nonMatchingArchCount++
		}
	}

	// If package not found for the system's architecture
	if !found {
		if nonMatchingArchCount > 0 {
			fmt.Printf("Package '%s' found, but not for the system's architecture.\n", packageName)
		} else {
			fmt.Printf("Info for the requested binary ('%s') not found in the metadata.json file. Please contribute to %s.\n", packageName, RMetadataURL)
		}
	}
}
