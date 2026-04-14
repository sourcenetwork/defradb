# Index: `tests/integration/signature`

## Overview

This folder contains integration tests for DefraDB's block-signing feature, which cryptographically signs CRDT commit blocks using node or client identities. The tests cover signature data inclusion in commit queries, key-type variations (Secp256k1 and Ed25519), fallback to the node identity when a client lacks a private key, ACP-gated signature verification, signature propagation across P2P peers with mixed key types, branchable-collection signing, hex encoding of identity fields, and direct block-signature verification via `VerifyBlockSignature`. A shared `utils.go` provides custom Gomega matchers (`signatureMatcher`, `identityMatcher`) used throughout the suite.

## Test Index

### `acp_test.go`

Tests that ACP read permissions are enforced during block signature verification, distinguishing between identities with and without document access.

| Test Function | Line | Description |
|---|---|---|
| `TestSignatureACP_IfHasNoAccessToDoc_ShouldError` | 49-91 | Block signature verification fails when the requesting identity lacks ACP read access. |
| `TestSignatureACP_IfHasAccessToDoc_ValidateSignature` | 93-134 | Block signature is validated successfully when the requesting identity has ACP read access. |

### `branchable_test.go`

Tests that branchable collection commits include valid signature data on all block types.

| Test Function | Line | Description |
|---|---|---|
| `TestSignature_WithBranchableCollection_ShouldSignCollectionBlocks` | 27-95 | All commit blocks for a branchable collection include valid ECDSA256K signature data. |

### `commit_test.go`

Tests signature data returned in commit queries, covering key types, update/delete behaviour, client identity usage, and identity encoding format.

| Test Function | Line | Description |
|---|---|---|
| `TestSignature_WithCommitQuery_ShouldIncludeSignatureData` | 49-120 | Commit query returns ECDSA256K signature type, identity, and value for each field block. |
| `TestSignature_WithUpdatedDocsAndCommitQuery_ShouldSignOnlyFirstFieldBlocks` | 122-220 | Only the initial field blocks at height 1 are signed; composite blocks at all heights are signed. |
| `TestSignature_WithDeletedDocAndCommitQuery_ShouldIncludeSignatureData` | 222-282 | Composite commit blocks for a deleted document include valid node-identity signature data. |
| `TestSignature_WithEd25519KeyType_ShouldIncludeSignatureData` | 284-352 | Commit query returns Ed25519 signature type and correctly matched identity and value fields. |
| `TestSignature_WithClientIdentity_ShouldUseItForSigning` | 356-433 | Commit blocks are signed with the client identity's key type when a client identity is supplied. |
| `TestSignature_WithCommitQuery_ShouldBeHexEncoded` | 434-501 | Commit query returns the signer identity as a hex-encoded public key string. |

### `fallback_signer_test.go`

Tests the fallback behaviour where the node identity is used as the signer when the client identity carries no private key.

| Test Function | Line | Description |
|---|---|---|
| `TestSignature_IfIdentityHasNoPrivateKey_ShouldUseNodeIdentity` | 24-67 | When a client identity lacks a private key, blocks are signed with the node identity instead. |

### `peer_test.go`

Tests that signed documents sync correctly between peers, including scenarios where each peer uses a different key type or both peers update the same document.

| Test Function | Line | Description |
|---|---|---|
| `TestDocSignature_WithPeersAndSecp256k1KeyType_ShouldSync` | 26-82 | Documents signed with Secp256k1 keys sync successfully between two peers. |
| `TestDocSignature_WithPeersAndEd25519KeyType_ShouldSync` | 84-140 | Documents signed with Ed25519 keys sync successfully between two peers. |
| `TestDocSignature_WithPeersAnDifferentKeyTypes_ShouldSync` | 142-239 | Documents from peers using different key types (Secp256k1 and Ed25519) sync and retain correct signatures. |
| `TestDocSignature_WithPeersAnDifferentKeyTypesUpdatingSameDoc_ShouldSync` | 241-350 | Peers with different key types updating the same document each preserve their own signature type. |

### `query_test.go`

Tests that normal document and commit queries function correctly when block signing is enabled.

| Test Function | Line | Description |
|---|---|---|
| `TestDocSignature_WithEnabledSigning_ShouldQuery` | 21-62 | A basic document query returns correct results when block signing is enabled. |
| `TestDocSignature_WithEnabledSigning_ShouldQueryCommitsWithoutSignature` | 64-114 | Querying commit metadata without the signature field succeeds when signing is enabled. |

### `verify_test.go`

Tests the `VerifyBlockSignature` action for valid signatures, alternative key types, wrong identity, and non-existent CIDs.

| Test Function | Line | Description |
|---|---|---|
| `TestSignatureVerify_WithValidData_ShouldVerify` | 24-64 | Block signatures verify successfully for create, update, and delete operations using the node identity. |
| `TestSignatureVerify_WithDifferentKeyType_ShouldVerify` | 66-95 | Block signature verification succeeds when the node identity uses an Ed25519 key. |
| `TestSignatureVerify_WithWrongIdentity_ShouldError` | 97-124 | Block signature verification returns a public key mismatch error when verified with the wrong identity. |
| `TestSignatureVerify_WithWrongCid_ShouldError` | 126-153 | Block signature verification returns a not-found error when given a non-existent CID. |
