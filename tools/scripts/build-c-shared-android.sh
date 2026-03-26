#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <ANDROID_NDK_PATH> [BUILD_FLAGS] [API LEVEL]"
  exit 1
fi

API_LEVEL=21
BUILD_FLAGS=""

# Parse remaining arguments
for arg in "$@"; do
  if [[ "$arg" =~ ^[0-9]+$ ]] && [ "$API_LEVEL" = 21 ]; then
    API_LEVEL="$arg"
  else
    BUILD_FLAGS="$BUILD_FLAGS $arg"
  fi
done

echo "Building c-shared library for Android (arm64) using NDK at: $ANDROID_NDK (API level $API_LEVEL)"

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

BUILD_DIR="build"

# Pre-build cleanup
echo "Removing existing .so and .h files..."
rm -f "$BUILD_DIR"/*.h
rm -f "$BUILD_DIR"/arm64-v8a/*.so
rm -f "$BUILD_DIR"/x86_64/*.so

# Ensure directories exist
mkdir -p "$BUILD_DIR/x86_64"
mkdir -p "$BUILD_DIR/arm64-v8a"

# Detect host platform
case "$(uname -s)" in
    Linux*)   HOST_TAG=linux-x86_64 ;;
    Darwin*)  HOST_TAG=darwin-x86_64 ;;
    MINGW*|MSYS*|CYGWIN*) HOST_TAG=windows-x86_64 ;;
    *)        echo "Unknown host: $(uname -s)"; exit 1 ;;
esac

# Build arm64-v8a
echo "Building arm64-v8a shared object..."
CGO_ENABLED=1 \
GOOS=android \
GOARCH=arm64 \
CC="$ANDROID_NDK/toolchains/llvm/prebuilt/$HOST_TAG/bin/aarch64-linux-android${API_LEVEL}-clang" \
go build -tags "cshared android" -buildmode=c-shared \
    -ldflags='-extldflags "-Wl,-soname,libdefradb.so"' \
    -o "$BUILD_DIR/arm64-v8a/libdefradb.so" ./cbindings

# Build x86_64
echo "Building x86_64 shared object..."
CGO_ENABLED=1 \
GOOS=android \
GOARCH=amd64 \
CC="$ANDROID_NDK/toolchains/llvm/prebuilt/$HOST_TAG/bin/x86_64-linux-android${API_LEVEL}-clang" \
go build -tags "cshared android" -buildmode=c-shared \
    -ldflags='-extldflags "-Wl,-soname,libdefradb.so"' \
    -o "$BUILD_DIR/x86_64/libdefradb.so" ./cbindings

echo "Build finished. Files in:"
echo "  $BUILD_DIR/arm64-v8a/libdefradb.so"
echo "  $BUILD_DIR/x86_64/libdefradb.so"


# Copy and clean up headers
cp "$BUILD_DIR/arm64-v8a/libdefradb.h" "$BUILD_DIR/"
cp ./cbindings/defra_structs.h "$BUILD_DIR/"
rm -f "$BUILD_DIR"/arm64-v8a/libdefradb.h "$BUILD_DIR"/x86_64/libdefradb.h

echo "Build complete: build/libdefradb.so"