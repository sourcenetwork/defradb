#!/usr/bin/env bash
set -euo pipefail

# Usage: tools/scripts/build-c-shared-macos.sh [go build flags...]
# Use the Makefile target to include version metadata.
if [[ "$(uname -s)" != Darwin ]]; then
  echo "The macOS C shared library requires a macOS host." >&2
  exit 1
fi

echo "Building c-shared library for macOS..."

mkdir -p build

# The cbindings package must become the main package for this to work, but
# we also need to change it back afterwards, whether this succeeds or fails
search="package cbindings"
replace="package main"
search_escaped=$(echo "$search" | sed 's/[\/&]/\\&/g')
replace_escaped=$(echo "$replace" | sed 's/[\/&]/\\&/g')

# BSD sed requires an explicit empty backup suffix for in-place edits.
trap '
  echo "Restoring package names..."
  find ./cbindings -type f -name "*.go" ! -path "*/.git/*" ! -path "*/vendor/*" \
    -exec sed -i "" "s/$replace_escaped/$search_escaped/g" {} +
' EXIT

echo "Temporarily replacing '$search' with '$replace'..."
find ./cbindings -type f -name "*.go" ! -path "*/.git/*" ! -path "*/vendor/*" \
  -exec sed -i "" "s/$search_escaped/$replace_escaped/g" {} +

# Remove the existing library and header artifacts.
rm -f build/libdefradb.dylib build/libdefradb.h

CGO_ENABLED=1 GOOS=darwin GOARCH="${GOARCH:-$(go env GOHOSTARCH)}" CC="${CC:-cc}" \
go build "$@" \
  -tags "cshared ${BUILD_TAGS:-}" \
  -buildmode=c-shared \
  -o build/libdefradb.dylib \
  ./cbindings

install_name_tool -id '@rpath/libdefradb.dylib' build/libdefradb.dylib
# Changing the install name invalidates Go's ad-hoc signature on Apple Silicon.
codesign --force --sign - build/libdefradb.dylib
cp ./cbindings/defra_structs.h ./build/

echo "Build complete: build/libdefradb.dylib"
echo "Headers: build/libdefradb.h, build/defra_structs.h"
