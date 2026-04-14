# Index: `tests/integration/collection_version/updates/add/field/crdt`

## Overview

This folder contains integration tests that verify CRDT-type validation and behaviour when adding new fields to a collection version via JSON Patch. Tests cover the full range of supported CRDT types (LWW, PNCounter, PCounter, and the implicit default/none), as well as error cases for unsupported types such as `composite`, `object`, and completely unknown numeric type codes, and kind-mismatch errors when a counter CRDT is paired with an incompatible field kind.

## Test Index

### `composite_test.go`

Tests that adding a field with the `composite` CRDT type is rejected, both for a single field and for multiple fields in the same patch.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldCRDTCompositeErrors` | 21-43 | Adding a field with composite CRDT type returns an unsupported CRDT error. |
| `TestCollectionVersionUpdatesAddFieldCRDTCompositeErrorsMultiple` | 45-68 | Adding multiple fields with composite CRDT type returns aggregated unsupported CRDT errors. |

### `invalid_test.go`

Tests that adding a field with a completely unknown CRDT type code (99) is rejected, both for a single field and for multiple fields in the same patch.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldCRDTInvalidErrors` | 21-43 | Adding a field with an unknown CRDT type returns an unsupported CRDT error. |
| `TestCollectionVersionUpdatesAddFieldCRDTInvalidErrorsMultiple` | 45-68 | Adding multiple fields with unknown CRDT types returns aggregated unsupported CRDT errors. |

### `lww_test.go`

Tests that adding a field with the LWW (Last-Write-Wins) CRDT type succeeds and the field becomes queryable.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldCRDTLWW` | 21-53 | Adding a field with LWW CRDT type succeeds and the field is queryable. |

### `none_test.go`

Tests that adding a field with no explicit CRDT type (default) or with type 0 succeeds and leaves the field queryable.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldCRDTDefault` | 21-53 | Adding a field without specifying a CRDT type uses the default and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldCRDTNone` | 55-87 | Adding a field with CRDT type 0 (none) succeeds and the field is queryable. |

### `object_bool_test.go`

Tests that adding a Boolean field with the `object` CRDT type is rejected, both for a single field and for multiple fields in the same patch.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldCRDTObjectWithBoolFieldErrors` | 21-43 | Adding a Boolean field with object CRDT type returns an unsupported CRDT error. |
| `TestCollectionVersionUpdatesAddFieldCRDTObjectWithBoolFieldErrorsMultiple` | 45-68 | Adding multiple Boolean fields with object CRDT type returns aggregated unsupported CRDT errors. |

### `pcounter_test.go`

Tests the PCounter (positive-only counter) CRDT type: confirms it succeeds with a compatible Int field and fails with an incompatible Boolean field.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdates_AddFieldCRDTPCounter_NoError` | 21-53 | Adding an Int field with pcounter CRDT type succeeds and the field is queryable. |
| `TestCollectionVersionUpdates_AddFieldCRDTPCounterWithMismatchKind_Error` | 55-77 | Adding a Boolean field with pcounter CRDT type returns a kind mismatch error. |

### `pncounter_test.go`

Tests the PNCounter (positive-negative counter) CRDT type: confirms it succeeds with a compatible Int field and fails with an incompatible Boolean field.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdates_AddFieldCRDTPNCounter_NoError` | 21-53 | Adding an Int field with pncounter CRDT type succeeds and the field is queryable. |
| `TestCollectionVersionUpdates_AddFieldCRDTPNCounterWithMismatchKind_Error` | 55-77 | Adding a Boolean field with pncounter CRDT type returns a kind mismatch error. |
