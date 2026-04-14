# Index: `tests/integration/collection_version/updates/add`

## Overview

This directory contains integration tests for JSON Patch `add` operations applied to collection versions via `PatchCollection`. The subdirectories verify that new fields can be appended to an existing collection schema with correct kind, CRDT type, and constraint settings, and that invalid combinations are rejected with appropriate errors.

## Subdirectories

| Directory | Summary |
|---|---|
| [`field/`](field/INDEX.md) | Tests adding new fields to a collection version, covering field kind, CRDT type, and constraint validation. |
