// info.go // this file implements the functionality of 'info' //>
package main

import (
	"fmt"
	"github.com/goccy/go-json"
	"io/ioutil"
	"net/http"
)

// BinaryInfo represents the structure of binary information including description.
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

// getBinaryInfo fetches a binary's information from RNMetadataURL and RMetadataURL and returns it as a BinaryInfo struct.
func getBinaryInfo(binaryName string) (*BinaryInfo, error) {
	// Fetch a binary's details from RNMetadataURL
	response, err := http.Get(RNMetadataURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching metadata from %s: %v", RNMetadataURL, err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading metadata from %s: %v", RNMetadataURL, err)
	}

	var metadata [][]map[string]interface{}
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("error decoding metadata from %s: %v", RNMetadataURL, err)
	}

	var binInfo BinaryInfo
	var found bool

	for _, hostInfoArray := range metadata {
		for _, hostInfo := range hostInfoArray {
			if hostInfo["host"].(string) == validatedArch[2] {
				mainBins, ok := hostInfo["Main"].([]interface{})
				if !ok {
					return nil, fmt.Errorf("error decoding Main field from %s: %v", RNMetadataURL, err)
				}
				for _, bin := range mainBins {
					binMap, ok := bin.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("error decoding binary details from %s: %v", RNMetadataURL, err)
					}
					if binMap["Name"].(string) == binaryName {
						binInfo = BinaryInfo{
							Name:    binMap["Name"].(string),
							Size:    binMap["Size"].(string),
							ModTime: binMap["ModTime"].(string),
							Source:  binMap["Source"].(string),
							B3SUM:   binMap["B3SUM"].(string),
							SHA256:  binMap["SHA256"].(string),
							Repo:    binMap["Repo"].(string),
						}
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		}
		if found {
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("info for the requested binary ('%s') not found in the metadata.json file for architecture '%s'", binaryName, validatedArch[2])
	}

	// Fetch binary description from RMetadataURL
	response, err = http.Get(RMetadataURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching description from %s: %v", RMetadataURL, err)
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading description from %s: %v", RMetadataURL, err)
	}

	var descriptionMetadata map[string][]BinaryInfo
	if err := json.Unmarshal(body, &descriptionMetadata); err != nil {
		return nil, fmt.Errorf("error decoding description from %s: %v", RMetadataURL, err)
	}

	binaries, exists := descriptionMetadata["packages"]
	if !exists {
		return nil, fmt.Errorf("invalid description metadata format. No 'packages' field found.")
	}

	var description string
	for _, binInfo := range binaries {
		if binInfo.Name == binaryName {
			description = binInfo.Description
			break
		}
	}

	// Combine the technical details and description into a single BinaryInfo struct
	combinedInfo := BinaryInfo{
		Name:        binInfo.Name,
		Description: description,
		Repo:        binInfo.Repo,
		ModTime:     binInfo.ModTime,
		Size:        binInfo.Size,
		SHA256:      binInfo.SHA256,
		B3SUM:       binInfo.B3SUM,
		Source:      binInfo.Source,
	}

	return &combinedInfo, nil
}
