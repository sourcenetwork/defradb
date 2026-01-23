#!/bin/bash
set -e

DEST_DIR="${1:-/usr/local/lib}"

WASMER_DIR=$(go list -m -f '{{.Dir}}' github.com/wasmerio/wasmer-go)
if [ -z "$WASMER_DIR" ]; then
    echo "Error: Could not find wasmer-go module."
    exit 1
fi

HOST_OS="${GOOS:-$(go env GOHOSTOS)}"
HOST_ARCH="${GOARCH:-$(go env GOHOSTARCH)}"

WASMER_ARCH=""
if [ "$HOST_OS" == "linux" ]; then
    if [ "$HOST_ARCH" == "amd64" ]; then
        WASMER_ARCH="linux-amd64"
    elif [ "$HOST_ARCH" == "arm64" ]; then
        WASMER_ARCH="linux-aarch64"
    fi
elif [ "$HOST_OS" == "darwin" ]; then
    if [ "$HOST_ARCH" == "amd64" ]; then
        WASMER_ARCH="darwin-amd64"
    elif [ "$HOST_ARCH" == "arm64" ]; then
        WASMER_ARCH="darwin-aarch64"
    fi
fi

if [ -z "$WASMER_ARCH" ]; then
    echo "Warning: Unsupported OS/Arch for auto-installing libwasmer: $HOST_OS/$HOST_ARCH"
    exit 0
fi

LIB_PATH="$WASMER_DIR/wasmer/packaged/lib/$WASMER_ARCH"
LIB_NAME="libwasmer.so"
if [ "$HOST_OS" == "darwin" ]; then
    LIB_NAME="libwasmer.dylib"
fi

if [ ! -f "$LIB_PATH/$LIB_NAME" ]; then
    echo "Error: Could not find $LIB_NAME in $LIB_PATH"
    exit 1
fi

echo "Found $LIB_NAME at $LIB_PATH"
echo "Installing to $DEST_DIR..."

if [ ! -d "$DEST_DIR" ]; then
    mkdir -p "$DEST_DIR" || { echo "Failed to create $DEST_DIR"; exit 1; }
fi

if [ -f "$DEST_DIR/$LIB_NAME" ]; then
    rm -f "$DEST_DIR/$LIB_NAME"
fi

cp "$LIB_PATH/$LIB_NAME" "$DEST_DIR/" || {
    echo "Failed to copy $LIB_NAME to $DEST_DIR. Permission denied?"
    echo "Try running with sudo or checking permissions."
    exit 1
}

echo "Successfully installed $LIB_NAME to $DEST_DIR"
