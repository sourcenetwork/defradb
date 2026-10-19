#!/usr/bin/env bash

# Download static assets from: `github.com/sourcenetwork/defradb-explorer`.
#
# Bump the release tag in the URL below to change versions.

curl -fsSL https://github.com/sourcenetwork/defradb-explorer/releases/download/v1.0.0/dist.tar.gz | tar xzf -
