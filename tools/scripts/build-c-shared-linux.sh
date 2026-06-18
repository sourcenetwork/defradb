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
  find ./cbindings -type f -name "*.go" \
    ! -path "*/.git/*" \
    ! -path "*/vendor/*" \
    -exec sed -i "s/$replace_escaped/$search_escaped/g" {} +
' EXIT

echo "Temporarily replacing '$search' with '$replace'..."
find ./cbindings -type f -name "*.go" \
  ! -path "*/.git/*" \
  ! -path "*/vendor/*" \
  -exec sed -i "s/$search_escaped/$replace_escaped/g" {} +

echo "Removing existing build artifacts..."
rm -rf build
mkdir -p build

echo "Building shared object..."
CGO_ENABLED=1 GOARCH=amd64 GOOS=linux \
go build \
  -tags "cshared ${BUILD_TAGS:-}" \
  $BUILD_FLAGS \
  -buildmode=c-shared \
  -o build/libdefradb.so \
  ./cbindings

# Copy extra headers if needed
cp ./cbindings/defra_structs.h ./build/

echo "Build complete: build/libdefradb.so"

if [[ "${MAKE_DEB:-}" == "1" ]]; then
  # Determine version from git tag, falling back to 0.0.0
  DEB_VERSION=$(
    git describe --tags --match 'v[0-9]*' 2>/dev/null \
      | sed 's/^v//' \
      | sed 's/-[0-9]*-g[0-9a-f]*//' \
    || echo "0.0.0"
  )

  # Stage and build entirely in /tmp so that chmod works (bind mounts such as
  # Docker volumes, WSL mounts, and CIFS shares ignore permission changes).
  # Only move the finished .deb back to build/ at the very end.
  DEB_DIR="/tmp/libdefradb_${DEB_VERSION}_amd64"
  DEB_TMP="/tmp/libdefradb_${DEB_VERSION}_amd64.deb"

  echo "Building .deb package (version ${DEB_VERSION})..."

  rm -rf "${DEB_DIR}"
  mkdir -p \
    "${DEB_DIR}/DEBIAN" \
    "${DEB_DIR}/usr/lib" \
    "${DEB_DIR}/usr/include"
  chmod 0755 "${DEB_DIR}" "${DEB_DIR}/DEBIAN" "${DEB_DIR}/usr" "${DEB_DIR}/usr/lib" "${DEB_DIR}/usr/include"

  cp build/libdefradb.so   "${DEB_DIR}/usr/lib/libdefradb.so"
  cp build/libdefradb.h    "${DEB_DIR}/usr/include/libdefradb.h"
  cp build/defra_structs.h "${DEB_DIR}/usr/include/defra_structs.h"

  cat > "${DEB_DIR}/DEBIAN/control" <<EOF
Package: libdefradb
Version: ${DEB_VERSION}
Architecture: amd64
Maintainer: Democratized Data Foundation <support@source.network>
Description: DefraDB C shared library
 Provides the libdefradb shared library and C headers for embedding
 DefraDB in applications via the C bindings API.
EOF
  chmod 0644 "${DEB_DIR}/DEBIAN/control"

  fakeroot dpkg-deb --build "${DEB_DIR}" "${DEB_TMP}"
  mv "${DEB_TMP}" "build/libdefradb_${DEB_VERSION}_amd64.deb"
  rm -rf "${DEB_DIR}"

  echo "Build complete: build/libdefradb_${DEB_VERSION}_amd64.deb"
fi