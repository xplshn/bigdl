// info.go // this file implements the functionality of 'info' //>
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
	Description string `json:"Description"`
	Repo        string `json:"Repo"`
	ModTime     string `json:"ModTime"`
	Size        string `json:"Size"`
	SHA256      string `json:"SHA256"`
	B3SUM       string `json:"B3SUM"`
	Source      string `json:"Source"`
}

// BinaryMetadata represents the structure of the metadata for a binary.
type BinaryMetadata struct {
	Binaries []BinaryInfo `json:"binaries"`
}

// getBinaryInfo fetches binary information from MetadataURLs and returns it as a BinaryInfo struct.
func getBinaryInfo(binaryName string) (*BinaryInfo, error) {
	for i, metadataURL := range MetadataURLs {
		if i >= 2 { // TODO: Correctly unmarshal Github's REST API's "contents" endpoint. In order not to do this ugly thing.
			break
		}
		response, err := http.Get(metadataURL)
		if err != nil {
			return nil, fmt.Errorf("error fetching metadata from %s: %v", metadataURL, err)
		}
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading metadata from %s: %v", metadataURL, err)
		}

		var metadata []BinaryInfo
		if err := json.Unmarshal(body, &metadata); err != nil {
			return nil, fmt.Errorf("error decoding metadata from %s: %v", metadataURL, err)
		}

		for _, bin := range metadata {
			if bin.Name == binaryName {
				// Fetch the description from RMetadataURL
				response, err = http.Get(RMetadataURL)
				if err != nil {
					return nil, fmt.Errorf("error fetching description from %s: %v", RMetadataURL, err)
				}
				defer response.Body.Close()

				body, err = ioutil.ReadAll(response.Body)
				if err != nil {
					return nil, fmt.Errorf("error reading description from %s: %v", RMetadataURL, err)
				}

				// Unmarshal the description as a BinaryMetadata object
				var binaryMetadata BinaryMetadata
				if err := json.Unmarshal(body, &binaryMetadata); err != nil {
					return nil, fmt.Errorf("error decoding description from %s: %v", RMetadataURL, err)
				}

				// Find the binary in the metadata and set the description
				for _, binInfo := range binaryMetadata.Binaries {
					if binInfo.Name == binaryName {
						bin.Description = binInfo.Description
						break
					}
				}

				return &bin, nil
			}
		}
	}

	return nil, fmt.Errorf("info for the requested binary ('%s') not found in the metadata.json files", binaryName)
}
