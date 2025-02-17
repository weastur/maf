#!/bin/sh

set -e

rm -rf build/manpages
rm -rf man

mkdir man
mkdir build/manpages

maf gen man

for man in man/*; do
    gzip -c -9 >"build/manpages/$(basename "$man").gz" "$man"
done
