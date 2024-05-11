# BigDL: Advanced Binary Management Tool
[![Go Report Card](https://goreportcard.com/badge/github.com/xplshn/bigdl)](https://goreportcard.com/report/github.com/xplshn/bigdl)
[![License](https://img.shields.io/badge/license-%20RABRMS-green)](https://github.com/xplshn/bigdl/blob/master/LICENSE)
[![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/xplshn/bigdl?include_prereleases)](https://github.com/xplshn/bigdl/releases/latest)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/xplshn/bigdl)

BigDL is a sophisticated, Golang-based rewrite of the original BDL, it is like a package manager, but without the hassle of dependencies nor the bloat, every binary provided is statically linked. This tool is made to operate on Linux systems, BigDL is particularly well-suited for embedded systems, with support for both Amd64 AND Aarch64.

##### Why?
 > “I tend to think the drawbacks of dynamic linking outweigh the advantages for many (most?) applications.” – John Carmack

## Features

```
$ bigdl --help
Usage:
 [-v|-h] [list|install|remove|update|run|info|search|tldr] <{args}>

Options:
 -h, --help       Show this help message
 -v, --version    Show the version number

Commands:
 list             List all available binaries
 install, add     Install a binary to $INSTALL_DIR
 remove, del      Remove a binary from the $INSTALL_DIR
 update           Update binaries, by checking their SHA against the repo's SHA
 run              Run a binary from cache
 info             Show information about a specific binary OR display installed binaries
 search           Search for a binary - (not all binaries have metadata. Use list to see all binaries)
 tldr             Show a brief description & usage examples for a given program/command. This is an alias equivalent to using "run" with "tlrc" as argument.
```

### Examples
```
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
```

#### What are these optional flags?
##### Flags that correspond to the `run` functionality
In the case of `--transparent`, it runs the program from $PATH and if it isn't available in the user's $PATH it will pull the binary from `bigdl`'s repos and run it from cache.
In the case of `--silent`, it simply hides the progressbar and all optional messages (warnings) that `bigdl` can show, as oppossed to `--verbose`, which will always report if the binary is found on cache + the return code of the binary to be ran if it differs from 0.
##### Flags that correspond to the `install` functionality
Install accepts three optional flags too. These affect the install message that will be displayed.
 1. `"You can put a custom installation complete message just after the name of the binary, --newline will be assumed"`
 2. `--fancy "This one allows exactly ONE formatting parameter (%s), which will be replaced with installation path, it MUST be used." --newline`
 3. `--fancy --truncate "Same as before, but will be truncated once the text wraps around" --newline`
 4. `--fancy "Text, but without the --newline, for some reason."`
 5. `--truncate "This will truncate text when it overflows the terminal's size. --new-line is assumed."`
`--silent` will work the same as `run`'s `--silent` option
##### `Update` arguments:
Update can receive an optional list of specific binaries to update OR no arguments at all. When `update` receives no arguments it updates everything that is both found in the repos and in your `$INSTALL_DIR`.
###### NOTE: I may remove this at some point and instead make the `install`/`add` functionality be able to install multiple binaries at the same time.
##### Arguments of `info`
When `info` is called with no arguments, it displays binaries which are part of the `list` and are also found on your `$INSTALL_DIR`. If `info` is called with a binary's name as argument, `info` will display as much information of it as is available. The "Size", "SHA256", "B3SUM" fields may not match your local installation.
###### Example:
```
$ bigdl info jq
Name: jq
Description: Command-line JSON processor
Repo: https://github.com/jqlang/jq
Updated: 2023-12-13T19:56:17Z
Version: jq-1.7.1
Size: 2.32 MB
Source: https://bin.ajam.dev/x86_64_Linux/jq
SHA256: 5942c9b0934e510ee61eb3e30273f1b3fe2590df93933a93d7c58b81d19c8ff5
B3SUM: f4f456f3a1a9a0dbcd9b0c2a77e29d14bc1f8bb036db4f6ff06d8c76a99e5ef2
```
##### Arguments of `list`
`list` can receive the optional argument `--described`/`-d`. It will display all binaries that have a description in their metadata.
##### Arguments of `search`
`search` can only receive ONE search term, if the name of a binary or a description of a binary contains the term, it is shown as a search result.
`search` can optionally receive a `--limit` argument, which changes the limit on how many search results can be displayed (default is 90).

## Getting Started

To begin using BigDL, simply download and install it on your Linux system. No additional setup is required.
###### Use without installing:
```
wget -qO- "https://raw.githubusercontent.com/xplshn/bigdl/master/stubdl" | sh -s -- --help
```
###### Install to `~/.local/bin`
```
wget -qO- "https://raw.githubusercontent.com/xplshn/bigdl/master/stubdl" | sh -s -- --install "$HOME/.local/bin/bigdl"
```

#### Example of one use case of bigdl | Inside of a SH script
Whenever you want to pull a specific GNU coreutil, busybox, toybox, etc, insert a bash snippet, use a *fetch tool, etc, you can use bigdl for the job! There's also a `--transparent` flag for `run`, which will use the users' installed version of the program you want to run, and if it is not found in the `$PATH` bigdl will fetch it and run it from `/tmp/bigdl_cached`.
```sh
system_info=$(wget -qO- "https://raw.githubusercontent.com/xplshn/bigdl/master/stubdl" | sh -s -- run --silent albafetch --no-logo - || curl -qsfSL "https://raw.githubusercontent.com/xplshn/bigdl/master/stubdl" | sh -s -- run --silent albafetch --no-logo -)
```

### Where do these binaries come from?
- https://github.com/Azathothas/Toolpacks [https://bin.ajam.dev]
- https://github.com/Azathothas/Static-Binaries [https://bin.ajam.dev/*/Baseutils/]
- https://github.com/xplshn/Handyscripts
>Hmm, can I add my own repos?

Yes! Absolutely. The repo's URL's are declared in main.go, simply add another one if your repo is hosted at Github or your endpoint follows the same JSON format that Github's endpoint provides. You can also provide a repo URL in the same format that the [Toolpacks](https://github.com/Azathothas/Toolpacks) repo uses.

>Good to hear, now... What about the so-called MetadataURLs?

MetadataURLs provide info about the binaries, which is used to `search` and update `binaries`, also for the functionality of `info` in both of its use-cases(showing the binaries which were installed to $INSTALL_DIR from the [Toolpacks](https://github.com/Azathothas/Toolpacks) repo).

## Contributing

Contributions are welcome! Whether you've found a bug, have a feature request, or wish to improve the documentation, your input is valuable. Fork the repository, make your changes, and submit a pull request. Together, we can make BigDL even more powerful and user-friendly. If you can provide repos that meet the requirements to add them to `bigdl`, I'd be grateful.

## License

BigDL is licensed under the RABRMS License. This allows for the use, modification, and distribution of the software under certain conditions. For more details, please refer to the [LICENSE](LICENSE) file. This license is equivalent to the New or Revised BSD License.
