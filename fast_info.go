// fast_info.go // this file implements the functionality of 'f_info' //>
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// fast_BinaryInfo represents the structure of binary information.
type fast_BinaryInfo struct {
	Description  string `json:"description"`
	Name         string `json:"name"`
	Architecture string `json:"architecture"`
	Version      string `json:"version"`
	Updated      string `json:"updated"`
	Size         string `json:"size"`
	SHA          string `json:"sha"`
	Source       string `json:"source"`
}

// fast_showBinaryInfo fetches a binary's information from RMetadataURL and prints it.
func fast_showBinaryInfo(binaryName string, validatedArch string) {
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

	var metadata map[string][]fast_BinaryInfo
	if err := json.Unmarshal(body, &metadata); err != nil {
		fmt.Printf("Error decoding metadata: %v\n", err)
		return
	}

	binaries, exists := metadata["packages"]
	if !exists {
		fmt.Println("Invalid metadata format. No fields found. Check fast_info.go:44:31")
		return
	}

	var found bool
	var nonMatchingArchCount int

	for _, bin := range binaries {
		if bin.Name == binaryName && bin.Architecture == validatedArch {
			// Print the binary information
			fmt.Printf("Name: %s\n", bin.Name)
			fmt.Printf("Description: %s\n", bin.Description)
			if bin.Version != "" {
				fmt.Printf("Version: %s\n", bin.Version)
			}
			if bin.Updated != "" {
				fmt.Printf("Updated: %s\n", bin.Updated)
			}
			if bin.Size != "" {
				fmt.Printf("Size: %s\n", bin.Size)
			}
			if bin.SHA != "" {
				fmt.Printf("SHA: %s\n", bin.SHA)
			}
			fmt.Printf("Source: %s\n", bin.Source)
			found = true
			break
		} else if bin.Name == binaryName {
			// Binary found but not for the system's architecture
			nonMatchingArchCount++
		}
	}

	// If binary not found for the system's architecture
	if !found {
		if nonMatchingArchCount > 0 {
			fmt.Printf("Binary '%s' found, but not for the system's architecture.\n", binaryName)
		} else {
			fmt.Printf("Info for the requested binary ('%s') not found in the metadata.json file. Please contribute to %s.\n", binaryName, RMetadataURL)
		}
	}
}
