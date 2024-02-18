// info.go // this file implements the functionality of 'info'

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// BinaryInfo represents the structure of binary information.
type BinaryInfo struct {
	Name        string `json:"Name"`
	Repo        string `json:"Repo"`
	Size        string `json:"Size"`
	SHA256      string `json:"SHA256"`
	B3SUM       string `json:"B3SUM"`
	Description string `json:"Description"`
}

// BinaryMetadata represents the structure of the metadata for a binary.
type BinaryMetadata struct {
	Binaries []BinaryInfo `json:"binaries"`
}

// showBinaryInfo fetches binary information from MetadataURLs and prints it.
func showBinaryInfo(binaryName string) {
	for i, metadataURL := range MetadataURLs {
		if i >= 2 { // TODO: Correctly unmarshal Github's REST API's "contents" endpoint.
			break
		}

		response, err := http.Get(metadataURL)
		if err != nil {
			fmt.Printf("Error fetching metadata from %s: %v\n", metadataURL, err)
			continue
		}
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("Error reading metadata from %s: %v\n", metadataURL, err)
			continue
		}

		var metadata []BinaryInfo
		if err := json.Unmarshal(body, &metadata); err != nil {
			fmt.Printf("Error decoding metadata from %s: %v\n", metadataURL, err)
			continue
		}

		for _, bin := range metadata {
			if bin.Name == binaryName {
				fmt.Printf("Binary Name: %s\n", bin.Name)
				if bin.Repo != "" {
					fmt.Printf("Repo: %s\n", bin.Repo)
				}
				if bin.Size != "" {
					fmt.Printf("Size: %s\n", bin.Size)
				}
				if bin.SHA256 != "" {
					fmt.Printf("SHA256: %s\n", bin.SHA256)
				}
				if bin.B3SUM != "" {
					fmt.Printf("B3SUM: %s\n", bin.B3SUM)
				}

				// Fetch the description from RMetadataURL
				response, err = http.Get(RMetadataURL)
				if err != nil {
					fmt.Printf("Error fetching description from %s: %v\n", RMetadataURL, err)
					return
				}
				defer response.Body.Close()

				body, err = ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Printf("Error reading description from %s: %v\n", RMetadataURL, err)
					return
				}

				// Unmarshal the description as a BinaryMetadata object
				var binaryMetadata BinaryMetadata
				if err := json.Unmarshal(body, &binaryMetadata); err != nil {
					fmt.Printf("Error decoding description from %s: %v\n", RMetadataURL, err)
					return
				}

				// Find the binary in the metadata and set the description
				for _, binInfo := range binaryMetadata.Binaries {
					if binInfo.Name == binaryName {
						bin.Description = binInfo.Description
						break
					}
				}

				if bin.Description != "" {
					fmt.Printf("Description: %s\n", bin.Description)
				}
				return
			}
		}
	}

	fmt.Printf("Info for the requested binary ('%s') not found in the metadata.json files. Please contribute to the repositories.\n", binaryName)
}
