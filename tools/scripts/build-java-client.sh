#!/usr/bin/env bash
set -euo pipefail

# Builds defradb.jar for the tests/clients/java client: clones (or updates)
# defradb-java-sdk into a local, gitignored checkout, pins it to the reviewed
# commit below (so identical DefraDB commits always build against the same SDK
# / ABI code), then runs its own
# build.sh, which in turn:
#   1. runs `make build-c-shared-linux` in this repo to build libdefradb.so
#   2. copies that .so + headers into the checkout
#   3. compiles nativewrapper.c against it (its own src/main/c/build.sh)
#   4. runs its Gradle build to produce build/libs/defradb.jar
#
# This only works on Linux (build-c-shared-linux cross-compiles nothing - it
# needs a real Linux gcc/cgo toolchain, and defradb-java-sdk's build.sh shells
# out to `make`/bash throughout) - run it from WSL on Windows.
#
# Requires: git, make, go, gcc, and a JDK (for Gradle) on PATH.

DEFRA_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WRAPPER_DIR="${DEFRA_JAVA_WRAPPER_DIR:-$DEFRA_DIR/.javaclient/defradb-java-sdk}"
WRAPPER_REPO="${DEFRA_JAVA_WRAPPER_REPO:-https://github.com/sourcenetwork/defradb-java-sdk.git}"

# The commit of the Java SDK to pin to
WRAPPER_COMMIT="${DEFRA_JAVA_WRAPPER_COMMIT:-92c52ca7b8571feea0c0b7373c0441cbbd24b749}"

if [ ! -d "$WRAPPER_DIR/.git" ]; then
  echo "Cloning defradb-java-sdk repo into $WRAPPER_DIR..."
  git clone "$WRAPPER_REPO" "$WRAPPER_DIR"
else
  echo "Updating existing defradb-java-sdk checkout at $WRAPPER_DIR..."
  git -C "$WRAPPER_DIR" fetch --quiet origin
fi

# Fetch the pinned commit explicitly in case it is not reachable from a branch head, then check it
# out detached rather than following any moving branch.
if ! git -C "$WRAPPER_DIR" cat-file -e "$WRAPPER_COMMIT^{commit}" 2>/dev/null; then
  git -C "$WRAPPER_DIR" fetch --quiet origin "$WRAPPER_COMMIT"
fi
git -C "$WRAPPER_DIR" checkout --quiet --detach "$WRAPPER_COMMIT"

ACTUAL_COMMIT="$(git -C "$WRAPPER_DIR" rev-parse HEAD)"
if [ "$ACTUAL_COMMIT" != "$WRAPPER_COMMIT" ]; then
  echo "defradb-java-sdk is at $ACTUAL_COMMIT, expected pinned commit $WRAPPER_COMMIT" >&2
  exit 1
fi
echo "Using defradb-java-sdk commit $ACTUAL_COMMIT"

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
echo "  CGO_ENABLED=1 \\"
echo "  CGO_CFLAGS=\"-I\$JAVA_HOME/include -I\$JAVA_HOME/include/linux\" \\"
echo "  DEFRA_CLIENT_JAVA=true \\"
echo "  DEFRA_JAVA_JAR=\"$JAR_PATH\" \\"
echo "  go test -tags javaclient ./tests/integration/..."
