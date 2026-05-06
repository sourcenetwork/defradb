#!/usr/bin/env bash
set -euo pipefail

# Optional: allow passing in build flags from the Makefile
BUILD_FLAGS="${1:-}"

echo "Building c-shared library for Linux..."

search="package cbindings"
replace="package main"

# Escape special characters for sed
search_escaped=$(echo "$search" | sed 's/[\/&]/\\&/g')
replace_escaped=$(echo "$replace" | sed 's/[\/&]/\\&/g')

# Always restore package names on exit
trap '
  echo "Restoring package names..."
  find ./cbindings -type f -name "*.go" ! -path "*/.git/*" ! -path "*/vendor/*" \
    -exec sed -i "s/$replace_escaped/$search_escaped/g" {} +
' EXIT

echo "Temporarily replacing '$search' with '$replace'..."
find ./cbindings -type f -name "*.go" ! -path "*/.git/*" ! -path "*/vendor/*" \
  -exec sed -i "s/$search_escaped/$replace_escaped/g" {} +

echo "Removing existing .so and .h files..."
rm -f build/libdefradb.so build/libdefradb.h

echo "Building shared object..."
CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -tags "cshared ${BUILD_TAGS:-}" $BUILD_FLAGS \
  -buildmode=c-shared -o build/libdefradb.so ./cbindings

# Copy extra headers if needed
cp ./cbindings/defra_structs.h ./build/

echo "Build complete: build/libdefradb.so"
