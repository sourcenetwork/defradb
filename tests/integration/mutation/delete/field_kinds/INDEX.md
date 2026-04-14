# Index: `tests/integration/mutation/delete/field_kinds`

## Overview

These tests cover `delete` mutations on relation field kinds — one-to-many and one-to-one-to-one chains. They verify that documents can be deleted by docID from either side of a relation, that soft-deletion is correctly reflected via the `showDeleted` query flag, that result fields can be aliased, and that transactional isolation is properly maintained so that concurrent transactions can read records deleted but not yet committed in a sibling transaction.

## Subdirectories

| Directory | Summary |
|---|---|
| [`one_to_many/`](one_to_many/INDEX.md) | Tests that deleting a document on the many-side of a one-to-many relation is correctly reflected when querying with the `showDeleted` flag. |
| [`one_to_one_to_one/`](one_to_one_to_one/INDEX.md) | Tests for deleting documents by docID across a three-type one-to-one-to-one chain, covering plain deletion, aliased result fields, and transactional read isolation before commit. |
