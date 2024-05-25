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
		if name, ok := binMap["name"].(string); ok && name == binaryName {
			description, _ := binMap["description"].(string)
			repo, _ := binMap["repo_url"].(string)
			build_date, _ := binMap["build_date"].(string)
			version, _ := binMap["repo_version"].(string)
			updated, _ := binMap["repo_updated"].(string)
			size, _ := binMap["size"].(string)
			sha256, _ := binMap["sha256"].(string)
			source, _ := binMap["download_url"].(string)

			return BinaryInfo{
				Name:        name,
				Description: description,
				Repo:        repo,
				ModTime:     build_date,
				Version:     version,
				Updated:     updated,
				Size:        size,
				SHA256:      sha256,
				Source:      source,
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
		return nil, fmt.Errorf("error: info for the requested binary ('%s') not found in the metadata.json file", binaryName)
	}

	return &binInfo, nil
}
