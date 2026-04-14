# Index: `tests/integration/acp/dac`

## Overview

This directory tests the complete Document Access Control (DAC) system in DefraDB. DAC is the per-document access control mechanism where each document is owned by the identity that created it and can optionally be shared with other actors via policy-defined relations. The top-level tests in this directory exercise the core read, write, delete, and aggregate operations across all combinations of identity presence (none, owner, wrong identity), and verify that relation-traversal queries (one-to-many and many-to-one) respect per-document visibility rules. The subdirectories cover the full lifecycle: registering DAC policies, linking collections to those policies, managing fine-grained actor relationships (granting and revoking access), index-backed queries under DAC, and P2P replication and subscription behaviour when collections are policy-protected.

## Test Index

### `avg_test.go`

Tests that the `AVG` aggregate respects DAC permissions, averaging only the documents visible to the requesting identity.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_QueryAverageWithoutIdentity` | 21-42 | DAC average aggregate without identity only includes public documents. |
| `TestACP_QueryAverageWithIdentity` | 44-66 | DAC average aggregate with owner identity includes all owned documents. |
| `TestACP_QueryAverageWithWrongIdentity` | 68-90 | DAC average aggregate with wrong identity only includes public documents. |

### `count_test.go`

Tests that the `COUNT` aggregate for both top-level documents and related nested objects respects DAC permissions based on the caller's identity.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_QueryCountDocumentsWithoutIdentity` | 21-41 | DAC count aggregate without identity only counts public top-level documents. |
| `TestACP_QueryCountRelatedObjectsWithoutIdentity` | 43-70 | DAC count of related objects without identity only includes public nested documents. |
| `TestACP_QueryCountDocumentsWithIdentity` | 72-93 | DAC count aggregate with owner identity counts all owned and public documents. |
| `TestACP_QueryCountRelatedObjectsWithIdentity` | 95-125 | DAC count of related objects with owner identity counts all accessible nested documents. |
| `TestACP_QueryCountDocumentsWithWrongIdentity` | 127-148 | DAC count aggregate with wrong identity only counts public documents. |
| `TestACP_QueryCountRelatedObjectsWithWrongIdentity` | 150-178 | DAC count of related objects with wrong identity only includes public nested documents. |

### `register_and_delete_test.go`

Tests the delete permission matrix: public documents are deletable by anyone, while private documents can only be deleted by their owner identity.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddWithoutIdentityAndDeleteWithoutIdentity_CanDelete` | 21-103 | Document added without identity can be deleted without identity (public document). |
| `TestACP_AddWithoutIdentityAndDeleteWithIdentity_CanDelete` | 105-188 | Public document added without identity can be deleted by an identity. |
| `TestACP_AddWithIdentityAndDeleteWithIdentity_CanDelete` | 190-273 | Private document owner can delete their own document using the same identity. |
| `TestACP_AddWithIdentityAndDeleteWithoutIdentity_CanNotDelete` | 275-364 | Private document cannot be deleted without an identity. |
| `TestACP_AddWithIdentityAndDeleteWithWrongIdentity_CanNotDelete` | 366-457 | Private document cannot be deleted by a different identity than the owner. |

### `register_and_read_test.go`

Tests the read permission matrix: public documents are readable by anyone, while private documents can only be read by their owner identity.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddWithoutIdentityAndReadWithoutIdentity_CanRead` | 21-98 | Public document added without identity is readable without identity. |
| `TestACP_AddWithoutIdentityAndReadWithIdentity_CanRead` | 100-179 | Public document added without identity is readable with any identity. |
| `TestACP_AddWithIdentityAndReadWithIdentity_CanRead` | 181-262 | Private document owner can read their own document using the same identity. |
| `TestACP_AddWithIdentityAndReadWithoutIdentity_CanNotRead` | 264-337 | Private document is not readable without an identity. |
| `TestACP_AddWithIdentityAndReadWithWrongIdentity_CanNotRead` | 339-414 | Private document is not readable by a different identity than the owner. |

### `register_and_update_test.go`

Tests the update permission matrix across all mutation types: public documents are updatable by anyone, private documents only by their owner, and GQL mutations silently ignore unauthorized updates instead of returning an error.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddWithoutIdentityAndUpdateWithoutIdentity_CanUpdate` | 24-118 | Public document added without identity can be updated without identity. |
| `TestACP_AddWithoutIdentityAndUpdateWithIdentity_CanUpdate` | 120-215 | Public document added without identity can be updated by an identity. |
| `TestACP_AddWithIdentityAndUpdateWithIdentity_CanUpdate` | 217-312 | Private document owner can update their own document using the same identity. |
| `TestACP_AddWithIdentityAndUpdateWithoutIdentity_CanNotUpdate` | 314-415 | Private document cannot be updated without an identity (collection and save mutation types). |
| `TestACP_AddWithIdentityAndUpdateWithWrongIdentity_CanNotUpdate` | 417-520 | Private document cannot be updated by a different identity than the owner (collection and save mutation types). |
| `TestACP_AddWithIdentityAndUpdateWithoutIdentityGQL_CanNotUpdate` | 524-624 | Private document update without identity via GQL silently fails (no error, update is ignored). |
| `TestACP_AddWithIdentityAndUpdateWithWrongIdentityGQL_CanNotUpdate` | 628-730 | Private document update with wrong identity via GQL silently fails (no error, update is ignored). |

### `relation_objects_test.go`

Tests that one-to-many and many-to-one relation queries respect DAC permissions, hiding private documents from callers who do not own them.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_QueryManyToOneRelationObjectsWithoutIdentity` | 21-55 | DAC many-to-one relation query without identity hides private parent documents. |
| `TestACP_QueryOneToManyRelationObjectsWithoutIdentity` | 57-89 | DAC one-to-many relation query without identity only returns public parent documents with their public children. |
| `TestACP_QueryManyToOneRelationObjectsWithIdentity` | 91-135 | DAC many-to-one relation query with owner identity returns all documents with their parent. |
| `TestACP_QueryOneToManyRelationObjectsWithIdentity` | 137-179 | DAC one-to-many relation query with owner identity returns all documents with all their children. |
| `TestACP_QueryManyToOneRelationObjectsWithWrongIdentity` | 181-216 | DAC many-to-one relation query with wrong identity hides private parent documents. |
| `TestACP_QueryOneToManyRelationObjectsWithWrongIdentity` | 218-251 | DAC one-to-many relation query with wrong identity only returns public parent documents with their public children. |

## Subdirectories

| Directory | Summary |
|---|---|
| [`add_policy/`](add_policy/INDEX.md) | Tests the `AddDACPolicy` action, covering valid policy acceptance, non-DRI-compliant but sourcehub-valid policies, and rejection of malformed YAML, duplicate keys, invalid expressions, and cross-resource relation references. |
| [`index/`](index/INDEX.md) | Tests index creation on DAC-protected collections and verifies that index-backed queries respect DAC read permissions for single collections and across one-to-many and many-to-one relations. |
| [`link_collection/`](link_collection/INDEX.md) | Tests linking a collection to a DAC policy via the `@policy` SDL directive, verifying acceptance of valid DRI resources and rejection of missing policies, invalid resource names, absent required permissions, and malformed directive arguments. |
| [`p2p/`](p2p/INDEX.md) | Tests P2P replication and pub-sub subscription behaviour under DAC policies, verifying that private documents remain hidden from unauthorised actors and that access granted or revoked via DAC actor relationships is honoured across all nodes. |
| [`relationship/`](relationship/INDEX.md) | Tests the full lifecycle of DAC access relationships between documents and actors, including granting and revoking permissions, input validation, authorization enforcement, delegation through managers, and edge cases such as public documents and wildcard actor targets. |
