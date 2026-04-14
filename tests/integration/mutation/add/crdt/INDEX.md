# Index: `tests/integration/mutation/add/crdt`

## Overview

These tests verify that adding documents with CRDT counter fields (`pcounter` and `pncounter`) works correctly across supported numeric kinds (Int, Float32, Float64). Each test adds a single document with a positive counter value and confirms the stored value is returned accurately by a query. The tests exclude secondary-index multipliers due to a known limitation where accumulated CRDT fields cannot be indexed.

## Test Index

### `pcounter_test.go`

Tests that adding documents with `pcounter` CRDT fields correctly stores and retrieves positive values across Int, Float32, and Float64 numeric types.

| Test Function | Line | Description |
|---|---|---|
| `TestPCounterAdd_IntKindWithPositiveValue_NoError` | 22-63 | Adding a document with a pcounter Int field stores and returns the positive value. |
| `TestPCounterAdd_Float32KindWithPositiveValue_NoError` | 65-106 | Adding a document with a pcounter Float32 field stores and returns the positive value. |
| `TestPCounterAdd_Float64KindWithPositiveValue_NoError` | 108-149 | Adding a document with a pcounter Float64 field stores and returns the positive value. |

### `pncounter_test.go`

Tests that adding documents with `pncounter` CRDT fields correctly stores and retrieves positive values across Int, Float32, and Float64 numeric types.

| Test Function | Line | Description |
|---|---|---|
| `TestPNCounterAdd_IntKindWithPositiveValue_NoError` | 22-63 | Adding a document with a pncounter Int field stores and returns the positive value. |
| `TestPNCounterAdd_Float32KindWithPositiveValue_NoError` | 65-106 | Adding a document with a pncounter Float32 field stores and returns the positive value. |
| `TestPNCounterAdd_Float64KindWithPositiveValue_NoError` | 108-149 | Adding a document with a pncounter Float64 field stores and returns the positive value. |
