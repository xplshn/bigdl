// main.go // This is the main entrypoint, which calls all the different functions //

package main

import (
	"fmt"
	"os"
	"runtime"
)

// Repositories contains the URLs for fetching metadata.
var Repositories []string

// MetadataURLs contains the URLs for fetching metadata.
var MetadataURLs []string

// Array for storing a variable that fsearch and info use.
var validatedArch = [2]string{}

func init() {
	switch runtime.GOARCH {
	case "amd64":
		Repositories = append(Repositories, "https://bin.ajam.dev/x86_64_Linux/")
		Repositories = append(Repositories, "https://raw.githubusercontent.com/xplshn/Handyscripts/master/")
		Repositories = append(Repositories, "https://bin.ajam.dev/x86_64_Linux/Baseutils/")
		MetadataURLs = append(MetadataURLs, "https://api.github.com/repos/xplshn/Handyscripts/contents")
		MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/x86_64_Linux/METADATA.json")
		MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/x86_64_Linux/Baseutils/METADATA.json")
		validatedArch = [2]string{"x86_64_Linux", "x86_64"}
	case "arm64":
		Repositories = append(Repositories, "https://bin.ajam.dev/aarch64_arm64_Linux/")
		Repositories = append(Repositories, "https://raw.githubusercontent.com/xplshn/Handyscripts/master/")
		Repositories = append(Repositories, "https://bin.ajam.dev/aarch64_arm64_Linux/Baseutils/")
		MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/aarch64_arm64_Linux/METADATA.json")
		MetadataURLs = append(MetadataURLs, "https://api.github.com/repos/xplshn/Handyscripts/contents")
		MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/aarch64_arm64_Linux/Baseutils/METADATA.json")
		validatedArch = [2]string{"aarch64_arm64_Linux", "aarch64_arm64"}
	default:
		fmt.Println("Unsupported architecture:", runtime.GOARCH)
		os.Exit(1)
	}
}

const RMetadataURL = "https://raw.githubusercontent.com/metis-os/hysp-pkgs/main/data/metadata.json"
const VERSION = "1.1"

///// YOU MAY CHANGE THESE TO POINT TO ANOTHER PLACE.

const (
	// Cache size limit & handling.
	MaxCacheSize     = 10
	BinariesToDelete = 5
	// TMPDIR is the directory for storing temporary files.
	TEMP_DIR = "/tmp/bigdl_cached"
)

func printHelp() {
	fmt.Println("Usage: bigdl [-vh] {list|install|remove|run|info|search|tldr} [args...]")
	fmt.Println("\nOptions:")
	fmt.Println("  -h, --help    Show this help message")
	fmt.Println("  -v, --version Show the version number")
	fmt.Println("\nCommands:")
	fmt.Println("  list          List all available binaries")
	fmt.Println("  install       Install a binary")
	fmt.Println("  remove        Remove a binary")
	fmt.Println("  run           Run a binary")
	fmt.Println("  info          Show information about a package")
	fmt.Println("  search        Search for a binary")
	fmt.Println("  tldr          Show a brief description & usage examples for a given program/command")
	fmt.Println("\nExamples:")
	fmt.Println("  bigdl install micro")
	fmt.Println("  bigdl remove bed")
	fmt.Println("  bigdl info jq")
	fmt.Println("  bigdl search fzf")
	fmt.Println("  bigdl tldr gum")
	fmt.Println("\nVersion:", VERSION)
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
		fmt.Println("Usage: bigdl [-vh] {list|install|remove|run|info|search|tldr} [args...]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "find_url":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bigdl find_url <binary>")
			os.Exit(1)
		}
		findURLCommand(os.Args[2])
	case "list":
		listBinaries()
	case "install":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bigdl install <binary> [install_dir] [install_message]")
			os.Exit(1)
		}
		binaryName := os.Args[2]
		var installDir, installMessage string
		if len(os.Args) > 3 {
			installDir = os.Args[3]
		}
		if len(os.Args) > 4 {
			installMessage = os.Args[4]
		}
		installCommand(binaryName, []string{installDir, installMessage})
	case "remove":
		if len(os.Args) != 3 {
			fmt.Println("Usage: bigdl remove <binary>")
			os.Exit(1)
		}
		binaryToRemove := os.Args[2]
		remove(binaryToRemove)
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bigdl run [--verbose] <binary> [args...]")
			os.Exit(1)
		}
		RunFromCache(os.Args[2], os.Args[3:])
	case "tldr":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bigdl tldr <page> [args...]")
			os.Exit(1)
		}
		RunFromCache("tlrc", os.Args[2:])
	case "info":
		if len(os.Args) != 3 {
			fmt.Println("Usage: bigdl info <package-name>")
			os.Exit(1)
		}
		packageName := os.Args[2]
		showPackageInfo(packageName, validatedArch[1])
	case "search":
		if len(os.Args) != 3 {
			fmt.Println("Usage: bigdl search <search-term>")
			os.Exit(1)
		}
		searchTerm := os.Args[2]
		fSearch(searchTerm, validatedArch[1])
	default:
		fmt.Printf("bigdl: Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
