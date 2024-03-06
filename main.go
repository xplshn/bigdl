// main.go // This is the main entrypoint, which calls all the different functions //>
package main

import (
	"flag"
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

const (
	RMetadataURL  = "https://raw.githubusercontent.com/Azathothas/Toolpacks/main/metadata.json" // This is the file from which we extract descriptions for different binaries
	RNMetadataURL = "https://bin.ajam.dev/METADATA.json"                                        // This is the file which contains a concatenation of all metadata in the different repos, this one also contains sha256 checksums.
	VERSION       = "1.3.1"
	usagePage     = " [-vh] [list|install|remove|update|run|info|search|tldr] <args>"
	// Truncation indicator
	indicator = "...>"
	// Cache size limit & handling.
	MaxCacheSize     = 10
	BinariesToDelete = 5
	TEMP_DIR         = "/tmp/bigdl_cached"
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

func printHelp() {
	helpMessage := "Usage:\n" + usagePage + `
	
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
 bigdl search editor
 bigdl install micro
 bigdl remove bed
 bigdl info jq
 bigdl tldr gum
 bigdl run --verbose curl -qsfSL "https://raw.githubusercontent.com/xplshn/bigdl/master/stubdl" | sh -
 bigdl run --silent elinks -no-home "https://fatbuffalo.neocities.org/def"
 bigdl run btop

Version: ` + VERSION

	fmt.Println(helpMessage)
}

func main() {

	errorOutInsufficientArgs := func() { os.Exit(errorEncoder("Error: Insufficient parameters")) }

	version := flag.Bool("v", false, "Show the version number")
	help := flag.Bool("h", false, "Show this help message")
	flag.Parse()

	if *version {
		fmt.Println("bigdl", VERSION)
		os.Exit(0)
	}

	if *help {
		printHelp()
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		fmt.Printf(" bigdl:%s\n", usagePage)
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "find_url":
		binaryName := flag.Arg(1)
		if binaryName == "" {
			fmt.Println("Usage: bigdl find_url [binary]")
			errorOutInsufficientArgs()
		}
		findURLCommand(binaryName)
	case "list":
		listBinaries()
	case "install", "add":
		// Assuming install requires a binary name and optionally an install directory and message
		binaryName := flag.Arg(1)
		if binaryName == "" {
			fmt.Printf("Usage: bigdl %s [binary] <install_dir> <install_message>\n", flag.Arg(0))
			errorOutInsufficientArgs()
		}
		var installDir, installMessage string
		if flag.NArg() > 2 {
			installDir = flag.Arg(2)
		}
		if flag.NArg() > 3 {
			installMessage = flag.Arg(3)
		}
		installCommand(binaryName, installDir, installMessage)
	case "remove", "del":
		if flag.NArg() < 2 {
			fmt.Printf("Usage: bigdl %s [binar|y|ies]\n", flag.Arg(0))
			errorOutInsufficientArgs()
		}
		remove(flag.Args()[1:])
	case "run":
		if flag.NArg() < 2 {
			fmt.Println("Usage: bigdl run <--verbose, --silent, --transparent> [binary] <args>")
			errorOutInsufficientArgs()
		}
		RunFromCache(flag.Arg(1), flag.Args()[2:])
	case "tldr":
		if flag.NArg() < 2 {
			fmt.Println("Usage: bigdl tldr <args> [page]")
			errorOutInsufficientArgs()
		}
		RunFromCache("tlrc", flag.Args()[1:])
	case "info":
		binaryName := flag.Arg(1)
		if binaryName == "" {
			fmt.Println("Usage: bigdl info [binary]")
			errorOutInsufficientArgs()
		}
		getBinaryInfo(binaryName)
	case "search":
		query := flag.Arg(1)
		if query == "" {
			fmt.Println("Usage: bigdl search [query]")
			errorOutInsufficientArgs()
		}
		fSearch(query)
	case "update":
		if flag.NArg() < 2 {
			fmt.Println("Usage: bigdl update [binar|y|ies]")
			errorOutInsufficientArgs()
		}
		programsToUpdate := flag.Args()[1:]
		update(programsToUpdate)
	default:
		errorOut("bigdl: Unknown command: %s\n", flag.Arg(0))
	}
}
