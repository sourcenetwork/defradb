# Index: `tests/integration/mutation/update/field_kinds`

## Overview

These tests cover `update` mutations on relation field kinds — one-to-many and one-to-one. They verify that existing relational links can be updated or re-established using both raw relation ID fields and aliased relation name fields, that mutations from the wrong side of the relation are correctly rejected, and that errors are returned for malformed document IDs, non-existent fields, and unique-index violations. Self-referencing relations and GQL mutation result shapes are also exercised.

## Subdirectories

| Directory | Summary |
|---|---|
| [`one_to_many/`](one_to_many/INDEX.md) | Tests for re-linking documents in a one-to-many relation via raw relation ID fields and alias relation name fields, covering error cases from the one-side and valid re-linking from the many-side via both Collection API and GQL. |
| [`one_to_one/`](one_to_one/INDEX.md) | Tests for updating one-to-one relation links via alias and raw relation ID fields, covering primary-side linking, secondary-side rejection, unique-index violations, self-referencing relations, invalid IDs, and GQL result shape. |
