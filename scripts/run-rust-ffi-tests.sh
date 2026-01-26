#!/bin/bash
# Run Go integration tests against Rust FFI and generate failure reports
# Usage: ./scripts/run-rust-ffi-tests.sh <package> [output_dir]
# Example: ./scripts/run-rust-ffi-tests.sh query/simple ./reports
#
# Required environment variable:
#   DEFRA_RS_PATH - path to defradb.rs repo (e.g., /path/to/defradb.rs)

set -e

# Check for required env var
if [ -z "$DEFRA_RS_PATH" ]; then
    echo "Error: DEFRA_RS_PATH environment variable must be set"
    echo "Example: export DEFRA_RS_PATH=/path/to/defradb.rs"
    exit 1
fi

# Verify the path exists
if [ ! -d "$DEFRA_RS_PATH/crates/ffi" ]; then
    echo "Error: $DEFRA_RS_PATH/crates/ffi not found"
    echo "Make sure DEFRA_RS_PATH points to the defradb.rs repo root"
    exit 1
fi

# Set CGO flags
export CGO_CFLAGS="-I$DEFRA_RS_PATH/crates/ffi"
export CGO_LDFLAGS="-L$DEFRA_RS_PATH/target/release"
export CGO_ENABLED=1
export DEFRA_CLIENT_RUST_FFI=true

PACKAGE="${1:-query/simple}"
OUTPUT_DIR="${2:-./reports}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${OUTPUT_DIR}/${PACKAGE//\//_}_${TIMESTAMP}.md"

mkdir -p "$OUTPUT_DIR"

echo "Running tests for: ./tests/integration/${PACKAGE}/..."
echo "Using Rust FFI from: $DEFRA_RS_PATH"
echo "Output: ${REPORT_FILE}"

# Run tests and capture output
RAW_OUTPUT=$(go test "./tests/integration/${PACKAGE}/..." -v -count=1 2>&1) || true

# Count results
TOTAL=$(echo "$RAW_OUTPUT" | grep -cE "^(---|===) (RUN|PASS|FAIL)" || echo 0)
PASSED=$(echo "$RAW_OUTPUT" | grep -c "^--- PASS" || echo 0)
FAILED=$(echo "$RAW_OUTPUT" | grep -c "^--- FAIL" || echo 0)

# Generate report
cat > "$REPORT_FILE" << EOF
# Rust FFI Test Report: ${PACKAGE}

**Generated:** $(date)
**Test Package:** \`./tests/integration/${PACKAGE}/...\`

## Summary

| Status | Count |
|--------|-------|
| Total  | ${TOTAL} |
| Passed | ${PASSED} |
| Failed | ${FAILED} |

## Failed Tests

EOF

# Extract failing tests with their errors
echo "$RAW_OUTPUT" | awk '
/^=== RUN/ {
    if (in_fail && buffer != "") {
        print "```"; print "";
    }
    in_fail = 0; buffer = ""; current_test = $3
}
/Error:/ || /error:/ || /panic/ { buffer = buffer $0 "\n" }
/^--- FAIL:/ {
    print "### " current_test
    print "```"
    if (buffer != "") print buffer
    else print "(no error details captured)"
    print "```"
    print ""
    in_fail = 0; buffer = ""
}
' >> "$REPORT_FILE"

# Add raw error patterns for quick reference
cat >> "$REPORT_FILE" << EOF

## Error Patterns

\`\`\`
EOF

echo "$RAW_OUTPUT" | grep -E "Error:|parse error|not implemented|panic" | sort | uniq -c | sort -rn | head -20 >> "$REPORT_FILE"

echo '```' >> "$REPORT_FILE"

echo ""
echo "=== Summary ==="
echo "Package: ${PACKAGE}"
echo "Passed: ${PASSED}/${TOTAL}"
echo "Failed: ${FAILED}/${TOTAL}"
echo "Report: ${REPORT_FILE}"
