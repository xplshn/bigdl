#!/bin/sh
curl -qfsL "https://github.com/xplshn/bigdl/releases/download/1.3.1/bigdl_amd64" -o bigdl_amd64 && \
 chmod +x bigdl_amd64 ; \
  command -v ./bigdl_amd64 && \
   ./bigdl_amd64 run --silent gum confirm "Do you want to test $PWD/bigdl_amd64?" --negative="No, remove it!" || \
     rm ./bigdl_amd64 ; \
   command -v ./bigdl_amd64 && \
    X=$(./bigdl_amd64 run --silent gum input --placeholder="Usage: bigdl [-vh] {list|install|remove|update|run|info|search|tldr} [args...]") && \
     echo "$X" | awk '{ for (i = 1; i <= NF; i++) { if ($i ~ /^(--help|--version|-v|-h|list|install|remove|update|run|info|search|tldr)$/) { for (j = i; j <= NF; j++) { printf "%s ", $j }; print ""; break } } }' | \
      xargs -r ./bigdl_amd64
