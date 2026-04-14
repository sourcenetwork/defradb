# Index: `tests/integration/mutation/add`

## Overview

This directory contains integration tests for the `add` (create/insert) mutation in DefraDB. The direct test files cover fundamental add behaviour — basic document creation across one or multiple documents in a single call, error handling for unknown fields and duplicate documents, empty and null input handling, GQL variable support, `@default` directive resolution across all supported field kinds, and retrieval of the commit CID after insertion. The subdirectories extend this coverage to schema-level constraints (`@constraints`), CRDT counter fields (`pcounter`/`pncounter`), automatic vector embedding generation (`@embedding`), and relational field kinds (one-to-many, one-to-one, and chained one-to-one-to-one relations).

## Test Index

### `simple_test.go`

Basic document creation tests covering the success path, non-existent field errors, duplicate document errors, empty input, and multi-collection schemas.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAdd_GivenNonExistantField_Errors` | 24-64 | Adding a document with a field that does not exist on the schema returns an error. |
| `TestMutationAdd` | 66-107 | Adding a document with valid fields stores and returns the document with its docID. |
| `TestMutationAdd_GivenDuplicate_Errors` | 109-143 | Adding a document whose content produces a docID that already exists returns an error. |
| `TestMutationAdd_GivenEmptyInput` | 145-173 | Adding a document with an empty input object succeeds and returns a stable docID. |
| `TestMutationAdd_With10Collections` | 175-235 | Adding a document works correctly when the schema defines ten collections. |

### `simple_add_many_test.go`

Tests that a batch JSON array input adds multiple documents in a single call and returns all stored documents.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAddMany` | 21-73 | Adding multiple documents via a JSON array in one call stores all documents correctly. |

### `with_default_values_test.go`

Tests for the `@default` directive during document creation, covering static defaults for all supported field types, the `UTC_NOW` timestamp sentinel, explicit nil overrides, user-supplied value precedence, duplicate-document error paths, unique-index enforcement, and JSON scalar/object default values.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAdd_WithDefaultValues_NoValuesProvided_SetsDefaultValue` | 25-81 | Omitting all fields uses each field's `@default` value for the created document. |
| `TestMutationAdd_WithDefaultValues_NoValuesProvided_SetsUTCNowDefaultValue` | 83-115 | A `DateTime` field with `@default(dateTime: UTC_NOW)` receives the current UTC time. |
| `TestMutationAdd_WithDefaultValues_NilValuesProvided_SetsNilValue` | 117-182 | Explicitly passing nil for a field with a `@default` stores nil rather than the default. |
| `TestMutationAdd_WithDefaultValues_ValuesProvided_SetsValue` | 184-249 | Providing explicit values for fields with `@default` stores the provided values. |
| `TestMutationAdd_WithDefaultValue_NoValueProvided_AddedTwice_ReturnsError` | 251-283 | Adding two documents with identical default values returns a duplicate-docID error. |
| `TestMutationAdd_WithDefaultValue_NoValueProvided_AddedTwice_UniqueIndex_ReturnsError` | 285-318 | Adding a second document that shares a unique-indexed default value returns a unique-index violation error. |
| `TestMutationAdd_WithDefaultJSONIntValue_ShouldBeSet` | 320-354 | A JSON field with an integer `@default` stores the integer value. |
| `TestMutationAdd_WithDefaultJSONFloatValue_ShouldBeSet` | 356-390 | A JSON field with a float `@default` stores the float value. |
| `TestMutationAdd_WithDefaultJSONBoolValue_ShouldBeSet` | 392-426 | A JSON field with a boolean `@default` stores the boolean value. |
| `TestMutationAdd_WithDefaultJSONNullValue_ReturnError` | 428-444 | Defining `@default(json: null)` on a field during schema registration returns an error. |
| `TestMutationAdd_WithDefaultJSONObjectValues_ShouldBeSet` | 446-480 | A JSON field with an object `@default` stores the serialised object. |
| `TestMutationAdd_WithDefaultJSONDeepObjectValue_ShouldBeSet` | 482-516 | A JSON field with a deeply nested object `@default` stores the full serialised structure. |
| `TestMutationAdd_WithDefaultValues_NoValuesProvided_SetsTwoEqualUTCNowDefaultValue` | 518-550 | Two documents added in the same mutation with `UTC_NOW` defaults receive the same timestamp. |

### `with_null_input_test.go`

Tests that null values for the `encrypt`, `encryptFields`, and `input` GQL arguments are handled gracefully.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAdd_WithNullEncrypt_Succeeds` | 21-49 | Passing `encrypt: null` to an add mutation succeeds and inserts the document. |
| `TestMutationAdd_WithNullInput_Succeeds` | 51-75 | Passing `input: null` to an add mutation succeeds and returns an empty result set. |
| `TestMutationAdd_WithNullInputEntry_ReturnsError` | 77-99 | Passing a null element inside the input array returns a GraphQL type error. |
| `TestMutationAdd_WithNullEncryptFields_Succeeds` | 101-129 | Passing `encryptFields: null` to an add mutation succeeds and inserts the document. |

### `with_null_value_test.go`

Tests that explicitly providing null for an optional field and omitting it produce the same docID, making a second add a duplicate.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAdd_WithOmittedValueAndExplicitNullValue` | 24-57 | Adding a document with an omitted field and then a document with that field set to null produces a duplicate-docID error. |

### `with_variables_test.go`

Tests for GQL variable support in add mutations, covering non-null variable types, default variable values, and variables embedded in JSON object fields.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAddWithNonNullVariable` | 23-56 | An add mutation using a non-null variable binding inserts the document correctly. |
| `TestMutationAddWithDefaultVariable` | 58-86 | An add mutation using a variable with an inline GQL default inserts the document correctly. |
| `TestMutationAdd_WithVariableInJSONObject_Succeeds` | 88-121 | A variable interpolated inside a JSON object literal in the input is stored correctly. |
| `TestMutationAdd_WithJSONVariable_Succeeds` | 123-158 | A JSON-typed variable passed directly to a JSON field is stored as the field value. |

### `with_version_test.go`

Tests that the `_version { cid }` field returned by an add mutation contains the correct commit CID for the new document.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAdd_ReturnsVersionCID` | 21-55 | An add mutation selecting `_version { cid }` returns the expected CID for the new commit. |

## Subdirectories

| Directory | Summary |
|---|---|
| [`constraints/`](constraints/INDEX.md) | Tests that the `@constraints(size: N)` directive correctly allows or rejects array fields based on element count during document creation. |
| [`crdt/`](crdt/INDEX.md) | Tests that adding documents with `pcounter` and `pncounter` CRDT fields correctly stores and retrieves positive values across Int, Float32, and Float64 numeric types. |
| [`embeddings/`](embeddings/INDEX.md) | Tests that the `@embedding` directive correctly triggers automatic vector embedding generation during document creation or skips it when the user supplies the vector directly. |
| [`field_kinds/`](field_kinds/INDEX.md) | Tests for creating and linking documents across one-to-many, one-to-one, and one-to-one-to-one relations using both raw foreign-key IDs and aliased relation names. |
