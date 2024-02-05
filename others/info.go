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
	Description string `json:"description"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Updated     string `json:"updated"`
	Size        string `json:"size"`
	SHA         string `json:"sha"`
	Source      string `json:"source"`
}

// showPackageInfo fetches package information from RMetadataURL and prints it.
func showPackageInfo(packageName string) {
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

	var packageInfo PackageInfo
	found := false

	for _, pkg := range packages {
		if pkg.Name == packageName {
			packageInfo = pkg
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Info for package '%s' not found in the metadata.json file. Please contribute to %s.\n", packageName, RMetadataURL)
		return
	}

	// Print the package information
	fmt.Printf("Package Name: %s\n", packageInfo.Name)
	fmt.Printf("Description: %s\n", packageInfo.Description)
	fmt.Printf("Version: %s\n", packageInfo.Version)
	fmt.Printf("Updated: %s\n", packageInfo.Updated)
	fmt.Printf("Size: %s\n", packageInfo.Size)
	fmt.Printf("SHA: %s\n", packageInfo.SHA)
	fmt.Printf("Source: %s\n", packageInfo.Source)
}
