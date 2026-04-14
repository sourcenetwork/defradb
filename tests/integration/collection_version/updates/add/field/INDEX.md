# Index: `tests/integration/collection_version/updates/add/field`

## Overview

This directory contains integration tests for adding new fields to a collection version via JSON Patch. The subdirectories collectively verify three orthogonal dimensions of field addition: the field kind (every supported scalar, array, and relational type, plus invalid kinds), the CRDT type assigned to the field (LWW, PNCounter, PNCounter, none/default, and unsupported types), and field constraints (such as the size constraint for array fields). Together they ensure that field additions are validated thoroughly and that successfully added fields behave correctly at query and write time.

## Subdirectories

| Directory | Summary |
|---|---|
| [`constraint/`](constraint/INDEX.md) | Tests that field constraints (e.g., the `size` array-length constraint) are validated and enforced when adding new fields. |
| [`crdt/`](crdt/INDEX.md) | Tests CRDT-type validation when adding new fields, covering supported types (LWW, PNCounter, PCounter, default) and error cases for unsupported or mismatched types. |
| [`kind/`](kind/INDEX.md) | Tests field-kind validation and data round-tripping for every supported scalar, array, and relational kind, as well as rejection of invalid kind values. |
