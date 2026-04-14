# Index: `tests/integration/collection_version/client_introspection`

## Overview

This folder verifies that the full client-facing GraphQL introspection query (the canonical query used by tools such as Altair, GraphiQL, and Postman) executes without error against DefraDB both on an empty schema and after collections with relational types have been registered. The embedded `.gql` file contains the standard 2023 introspection query shared across all tests in the package.

## Test Index

### `one_many_test.go`

Tests that the client introspection query works correctly when a one-to-many relational schema is present.

| Test Function | Line | Description |
|---|---|---|
| `TestClientIntrospectionWithOneToManyCollection` | 22-44 | A client introspection query succeeds after adding a one-to-many relational collection. |

### `simple_test.go`

Tests that the client introspection query works correctly against a default, schema-less DefraDB instance.

| Test Function | Line | Description |
|---|---|---|
| `TestClientIntrospectionBasic` | 24-34 | A standard client introspection query succeeds against an empty schema. |
