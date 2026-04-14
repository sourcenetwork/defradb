# Index: `tests/integration/encryption`

## Overview

This folder contains integration tests for DefraDB's at-rest document encryption feature. The tests cover how encryption keys are generated and applied at the document and field levels, how commit deltas are stored in encrypted form across different CRDT types (LWW and counter), how encrypted documents are correctly decrypted on query, and how encryption interacts with peer-to-peer sync, access-control policies (ACP), and secondary indexes.

## Test Index

### `commit_test.go`

Tests that commit deltas are stored in encrypted form for both LWW and counter CRDTs, that updates continue to encrypt deltas, and that per-document keys produce distinct cipher texts.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryption_WithEncryptionOnLWWCRDT_ShouldStoreCommitsDeltaEncrypted` | 25-93 | Encrypted LWW CRDT document stores all field commit deltas in encrypted form. |
| `TestDocEncryption_UponUpdateOnLWWCRDT_ShouldEncryptCommitDelta` | 95-133 | Updating an encrypted LWW CRDT document produces encrypted commit deltas. |
| `TestDocEncryption_WithMultipleDocsUponUpdate_ShouldEncryptOnlyRelevantDocs` | 135-194 | Only the encrypted document has encrypted commit deltas; plaintext doc remains unencrypted. |
| `TestDocEncryption_WithEncryptionOnCounterCRDT_ShouldStoreCommitsDeltaEncrypted` | 196-241 | Encrypted counter CRDT document stores commit deltas in encrypted form. |
| `TestDocEncryption_UponUpdateOnCounterCRDT_ShouldEncryptedCommitDelta` | 243-290 | Updating an encrypted counter CRDT produces encrypted commit deltas for each increment. |
| `TestDocEncryption_UponEncryptionSeveralDocs_ShouldStoreAllCommitsDeltaEncrypted` | 292-344 | Batch-inserting multiple encrypted documents stores all their commit deltas encrypted. |
| `TestDocEncryption_IfTwoDocsHaveSameFieldValue_CipherTextShouldBeDifferent` | 346-388 | Two documents with identical field values produce distinct cipher texts due to per-doc keys. |

### `commit_relation_test.go`

Tests that commit deltas are stored in encrypted form when encryption is applied across related collections.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryption_WithEncryptionSecondaryRelations_ShouldStoreEncryptedCommit` | 21-107 | Encrypting documents across a relation stores all field commit deltas encrypted. |

### `field_commit_test.go`

Tests that field-level encryption targets only the specified fields and uses dedicated per-field encryption keys, including after updates.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryptionField_WithEncryptionOnField_ShouldStoreOnlyFieldsDeltaEncrypted` | 23-67 | Field-level encryption stores only the specified field's delta encrypted; others remain plaintext. |
| `TestDocEncryptionField_WithDocAndFieldEncryption_ShouldUseDedicatedEncKeyForIndividualFields` | 69-129 | Doc-level and field-level encryption use dedicated keys per individually encrypted field. |
| `TestDocEncryptionField_UponUpdateWithDocAndFieldEncryption_ShouldUseDedicatedEncKeyForIndividualFields` | 131-200 | After update, individually encrypted fields continue to use dedicated per-field encryption keys. |

### `field_query_test.go`

Tests that querying a document with an encrypted field transparently returns the decrypted value.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryptionField_WithEncryption_ShouldFetchDecrypted` | 21-59 | Querying a document with an encrypted field returns the decrypted field value. |

### `field_test.go`

Tests error handling when invalid or built-in fields are specified for field-level encryption via both GQL and collection APIs.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryptionField_IfFieldDoesNotExistInGQLSchema_ReturnError` | 26-49 | Encrypting a non-existent field via GQL mutation returns an invalid argument error. |
| `TestDocEncryptionField_IfAttemptToEncryptBuiltinFieldInGQLSchema_ReturnError` | 51-79 | Attempting to encrypt built-in fields via GQL mutation returns an invalid argument error. |
| `TestDocEncryptionField_IfFieldDoesNotExist_ReturnError` | 81-105 | Specifying a non-existent field for encryption via collection API returns a field-not-exist error. |
| `TestDocEncryptionField_IfAttemptToEncryptBuiltinField_ReturnError` | 107-136 | Attempting to encrypt built-in fields via collection API returns a cannot-encrypt built-in error. |

### `peer_acp_test.go`

Tests encrypted document access during peer sync when combined with access-control policies (ACP), covering scenarios where the node or user may or may not have the required permissions.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryptionACP_IfUserAndNodeHaveAccess_ShouldFetch` | 62-141 | Peer with both user and node ACP reader access can fetch and decrypt the document. |
| `TestDocEncryptionACP_IfUserHasAccessButNotNode_ShouldNotFetch` | 143-231 | User with ACP access but without node-level permission cannot decrypt or fetch the document. |
| `TestDocEncryptionACP_IfNodeHasAccessToSomeDocs_ShouldFetchOnlyThem` | 233-372 | Node fetches only documents it has access to across mixed encrypted and plaintext access patterns. |
| `TestDocEncryptionACP_IfClientNodeHasDocPermissionButServerNodeIsNotAvailable_ShouldNotFetch` | 374-460 | When the origin node is offline, a permitted peer cannot fetch the encrypted document. |

### `peer_sec_index_test.go`

Tests that secondary indexes on encrypted fields are correctly rebuilt after peer sync and decryption.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryptionPeer_IfEncryptedDocHasIndexedField_ShouldIndexAfterDecryption` | 23-102 | Synced encrypted documents with indexed fields are decrypted and correctly indexed on the peer. |
| `TestDocEncryptionPeer_IfDocDocHasEncryptedIndexedField_ShouldIndexAfterDecryption` | 104-165 | Synced documents with individually encrypted indexed fields are decrypted and indexed on the peer. |

### `peer_share_test.go`

Tests that peers correctly fetch encryption keys from the KMS and decrypt documents after sync, including public docs, field-level encryption, counter CRDTs, and edge cases like empty or null values.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryptionPeer_IfDocIsPublic_ShouldFetchKeyAndDecrypt` | 24-72 | A peer syncing a public encrypted document fetches the key and returns the decrypted value. |
| `TestDocEncryptionPeer_IfPublicDocHasEncryptedField_ShouldFetchKeyAndDecrypt` | 74-124 | A peer syncing a public doc with one encrypted field fetches the field key and decrypts it. |
| `TestDocEncryptionPeer_IfEncryptedPublicDocHasEncryptedField_ShouldFetchKeysAndDecrypt` | 126-177 | A peer fetches both doc-level and field-level keys and decrypts all fields successfully. |
| `TestDocEncryptionPeer_IfAllFieldsOfEncryptedPublicDocAreIndividuallyEncrypted_ShouldFetchKeysAndDecrypt` | 179-230 | Peer decrypts an encrypted public doc where every field is also individually encrypted. |
| `TestDocEncryptionPeer_IfAllFieldsOfPublicDocAreIndividuallyEncrypted_ShouldFetchKeysAndDecrypt` | 232-282 | Peer fetches per-field keys and decrypts a public doc where all fields are individually encrypted. |
| `TestDocEncryptionPeer_WithUpdatesOnEncryptedDeltaBasedCRDTField_ShouldDecryptAndCorrectlyMerge` | 284-345 | Peer correctly decrypts and merges accumulated updates on an encrypted counter CRDT field. |
| `TestDocEncryptionPeer_WithUpdatesOnDeltaBasedCRDTFieldOfEncryptedDoc_ShouldDecryptAndCorrectlyMerge` | 347-408 | Peer decrypts and correctly merges counter CRDT updates on a fully encrypted document. |
| `TestDocEncryptionPeer_WithUpdatesThatSetsEmptyString_ShouldDecryptAndCorrectlyMerge` | 410-478 | Peer correctly decrypts and merges updates that set an encrypted field to an empty string. |
| `TestDocEncryptionPeer_WithUpdatesThatSetsStringToNull_ShouldDecryptAndCorrectlyMerge` | 480-548 | Peer correctly decrypts and merges updates that set an encrypted field to null. |

### `peer_test.go`

Tests basic peer sync behavior for encrypted documents, including DAG replication and key availability timing.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryptionPeer_UponSync_ShouldSyncEncryptedDAG` | 23-105 | Peer sync replicates an encrypted document's DAG commits with encrypted deltas intact. |
| `TestDocEncryptionPeer_IfPeerDidNotReceiveKey_ShouldNotFetch` | 107-156 | A peer that has not yet received the encryption key cannot fetch decrypted document data. |

### `query_relation_test.go`

Tests that queries on encrypted related documents correctly return decrypted values across both sides of a relation.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryption_WithEncryptionOnBothRelations_ShouldFetchDecrypted` | 21-83 | Querying encrypted documents on both sides of a relation returns all fields decrypted. |
| `TestDocEncryption_WithEncryptionOnPrimaryRelations_ShouldFetchDecrypted` | 85-146 | Querying an encrypted primary-side document with its related documents returns decrypted fields. |
| `TestDocEncryption_WithEncryptionOnSecondaryRelations_ShouldFetchDecrypted` | 148-209 | Querying an encrypted secondary-side document with its related documents returns decrypted fields. |

### `query_test.go`

Tests that queries on encrypted documents return correctly decrypted values, including counter CRDT accumulation after updates.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryption_WithEncryption_ShouldFetchDecrypted` | 22-60 | Querying a fully encrypted document returns all fields correctly decrypted. |
| `TestDocEncryption_WithEncryptionOnCounterCRDT_ShouldFetchDecrypted` | 62-121 | Querying an encrypted counter CRDT document returns the correctly decrypted and accumulated value. |
