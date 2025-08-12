#!/usr/bin/env bash
# Usage: ./replace.sh "search_string" "replace_string"

search="package cbindings"
replace="package main"

# Recursively find all regular files and replace in place
find . -type f -name '*.go' ! -path '*/.git/*' ! -path '*/vendor/*' \
  -exec sed -i '' "s/${search//\//\\/}/${replace//\//\\/}/g" {} +

go build --buildmode=c-archive .

# Recursively find all regular files and replace in place
find . -type f -name '*.go' ! -path '*/.git/*' ! -path '*/vendor/*' \
  -exec sed -i '' "s/${replace//\//\\/}/${search//\//\\/}/g" {} +