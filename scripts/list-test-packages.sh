#!/bin/bash
# List all available test packages with test counts
# Output format suitable for coordinator consumption

echo "# DefraDB Integration Test Packages"
echo ""
echo "| Package | Tests | Priority | Notes |"
echo "|---------|-------|----------|-------|"

# Core packages first (foundational)
for pkg in "query/simple" "mutation/create" "mutation/update" "mutation/delete"; do
  count=$(find "./tests/integration/${pkg}" -name "*_test.go" -exec grep -h "^func Test" {} \; 2>/dev/null | wc -l | tr -d ' ')
  echo "| \`${pkg}\` | ${count} | P0 | Foundation |"
done

echo ""

# Query packages
for pkg in "query/inline_array" "query/commits" "query/json" "query/one_to_many" "query/one_to_one"; do
  count=$(find "./tests/integration/${pkg}" -name "*_test.go" -exec grep -h "^func Test" {} \; 2>/dev/null | wc -l | tr -d ' ')
  echo "| \`${pkg}\` | ${count} | P1 | Query |"
done

echo ""

# Advanced query (relations)
for pkg in "query/one_to_many_to_one" "query/one_to_many_multiple" "query/one_to_one_to_many" "query/one_to_one_multiple" "query/many_to_many"; do
  count=$(find "./tests/integration/${pkg}" -name "*_test.go" -exec grep -h "^func Test" {} \; 2>/dev/null | wc -l | tr -d ' ')
  echo "| \`${pkg}\` | ${count} | P2 | Relations |"
done

echo ""

# Index and explain
for pkg in "index" "explain"; do
  count=$(find "./tests/integration/${pkg}" -name "*_test.go" -exec grep -h "^func Test" {} \; 2>/dev/null | wc -l | tr -d ' ')
  echo "| \`${pkg}\` | ${count} | P2 | Advanced |"
done

echo ""

# Views and subscriptions
for pkg in "view" "subscription" "collection"; do
  count=$(find "./tests/integration/${pkg}" -name "*_test.go" -exec grep -h "^func Test" {} \; 2>/dev/null | wc -l | tr -d ' ')
  echo "| \`${pkg}\` | ${count} | P3 | Features |"
done

echo ""
echo "## Usage"
echo ""
echo "First, set the path to defradb.rs:"
echo "\`\`\`bash"
echo "export DEFRA_RS_PATH=/path/to/defradb.rs"
echo "\`\`\`"
echo ""
echo "Run tests for a package:"
echo "\`\`\`bash"
echo "./scripts/run-rust-ffi-tests.sh query/simple ./reports"
echo "\`\`\`"
echo ""
echo "Run all P0 tests:"
echo "\`\`\`bash"
echo "for pkg in query/simple mutation/create mutation/update mutation/delete; do"
echo "  ./scripts/run-rust-ffi-tests.sh \$pkg ./reports"
echo "done"
echo "\`\`\`"
