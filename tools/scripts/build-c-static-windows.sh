#!/usr/bin/env bash
set -euo pipefail
BUILD_FLAGS=("$@")
echo "Building Windows static library..."

mkdir -p build

# The cbindings package must become the main package for this to work, but
# we also need to change it back afterwards, whether this succeeds or fails
search="package cbindings"
replace="package main"
search_escaped=$(echo "$search" | sed 's/[\/&]/\\&/g')
replace_escaped=$(echo "$replace" | sed 's/[\/&]/\\&/g')

trap '
echo "Restoring package names..."
find ./cbindings -type f -name "*.go" ! -path "*/.git/*" ! -path "*/vendor/*" \
  -exec sed -i "s/$replace_escaped/$search_escaped/g" {} +
' EXIT

echo "Temporarily replacing '$search' with '$replace'..."
find ./cbindings -type f -name "*.go" ! -path "*/.git/*" ! -path "*/vendor/*" \
  -exec sed -i "s/$search_escaped/$replace_escaped/g" {} +

# Remove the existing lib and header artifacts
rm -f build/libdefradb.lib build/libdefradb.h

# Build the lib file
GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 \
go build -tags cshared "${BUILD_FLAGS[@]}" -buildmode=c-archive -o build/libdefradb.lib ./cbindings

# Copy over the structs header the user will need
cp ./cbindings/defra_structs.h ./build/

echo "Build complete"
echo "Static library: build/libdefradb.lib"
echo "Headers: build/libdefradb.h, build/defra_structs.h"