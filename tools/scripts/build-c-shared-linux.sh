#!/usr/bin/env bash
set -euo pipefail

# Optional: allow passing in build flags from the Makefile
BUILD_FLAGS="${1:-}"

# Default to the host architecture. A native build uses the system C compiler,
# a cross build uses the matching GNU cross compiler. Either can be overridden
# by setting GOARCH or CC.
HOST_ARCH=$(go env GOHOSTARCH)
GOARCH="${GOARCH:-$HOST_ARCH}"
case "$GOARCH" in
  amd64) CROSS_CC=x86_64-linux-gnu-gcc ;;
  arm64) CROSS_CC=aarch64-linux-gnu-gcc ;;
  *)
    echo "Unsupported architecture: $GOARCH (expected amd64 or arm64)" >&2
    exit 1
    ;;
esac
if [[ "$GOARCH" == "$HOST_ARCH" ]]; then
  CC="${CC:-cc}"
else
  CC="${CC:-$CROSS_CC}"
fi

echo "Building c-shared library for Linux ($GOARCH, using $CC)..."

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
CGO_ENABLED=1 GOARCH="$GOARCH" GOOS=linux CC="$CC" \
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
  # Debian's architecture names match Go's for the supported architectures.
  DEB_DIR="/tmp/libdefradb_${DEB_VERSION}_${GOARCH}"
  DEB_TMP="/tmp/libdefradb_${DEB_VERSION}_${GOARCH}.deb"

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
Architecture: ${GOARCH}
Maintainer: Democratized Data Foundation <support@source.network>
Description: DefraDB C shared library
 Provides the libdefradb shared library and C headers for embedding
 DefraDB in applications via the C bindings API.
EOF
  chmod 0644 "${DEB_DIR}/DEBIAN/control"

  fakeroot dpkg-deb --build "${DEB_DIR}" "${DEB_TMP}"
  mv "${DEB_TMP}" "build/libdefradb_${DEB_VERSION}_${GOARCH}.deb"
  rm -rf "${DEB_DIR}"

  echo "Build complete: build/libdefradb_${DEB_VERSION}_${GOARCH}.deb"
fi