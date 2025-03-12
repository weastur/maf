#!/bin/sh

set -e

rm -rf build/completions
mkdir build/completions

for sh in bash zsh fish; do
	./bin/maf completion "$sh" >"build/completions/maf.$sh"
done
