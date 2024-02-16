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

///// YOU MAY CHANGE THESE TO POINT TO ANOTHER PLACE.

const (
	// Cache size limit & handling.
	MaxCacheSize     = 10
	BinariesToDelete = 5
	// TMPDIR is the directory for storing temporary files.
	TEMP_DIR = "/tmp/bigdl_cached"
	// CACHE_FILE is the file path for caching installation information.
	CACHE_FILE = TEMP_DIR + "/bigdl_cache.log"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: bigdl {list|install|remove|run|info|search|tldr} [args...]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "find_url":
		if len(os.Args) != 3 {
			fmt.Println("Usage: bigdl find_url <binary>")
			os.Exit(1)
		}
		findURLCommand(os.Args[2])
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
	case "return_cached_file":
		if len(os.Args) < 2 {
			fmt.Println("Usage: bigdl return_cached_file <binary>")
			os.Exit(1)
		}
		fmt.Println(ReturnCachedFile(os.Args[2]))
	case "list":
		listBinaries()
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
		RunFromCache("tlrc", os.Args[2:]) // Rust version of tldr.sh (its called tldr on the repo.) | I'd like to use something lighter tho.
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
	case "remove":
		if len(os.Args) != 3 {
			fmt.Println("Usage: bigdl remove <binary>")
			os.Exit(1)
		}
		binaryToRemove := os.Args[2]
		remove(binaryToRemove)
	default:
		fmt.Printf("bigdl: Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
