# BigDL: Advanced Binary Management Tool

BigDL is a sophisticated, Golang-based rewrite of the original BDL, designed to enhance the management and downloading of static binaries with minimal effort. This tool is made to operate on Linux systems, BigDL is particularly well-suited for embedded systems, requiring a minimum of 8MB RAM and storage, with support for both Amd64 AND Aarch64 architectures.

## Features

- **Minimal Dependencies**: BigDL is designed with simplicity and efficiency in mind, boasting a slim dependency footprint.
- **Versatile Commands**:
 - `list`: Browse through all available binaries across repositories.
 - `install`: Effortlessly add your desired programs to your system.
 - `remove`: Uninstall programs that are no longer needed.
 - `update`: Keep your system up-to-date with new features or updates for selected programs.
 - `run`: Execute programs directly without the need for installation.
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

###### Expample of one use case of bigdl | Inside of a SH script
Whenever you want to pull a specific GNU coreutil, insert a bash snippet, use a *fetch tool, etc, you can use bigdl for the job! There's also a `--transparent` flag for `run`, which will use the users' installed version of the program you want to run, and if it is not found in the `$PATH` bigdl will fetch it and run it from `/tmp/bigdl_cached`.
```sh
system_info=$(wget -qO- "https://raw.githubusercontent.com/xplshn/bigdl/master/stubdl" | sh -s -- run --silent albafetch --no-logo - || curl -qsfSL "https://raw.githubusercontent.com/xplshn/bigdl/master/stubdl" | sh -s -- run --silent albafetch --no-logo -)
```

## Contributing

Contributions are welcome! Whether you've found a bug, have a feature request, or wish to improve the documentation, your input is valuable. Fork the repository, make your changes, and submit a pull request. Together, we can make BigDL even more powerful and user-friendly.

## License

BigDL is licensed under the New BSD License. This allows for the use, modification, and distribution of the software under certain conditions. For more details, please refer to the [LICENSE](LICENSE) file.
