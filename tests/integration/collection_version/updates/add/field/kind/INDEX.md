# Index: `tests/integration/collection_version/updates/add/field/kind`

## Overview

This folder contains integration tests that verify field-kind validation and data round-tripping when adding new fields to a collection version via JSON Patch. Tests cover every supported scalar kind (Bool, Int, Float32, Float64, DateTime, String, Blob, JSON, DocID) as well as their array variants (non-nullable and nillable), foreign-object (relation) fields, and invalid / unsupported kind values. For each supported kind, three patterns are typically tested: querying the empty schema after the field is added, adding a document and reading it back, and using the string-name kind substitution instead of the numeric kind code.

## Test Index

### `blob_test.go`

Tests that a Blob field can be added to a collection version by numeric kind or by string substitution, and that blob hex values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindBlob` | 21-53 | Adding a Blob field to a collection version succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindBlobWithAdd` | 55-99 | Adding a Blob field and inserting a document stores and retrieves the blob value correctly. |
| `TestCollectionVersionUpdatesAddFieldKindBlobSubstitutionWithAdd` | 101-145 | Adding a Blob field using string kind substitution stores and retrieves the blob value correctly. |

### `bool_array_test.go`

Tests that a non-nullable Boolean array field (`[Boolean!]`) can be added by numeric kind or by string substitution, and that boolean array values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindBoolArray` | 21-53 | Adding a non-nullable Boolean array field succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindBoolArrayWithAdd` | 55-99 | Adding a non-nullable Boolean array field and inserting a document stores and retrieves the values. |
| `TestCollectionVersionUpdatesAddFieldKindBoolArraySubstitutionWithAdd` | 101-145 | Adding a [Boolean!] field using string kind substitution stores and retrieves the array values. |

### `bool_nil_array_test.go`

Tests that a nillable Boolean array field (`[Boolean]`) can be added by numeric kind or by string substitution, and that nullable boolean array values (including null elements) are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindNillableBoolArray` | 23-55 | Adding a nillable Boolean array field succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindNillableBoolArrayWithAdd` | 57-101 | Adding a nillable Boolean array field and inserting a document stores and retrieves nullable values. |
| `TestCollectionVersionUpdatesAddFieldKindNillableBoolArraySubstitutionWithAdd` | 103-147 | Adding a [Boolean] field using string kind substitution stores and retrieves nullable array values. |

### `bool_test.go`

Tests that a Boolean field can be added to a collection version by numeric kind or by string substitution, and that boolean values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindBool` | 21-53 | Adding a Boolean field to a collection version succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindBoolWithAdd` | 55-99 | Adding a Boolean field and inserting a document stores and retrieves the boolean value correctly. |
| `TestCollectionVersionUpdatesAddFieldKindBoolSubstitutionWithAdd` | 101-145 | Adding a Boolean field using string kind substitution stores and retrieves the boolean value. |

### `datetime_test.go`

Tests that a DateTime field can be added to a collection version by numeric kind or by string substitution, and that RFC3339 timestamps are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindDateTime` | 21-53 | Adding a DateTime field to a collection version succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindDateTimeWithAdd` | 55-99 | Adding a DateTime field and inserting a document stores and retrieves the timestamp correctly. |
| `TestCollectionVersionUpdatesAddFieldKindDateTimeSubstitutionWithAdd` | 101-145 | Adding a DateTime field using string kind substitution stores and retrieves the timestamp correctly. |

### `doc_id_test.go`

Tests that a DocID field can be added to a collection version by numeric kind or by string substitution, and that document identifier values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindDocID` | 21-53 | Adding a DocID field to a collection version succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindDocIDWithAdd` | 55-99 | Adding a DocID field and inserting a document stores and retrieves the document identifier. |
| `TestCollectionVersionUpdatesAddFieldKindDocIDSubstitutionWithAdd` | 101-145 | Adding a DocID field using string kind substitution stores and retrieves the document identifier. |

### `float32_array_test.go`

Tests that a non-nullable Float32 array field (`[Float32!]`) can be added by numeric kind or by string substitution, and that float32 array values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindFloat32Array` | 21-53 | Adding a non-nullable Float32 array field succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindFloat32ArrayWithAdd` | 55-99 | Adding a non-nullable Float32 array field and inserting a document stores and retrieves the values. |
| `TestCollectionVersionUpdatesAddFieldKindFloat32ArraySubstitutionWithAdd` | 101-145 | Adding a [Float32!] field using string kind substitution stores and retrieves the array values. |

### `float32_nil_array_test.go`

Tests that a nillable Float32 array field (`[Float32]`) can be added by numeric kind or by string substitution, and that nullable float32 array values (including null elements) are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindNillableFloat32Array` | 23-55 | Adding a nillable Float32 array field succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindNillableFloat32ArrayWithAdd` | 57-105 | Adding a nillable Float32 array field and inserting a document stores and retrieves nullable values. |
| `TestCollectionVersionUpdatesAddFieldKindNillableFloat32ArraySubstitutionWithAdd` | 107-155 | Adding a [Float32] field using string kind substitution stores and retrieves nullable array values. |

### `float32_test.go`

Tests that a Float32 field can be added to a collection version by numeric kind or by string substitution, and that float32 values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindFloat32` | 21-53 | Adding a Float32 field to a collection version succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindFloat32WithAdd` | 55-99 | Adding a Float32 field and inserting a document stores and retrieves the float32 value correctly. |
| `TestCollectionVersionUpdatesAddFieldKindFloat32SubstitutionWithAdd` | 101-145 | Adding a Float32 field using string kind substitution stores and retrieves the float32 value. |

### `float_array_test.go`

Tests that a non-nullable Float64 array field (`[Float!]`) can be added by numeric kind or by string substitution, and that float64 array values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindFloatArray` | 21-53 | Adding a non-nullable Float array field succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindFloatArrayWithAdd` | 55-99 | Adding a non-nullable Float array field and inserting a document stores and retrieves float values. |
| `TestCollectionVersionUpdatesAddFieldKindFloatArraySubstitutionWithAdd` | 101-145 | Adding a [Float!] field using string kind substitution stores and retrieves the float array values. |

### `float_nil_array_test.go`

Tests that a nillable Float64 array field (`[Float]`) can be added by numeric kind or by string substitution, and that nullable float64 array values (including null elements) are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindNillableFloatArray` | 23-55 | Adding a nillable Float array field succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindNillableFloatArrayWithAdd` | 57-105 | Adding a nillable Float array field and inserting a document stores and retrieves nullable values. |
| `TestCollectionVersionUpdatesAddFieldKindNillableFloatArraySubstitutionWithAdd` | 107-155 | Adding a [Float] field using string kind substitution stores and retrieves nullable float array values. |

### `float_test.go`

Tests that a Float64 field can be added to a collection version by numeric kind or by string substitution, and that float64 values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindFloat` | 21-53 | Adding a Float64 field to a collection version succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindFloatWithAdd` | 55-99 | Adding a Float64 field and inserting a document stores and retrieves the float64 value correctly. |
| `TestCollectionVersionUpdatesAddFieldKindFloatSubstitutionWithAdd` | 101-145 | Adding a Float64 field using string kind substitution stores and retrieves the float64 value. |

### `foreign_object_array_test.go`

Tests error cases when adding a foreign object array (one-to-many relation) field: unknown collection name, missing relation name, and a known collection with an incomplete relation definition.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindForeignObjectArray_UnknownCollection` | 21-45 | Adding a foreign object array field with an unknown collection name returns an error. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObjectArray_NoRelationName` | 47-71 | Adding a foreign object array field without a relation name returns an error. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObjectArray_KnownCollection` | 73-97 | Adding a foreign object array field with a known collection but incomplete relation returns an error. |

### `foreign_object_test.go`

Tests adding foreign object (relation) fields: covers numeric kind errors, unknown collection names, invalid ID companion field configurations, a successful self-referential one-to-one relation, and cross-collection one-to-one and one-to-many relations added in a single patch batch.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindForeignObject` | 21-43 | Adding a foreign object field with an unresolvable numeric kind returns a no-type-found error. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObject_UnknownCollection` | 45-69 | Adding a foreign object field referencing an unknown collection name returns an error. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObject_IDFieldMissingKind` | 71-96 | Adding a relational field when the ID companion field has no kind returns an invalid kind error. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObject_IDFieldInvalidKind` | 98-123 | Adding a relational field when the ID companion field has a non-DocID kind returns an error. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObject_Succeeds` | 125-187 | Adding a valid one-to-one relational field with a proper ID companion field succeeds. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObject_WithPatchAddingOneToOneRelationInSameBatch_ShouldSucceed` | 189-274 | Adding a one-to-one relation across two collections in a single patch batch succeeds. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObject_WithPatchAddingOneToManyRelationInSameBatch_ShouldSucceed` | 276-375 | Adding a one-to-many relation across two collections in a single patch batch succeeds. |

### `int_array_test.go`

Tests that a non-nullable Int array field (`[Int!]`) can be added by numeric kind or by string substitution, and that integer array values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindIntArray` | 21-53 | Adding a non-nullable Int array field succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindIntArrayWithAdd` | 55-99 | Adding a non-nullable Int array field and inserting a document stores and retrieves integer values. |
| `TestCollectionVersionUpdatesAddFieldKindIntArraySubstitutionWithAdd` | 101-145 | Adding a [Int!] field using string kind substitution stores and retrieves the integer array values. |

### `int_nil_array_test.go`

Tests that a nillable Int array field (`[Int]`) can be added by numeric kind or by string substitution, and that nullable integer array values (including null elements) are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindNillableIntArray` | 23-55 | Adding a nillable Int array field succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindNillableIntArrayWithAdd` | 57-105 | Adding a nillable Int array field and inserting a document stores and retrieves nullable int values. |
| `TestCollectionVersionUpdatesAddFieldKindNillableIntArraySubstitutionWithAdd` | 107-155 | Adding a [Int] field using string kind substitution stores and retrieves nullable integer array values. |

### `int_test.go`

Tests that an Int field can be added to a collection version by numeric kind or by string substitution, and that integer values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindInt` | 21-53 | Adding an Int field to a collection version succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindIntWithAdd` | 55-99 | Adding an Int field and inserting a document stores and retrieves the integer value correctly. |
| `TestCollectionVersionUpdatesAddFieldKindIntSubstitutionWithAdd` | 101-145 | Adding an Int field using string kind substitution stores and retrieves the integer value. |

### `invalid_test.go`

Tests that unsupported numeric kind values (reserved, out-of-range, or high arbitrary values) and unrecognized string kind names are rejected with appropriate errors.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKind15` | 21-43 | Adding a field with reserved numeric kind 15 returns a no-type-found error. |
| `TestCollectionVersionUpdatesAddFieldKind25` | 47-69 | Adding a field with the first unsupported numeric kind above the valid range returns an error. |
| `TestCollectionVersionUpdatesAddFieldKind198` | 73-95 | Adding a field with a high arbitrary unsupported numeric kind returns a no-type-found error. |
| `TestCollectionVersionUpdatesAddFieldKindInvalid` | 97-119 | Adding a field with an unrecognized string kind value returns a no-type-found error. |

### `json_test.go`

Tests that a JSON field can be added to a collection version by numeric kind or by string substitution, and that JSON object values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindJSON` | 21-53 | Adding a JSON field to a collection version succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindJSONWithAdd` | 55-99 | Adding a JSON field and inserting a document stores and retrieves the JSON object correctly. |
| `TestCollectionVersionUpdatesAddFieldKindJSONSubstitutionWithAdd` | 101-145 | Adding a JSON field using string kind substitution stores and retrieves the JSON object correctly. |

### `none_test.go`

Tests that kind 0 (unset/none) is rejected since it has no valid type mapping.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindNone` | 21-43 | Adding a field with kind 0 (unset/none) returns a no-type-found error. |

### `string_array_test.go`

Tests that a non-nullable String array field (`[String!]`) can be added by numeric kind or by string substitution, and that string array values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindStringArray` | 21-53 | Adding a non-nullable String array field succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindStringArrayWithAdd` | 55-99 | Adding a non-nullable String array field and inserting a document stores and retrieves string values. |
| `TestCollectionVersionUpdatesAddFieldKindStringArraySubstitutionWithAdd` | 101-145 | Adding a [String!] field using string kind substitution stores and retrieves the string array values. |

### `string_nil_array_test.go`

Tests that a nillable String array field (`[String]`) can be added by numeric kind or by string substitution, and that nullable string array values (including null elements) are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindNillableStringArray` | 23-55 | Adding a nillable String array field succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindNillableStringArrayWithAdd` | 57-105 | Adding a nillable String array field and inserting a document stores and retrieves nullable values. |
| `TestCollectionVersionUpdatesAddFieldKindNillableStringArraySubstitutionWithAdd` | 107-155 | Adding a [String] field using string kind substitution stores and retrieves nullable string array values. |

### `string_test.go`

Tests that a String field can be added to a collection version by numeric kind or by string substitution, and that string values are stored and retrieved correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindString` | 21-53 | Adding a String field to a collection version succeeds and the field is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindStringWithAdd` | 55-99 | Adding a String field and inserting a document stores and retrieves the string value correctly. |
| `TestCollectionVersionUpdatesAddFieldKindStringSubstitutionWithAdd` | 101-145 | Adding a String field using string kind substitution stores and retrieves the string value. |
