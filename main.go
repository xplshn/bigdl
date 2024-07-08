package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	// Repositories contains all available repos - This variable is used by findURL.go
	Repositories []string
	// MetadataURLs are used for listing the binaries themselves. Not to be confused with R*MetadataURLs.
	MetadataURLs []string
	// RNMetadataURL should contain the JSON that describes available binaries for your architecture
	RNMetadataURL string
	// ValidatedArch is used in fsearch.go, info.go and main.go to determine which repos to use.
	ValidatedArch = [3]string{}
	// InstallDir holds the directory that shall be used for installing, removing, updating, listing with `info`. It takes the value of $INSTALL_DIR if it is set in the user's env, otherwise it is set to have a default value
	InstallDir = os.Getenv("INSTALL_DIR")
	// TEMPDIR will be used as the dir to download files to before moving them to a final destination AND as the place that will hold cached binaries downloaded by `run`
	TEMPDIR = os.Getenv("BIGDL_CACHEDIR")
	// InstallMessage will be printed when installCommand() succeeds
	InstallMessage = "disabled"
	// InstallUseCache determines if cached files should be used when requesting an install
	InstallUseCache = true
	// UseProgressBar determines if the progressbar is shown or not
	UseProgressBar = true
	// DisableTruncation determines if update.go, fsearch.go, etc, truncate their messages or not
	DisableTruncation = false
	// Always adds a NEWLINE to text truncated by the truncateSprintf/truncatePrintf function
	AddNewLineToTruncateFn = true
)

const (
	VERSION   = "1.6.9"                                                               // VERSION to be displayed
	usagePage = " [-v|-h] [list|install|remove|update|run|info|search|tldr] <-args->" // usagePage to be shown
	// Truncation indicator
	indicator = "...>"
	// MaxCacheSize is the limit of binaries which can be stored at TEMP_DIR
	MaxCacheSize = 10
	// BinariesToDelete - Once the cache is filled - The programs populate the list of binaries to be removed in order of least used. This variable sets the amount of binaries that should be deleted
	BinariesToDelete = 5
)

// Exclude specified file types and file names, these shall not appear in Lists nor in the Search Results
var excludedFileTypes = map[string]struct{}{
	".7z":   {},
	".bz2":  {},
	".json": {},
	".gz":   {},
	".md":   {},
	".txt":  {},
	".tar":  {},
	".zip":  {},
	".cfg":  {},
	".dir":  {},
	".test": {},
}

var excludedFileNames = map[string]struct{}{
	"TEST":                     {},
	"LICENSE":                  {},
	"experimentalBinaries_dir": {},
	"bundles_dir":              {},
	"blobs_dir":                {},
	"robotstxt":                {},
	"bdl.sh":                   {},
	// Because the repo contains duplicated files. And I do not manage the repo nor plan to implement sha256 filtering :
	"uroot":             {},
	"uroot-busybox":     {},
	"gobusybox":         {},
	"sysinfo-collector": {},
	"neofetch":          {},
}

func init() {
	if InstallDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			errorOut("error: Failed to get user's Home directory. Maybe set $BIGDL_CACHEDIR? %v\n", err)
			os.Exit(1)
		}
		InstallDir = filepath.Join(homeDir, ".local", "bin")
	}
	if TEMPDIR == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			errorOut("error: Failed to get user's Cache directory. Maybe set $BIGDL_CACHEDIR? %v\n", err)
			os.Exit(1)
		}
		TEMPDIR = filepath.Join(cacheDir, "bigdl_cache")
	}
	if os.Getenv("BIGDL_TRUNCATION") == "0" {
		DisableTruncation = true
	}
	if os.Getenv("BIGDL_ADDNEWLINE") == "1" {
		AddNewLineToTruncateFn = true
	}
	if os.Getenv("BIGDL_PRBAR") == "0" {
		UseProgressBar = false
	}

	// The repos are a mess. So we need to do this. Sorry
	arch := runtime.GOARCH + "_" + runtime.GOOS
	switch arch {
	case "amd64_linux":
		ValidatedArch = [3]string{"x86_64_Linux", "x86_64", "x86_64-Linux"}
	case "arm64_linux":
		ValidatedArch = [3]string{"aarch64_arm64_Linux", "aarch64_arm64", "aarch64-Linux"}
	case "arm64_android":
		ValidatedArch = [3]string{"arm64_v8a_Android", "arm64_v8a_Android", "arm64-v8a-Android"}
	case "amd64_windows":
		ValidatedArch = [3]string{"x64_Windows", "x64_Windows", "AMD64-Windows_NT"}
	default:
		fmt.Println("Unsupported architecture:", arch)
		os.Exit(1)
	}
	arch = ValidatedArch[0]
	Repositories = append(Repositories, "https://bin.ajam.dev/"+arch+"/")
	Repositories = append(Repositories, "https://bin.ajam.dev/"+arch+"/Baseutils/")
	Repositories = append(Repositories, "https://raw.githubusercontent.com/xplshn/Handyscripts/master/")
	// Binaries that are available in the Repositories but aren't described in any MetadataURLs will not be updated, nor listed with `info` nor `list`
	RNMetadataURL = "https://bin.ajam.dev/" + arch + "/METADATA.json" // RNMetadataURL is the file which contains a concatenation of all metadata in the different repos, this one also contains sha256 checksums
	MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/"+arch+"/METADATA.json")
	MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/"+arch+"/Baseutils/METADATA.json")
	MetadataURLs = append(MetadataURLs, "https://api.github.com/repos/xplshn/Handyscripts/contents")
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
 run              Run a specified binary from cache
 info             Show information about a specific binary OR display installed binaries
 search           Search for a binary - (not all binaries have metadata. Use list to see all binaries)
 tldr             Equivalent to "run --transparent --verbose tlrc" as argument

Variables:
 BIGDL_PRBAR      If present, and set to ZERO (0), the download progressbar will be disabled
 BIGDL_TRUNCATION If present, and set to ZERO (0), string truncation will be disabled
 BIGDL_ADDNEWLINE If present, and set to ONE  (1), truncated strings will always be on a new line
 BIGDL_CACHEDIR   If present, it must contain a valid directory
 INSTALL_DIR      If present, it must contain a valid directory

Examples:
 bigdl search editor
 bigdl install micro
 bigdl install lux kakoune aretext shfmt
 bigdl install --silent bed && echo "[bed] was installed to $INSTALL_DIR/bed"
 bigdl del bed
 bigdl del orbiton tgpt lux
 bigdl info
 bigdl info jq
 bigdl list --described
 bigdl tldr gum
 bigdl run --verbose curl -qsfSL "https://raw.githubusercontent.com/xplshn/bigdl/master/stubdl" | sh -
 bigdl run --silent elinks -no-home "https://fatbuffalo.neocities.org/def"
 bigdl run --transparent --silent micro ~/.profile
 bigdl run btop

Version: ` + VERSION

	fmt.Println(helpMessage)
}

func main() {
	errorOutInsufficientArgs := func() { errorOut("Error: Insufficient parameters\n") }
	version := flag.Bool("v", false, "Show the version number")
	versionLong := flag.Bool("version", false, "Show the version number")

	flag.Usage = printHelp
	flag.Parse()

	if *version || *versionLong {
		errorOut("bigdl %s\n", VERSION)
	}

	if flag.NArg() < 1 {
		errorOut(" bigdl:%s\n", usagePage)
	}

	if err := os.MkdirAll(InstallDir, os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get user's Home directory. %v\n", err)
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
				// Call fSearch with an empty query and a large limit to list all described binaries
				fSearch("", 99999)
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
		if flag.NArg() < 2 {
			fmt.Printf("Usage: bigdl %s <--silent> [binar|y|ies]\n", flag.Arg(0))
			os.Exit(1)
		}

		// Join the binary names into a single string separated by spaces
		binaries := strings.Join(flag.Args()[1:], " ")
		silent := false
		if flag.Arg(1) == "--silent" {
			silent = true
			// Skip the "--silent" flag when joining the binary names
			binaries = strings.Join(flag.Args()[2:], " ")
		}

		err := installCommand(silent, binaries)
		if err != nil {
			fmt.Printf("Installation failed: %v\n", err)
			os.Exit(1)
		}
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
		args := append([]string{"--transparent", "--verbose", "tlrc"}, flag.Args()[1:]...) // UGLY!
		RunFromCache(args[0], args[1:])
	case "info":
		binaryName := flag.Arg(1)
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
			if binaryInfo.Updated != "" {
				fmt.Printf("Updated: %s\n", binaryInfo.Updated)
			}
			if binaryInfo.Version != "" {
				fmt.Printf("Version: %s\n", binaryInfo.Version)
			}
			if binaryInfo.Size != "" {
				fmt.Printf("Size: %s\n", binaryInfo.Size)
			}
			if binaryInfo.Source != "" { // if binaryInfo.Extras != "" {
				fmt.Printf("Source: %s\n", binaryInfo.Source)
			}
			if binaryInfo.SHA256 != "" {
				fmt.Printf("SHA256: %s\n", binaryInfo.SHA256)
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
