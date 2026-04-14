# Index: `tests/integration/mutation/update/constraints`

## Overview

This folder contains integration tests that verify DefraDB enforces field-level schema constraints during document update mutations. The tests confirm that updates respecting a constraint succeed and that updates violating a constraint are rejected with an appropriate error.

## Test Index

### `size_constraint_test.go`

Tests that the `@constraints(size: N)` directive is enforced when updating array fields, covering both the success path and the error path for mismatched array lengths.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdate_WithSizeConstrain_ShouldSucceed` | 21-66 | Update succeeds when array value matches the size constraint. |
| `TestMutationUpdate_WithSizeConstrainMismatch_ShouldError` | 68-96 | Update errors when updated array length violates the size constraint. |
