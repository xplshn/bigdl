// findURL.go // This file implements the findURL function //>
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// BinaryInfo struct holds binary metadata used in main.go for the `info`, `update`, `list` functionality
type BinaryInfo struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Repo        string `json:"Repo"`
	ModTime     string `json:"ModTime"`
	Version     string `json:"Version"`
	Updated     string `json:"Updated"`
	Size        string `json:"Size"`
	SHA256      string `json:"SHA256"`
	B3SUM       string `json:"B3SUM"`
	Source      string `json:"Source"`
}

func fetchJSON(url string, v interface{}) error {
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error fetching from %s: %v", url, err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading from %s: %v", url, err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("error decoding from %s: %v", url, err)
	}

	return nil
}

func findBinaryInfo(metadata [][]map[string]interface{}, binaryName string) (BinaryInfo, bool) {
	for _, hostInfoArray := range metadata {
		for _, hostInfo := range hostInfoArray {
			if hostInfo["host"].(string) == ValidatedArch[2] {
				mainBins, ok := hostInfo["Main"].([]interface{})
				if !ok {
					continue
				}
				for _, bin := range mainBins {
					binMap, ok := bin.(map[string]interface{})
					if !ok {
						continue
					}
					if binMap["Name"].(string) == binaryName {
						return BinaryInfo{
							Name:    binMap["Name"].(string),
							Size:    binMap["Size"].(string),
							ModTime: binMap["ModTime"].(string),
							Source:  binMap["Source"].(string),
							B3SUM:   binMap["B3SUM"].(string),
							SHA256:  binMap["SHA256"].(string),
							Repo:    binMap["Repo"].(string),
						}, true
					}
				}
			}
		}
	}
	return BinaryInfo{}, false
}

func getBinaryInfo(binaryName string) (*BinaryInfo, error) {
	var metadata [][]map[string]interface{}
	if err := fetchJSON(RNMetadataURL, &metadata); err != nil {
		return nil, err
	}

	binInfo, found := findBinaryInfo(metadata, binaryName)
	if !found {
		return nil, fmt.Errorf("info for the requested binary ('%s') not found in the metadata.json file for architecture '%s'", binaryName, ValidatedArch[2])
	}

	var rMetadata map[string][]BinaryInfo
	if err := fetchJSON(RMetadataURL, &rMetadata); err != nil {
		return nil, err
	}

	binaries, exists := rMetadata["packages"]
	if !exists {
		return nil, fmt.Errorf("invalid description metadata format. No 'packages' field found")
	}

	var description, updated, version string
	for _, binInfo := range binaries {
		if binInfo.Name == binaryName {
			description = binInfo.Description
			updated = binInfo.Updated
			version = binInfo.Version
			break
		}
	}

	combinedInfo := BinaryInfo{
		Name:        binInfo.Name,
		Description: description,
		Repo:        binInfo.Repo,
		ModTime:     binInfo.ModTime,
		Version:     version,
		Updated:     updated,
		Size:        binInfo.Size,
		SHA256:      binInfo.SHA256,
		B3SUM:       binInfo.B3SUM,
		Source:      binInfo.Source,
	}

	return &combinedInfo, nil
}
