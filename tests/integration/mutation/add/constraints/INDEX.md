# Index: `tests/integration/mutation/add/constraints`

## Overview

This folder contains integration tests for schema-level field constraints applied during document mutation (add). The tests verify that DefraDB enforces constraints such as array size limits when creating documents. Both the success path (valid constraint satisfaction) and error path (constraint violation) are covered.

## Test Index

### `size_constraint_test.go`

Tests that the `@constraints(size: N)` directive correctly allows or rejects array fields based on element count during document creation.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAdd_WithSizeConstrain_ShouldSucceed` | 21-61 | Adding a document with an array matching the size constraint succeeds. |
| `TestMutationAdd_WithSizeConstrainMismatch_ShouldError` | 63-86 | Adding a document with an array exceeding the size constraint returns an error. |
