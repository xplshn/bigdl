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
