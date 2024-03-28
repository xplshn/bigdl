// main.go // This is the main entrypoint, which calls all the different functions //>
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

var (
	Repositories      []string
	MetadataURLs      []string
	validatedArch     = [3]string{}
	InstallDir        = os.Getenv("INSTALL_DIR")
	installUseCache   = true
	useProgressBar    = true
	disableTruncation = false
)

const (
	RMetadataURL  = "https://raw.githubusercontent.com/metis-os/hysp-pkgs/main/data/metadata.json" // This is the file from which we extract descriptions for different binaries //unreliable mirror: "https://raw.githubusercontent.com/Azathothas/Toolpacks/main/metadata.json"
	RNMetadataURL = "https://bin.ajam.dev/METADATA.json"                                           // This is the file which contains a concatenation of all metadata in the different repos, this one also contains sha256 checksums.
	VERSION       = "1.5.1"
	usagePage     = " [-v|-h] [list|install|remove|update|run|info|search|tldr] <{args}>"
	// Truncation indicator
	indicator = "...>"
	// Cache size limit & handling.
	MaxCacheSize     = 10
	BinariesToDelete = 5 // Once the cache is filled - The programs populate the list of binaries to be removed in order of least used.
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

	if os.Getenv("DISABLE_TRUNCATION") == "true" || os.Getenv("DISABLE_TRUNCATION") == "1" {
		disableTruncation = true
	}
	if os.Getenv("DISABLE_PRBAR") == "true" || os.Getenv("DISABLE_PRBAR") == "1" {
		useProgressBar = false
	}
}

func printHelp() {
	helpMessage := "Usage:\n" + usagePage + `

Options:
 -h, --help       Show this help message
 -v, --version    Show the version number

Commands:
 list             List all available binaries
 install, add     Install a binary
 remove, del      Remove a binary
 update           Update binaries, by checking their SHA against the repo's SHA
 run              Run a binary
 info             Show information about a specific binary
 search           Search for a binary - (not all binaries have metadata. Use list to see all binaries)
 tldr             Show a brief description & usage examples for a given program/command. This is an alias equivalent to using "run" with "tlrc" as argument.

Examples:
 bigdl search editor
 bigdl install micro
 bigdl install lux --fancy "%s was installed to $INSTALL_DIR." --newline
 bigdl install bed --fancy --truncate "%s was installed to $INSTALL_DIR." --newline
 bigdl install orbiton --truncate "installed Orbiton to $INSTALL_DIR."
 bigdl remove bed
 bigdl remove orbiton tgpt lux
 bigdl info jq
 bigdl tldr gum
 bigdl run --verbose curl -qsfSL "https://raw.githubusercontent.com/xplshn/bigdl/master/stubdl" | sh -
 bigdl run --silent elinks -no-home "https://fatbuffalo.neocities.org/def"
 bigdl run --transparent --silent micro ~/.profile
 bigdl run btop

Version: ` + VERSION

	fmt.Println(helpMessage)
}

func main() {

	errorOutInsufficientArgs := func() { os.Exit(errorEncoder("Error: Insufficient parameters\n")) }
	version := flag.Bool("v", false, "Show the version number")
	versionLong := flag.Bool("version", false, "Show the version number")

	flag.Usage = printHelp
	flag.Parse()

	if *version || *versionLong {
		fmt.Println("bigdl", VERSION)
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
		if len(os.Args) == 3 {
			if os.Args[2] == "--described" || os.Args[2] == "-d" {
				fSearch("", 99999) // Call fSearch with an empty query and a large limit to list all described binaries
			} else {
				errorOut("bigdl: Unknown command.\n")
			}
		} else {
			binaries, err := listBinaries()
			if err != nil {
				fmt.Println("Error listing binaries:", err)
				os.Exit(1)
			}
			for _, binary := range binaries {
				fmt.Println(binary)
			}
		}
	case "install", "add":
		// Check if the binary name is provided
		if flag.NArg() < 2 {
			fmt.Printf("Usage: bigdl %s [binary] <install_message>\n", flag.Arg(0))
			fmt.Println("Options:")
			fmt.Println(" --fancy <--truncate> : Will replace exactly ONE '%s' with the name of the requested binary in the install message <--newline>")
			fmt.Println(" --truncate: Truncates the message to fit into the terminal")
			errorOutInsufficientArgs()
		}

		binaryName := os.Args[2]
		installMessage := os.Args[3:]

		installCommand(binaryName, installMessage...)
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

		if len(os.Args) > 3 {
			fmt.Fprintln(os.Stderr, "Warning: The command contains more arguments than expected. Part of the input unused.")
		}

		if len(os.Args) < 3 {
			installedPrograms, err := validateProgramsFrom(InstallDir, nil)
			if err != nil {
				fmt.Println("Error validating programs:", err)
				return
			}
			for _, program := range installedPrograms {
				fmt.Println(program)
			}
		} else {

			binaryInfo, err := getBinaryInfo(binaryName)
			if err != nil {
				errorOut("%v\n", err)
			}
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
		}
	case "search":
		limit := 90
		queryIndex := 2

		if len(os.Args) < queryIndex+1 {
			fmt.Println("Usage: bigdl search <--limit||-l [int]> [query]")
			os.Exit(1)
		}

		if len(os.Args) > 2 && os.Args[queryIndex] == "--limit" || os.Args[queryIndex] == "-l" {
			if len(os.Args) > queryIndex+1 {
				var err error
				limit, err = strconv.Atoi(os.Args[queryIndex+1])
				if err != nil {
					errorOut("Error: 'limit' value is not an int.\n")
				}
				queryIndex += 2
			} else {
				errorOut("Error: Missing 'limit' value.\n")
			}
		}

		query := os.Args[queryIndex]
		fSearch(query, limit)
	case "update":
		var programsToUpdate []string
		if len(os.Args) > 2 {
			programsToUpdate = os.Args[2:]
		}
		update(programsToUpdate)
	default:
		errorOut("bigdl: Unknown command.\n")
	}
}
