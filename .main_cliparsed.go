// main.go // This is the main entrypoint, which calls all the different functions //>
package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
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
	usagePage     = "Usage: bigdl [-vh] [list|install|remove|update|run|info|search|tldr] <args>"
	// Truncation indicator
	indicator = "...>"
	// Cache size limit & handling.
	MaxCacheSize     = 10
	BinariesToDelete = 5
	TEMP_DIR         = "/tmp/bigdl_cached"
)

func main() {
	app := &cli.App{
		Name:  "bigdl",
		Usage: "bigdl is a command-line tool for managing binaries",
		Before: func(c *cli.Context) error {
			// Prepare global variables
			if InstallDir == "" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to get user's Home directory: %v", err)
				}
				InstallDir = filepath.Join(homeDir, ".local", "bin")
			}
			if err := os.MkdirAll(InstallDir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create install directory: %v", err)
			}
			switch runtime.GOARCH {
			case "amd64":
				validatedArch = [3]string{"x86_64_Linux", "x86_64", "x86_64-Linux"}
			case "arm64":
				validatedArch = [3]string{"aarch64_arm64_Linux", "aarch64_arm64", "aarch64-Linux"}
			default:
				return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
			}
			arch := validatedArch[0]
			Repositories = append(Repositories, "https://bin.ajam.dev/"+arch+"/")
			Repositories = append(Repositories, "https://bin.ajam.dev/"+arch+"/Baseutils/")
			Repositories = append(Repositories, "https://raw.githubusercontent.com/xplshn/Handyscripts/master/")
			// These are used for listing the binaries themselves
			MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/"+arch+"/METADATA.json")
			MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/"+arch+"/Baseutils/METADATA.json")
			MetadataURLs = append(MetadataURLs, "https://api.github.com/repos/xplshn/Handyscripts/contents")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "install",
				Aliases: []string{"add"},
				Usage:   "Install a binary",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return cli.NewExitError("Missing binary name", 1)
					}
					binaryName := c.Args().First()
					err := installCommand(binaryName)
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:    "remove",
				Aliases: []string{"del"},
				Usage:   "Remove a binary",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return cli.NewExitError("Missing binary name", 1)
					}
					remove(c.Args().Slice())
					return nil
				},
			},
			{
				Name:  "update",
				Usage: "Update binaries",
				Action: func(c *cli.Context) error {
					var binariesToUpdate []string
					if c.NArg() > 0 {
						binariesToUpdate = c.Args().Slice()
					}
					err := update(binariesToUpdate)
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:  "run",
				Usage: "Run a binary",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return cli.NewExitError("Missing binary name", 1)
					}
					binaryName := c.Args().First()
					RunFromCache(binaryName, c.Args().Tail())
					return nil
				},
			},
			{
				Name:  "info",
				Usage: "Show information about a specific binary",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return cli.NewExitError("Missing binary name", 1)
					}
					binaryName := c.Args().First()
					binaryInfo, err := getBinaryInfo(binaryName)
					if err != nil {
						return err
					}
					// Assuming binaryInfo is a struct with fields like Name, Description, etc.
					fmt.Printf("Name: %s\n", binaryInfo.Name)
					// Add other fields as necessary
					return nil
				},
			},
			{
				Name:  "search",
				Usage: "Search for a binary",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return cli.NewExitError("Missing search term", 1)
					}
					searchTerm := c.Args().First()
					fSearch(searchTerm)
					return nil
				},
			},
			{
				Name:  "tldr",
				Usage: "Show a brief description & usage examples for a given program/command",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return cli.NewExitError("Missing program/command name", 1)
					}
					// Directly use the first argument as the program/command name
					RunFromCache("tlrc", c.Args().Slice())
					return nil
				},
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show the version number",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Bool("version") {
				fmt.Println("bigdl", "1.3.1") // Use the actual version from your original main.go
				return nil
			}
			if c.Bool("help") {
				cli.ShowAppHelp(c)
				return nil
			}
			cli.ShowAppHelp(c)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
