// info.go // This file implements binInfo, which `info` and `update` use //>
package main

import (
	"fmt"
)

// BinaryInfo struct holds binary metadata used in main.go for the `info`, `update`, `list` functionality
type BinaryInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Repo        string `json:"repo_url"`
	ModTime     string `json:"build_date"`
	Version     string `json:"repo_version"`
	Updated     string `json:"repo_updated"`
	Size        string `json:"size"`
	SHA256      string `json:"sha256"`
	Source      string `json:"download_url"`
}

func findBinaryInfo(metadata []map[string]interface{}, binaryName string) (BinaryInfo, bool) {
	for _, binMap := range metadata {
		if binMap["name"].(string) == binaryName {
			return BinaryInfo{
				Name:        binMap["name"].(string),
				Description: binMap["description"].(string),
				Repo:        binMap["repo_url"].(string),
				ModTime:     binMap["build_date"].(string),
				Version:     binMap["repo_version"].(string),
				Updated:     binMap["repo_updated"].(string),
				Size:        binMap["size"].(string),
				SHA256:      binMap["sha256"].(string),
				Source:      binMap["download_url"].(string),
			}, true
		}
	}
	return BinaryInfo{}, false
}

func getBinaryInfo(binaryName string) (*BinaryInfo, error) {
	var metadata []map[string]interface{}
	if err := fetchJSON(RNMetadataURL, &metadata); err != nil {
		return nil, err
	}

	binInfo, found := findBinaryInfo(metadata, binaryName)
	if !found {
		return nil, fmt.Errorf("info for the requested binary ('%s') not found in the metadata.json file", binaryName)
	}

	return &binInfo, nil
}
