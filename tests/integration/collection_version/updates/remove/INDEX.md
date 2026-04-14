# Index: `tests/integration/collection_version/updates/remove`

## Overview

This directory tests the behaviour of JSON-Patch `remove` operations applied to collection schema versions via `PatchCollection`. The tests cover removing individual fields, removing all fields at once, and attempts to remove individual field properties (such as `Name`, `Kind`, and `Typ`), which are rejected because mutating field metadata in-place is not supported.

## Subdirectories

| Directory | Summary |
|---|---|
| [`fields/`](fields/INDEX.md) | Tests for JSON-Patch `remove` operations on collection fields, covering valid field removals and invalid attempts to remove individual field properties. |
