// main.go // This is the main entrypoint, which calls all the different functions //>
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	Repositories    []string
	MetadataURLs    []string
	validatedArch   = [3]string{}
	InstallDir      = os.Getenv("INSTALL_DIR")
	installUseCache = true
	useProgressBar  = true
)

func init() {
	if InstallDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get user's Home directory. %v\n", err)
			os.Exit(1)
		}
		InstallDir = filepath.Join(homeDir, ".local", "bin")
	}
	if err := os.MkdirAll(InstallDir, os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get user's Home directory. %v\n", err)
		os.Exit(1)
	}
	switch runtime.GOARCH {
	case "amd64":
		validatedArch = [3]string{"x86_64_Linux", "x86_64", "x86_64-Linux"}
	case "arm64":
		validatedArch = [3]string{"aarch64_arm64_Linux", "aarch64_arm64", "aarch64-Linux"}
	default:
		fmt.Println("Unsupported architecture:", runtime.GOARCH)
		os.Exit(1)
	}
	arch := validatedArch[0]
	Repositories = append(Repositories, "https://bin.ajam.dev/"+arch+"/")
	Repositories = append(Repositories, "https://bin.ajam.dev/"+arch+"/Baseutils/")
	Repositories = append(Repositories, "https://raw.githubusercontent.com/xplshn/Handyscripts/master/")
	// These are used for listing the binaries themselves
	MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/"+arch+"/METADATA.json")
	MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/"+arch+"/Baseutils/METADATA.json")
	MetadataURLs = append(MetadataURLs, "https://api.github.com/repos/xplshn/Handyscripts/contents") // You may add other repos if need be? bigdl is customizable, feel free to open a PR, ask questions, etc.
}

const (
	RMetadataURL  = "https://raw.githubusercontent.com/Azathothas/Toolpacks/main/metadata.json" // This is the file from which we extract descriptions for different binaries
	RNMetadataURL = "https://bin.ajam.dev/METADATA.json" // This is the file which contains a concatenation of all metadata in the different repos, this one also contains sha256 checksums.
	VERSION       = "1.3.1"
	usagePage     = "Usage: bigdl [-vh] [list|install|remove|update|run|info|search|tldr] <args>"
	// Truncation indicator
	indicator = "...>"
	// Cache size limit & handling.
	MaxCacheSize     = 10
	BinariesToDelete = 5
	TEMP_DIR         = "/tmp/bigdl_cached"
)

func printHelp() {
	helpMessage := usagePage + `
	
Options:
 -h, --help     Show this help message
 -v, --version Show the version number

Commands:
 list           List all available binaries
 install, add   Install a binary
 remove, del    Remove a binary
 update         Update binaries, by checking their SHA against the repo's SHA.
 run            Run a binary
 info           Show information about a specific binary
 search         Search for a binary - (not all binaries have metadata. Use list to see all binaries)
 tldr           Show a brief description & usage examples for a given program/command

Examples:
 bigdl install micro
 bigdl remove bed
 bigdl info jq
 bigdl search editor
 bigdl tldr gum
 bigdl run --verbose neofetch
 bigdl run --silent micro
 bigdl run btop

Version: ` + VERSION

	fmt.Println(helpMessage)
}

func main() {
	// Check for flags directly in the main function
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Println("bigdl", VERSION)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	// If no arguments are received, show the usage text
	if len(os.Args) < 2 {
		fmt.Printf("%s\n", usagePage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "find_url":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bigdl find_url [binary]")
			os.Exit(1)
		}
		findURLCommand(os.Args[2])
	case "list":
		binaries, err := listBinaries()
		if err != nil {
			fmt.Println("Error listing binaries:", err)
			os.Exit(1)
		}
		for _, binary := range binaries {
			fmt.Println(binary)
		}
	case "install", "add":
		if len(os.Args) < 3 {
			fmt.Printf("Usage: bigdl %s [binary] <install_dir> <install_message>\n", os.Args[1])
			os.Exit(1)
		}
		binaryName := os.Args[2]
		var installMessage string
		if len(os.Args) > 3 {
			InstallDir = os.Args[3]
		}
		if len(os.Args) > 4 {
			installMessage = os.Args[4]
		}
		err := installCommand(binaryName, installMessage)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			os.Exit(1)
		}
	case "remove", "del":
		if len(os.Args) < 3 {
			fmt.Printf("Usage: bigdl %s [binary]\n", os.Args[1])
			os.Exit(1)
		}
		remove(os.Args[2:])
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bigdl run <--verbose, --silent> [binary] <args>")
			os.Exit(1)
		}
		RunFromCache(os.Args[2], os.Args[3:])
	case "tldr":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bigdl tldr <args> [page]")
			os.Exit(1)
		}
		RunFromCache("tlrc", os.Args[2:])
	case "info":
		if len(os.Args) != 3 {
			fmt.Println("Usage: bigdl info [binary]")
			os.Exit(1)
		}
		binaryName := os.Args[2]
		binaryInfo, err := getBinaryInfo(binaryName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Print the fields
		fmt.Printf("Name: %s\n", binaryInfo.Name)
		if binaryInfo.Description != "" {
			fmt.Printf("Description: %s\n", binaryInfo.Description)
		}
		if binaryInfo.Repo != "" {
			fmt.Printf("Repo: %s\n", binaryInfo.Repo)
		}
		if binaryInfo.Size != "" {
			fmt.Printf("Size: %s\n", binaryInfo.Size)
		}
		if binaryInfo.SHA256 != "" {
			fmt.Printf("SHA256: %s\n", binaryInfo.SHA256)
		}
		if binaryInfo.B3SUM != "" {
			fmt.Printf("B3SUM: %s\n", binaryInfo.B3SUM)
		}
		if binaryInfo.Source != "" {
			fmt.Printf("Source: %s\n", binaryInfo.Source)
		}
	case "search":
		if len(os.Args) != 3 {
			fmt.Println("Usage: bigdl search [query]")
			os.Exit(1)
		}
		searchTerm := os.Args[2]
		fSearch(searchTerm)
	case "update":
		var programsToUpdate []string
		if len(os.Args) > 2 {
			// Bulk update with list of programs to update
			programsToUpdate = os.Args[2:]
		}
		if err := update(programsToUpdate); err != nil {
			fmt.Printf("Error updating programs: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("bigdl: Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
