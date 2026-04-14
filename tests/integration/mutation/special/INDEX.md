# Index: `tests/integration/mutation/special`

## Overview

This folder contains integration tests for edge-case and invalid mutation scenarios in DefraDB that do not fit neatly into the standard add, update, delete, or upsert categories. Currently it covers mutations that use operation names not recognized by the generated GraphQL schema, verifying that the database returns the correct schema validation error.

## Test Index

### `invalid_operation_test.go`

Tests that mutations referencing non-existent or malformed operation names are rejected with an appropriate GraphQL schema error.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationInvalidMutation` | 21-44 | Mutation with an unrecognized operation name returns a schema error. |
