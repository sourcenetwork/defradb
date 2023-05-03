#!/bin/bash

PROJECT_DIR="./"

# Use the find command to search for all files within the project directory
# and exclude files with names containing '*_test*'
# and directories with names such as './tests', './node_modules', ...
find "$PROJECT_DIR" \
    -type f \
    -not -name "*_test*" \
    -not -path "*.md" \
    -not -path "*.git*" \
    -not -path "./.git/*" \
    -not -path "./docs/*" \
    -not -path "./tests/*" \
    -not -path "./.github/*" \
    -not -path "./licenses/*" \
    -not -path "./examples/*" \
    -not -path "./build/*" \
    -print0 | \
    xargs -0 cat | \
    wc -l
