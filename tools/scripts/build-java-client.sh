#!/usr/bin/env bash
set -euo pipefail

# Builds defradb.jar for the tests/clients/java client: clones (or updates)
# DefraJavaWrapper into a local, gitignored checkout, then runs its own
# build.sh, which in turn:
#   1. runs `make build-c-shared-linux` in this repo to build libdefradb.so
#   2. copies that .so + headers into the checkout
#   3. compiles nativewrapper.c against it (its own src/main/c/build.sh)
#   4. runs its Gradle build to produce build/libs/defradb.jar
#
# This only works on Linux (build-c-shared-linux cross-compiles nothing - it
# needs a real Linux gcc/cgo toolchain, and DefraJavaWrapper's build.sh shells
# out to `make`/bash throughout) - run it from WSL on Windows.
#
# Requires: git, make, go, gcc, and a JDK (for Gradle) on PATH.

DEFRA_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WRAPPER_DIR="${DEFRA_JAVA_WRAPPER_DIR:-$DEFRA_DIR/.javaclient/DefraJavaWrapper}"
WRAPPER_REPO="${DEFRA_JAVA_WRAPPER_REPO:-https://github.com/sourcenetwork/DefraJavaWrapper.git}"

if [ ! -d "$WRAPPER_DIR/.git" ]; then
  echo "Cloning DefraJavaWrapper into $WRAPPER_DIR..."
  git clone "$WRAPPER_REPO" "$WRAPPER_DIR"
else
  echo "Updating existing DefraJavaWrapper checkout at $WRAPPER_DIR..."
  git -C "$WRAPPER_DIR" pull --ff-only
fi

chmod +x "$WRAPPER_DIR/build.sh" "$WRAPPER_DIR/gradlew" "$WRAPPER_DIR/src/main/c/build.sh"

echo "Building defradb.jar (this rebuilds libdefradb.so from the current working tree, then compiles+packages the Java bindings)..."
(cd "$WRAPPER_DIR" && ./build.sh --defra-dir "$DEFRA_DIR" --linux --cleanup)

JAR_PATH="$WRAPPER_DIR/build/libs/defradb.jar"
if [ ! -f "$JAR_PATH" ]; then
  echo "Build finished but $JAR_PATH was not produced" >&2
  exit 1
fi

echo ""
echo "Built $JAR_PATH"
echo ""
echo "Run 'make test:java' to run the integration tests against this jar - it derives"
echo "CGO_CFLAGS from JAVA_HOME and DEFRA_JAVA_JAR from this build for you (both are"
echo "otherwise required and are not set by the command below on their own)."
echo ""
echo "Equivalent by hand:"
echo "  CGO_CFLAGS=\"-I\$JAVA_HOME/include -I\$JAVA_HOME/include/linux\" \\"
echo "  DEFRA_CLIENT_JAVA=true \\"
echo "  DEFRA_JAVA_JAR=\"$JAR_PATH\" \\"
echo "  go test -tags javaclient ./tests/integration/..."
