#!/bin/bash
# Run Go integration tests against Rust FFI and generate failure reports
# Usage: ./scripts/run-rust-ffi-tests.sh <package> [output_dir]
# Example: ./scripts/run-rust-ffi-tests.sh query/simple ./reports

set -e

PACKAGE="${1:-query/simple}"
OUTPUT_DIR="${2:-./reports}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${OUTPUT_DIR}/${PACKAGE//\//_}_${TIMESTAMP}.md"

mkdir -p "$OUTPUT_DIR"

echo "Running tests for: ./tests/integration/${PACKAGE}/..."
echo "Output: ${REPORT_FILE}"

# Run tests and capture output
RAW_OUTPUT=$(DEFRA_CLIENT_RUST_FFI=true CGO_ENABLED=1 go test "./tests/integration/${PACKAGE}/..." -v -count=1 2>&1) || true

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
