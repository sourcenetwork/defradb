#!/usr/bin/env bash

# Cross-compilation script for DefraDB.
# We assume we run this from a position where DefraDB's build toolchain (go, make, ...) is available.

BUILD_DIR="build/"

platforms=$1
if [[ -z "${platforms}" ]]; then
    echo "Building for all platforms"
    # A subset of the comprehensive list found at https://go.dev/doc/install/source#environment
    platforms=(
        "windows/amd64"
        "windows/arm64"
        "windows/arm"
        "linux/amd64"
        "linux/arm64"
        "linux/arm"
        "darwin/amd64"
        "darwin/arm64"
        # "js/wasm"
    )
else
    platforms=(${platforms//,/ })
fi

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS="${platform_split[0]}"
    GOARCH="${platform_split[1]}"
    output_name=$BUILD_DIR'defradb-'$GOOS'-'$GOARCH
    if [ "$GOOS" = "windows" ]; then
        output_name+='.exe'
    fi
    if ! env GOOS="$GOOS" GOARCH="$GOARCH" make build path="$output_name"; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
    echo "Completed: ${output_name}"
done