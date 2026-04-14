# Index: `tests/integration/acp/dac/index`

## Overview

This folder contains integration tests that verify index behaviour on collections protected by DAC (Document Access Control) policies. The tests confirm that index creation works correctly on policy-guarded collections, and that index-assisted queries honour DAC permissions — returning only the documents (and related documents via one-to-many and many-to-one relations) that the requesting identity is authorised to read.

## Test Index

### `new_test.go`

Tests that indexes can be created on DAC-protected collections without error, both via a separate `NewIndex` action and via an inline `@index` SDL directive.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_IndexNewWithSeparateRequest_OnCollectionWithPolicy_NoError` | 21-66 | Creating an index via separate request on a DAC-protected collection succeeds. |
| `TestACP_IndexNewWithDirective_OnCollectionWithPolicy_NoError` | 68-106 | Creating an index via @index directive on a DAC-protected collection succeeds. |

### `query_test.go`

Tests that index-backed queries on a single DAC-protected collection respect read permissions based on the caller's identity.

| Test Function | Line | Description |
|---|---|---|
| `TestACPWithIndex_UponQueryingPrivateDocWithoutIdentity_ShouldNotFetch` | 21-72 | Index query without identity does not return private DAC-protected documents. |
| `TestACPWithIndex_UponQueryingPrivateDocWithIdentity_ShouldFetch` | 74-131 | Index query with the owner's identity returns both public and private DAC documents. |
| `TestACPWithIndex_UponQueryingPrivateDocWithWrongIdentity_ShouldNotFetch` | 133-187 | Index query with a different identity does not return another owner's private DAC documents. |

### `query_with_relation_test.go`

Tests that index-filtered queries across one-to-many and many-to-one relations enforce DAC permissions, hiding private documents from callers who lack the correct identity.

| Test Function | Line | Description |
|---|---|---|
| `TestACPWithIndex_UponQueryingPrivateOneToManyRelatedDocWithoutIdentity_ShouldNotFetch` | 96-130 | Index-filtered one-to-many query without identity excludes private related documents. |
| `TestACPWithIndex_UponQueryingPrivateOneToManyRelatedDocWithIdentity_ShouldFetch` | 132-179 | Index-filtered one-to-many query with owner identity returns all accessible related documents. |
| `TestACPWithIndex_UponQueryingPrivateOneToManyRelatedDocWithWrongIdentity_ShouldNotFetch` | 181-216 | Index-filtered one-to-many query with wrong identity excludes private related documents. |
| `TestACPWithIndex_UponQueryingPrivateManyToOneRelatedDocWithoutIdentity_ShouldNotFetch` | 218-250 | Index-filtered many-to-one query without identity excludes private related documents. |
| `TestACPWithIndex_UponQueryingPrivateManyToOneRelatedDocWithIdentity_ShouldFetch` | 252-298 | Index-filtered many-to-one query with owner identity returns all accessible related documents. |
| `TestACPWithIndex_UponQueryingPrivateManyToOneRelatedDocWithWrongIdentity_ShouldNotFetch` | 300-333 | Index-filtered many-to-one query with wrong identity excludes private related documents. |
