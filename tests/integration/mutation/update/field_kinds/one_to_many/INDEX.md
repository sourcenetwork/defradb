# Index: `tests/integration/mutation/update/field_kinds/one_to_many`

## Overview

This folder contains integration tests for mutation update operations on one-to-many relation fields in DefraDB. The tests exercise linking and re-linking documents across the `Book`–`Author` one-to-many relationship using both raw relation ID fields (e.g. `_authorID`) and alias relation name fields (e.g. `author`/`published`), covering both the "many" side (the foreign-key holder) and the "one" side (the secondary side), and asserting correct error behaviour when mutations are attempted from the wrong side, with malformed IDs, or with non-existent fields.

## Test Index

### `simple_test.go`

Tests that update one-to-many relation links via raw relation ID fields (`_authorID`), covering error cases from the single/one-side and valid re-linking from the many-side.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdateOneToMany_RelationIDToLinkFromSingleSide_Error` | 25-76 | Update one-to-many relation ID from single side returns field-not-found error. |
| `TestMutationUpdateOneToMany_InvalidRelationIDToLinkFromManySide` | 78-116 | Update book's relation ID with a malformed author ID returns UUID error. |
| `TestMutationUpdateOneToMany_RelationIDToLinkFromManySideWithWrongField_Error` | 118-169 | Update book with a non-existent field alongside a relation ID returns field error. |
| `TestMutationUpdateOneToMany_RelationIDToLinkFromManySide` | 171-261 | Update book's relation ID from the many-side re-links it to a different author. |

### `with_alias_test.go`

Tests that update one-to-many relation links via alias relation name fields (`author`/`published`), covering error cases from the one-side via both Collection API and GQL, invalid IDs, wrong fields, and successful re-linking from the many-side.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdateOneToMany_AliasRelationNameToLinkFromSingleSide_CollectionApi` | 25-73 | Update one-to-many alias relation from the one-side via Collection API returns error. |
| `TestMutationUpdateOneToMany_AliasRelationNameToLinkFromSingleSide_GQL` | 75-122 | Update one-to-many alias relation from the one-side via GQL returns invalid argument error. |
| `TestMutationUpdateOneToMany_InvalidAliasRelationNameToLinkFromManySide_GQL` | 126-164 | Update book's alias relation field with a malformed author ID via GQL returns UUID error. |
| `TestMutationUpdateOneToMany_InvalidAliasRelationNameToLinkFromManySide_Collection` | 166-204 | Update book's alias relation field with a malformed author ID via Collection API returns UUID error. |
| `TestMutationUpdateOneToMany_AliasRelationNameToLinkFromManySideWithWrongField_Error` | 206-257 | Update book with a non-existent field alongside an alias relation ID returns field error. |
| `TestMutationUpdateOneToMany_AliasRelationNameToLinkFromManySide` | 259-349 | Update book's alias relation field from the many-side re-links it to a different author. |
