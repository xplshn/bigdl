### BigDL: A (static)Binary manager

BigDL, the sizzling hot Golang rewrite of BDL! 🔥

BigDL is your sleek and efficient companion for managing and downloading static binaries effortlessly. It flaunts its independence, requiring nothing but a Linux system.
Unlike BDL which required either wget or curl, BigDL is suited for embedded systems (minimum needed: 8MBs of RAM and storage. Amd64/Aarch64).

#### Features:

1. **Minimal Dependencies**: BigDL prides itself on its self-sufficiency. No bloated dependencies, we like 'em slim.

2. **Versatile Commands**:
   - `list`: View all available binaries across all three repos.
   - `install`: Seamlessly add your desired program to your system.
   - `remove`: Say goodbye to programs you no longer need.
   - `update`: Get new features. Update all, or a select list of programs in an instant.
   - `run`: Execute programs instantly, without the hassle of installing.
   - `info`: Get detailed information about a specific program.
   - `search`: Find the perfect program to suit your needs.
   - `tldr`: Show a <abbr title="Details: its just an alias to 'bigdl run tlrc'">tldr</abbr> page (without installing any 'TLDR' client)


#### Usage:

```
$ Usage: bigdl [-vh] {list|install|remove|update|run|info|search|tldr} [args...]
```

#### BDL Compatibility:

Don't worry if you're accustomed to the old ways. BigDL maintains compatibility with BDL commands for a smooth transition:

```
$ Usage: bdl {list|install|remove|run|info|search|tldr} [args...]
```
See? Same syntax!

#### So why wait?

Get your hands on BigDL now and experience your binaries like never before. It's time to turn up the heat on your command line experience! 🚀

### Contribute:

Found a bug? Have a feature request? We welcome contributions of all kinds! Fork the repository, make your changes, and submit a pull request. Let's make BigDL even sexier together. 😏

### License:

BigDL is licensed under the [New BSD License](LICENSE), so feel free to use, modify, and distribute it as you please. Enjoy responsibly! 🍸

### Disclaimer:

BigDL is designed for those who like their command line experiences spicy. Use responsibly and at your own risk. We are not liable for any overheating systems or heart palpitations caused by the sheer awesomeness of BigDL. 😎

###### Special thanks to ChatGPT for the awesome (And very stupid) Readme
