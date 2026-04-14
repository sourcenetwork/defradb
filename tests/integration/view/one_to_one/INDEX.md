# Index: `tests/integration/view/one_to_one`

## Overview

This folder contains integration tests for views defined over one-to-one relations in DefraDB. The tests cover self-referential view schemas, persistence of embedded interface types across schema updates and node restarts, duplicate embedded schema error handling, and lens transforms applied to the outer type of a one-to-one view.

## Test Index

### `identical_schema_test.go`

Tests that a view can collapse a one-to-one self-referential relation into a single view type, and that embedded interface types survive subsequent schema updates without being lost.

| Test Function | Line | Description |
|---|---|---|
| `TestView_OneToOneSameSchema` | 21-97 | View collapses a one-to-one self-referential relation into a single view type. |
| `TestView_OneToOneEmbeddedSchemaIsNotLostOnNextUpdate` | 99-149 | Embedded view interface type is retained after a subsequent schema update. |

### `simple_test.go`

Tests the error behaviour when two views attempt to register an embedded interface type with the same name, and verifies that the first definition remains intact.

| Test Function | Line | Description |
|---|---|---|
| `TestView_OneToOneDuplicateEmbeddedSchema_Errors` | 21-96 | Creating a second view that redefines an existing embedded interface type errors. |

### `with_restart_test.go`

Tests that embedded interface types defined as part of a view are correctly reloaded and remain available following a node restart.

| Test Function | Line | Description |
|---|---|---|
| `TestView_OneToOneEmbeddedSchemaIsNotLostORestart` | 21-90 | Embedded view interface type is retained across a node restart. |

### `with_transform_test.go`

Tests that a lens transform applied to the outer type of a one-to-one view correctly maps source field values to the transformed destination field.

| Test Function | Line | Description |
|---|---|---|
| `TestView_OneToOneWithTransformOnOuter` | 25-115 | Lens transform on the outer type copies a field in a one-to-one view. |
