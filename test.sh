#!/bin/sh

#echo "override" >> ~/.local/bin/tgpt
#echo "override" >> ~/.local/bin/btop
#echo "override" >> ~/.local/bin/dd
#echo "override" >> ~/.local/bin/whoami
#echo "override" >> ~/.local/bin/who
#echo "override" >> ~/.local/bin/dircolors
#echo "override" >> ~/.local/bin/jq
./bigdl add tgpt jq btop aretext busybox/dd busybox/whoami coreutils/who coreutils/dircolors
./bigdl del aretext
