# Index: `tests/integration/collection_version/updates/add/field/constraint`

## Overview

This folder contains integration tests that verify field constraint validation when adding new fields to a collection version via JSON Patch. Currently it covers the `size` constraint, which limits the number of elements allowed in an array-typed field and is enforced at document-write time.

## Test Index

### `size_test.go`

Tests that a `Size` constraint added to a new array field is enforced when documents are inserted, rejecting arrays whose length does not match the declared size.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdates_AddFieldSizeContraint_ShouldSucceed` | 21-50 | Adding a field with a size constraint enforces the array size limit on document writes. |
