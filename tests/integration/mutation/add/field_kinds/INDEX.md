# Index: `tests/integration/mutation/add/field_kinds`

## Overview

These tests cover `add` (create/insert) mutations on relation field kinds — one-to-one, one-to-many, and one-to-one-to-one chains. They verify that documents can be correctly linked at creation time using both raw foreign-key ID fields and aliased relation names, that mutations from invalid sides are rejected, that uniqueness and null-value edge cases behave correctly, and that transactional snapshot isolation is properly enforced when concurrently creating and linking related documents.

## Subdirectories

| Directory | Summary |
|---|---|
| [`one_to_many/`](one_to_many/INDEX.md) | Tests for creating and linking documents on the many-side of a one-to-many relation using both raw foreign-key ID fields and aliased relation names, covering valid linking and expected error conditions. |
| [`one_to_one/`](one_to_one/INDEX.md) | Tests for creating and linking documents across a one-to-one relation using raw foreign-key fields, aliased relation names, and explicit null values, covering bidirectional traversal, secondary-side rejection, and uniqueness enforcement. |
| [`one_to_one_to_one/`](one_to_one_to_one/INDEX.md) | Tests for concurrent transactional creation and relational linking of documents across a three-type one-to-one-to-one chain, verifying correct link visibility and SSI conflict detection. |
