Convert this SH script which does the following to Golang.

```instructions_why_and_how_and_rules
Provide me with a JQ command that takes this JSON format/structure as input and:
 1. Marks duplicates by adding #1, #2, #3 and so on to their "name" field's value
 2. Adds a field called real_name after the "name" field that is the value of the filepath of download_url. (Convert the URL into a file path, forget about the protocol:// thingie and then use that. Remember to convert URI encoding to plain). This field should be NULL if the file path matches the "name" field

```
[
  {
    "name": "7z",
    "description": "Unarchiver",
    "download_url": "https://bin.ajam.dev/x86_64_Linux/7z",
    "size": "3.73 MB",
    "b3sum": "125acdc505ed6582ea1daec36c39d16749bbbf58ce2d19bdadaac27ff3b74f23",
    "sha256": "a2728a3dbd244cbb1a04f6a0998f53ec03abb7e3fb30e8e361fa22614c98e8d3",
    "build_date": "2024-06-24T03:39:33",
    "repo_url": "https://github.com/ip7z/7zip",
    "repo_author": "ip7z",
    "repo_info": "7-Zip",
    "repo_updated": "2024-07-05T14:39:53Z",
    "repo_released": "2024-06-19T10:45:51Z",
    "repo_version": "24.07",
    "repo_stars": "447",
    "repo_language": "C++",
    "repo_license": "",
    "repo_topics": "",
    "web_url": "https://www.7-zip.org",
    "extra_bins": ""
  },
```

If there were a duplicate of 7z, it would become:

```
[
  {
    "name": "7z#2",
    "description": "7z unarchiver but with additional support for PAX",
    "download_url": "https://bin.ajam.dev/x86_64_Linux/7z",
    "size": "3.73 MB",
    "b3sum": "125acdc505ed6582ea1daec36c39d16749bbbf58ce2d19bdadaac27ff3b74f23",
    "sha256": "a2728a3dbd244cbb1a04f6a0998f53ec03abb7e3fb30e8e361fa22614c98e8d3",
    "build_date": "2024-06-24T03:39:33",
    "repo_url": "https://github.com/ip7z/7zip",
    "repo_author": "ip7z",
    "repo_info": "7-Zip",
    "repo_updated": "2024-07-05T14:39:53Z",
    "repo_released": "2024-06-19T10:45:51Z",
    "repo_version": "24.07",
    "repo_stars": "447",
    "repo_language": "C++",
    "repo_license": "",
    "repo_topics": "",
    "web_url": "https://www.7-zip.org",
    "extra_bins": ""
  },
```

Since download_url points to a file which's file path is after removing the domain and the protocol is in the same dir as the value of the "name" field, the real_name is not added.

```

```sh
#!/bin/sh

# Function to URL decode
urldecode() {
  url_encoded=$(echo "$1" | sed -e 's/+/ /g' -e 's/%\([0-9a-fA-F][0-9a-fA-F]\)/\\x\1/g')
  printf '%b' "$url_encoded"
}

# Temporary file setup
tmp_dir="$(mktemp -d -t "jq_process_$(date +%s)_XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

# Parse arguments
input_file=""
output_file=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o) shift; output_file="$1";;
    *) input_file="$1";;
  esac
  shift
done

if [ -z "$input_file" ]; then
  echo "Input file is required." >&2
  exit 1
fi

# Process the JSON to handle duplicates and add a placeholder for real_name
jq '[
  group_by(.name)[] |
  to_entries[] |
  .value += {
    name: (.value.name + if .key > 0 then "#" + (.key + 1 | tostring) else "" end),
    real_name: null
  } | .value | {
    name: .name,
    real_name: .real_name,
    description: .description,
    download_url: .download_url,
    size: .size,
    b3sum: .b3sum,
    sha256: .sha256,
    build_date: .build_date,
    repo_url: .repo_url,
    repo_author: .repo_author,
    repo_info: .repo_info,
    repo_updated: .repo_updated,
    repo_released: .repo_released,
    repo_version: .repo_version,
    repo_stars: .repo_stars,
    repo_language: .repo_language,
    repo_license: .repo_license,
    repo_topics: .repo_topics,
    web_url: .web_url,
    extra_bins: .extra_bins
  }
] | flatten' "$input_file" > "$tmp_dir/processed.json"

# Iterate over the JSON objects and replace null with the correct real_name
jq -c '.[]' "$tmp_dir/processed.json" | while read -r item; do
  name=$(echo "$item" | jq -r '.name')
  download_url=$(echo "$item" | jq -r '.download_url')
  filepath=$(urldecode "${download_url#*://}")
  filepath="${filepath#/}"

  # Check if the filepath matches the name
  if [ "$filepath" = "$name" ]; then
    real_name="null"
  else
    real_name="\"$filepath\""
  fi

  # Update the real_name field in the item
  echo "$item" | jq --arg real_name "$real_name" '.real_name = ($real_name | fromjson)' >> "$tmp_dir/updated.json"
done

# Format the final JSON output
jq -s '.' "$tmp_dir/updated.json" > "$tmp_dir/final.json"

# Output to the specified file or stdout
if [ -n "$output_file" ]; then
  cat "$tmp_dir/final.json" > "$output_file"
else
  cat "$tmp_dir/final.json"
fi

```

Part from this template for the Golang version
```
package main

import (
	"fmt"
	"os"
	"runtime"
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
	"sh":                {}, // Because in the repo, it is a duplicate of bash and not a POSIX nor the original Thompshon Shell
}

func init() {
	// The repos are a mess. So we need to do this. Sorry
	arch := runtime.GOARCH + "_" + runtime.GOOS
	switch arch {
		ValidatedArch = [3]string{"x86_64_Linux", "aarch64_arm64_Linux", "arm64_v8a_Android"}
	default:
		fmt.Println("Unsupported architecture:", arch)
		os.Exit(1)
	}

  /* Use a FOR LOOP RANGE of ValidatedArch HERE to do this for each possible .json
	arch = ValidatedArch[0]
	RNMetadataURL = "https://bin.ajam.dev/" + arch + "/METADATA.json"
	MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/"+arch+"/METADATA.json")
	MetadataURLs = append(MetadataURLs, "https://bin.ajam.dev/"+arch+"/Baseutils/METADATA.json")
	*/
}
```

1. You must download ALL OF THE POSSIBLE COMBINATIONS of Metadata.json using the [0] for each of the archs of the ValidatedArch. Both Baseutils and bin.ajam.dev/+arch+/METADATA.json
2. You must output each of the files as The_Original_FileName_For_TheDownloaded_File.bigdl_ValidatedArch.json
3. Forget about input/output is optional and when it is used you process only the provided file. Use -i and -o for this.
4. Use existing Go libraries that are part of the standard library
5. The real_name cannot be, for example:
 ```
  "real_name": "x86_64_Linux/Baseutils/bash/bash"
 ```
  - Because the first element there is the ValidatedArch[]
  - And Also because what follows after the ValidatedArch is the basepath/basename of the MetadataURLs[1], which is "https://bin.ajam.dev/"+arch+"/Baseutils".
  - real_name's value can also NOT start with the domain of the URL.
