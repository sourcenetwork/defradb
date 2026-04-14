# Index: `tests/integration/collection_version`

## Overview

This directory contains integration tests for the `AddCollection` and `PatchCollection` APIs, covering the full lifecycle of collection schema versions in DefraDB. The tests validate that collections can be defined with any combination of scalar, relational, inline-array, CRDT-typed, and @embedding fields; that GraphQL introspection correctly reflects the schema (filter, groupBy, order, similarity, and mutation input types); that the `@branchable` and `@default` directives behave correctly; that branching, active-version switching, and `SetActiveCollectionVersion` work as expected; and that `GetCollections` correctly filters by version ID, collection ID, and name across active and inactive versions. The subdirectories cover aggregate introspection, client-facing introspection queries, lens-based migrations, and JSON Patch updates to collection versions.

## Test Index

### `add_one_one_data_test.go`

Tests that mutation input types for one-to-one relations are correctly shaped — the primary side exposes the foreign-key ID field while the secondary side does not.

| Test Function | Line | Description |
|---|---|---|
| `TestAddOneToOne_Input_PrimaryObject` | 21-89 | Adding a one-to-one collection exposes the foreign key ID field on the primary side's mutation input type. |
| `TestAddOneToOne_Input_SecondaryObject` | 91-146 | Adding a one-to-one collection does not expose relation fields on the secondary side's mutation input type. |

### `branchable_directive_test.go`

Tests that the `@branchable` directive correctly sets the `IsBranchable` flag on the registered collection version.

| Test Function | Line | Description |
|---|---|---|
| `TestColVersionBranchable_NoArguments_DefaultTrue` | 22-43 | The @branchable directive without arguments defaults the collection to branchable. |
| `TestColVersionBranchable_ArgumentIfTrue_ShouldBeTrue` | 45-66 | The @branchable directive with if: true marks the collection as branchable. |
| `TestColVersionBranchable_ArgumentIfFalse_ShouldBeFalse` | 68-89 | The @branchable directive with if: false marks the collection as not branchable. |

### `client_test.go`

Tests that the GraphQL schema introspection exposes system-level types such as the `ExplainType` enum.

| Test Function | Line | Description |
|---|---|---|
| `TestIntrospectionExplainTypeDefined` | 23-55 | The GraphQL schema introspection exposes the ExplainType enum with correct kind and description. |

### `crdt_type_test.go`

Tests that PN counter and P counter CRDT types can be assigned to numeric fields and are rejected for incompatible field kinds.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionAdd_ContainsPNCounterTypeWithIntKind_NoError` | 23-59 | Adding a collection with a PN counter field of Int kind succeeds and stores the correct CRDT type. |
| `TestCollectionAdd_ContainsPNCounterTypeWithFloatKind_NoError` | 61-97 | Adding a collection with a PN counter field of Float kind succeeds and stores the correct CRDT type. |
| `TestCollectionAdd_ContainsPNCounterTypeWithWrongKind_Error` | 99-118 | Adding a collection with a PN counter field on a non-numeric kind returns an error. |
| `TestCollectionAdd_ContainsPNCounterWithInvalidType_Error` | 120-139 | Adding a collection with an invalid CRDT type string returns a GraphQL argument error. |
| `TestCollectionAdd_ContainsPCounterTypeWithIntKind_NoError` | 141-177 | Adding a collection with a P counter field of Int kind succeeds and stores the correct CRDT type. |
| `TestCollectionAdd_ContainsPCounterTypeWithFloatKind_NoError` | 179-215 | Adding a collection with a P counter field of Float kind succeeds and stores the correct CRDT type. |
| `TestCollectionAdd_ContainsPCounterTypeWithFloat64Kind_NoError` | 217-253 | Adding a collection with a P counter field of Float64 kind succeeds and stores the correct CRDT type. |
| `TestCollectionAdd_ContainsPCounterTypeWithFloat32Kind_NoError` | 255-291 | Adding a collection with a P counter field of Float32 kind succeeds and stores the correct CRDT type. |
| `TestCollectionAdd_ContainsPCounterTypeWithWrongKind_Error` | 293-309 | Adding a collection with a P counter field on a non-numeric kind returns an error. |

### `filter_test.go`

Tests that the GraphQL filter argument for a collection is correctly shaped, including field-specific operator blocks and relational nested filter types.

| Test Function | Line | Description |
|---|---|---|
| `TestFilterForSimpleCollection` | 21-132 | Adding a simple collection exposes a correctly typed filter argument with field-specific operator blocks. |
| `TestFilterForOneToOneCollection` | 155-286 | Adding a one-to-one collection exposes relational filter arguments including foreign key and nested object filter types. |
| `TestCollectionVersionFilterInputs_WithJSONField_Succeeds` | 309-438 | Adding a collection with a JSON field exposes a JSON filter operator (not a block type) in the filter input. |

### `get_collection_version_test.go`

Tests that `GetCollections` correctly filters by version ID, collection ID, and name, and correctly includes or excludes inactive versions.

| Test Function | Line | Description |
|---|---|---|
| `TestGetCollectionVersion_GivenNonExistantCollectionVersionID_Errors` | 25-37 | Fetching a collection version with a non-existent version ID returns a collection-not-found error. |
| `TestGetCollectionVersion_GivenNoCollectionReturnsEmptySet` | 39-50 | Fetching collection versions when no collections have been added returns an empty result set. |
| `TestGetCollectionVersion_GivenNoCollectionGivenUnknownRoot` | 52-64 | Fetching collection versions filtered by an unknown collection ID returns an empty result set. |
| `TestGetCollectionVersion_GivenNoCollectionGivenUnknownName` | 66-78 | Fetching collection versions filtered by an unknown name returns an empty result set. |
| `TestGetCollectionVersion_ReturnsAllCollections` | 80-153 | Fetching all collection versions including inactive ones returns all registered versions across multiple collections. |
| `TestGetCollectionVersion_ReturnsCollectionForGivenRoot` | 155-224 | Fetching collection versions filtered by collection ID returns all versions sharing that collection root. |
| `TestGetCollectionVersion_ReturnsCollectionForGivenName` | 226-288 | Fetching collection versions filtered by name returns all active and inactive versions with that name. |

### `group_test.go`

Tests that the `groupBy` field enum for a collection exposes the correct enum values for both sides of a one-to-many relation.

| Test Function | Line | Description |
|---|---|---|
| `TestGroupByFieldForTheManySideInCollection` | 21-79 | Adding a one-to-many collection exposes the correct groupBy field enum values on the many side, including relation and internal fields. |
| `TestGroupByFieldForTheSingleSideInCollection` | 81-140 | Adding a one-to-many collection exposes the correct groupBy field enum values on the one side, including relation and internal fields. |

### `input_type_test.go`

Tests that the `order` argument on the `GROUP` field has the correct input object type for both many-relation and single-object-relation schemas.

| Test Function | Line | Description |
|---|---|---|
| `TestInputTypeOfOrderFieldWhereCollectionHasManyRelationType` | 21-117 | The GROUP field's order argument on a type with a many-relation is typed as the correct list-side order input object. |
| `TestInputTypeOfOrderFieldWhereCollectionHasRelationType` | 119-202 | The GROUP field's order argument on a type with a single-object relation is typed as the correct scalar-side order input object. |

### `nil_type_test.go`

Tests error handling when a collection field has no type specified in the SDL.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersion_WithMissingType_Errors` | 21-37 | Adding a collection with a field that has no type specified returns a field-type-not-specified error. |

### `one_many_test.go`

Tests that one-to-many and self-referencing one-to-many collections produce correct field descriptions and GraphQL output.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionOneMany_Primary` | 25-97 | Adding a one-to-many collection with @primary on the many side produces correct field descriptions including foreign key and relation name on both sides. |
| `TestCollectionVersionOneMany_SelfReferenceOneFieldLexographicallyFirst` | 99-148 | A self-referencing one-to-many collection where the scalar field is lexicographically first produces correct field descriptions. |
| `TestCollectionVersionOneMany_SelfReferenceManyFieldLexographicallyFirst` | 150-197 | A self-referencing one-to-many collection where the list field is lexicographically first produces correct field descriptions. |
| `TestCollectionVersionOneMany_SelfUsingActualName` | 199-293 | A self-referencing one-to-many collection with named relation fields produces correct field descriptions and correct GraphQL introspection output. |

### `one_one_test.go`

Tests that one-to-one collection definitions enforce correct @primary placement and produce correct field descriptions and GraphQL output for self-referencing schemas.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionOneOne_NoPrimary_Errors` | 25-48 | Adding a one-to-one collection without any @primary directive returns a relation-missing-field error. |
| `TestCollectionVersionOneOne_TwoPrimaries_Errors` | 50-71 | Adding a one-to-one collection with @primary on both sides returns a single-primary-field error. |
| `TestCollectionVersionOneOne_SelfUsingActualName` | 73-187 | A self-referencing one-to-one collection with named relation fields produces correct field descriptions, a unique index on the primary side, and correct GraphQL introspection output. |

### `self_ref_test.go`

Tests that circular self-referencing collection schemas produce stable, deterministic collection IDs and correctly use collection-set relative indices when multiple types form a shared circular dependency.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionSelfReferenceSimple_HasSimpleCollectionID` | 24-108 | A single-type self-referencing collection produces a stable, deterministic collection ID and correct field descriptions and GraphQL introspection output. |
| `TestCollectionVersionSelfReferenceTwoTypes_HasComplexCollectionID` | 110-353 | Two types that form a circular cross-collection self-reference produce stable, deterministic collection IDs that incorporate both types. |
| `TestCollectionVersionSelfReferenceTwoTypes_HasComplexCollectionID_SingleSidedRelations` | 355-446 | Two types with single-sided circular cross-references produce stable collection IDs encoded with relative-index suffixes distinguishing each type in the set. |
| `TestCollectionVersionSelfReferenceTwoPairsOfTwoTypes_HaveDifferentComplexCollectionID` | 448-736 | Two independent circular pairs (User/Dog and Cat/Mouse) produce different complex collection IDs even when a non-circular cross-pair relation exists. |
| `TestCollectionVersionSelfReferenceTwoPairsOfTwoTypesJoinedByThirdCircle_AllHaveSameBaseCollectionID` | 738-1050 | When a third circular dependency bridges two independent circles (User/Dog and Cat/Mouse), all four types merge into a single collection set sharing the same base ID. |
| `TestCollectionVersionSelfReferenceTwoPairsOfTwoTypesJoinedByThirdCircleAcrossAll_AllHaveSameBaseCollectionID` | 1052-1364 | A larger cross-pair circle bridging two independent circles at different entry points (User=>Dog=>Mouse=>Cat=>User) causes all four types to share the same base collection ID. |

### `similarity_test.go`

Tests that collections with non-null array fields (Int or Float32) expose correctly typed `SIMILARITY` selectors in GraphQL introspection.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionIntrospection_SimilarityCapableFieldIntArray` | 21-112 | Adding a collection with a non-null integer array field exposes a SIMILARITY selector with the correct Int vector input type. |
| `TestCollectionVersionIntrospection_SimilarityCapableFieldFloat32Array` | 114-205 | Adding a collection with a non-null Float32 array field exposes a SIMILARITY selector with the correct Float32 vector input type. |
| `TestCollectionVersionIntrospection_SimilarityCapableFieldsIntArrayAndFloat32Array` | 207-327 | Adding a collection with both non-null Int and Float32 array fields exposes a SIMILARITY selector with separate selectors for each vector field. |

### `simple_test.go`

Tests that collections can be added with every supported scalar and array type and are correctly reflected in GraphQL introspection, and that invalid SDLs are rejected with appropriate errors.

| Test Function | Line | Description |
|---|---|---|
| `TestColVersionSimpleAddsColGivenEmptyType` | 23-64 | Adding an empty collection type creates the collection with only the _docID field and exposes it in GraphQL introspection. |
| `TestCollectionVersionSimpleErrorsGivenDuplicateCollection` | 66-86 | Adding a collection with the same name as an already-registered collection returns a collection-already-exists error. |
| `TestCollectionVersionSimpleErrorsGivenDuplicateCollectionInSameSDL` | 88-103 | Providing an SDL with two identically named types in the same AddCollection call returns a collection-already-exists error. |
| `TestCollectionVersionSimpleErrorsGivenDuplicateCollectionInSameSDLMultiple` | 105-121 | Providing an SDL with three identically named types returns aggregated collection-already-exists errors for each duplicate. |
| `TestCollectionVersionSimpleAddsCollectionGivenNewTypes` | 123-155 | Adding multiple distinct collection types in separate calls makes all types accessible via GraphQL introspection. |
| `TestCollectionVersionSimpleAddsCollectionWithDefaultFieldsGivenEmptyType` | 157-192 | An empty collection type exposes only the default system fields in GraphQL introspection. |
| `TestCollectionVersionSimpleErrorsGivenTypeWithInvalidFieldType` | 194-210 | Adding a collection with a field of an unrecognised type name returns a no-type-found error. |
| `TestCollectionVersionSimpleErrorsGivenTypeWithInvalidFieldTypeMultiple` | 212-229 | Adding a collection with multiple fields of unrecognised types returns aggregated no-type-found errors for each field. |
| `TestCollectionVersionSimpleAddsCollectionGivenTypeWithStringField` | 231-276 | Adding a collection with a String field exposes that field as a SCALAR String in GraphQL introspection. |
| `TestCollectionVersionSimpleErrorsGivenNonNullField` | 278-294 | Adding a collection with a non-null scalar field (String!) returns an unsupported-non-null-field error. |
| `TestCollectionVersionSimpleErrorsGivenNonNullManyRelationField` | 296-316 | Adding a collection with a non-null element type in a list relation field returns an unsupported-non-null-variant error. |
| `TestCollectionVersionSimpleAddsCollectionGivenTypeWithBlobField` | 318-363 | Adding a collection with a Blob field exposes that field as a SCALAR Blob in GraphQL introspection. |
| `TestCollectionVersionSimple_WithJSONField_AddsCollectionGivenType` | 365-410 | Adding a collection with a JSON field exposes that field as a SCALAR JSON in GraphQL introspection. |
| `TestCollectionVersionSimple_WithFloat32Field_AddsCollectionGivenType` | 412-457 | Adding a collection with a Float32 field exposes that field as a SCALAR Float32 in GraphQL introspection. |
| `TestCollectionVersionSimple_WithFloat64Field_AddsCollectionGivenType` | 459-504 | Adding a collection with a Float64 field exposes that field as a SCALAR Float64 in GraphQL introspection. |
| `TestCollectionVersionSimple_WithFloatField_AddsCollectionGivenType` | 506-551 | Adding a collection with a Float field exposes that field as a SCALAR Float64 (the canonical float type) in GraphQL introspection. |
| `TestCollectionVersionSimple_WithAllTypes_AddsCollectionGivenTypes` | 557-761 | Adding a collection with fields of every supported scalar and array type correctly exposes each field with the expected GraphQL kind and name in introspection. |

### `type_explicit_fields_test.go`

Tests that user-defined fields are exposed in the `UserField` enum used to select which fields to encrypt during mutations.

| Test Function | Line | Description |
|---|---|---|
| `TestEncryptFieldsForAddMutation` | 21-60 | Adding a collection with user-defined fields exposes those fields in the UserField enum for use in mutation encryption selectors. |

### `vector_embedding_test.go`

Tests that the `@embedding` directive only accepts non-null Float32 array fields and correctly validates provider, model, and field-list constraints.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersion_WithStringForEmbedding_ShouldError` | 21-37 | Adding a collection with a String array field annotated with @embedding returns an invalid-type error. |
| `TestCollectionVersion_WithIntForEmbedding_ShouldError` | 39-55 | Adding a collection with an Int array field annotated with @embedding returns an invalid-type error. |
| `TestCollectionVersion_WithFloatForEmbedding_ShouldError` | 56-72 | Adding a collection with a Float array field annotated with @embedding returns an invalid-type error. |
| `TestCollectionVersion_WithFloat64ForEmbedding_ShouldError` | 74-90 | Adding a collection with a Float64 array field annotated with @embedding returns an invalid-type error. |
| `TestCollectionVersion_WithNillableFloat32ForEmbedding_ShouldError` | 92-108 | Adding a collection with a nillable Float32 array field annotated with @embedding returns an invalid-type error (only non-null [Float32!] is valid). |
| `TestCollectionVersion_WithFloat32ForEmbedding_ShouldSucceed` | 110-126 | Adding a collection with a non-null Float32 array field and a fully specified @embedding directive succeeds. |
| `TestCollectionVersion_WithNonExistantFieldForEmbedding_ShouldError` | 128-145 | Adding a collection where the @embedding fields list references a non-existent field returns a does-not-exist error. |
| `TestCollectionVersion_WithInvalidEmbeddingGenerationFieldType_ShouldError` | 147-165 | Adding a collection where the @embedding fields list includes a JSON field returns an invalid-field-type error. |
| `TestCollectionVersion_WithUnsupportedProviderForEmbedding_ShouldError` | 167-184 | Adding a collection with an unknown provider in the @embedding directive returns an unknown-provider error. |
| `TestCollectionVersion_WithMissingModelForEmbedding_ShouldError` | 186-203 | Adding a collection with a valid provider but no model in the @embedding directive returns an empty-model error. |
| `TestCollectionVersion_ReferenceToSelfForEmbedding_ShouldError` | 205-222 | Adding a collection where an @embedding field lists itself in its fields list returns a self-reference error. |
| `TestCollectionVersion_ReferenceToAnotherEmbedding_ShouldError` | 224-242 | Adding a collection where an @embedding field lists another embedding field in its fields list returns a cross-embedding-reference error. |

### `with_default_fields_test.go`

Tests that the `@default` directive stores correct typed defaults and rejects invalid argument types, mismatched types, multiple arguments, and unsupported field types.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersion_WithDefaultFieldValues` | 23-99 | Adding a collection with @default directives on fields stores the correct typed default values in the collection description. |
| `TestCollectionVersion_WithInvalidDefaultFieldValueType_ReturnsError` | 101-117 | Adding a collection with an invalid argument value in a @default directive returns a GraphQL argument error. |
| `TestCollectionVersion_WithIncorrectDefaultFieldValueType_ReturnsError` | 119-135 | Adding a collection where a @default argument type mismatches the field type returns a type-mismatch error. |
| `TestCollectionVersion_WithMultipleDefaultFieldValueTypes_ReturnsError` | 137-153 | Adding a collection where a @default directive specifies more than one argument type returns a must-specify-one-argument error. |
| `TestCollectionVersion_WithDefaultFieldValueOnRelation_ReturnsError` | 155-171 | Adding a collection with a @default directive on a relation field returns a not-allowed-for-type error. |
| `TestCollectionVersion_WithDefaultFieldValueOnList_ReturnsError` | 173-189 | Adding a collection with a @default directive on a list field returns a not-allowed-for-type error. |

### `with_inline_array_test.go`

Tests that collections with inline array fields can be registered and are accessible in GraphQL introspection.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionInlineArrayAddsCollectionGivenSingleType` | 21-50 | Adding a collection with an inline array field succeeds and makes the type accessible in GraphQL introspection. |
| `TestCollectionVersionInlineArrayAddsCollectionGivenSecondType` | 52-88 | Adding a second collection with an inline array field alongside an existing collection succeeds and makes both types accessible. |

### `with_update_set_default_test.go`

Tests `SetActiveCollectionVersion` — error handling for empty or unknown version IDs, and the effect on field queryability when switching active versions.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersion_WithUpdateAndSetDefaultVersionToEmptyString_Errors` | 21-46 | Setting the active collection version to an empty string returns an empty-ID error. |
| `TestCollectionVersion_WithUpdateAndSetDefaultVersionToUnknownVersion_Errors` | 48-73 | Setting the active collection version to an unknown version ID returns a collection-not-found error. |
| `TestCollectionVersion_WithUpdateAndSetDefaultVersionToOriginal_NewFieldIsNotQueriable` | 75-110 | Setting the active version back to the original after a patch makes the new field inaccessible for queries. |
| `TestCollectionVersion_WithUpdateAndSetDefaultVersionToNew_AllowsQueryingOfNewField` | 112-148 | Setting the active version to the newly patched version makes the new field accessible for queries. |

## Subdirectories

| Directory | Summary |
|---|---|
| [`aggregates/`](aggregates/INDEX.md) | Tests that adding a collection exposes correct COUNT, SUM, and AVG aggregate fields and their selector argument types in GraphQL introspection, covering both simple collections and inline scalar arrays. |
| [`client_introspection/`](client_introspection/INDEX.md) | Tests that the full client-facing GraphQL introspection query executes without error against an empty schema and after adding relational collections. |
| [`migrations/`](migrations/INDEX.md) | Tests for lens-based schema migrations applied to collection versions, covering forward/inverse migrations, indexes, transactions, P2P replication, and node restarts. |
| [`updates/`](updates/INDEX.md) | Tests for JSON Patch operations (`add`, `copy`, `move`, `remove`, `replace`, `test`) on collection schema versions, and for managing active collection version branches. |
