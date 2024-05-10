# BigDL: Advanced Binary Management Tool
[![Go Report Card](https://goreportcard.com/badge/github.com/xplshn/bigdl)](https://goreportcard.com/report/github.com/xplshn/bigdl)
[![License](https://img.shields.io/badge/license-%20RABRMS-green)](https://github.com/xplshn/bigdl/blob/master/LICENSE)
[![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/xplshn/bigdl?include_prereleases)](https://github.com/xplshn/bigdl/releases/latest)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/xplshn/bigdl)

BigDL is a sophisticated, Golang-based rewrite of the original BDL, it is like a package manager, but without the hassle of dependencies nor the bloat, every binary provided is statically linked. This tool is made to operate on Linux systems, BigDL is particularly well-suited for embedded systems, with support for both Amd64 AND Aarch64.

##### Why?
 > “I tend to think the drawbacks of dynamic linking outweigh the advantages for many (most?) applications.” – John Carmack

## Features
**Minimal Dependencies**: BigDL is designed with simplicity and efficiency in mind, boasting a slim dependency footprint.
 - `list`: Browse through all available binaries across repositories.
 - `install`: Effortlessly add your desired programs to your system.
 - `remove`: Uninstall programs that are no longer needed.
 - `update`: Keep your system up-to-date with new features or updates for selected programs.
 - `run`: Execute programs directly without the need for installation. It will also exit using the SAME exit code as the program you try to run, that along with the --silent/--verbose/--transparent flags allow for using bigdl inside of scripts (check out stubdl).
 - `info`: Obtain detailed information about specific programs.
 - `search`: Locate the perfect program to meet your requirements.
 - `tldr`: Access a quick reference guide without installing any additional software.

### Usage

```
$ Usage: bigdl [-v,-h] {list|install|remove|update|run|info|search|tldr} [args...]
```

## Getting Started

To begin using BigDL, simply download and install it on your Linux system. No additional setup is required.

```
wget -qO- "https://raw.githubusercontent.com/xplshn/bigdl/master/stubdl" | sh -s -- --install "$HOME/.local/bin/bigdl"
```

###### Example of one use case of bigdl | Inside of a SH script
Whenever you want to pull a specific GNU coreutil, insert a bash snippet, use a *fetch tool, etc, you can use bigdl for the job! There's also a `--transparent` flag for `run`, which will use the users' installed version of the program you want to run, and if it is not found in the `$PATH` bigdl will fetch it and run it from `/tmp/bigdl_cached`.
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
