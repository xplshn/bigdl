// findURL.go // This file implements the findURL function //>
package main

import (
	"fmt"
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

func findBinaryInfo(metadata [][]map[string]interface{}, binaryName string) (BinaryInfo, bool) {
	for _, hostInfoArray := range metadata {
		for _, hostInfo := range hostInfoArray {
			//			if hostInfo["host"].(string) == ValidatedArch[0] {
			mainBins, ok := hostInfo["Main"].([]interface{})
			if !ok {
				continue
			}
			for _, bin := range mainBins {
				binMap, ok := bin.(map[string]interface{})
				if !ok {
					continue
				}
				if binMap["name"].(string) == binaryName {
					return BinaryInfo{
						Name:    binMap["name"].(string),
						Size:    binMap["size"].(string),
						ModTime: binMap["build_date"].(string),
						Source:  binMap["download_url"].(string),
						B3SUM:   binMap["b3sum"].(string),
						SHA256:  binMap["sha256"].(string),
						Repo:    binMap["repo_url"].(string),
					}, true
				}
			}
			//			}
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
		return nil, fmt.Errorf("info for the requested binary ('%s') not found in the metadata.json file for architecture '%s'", binaryName, ValidatedArch[0])
	}

	combinedInfo := BinaryInfo{
		Name:        binInfo.Name,
		Description: binInfo.Description,
		Repo:        binInfo.Repo,
		ModTime:     binInfo.ModTime,
		Version:     binInfo.Version,
		Updated:     binInfo.Updated,
		Size:        binInfo.Size,
		SHA256:      binInfo.SHA256,
		B3SUM:       binInfo.B3SUM,
		Source:      binInfo.Source,
	}

	return &combinedInfo, nil
}
