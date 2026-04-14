# Index: `tests/integration/acp/nac`

## Overview

This directory tests the Node Access Control (NAC) system in DefraDB. NAC is a node-level access control layer that restricts all administrative and data operations to authorized identities when enabled. The tests cover NAC lifecycle management (starting, enabling, disabling, and re-enabling NAC; restart persistence), NAC access gates for every node operation (schema management, document CRUD, P2P networking, index management, lens migrations, signature verification, and DAC interaction), the relationship between NAC and DAC access patterns (empty users, node owners, non-node-owners, and the NAC admin DAC bypass), and the management of NAC actor relationships (adding and deleting, including for wildcard all-identity access). The `relation_admin/` subdirectory tests the NAC `admin` relation specifically — a delegated-authority mechanism that grants a user the same permissions as the node owner.

## Test Index

### `add_collection_test.go`

Tests that NAC gates the `AddCollection` operation: the node owner succeeds while no-identity and wrong-identity callers receive a `NotAuthorizedError`.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesAddCollection_AllowIfAuthorizedElseError` | 22-70 | NAC gates AddCollection: authorized node owner succeeds, no or wrong identity returns NotAuthorizedError. |

### `add_dac_policy_test.go`

Tests that NAC gates the `AddDACPolicy` operation: the node owner can add policies, while no-identity and wrong-identity callers are rejected.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesAddDACPolicy_AuthorizedIdentity_AllowAccess` | 21-41 | NAC gates AddDACPolicy: authorized node owner identity can add a DAC policy. |
| `TestNAC_GatesAddDACPolicy_NoIdentity_NotAuthorizedError` | 43-64 | NAC gates AddDACPolicy: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesAddDACPolicy_WrongIdentity_NotAuthorizedError` | 66-87 | NAC gates AddDACPolicy: request with wrong identity returns NotAuthorizedError. |

### `add_dac_relationship_test.go`

Tests that NAC gates the `AddDACRelationship` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesAddDACRelationship_AuthorizedIdentity_AllowAccess` | 22-61 | NAC gates AddDACRelationship: authorized node owner identity can add a DAC relationship. |
| `TestNAC_GatesAddDACRelationship_NoIdentity_NotAuthorizedError` | 63-102 | NAC gates AddDACRelationship: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesAddDACRelationship_WrongIdentity_NotAuthorizedError` | 104-143 | NAC gates AddDACRelationship: request with wrong identity returns NotAuthorizedError. |

### `add_document_test.go`

Tests that NAC gates the `AddDocument` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesAddDocument_AuthorizedIdentity_AllowAccess` | 21-55 | NAC gates AddDocument: authorized node owner identity can add a document. |
| `TestNAC_GatesAddDocument_NoIdentity_NotAuthorizedError` | 57-96 | NAC gates AddDocument: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesAddDocument_WrongIdentity_NotAuthorizedError` | 98-137 | NAC gates AddDocument: request with wrong identity returns NotAuthorizedError. |

### `add_lens_test.go`

Tests that NAC gates the `AddLens` (lens migration) operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesAddLens_AuthorizedIdentity_AllowAccess` | 25-55 | NAC gates AddLens: authorized node owner identity can add a lens migration. |
| `TestNAC_GatesAddLens_NoIdentity_NotAuthorizedError` | 57-88 | NAC gates AddLens: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesAddLens_WrongIdentity_NotAuthorizedError` | 90-121 | NAC gates AddLens: request with wrong identity returns NotAuthorizedError. |

### `add_nac_relationship_test.go`

Tests the full lifecycle of adding NAC actor relationships, covering configuration preconditions, invalid identity combinations, invalid relation names, successful grant, all-identity wildcard access, and the caveat that explicit identity is always required.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AddRelationshipWhenNACNotConfiguredBefore_Error` | 22-53 | Adding a NAC relationship when NAC has not been configured returns an error. |
| `TestNAC_AddRelationshipWhenNACIsEnabledWithInvalidIdentities_Error` | 55-109 | Adding a NAC relationship with no or wrong identity when NAC is enabled returns an error. |
| `TestNAC_AddRelationshipWhenNACIsDisabledWithInvalidIdentities_Error` | 111-167 | Adding a NAC relationship with no or wrong identity when NAC is temporarily disabled returns an error. |
| `TestNAC_AddRelationshipWithInvalidRelationName_Error` | 169-189 | Adding a NAC relationship with an invalid relation name returns an error. |
| `TestNAC_AddRelationshipWithValidIdentity_RelationshipAdded` | 191-236 | Adding a NAC relationship with a valid node owner identity succeeds. |
| `TestNAC_AddRelationshipForAllIdentities_AllIdentitiesCanAccess` | 238-271 | Adding a NAC relationship for all identities grants access to any identity. |
| `TestNAC_AddRelationshipStillRequiresIdentityEvenIfAllIdentitiesGivenAccess_StillNeedIdentity` | 273-310 | Even with all-identity NAC access, an explicit identity is still required to authenticate. |

### `add_p2p_collection_test.go`

Tests that NAC gates the `AddP2PCollection` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesAddP2PCollection_AuthorizedIdentity_AllowAccess` | 25-72 | NAC gates AddP2PCollection: authorized node owner identity can add a P2P collection subscription. |
| `TestNAC_GatesAddP2PCollection_NoIdentity_NotAuthorizedError` | 74-122 | NAC gates AddP2PCollection: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesAddP2PCollection_WrongIdentity_NotAuthorizedError` | 124-172 | NAC gates AddP2PCollection: request with wrong identity returns NotAuthorizedError. |

### `add_p2p_document_test.go`

Tests that NAC gates the `AddP2PDocument` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesAddP2PDocument_AuthorizedIdentity_AllowAccess` | 25-80 | NAC gates AddP2PDocument: authorized node owner identity can add a P2P document subscription. |
| `TestNAC_GatesAddP2PDocument_NoIdentity_NotAuthorizedError` | 82-138 | NAC gates AddP2PDocument: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesAddP2PDocument_WrongIdentity_NotAuthorizedError` | 140-196 | NAC gates AddP2PDocument: request with wrong identity returns NotAuthorizedError. |

### `add_p2p_replicator_test.go`

Tests that NAC gates the `AddP2PReplicator` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesAddP2PReplicator_AuthorizedIdentity_AllowAccess` | 24-56 | NAC gates AddP2PReplicator: authorized node owner identity can add a P2P replicator. |
| `TestNAC_GatesAddP2PReplicator_NoIdentity_NotAuthorizedError` | 58-83 | NAC gates AddP2PReplicator: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesAddP2PReplicator_WrongIdentity_NotAuthorizedError` | 85-110 | NAC gates AddP2PReplicator: request with wrong identity returns NotAuthorizedError. |

### `add_view_test.go`

Tests that NAC gates the `AddView` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesAddView_AuthorizedIdentity_AllowAccess` | 22-60 | NAC gates AddView: authorized node owner identity can add a view. |
| `TestNAC_GatesAddView_NoIdentity_NotAuthorizedError` | 62-101 | NAC gates AddView: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesAddView_WrongIdentity_NotAuthorizedError` | 103-142 | NAC gates AddView: request with wrong identity returns NotAuthorizedError. |

### `connect_p2p_peer_test.go`

Tests that NAC gates the `ConnectP2PPeer` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesConnectP2PPeer_AuthorizedIdentity_AllowAccess` | 24-56 | NAC gates ConnectP2PPeer: authorized node owner identity can connect to a P2P peer. |
| `TestNAC_GatesConnectP2PPeer_NoIdentity_NotAuthorizedError` | 58-83 | NAC gates ConnectP2PPeer: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesConnectP2PPeer_WrongIdentity_NotAuthorizedError` | 85-110 | NAC gates ConnectP2PPeer: request with wrong identity returns NotAuthorizedError. |

### `dac_access_by_empty_user_nac_off_test.go`

Tests that when NAC is temporarily disabled, empty-user access to documents falls back entirely to DAC rules (private documents are still blocked by DAC, public documents are accessible).

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_Disabled_WithDACEnabled_AccessByEmptyUser_PrivateDocumentOwnedByNodeOwner_CanNotAccess` | 21-72 | With NAC temporarily disabled and DAC enabled, an empty user cannot access a private document owned by the node owner. |
| `TestNAC_Disabled_WithDACEnabled_AccessByEmptyUser_PrivateDocumentOwnedByNonNodeOwner_CanNotAccess` | 74-124 | With NAC temporarily disabled and DAC enabled, an empty user cannot access a private document owned by a non-node-owner. |
| `TestNAC_Disabled_WithDACEnabled_AccessEmptyUser_PublicDocument_CanAccess` | 126-179 | With NAC temporarily disabled and DAC enabled, an empty user can access a public document. |

### `dac_access_by_empty_user_nac_on_test.go`

Tests that when NAC is enabled, empty (unauthenticated) users are blocked by NAC for all document types, including public documents and via both cacheless and materialized view paths.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_WithDACEnabled_AccessByEmptyUser_PrivateDocumentOwnedByNodeOwner_NotAuthorizedError` | 24-70 | With NAC enabled, an empty user attempting to access a private node-owner document receives a NotAuthorizedError. |
| `TestNAC_WithDACEnabled_AccessByEmptyUser_PrivateDocumentOwnedByNodeOwner_MaterializedView_NotAuthorizedError` | 72-118 | With NAC enabled, an empty user accessing a private node-owner document via materialized view gets a refresh-view NotAuthorizedError. |
| `TestNAC_WithDACEnabled_AccessByEmptyUser_PrivateDocumentOwnedByNonNodeOwner_NotAuthorizedError` | 120-173 | With NAC enabled, an empty user attempting to access a private non-node-owner document receives a NotAuthorizedError. |
| `TestNAC_WithDACEnabled_AccessByEmptyUser_PrivateDocumentOwnedByNonNodeOwner_MaterializedView_NotAuthorizedError` | 175-228 | With NAC enabled, an empty user accessing a private non-node-owner document via materialized view gets a refresh-view NotAuthorizedError. |
| `TestNAC_WithDACEnabled_AccessEmptyUser_PublicDocument_NotAuthorizedError` | 230-283 | With NAC enabled, even a public document is inaccessible to an empty user, returning NotAuthorizedError. |
| `TestNAC_WithDACEnabled_AccessEmptyUser_PublicDocument_MaterializedView_NotAuthorizedError` | 285-338 | With NAC enabled, accessing a public document via materialized view without identity gets a refresh-view NotAuthorizedError. |

### `dac_access_by_node_owner_nac_on_test.go`

Tests that when NAC is enabled, the node owner (the identity that configured NAC) has implicit DAC bypass and can access all documents regardless of ownership.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_WithDACEnabled_AccessByNodeOwner_DoesNotOwnTheDocument_AllowAccess` | 21-73 | With NAC enabled, the node owner can bypass DAC and access a document they do not own. |
| `TestNAC_WithDACEnabled_AccessByNodeOwner_OwnsTheDocument_AllowAccess` | 75-120 | With NAC enabled, the node owner can access their own private document. |
| `TestNAC_WithDACEnabled_AccessByNodeOwner_PublicDocument_AllowAccess` | 121-173 | With NAC enabled, the node owner can access a public document. |

### `dac_access_by_owner_nac_off_test.go`

Tests that when NAC is disabled, the node owner identity has no special bypass and its document access is governed solely by DAC rules.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_Disabled_WithDACEnabled_AccessByNodeOwner_DoesNotOwnTheDocument_CanNotAccess` | 21-72 | With NAC disabled, the node owner identity cannot bypass DAC and cannot access a document they do not own. |
| `TestNAC_Disabled_WithDACEnabled_AccessByNodeOwner_OwnsTheDocument_CanAccess` | 74-128 | With NAC disabled, the node owner can access their own private document via DAC. |
| `TestNAC_Disabled_WithDACEnabled_AccessByNodeOwner_PublicDocument_CanAccess` | 129-182 | With NAC disabled, the node owner can access public documents. |

### `dac_access_by_wrong_user_nac_off_test.go`

Tests that when NAC is disabled, non-node-owner identities access documents via DAC only, with no NAC interference.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_Disabled_WithDACEnabled_AccessByNonNodeOwner_OwnsTheDocument_CanAccess` | 21-74 | With NAC disabled, a non-node-owner can access their own private documents via DAC. |
| `TestNAC_Disabled_WithDACEnabled_AccessByNonNodeOwner_DoesNotOwnTheDocument_CanNotAccess` | 76-126 | With NAC disabled, a non-node-owner cannot access documents they do not own. |
| `TestNAC_Disabled_WithDACEnabled_AccessByNonNodeOwner_PublicDocument_CanAccess` | 128-181 | With NAC disabled, a non-node-owner can access public documents. |

### `dac_access_by_wrong_user_nac_on_test.go`

Tests that when NAC is enabled, non-node-owner identities are blocked by NAC for private documents but may access public documents, across both cacheless and materialized view paths.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_WithDACEnabled_AccessByNonNodeOwner_OwnsTheDocument_NotAuthorizedError` | 24-77 | With NAC enabled, a non-node-owner who owns the document gets NotAuthorizedError from NAC. |
| `TestNAC_WithDACEnabled_AccessByNonNodeOwner_OwnsTheDocument_MaterializedView_NotAuthorizedError` | 79-132 | With NAC enabled, a non-node-owner accessing their own document via materialized view gets a refresh-view NotAuthorizedError. |
| `TestNAC_WithDACEnabled_AccessByNonNodeOwner_DoesNotOwnTheDocument_NotAuthorizedError` | 134-187 | With NAC enabled, a non-node-owner who does not own the document gets NotAuthorizedError. |
| `TestNAC_WithDACEnabled_AccessByNonNodeOwner_DoesNotOwnTheDocument_MaterializedView_NotAuthorizedError` | 189-242 | With NAC enabled, a non-node-owner accessing a document they do not own via materialized view gets a refresh-view NotAuthorizedError. |
| `TestNAC_WithDACEnabled_AccessByNonNodeOwner_PublicDocument_AllowAccess` | 244-297 | With NAC enabled, a non-node-owner can access a public document without restriction. |
| `TestNAC_WithDACEnabled_AccessByNonNodeOwner_PublicDocument_MaterializedView_NotAuthorizedError` | 299-352 | With NAC enabled, a non-node-owner accessing a public document via materialized view gets a refresh-view NotAuthorizedError. |

### `delete_dac_relationship_test.go`

Tests that NAC gates the `DeleteDACRelationship` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesDeleteDACRelationship_AuthorizedIdentity_AllowAccess` | 22-69 | NAC gates DeleteDACRelationship: authorized node owner identity can delete a DAC relationship. |
| `TestNAC_GatesDeleteDACRelationship_NoIdentity_NotAuthorizedError` | 71-118 | NAC gates DeleteDACRelationship: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesDeleteDACRelationship_WrongIdentity_NotAuthorizedError` | 120-167 | NAC gates DeleteDACRelationship: request with wrong identity returns NotAuthorizedError. |

### `delete_document_test.go`

Tests that NAC gates the `DeleteDocument` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesDeleteDocument_AuthorizedIdentity_AllowAccess` | 24-70 | NAC gates DeleteDocument: authorized node owner identity can delete a document. |
| `TestNAC_GatesDeleteDocument_NoIdentity_NotAuthorizedError` | 72-127 | NAC gates DeleteDocument: request with no identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesDeleteDocument_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 129-185 | NAC gates DeleteDocument: request with no identity returns NotAuthorizedError (CLI, C, HTTP clients). |
| `TestNAC_GatesDeleteDocument_WrongIdentity_NotAuthorizedError` | 187-242 | NAC gates DeleteDocument: request with wrong identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesDeleteDocument_WrongIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 244-300 | NAC gates DeleteDocument: request with wrong identity returns NotAuthorizedError (CLI, C, HTTP clients). |

### `delete_encrypted_index_test.go`

Tests that NAC gates the `DeleteEncryptedIndex` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesDeleteEncryptedIndex_AuthorizedIdentity_AllowAccess` | 25-68 | NAC gates DeleteEncryptedIndex: authorized node owner identity can delete an encrypted index. |
| `TestNAC_GatesDeleteEncryptedIndex_NoIdentity_NotAuthorizedError` | 70-107 | NAC gates DeleteEncryptedIndex: request with no identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesDeleteEncryptedIndex_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 109-147 | NAC gates DeleteEncryptedIndex: request with no identity returns NotAuthorizedError (CLI, C, HTTP clients). |
| `TestNAC_GatesDeleteEncryptedIndex_WrongIdentity_NotAuthorizedError` | 149-186 | NAC gates DeleteEncryptedIndex: request with wrong identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesDeleteEncryptedIndex_WrongIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 188-226 | NAC gates DeleteEncryptedIndex: request with wrong identity returns NotAuthorizedError (CLI, C, HTTP clients). |

### `delete_index_test.go`

Tests that NAC gates the `DeleteIndex` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesDeleteIndex_AuthorizedIdentity_AllowAccess` | 25-63 | NAC gates DeleteIndex: authorized node owner identity can delete an index. |
| `TestNAC_GatesDeleteIndex_NoIdentity_NotAuthorizedError` | 65-104 | NAC gates DeleteIndex: request with no identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesDeleteIndex_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 106-146 | NAC gates DeleteIndex: request with no identity returns NotAuthorizedError (CLI, C, HTTP clients). |
| `TestNAC_GatesDeleteIndex_WrongIdentity_NotAuthorizedError` | 148-187 | NAC gates DeleteIndex: request with wrong identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesDeleteIndex_WrongIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 189-229 | NAC gates DeleteIndex: request with wrong identity returns NotAuthorizedError (CLI, C, HTTP clients). |

### `delete_nac_relationship_test.go`

Tests the full lifecycle of deleting NAC actor relationships, covering configuration preconditions, invalid identity combinations, invalid relation names, successful deletion, all-identity wildcard revocation, and the explicit-identity requirement.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_DeleteRelationshipWhenNACNotConfiguredBefore_Error` | 22-53 | Deleting a NAC relationship when NAC has not been configured returns an error. |
| `TestNAC_DeleteRelationshipWhenNACIsEnabledWithInvalidIdentities_Error` | 55-109 | Deleting a NAC relationship with no or wrong identity when NAC is enabled returns an error. |
| `TestNAC_DeleteRelationshipWhenNACIsDisabledWithInvalidIdentities_Error` | 111-167 | Deleting a NAC relationship with no or wrong identity when NAC is temporarily disabled returns an error. |
| `TestNAC_DeleteRelationshipWithInvalidRelationName_Error` | 169-189 | Deleting a NAC relationship with an invalid relation name returns an error. |
| `TestNAC_DeleteRelationshipWithValidIdentity_RelationshipDeleted` | 191-235 | Deleting a NAC relationship with a valid node owner identity succeeds. |
| `TestNAC_DeleteRelationshipForAllIdentities_AllImplicitIdentitiesAccessRevoked` | 237-303 | Deleting the all-identity NAC relationship revokes implicit access for all identities. |
| `TestNAC_DeleteRelationshipStillRequiresIdentityEvenIfAllIdentitiesGivenAccess_StillNeedIdentity` | 305-347 | Even with all-identity NAC access granted, explicit identity is still required after deletion of another relation. |

### `delete_p2p_collection_test.go`

Tests that NAC gates the `DeleteP2PCollection` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesDeleteP2PCollection_AuthorizedIdentity_AllowAccess` | 25-77 | NAC gates DeleteP2PCollection: authorized node owner identity can delete a P2P collection subscription. |
| `TestNAC_GatesDeleteP2PCollection_NoIdentity_NotAuthorizedError` | 79-132 | NAC gates DeleteP2PCollection: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesDeleteP2PCollection_WrongIdentity_NotAuthorizedError` | 134-187 | NAC gates DeleteP2PCollection: request with wrong identity returns NotAuthorizedError. |

### `delete_p2p_document_test.go`

Tests that NAC gates the `DeleteP2PDocument` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesDeleteP2PDocument_AuthorizedIdentity_AllowAccess` | 25-87 | NAC gates DeleteP2PDocument: authorized node owner identity can delete a P2P document subscription. |
| `TestNAC_GatesDeleteP2PDocument_NoIdentity_NotAuthorizedError` | 89-152 | NAC gates DeleteP2PDocument: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesDeleteP2PDocument_WrongIdentity_NotAuthorizedError` | 154-217 | NAC gates DeleteP2PDocument: request with wrong identity returns NotAuthorizedError. |

### `delete_p2p_replicator_test.go`

Tests that NAC gates the `DeleteP2PReplicator` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesDeleteP2PReplicator_AuthorizedIdentity_AllowAccess` | 24-61 | NAC gates DeleteP2PReplicator: authorized node owner identity can delete a P2P replicator. |
| `TestNAC_GatesDeleteP2PReplicator_NoIdentity_NotAuthorizedError` | 63-101 | NAC gates DeleteP2PReplicator: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesDeleteP2PReplicator_WrongIdentity_NotAuthorizedError` | 103-141 | NAC gates DeleteP2PReplicator: request with wrong identity returns NotAuthorizedError. |

### `disable_nac_test.go`

Tests the full range of NAC disable behaviour: preconditions (not configured, not enabled), error cases (no identity, wrong identity, already disabled), the success case, and restart persistence.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_DisableNotConfiguredBefore_Error` | 25-36 | Disabling NAC when it has not been configured returns an error. |
| `TestNAC_DisableNotConfiguredBeforeWithIdentity_Error` | 38-50 | Disabling NAC with an identity when it has not been configured returns an error. |
| `TestNAC_DisableWithoutIdentityOnNodeThatHasConfigured_Error` | 52-74 | Disabling NAC without an identity on a configured node returns an error. |
| `TestNAC_DisableWithWrongIdentityOnNodeThatHasConfigured_Error` | 76-99 | Disabling NAC with a wrong identity on a configured node returns an error. |
| `TestNAC_DisableWithIdentityOnNodeThatHasNACConfiguredAndEnabled_Successful` | 101-139 | Disabling NAC with the correct node owner identity on an enabled-NAC node succeeds. |
| `TestNAC_DisableNoIdentityWhenConfiguredAndAlreadyDisabledBefore_Error` | 141-166 | Disabling NAC without an identity when NAC is already disabled returns an error. |
| `TestNAC_DisableWithIdentityWhenConfiguredAndAlreadyDisabledBefore_Error` | 168-194 | Disabling NAC with an identity when NAC is already disabled returns an error. |
| `TestNAC_DisableSuccessfullyThenRestartWithNoArgs_RemainsDisabled` | 196-235 | Disabling NAC successfully then restarting the node without NAC args keeps NAC disabled. |
| `TestNAC_DisableSuccessfullyThenRestartWithStartArgs_RemainsDisabled` | 237-282 | Disabling NAC successfully then restarting with start args keeps NAC disabled. |

### `get_collection_by_id_test.go`

Tests that NAC gates the `GetCollectionByID` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesGetCollectionByID_AuthorizedIdentity_AllowAccess` | 24-45 | NAC gates GetCollectionByID: authorized node owner identity can retrieve a collection by ID. |
| `TestNAC_GatesGetCollectionByID_NoIdentity_NotAuthorizedError` | 47-68 | NAC gates GetCollectionByID: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesGetCollectionByID_WrongIdentity_NotAuthorizedError` | 70-91 | NAC gates GetCollectionByID: request with wrong identity returns NotAuthorizedError. |

### `get_collection_by_name_test.go`

Tests that NAC gates the `GetCollectionByName` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesGetCollectionByName_AuthorizedIdentity_AllowAccess` | 24-65 | NAC gates GetCollectionByName: authorized node owner identity can retrieve a collection by name. |
| `TestNAC_GatesGetCollectionByName_NoIdentity_NotAuthorizedError` | 67-88 | NAC gates GetCollectionByName: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesGetCollectionByName_WrongIdentity_NotAuthorizedError` | 90-111 | NAC gates GetCollectionByName: request with wrong identity returns NotAuthorizedError. |

### `get_collection_by_version_test.go`

Tests that NAC gates the `GetCollectionByVersion` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesGetCollectionByVersion_AuthorizedIdentity_AllowAccess` | 24-45 | NAC gates GetCollectionByVersion: authorized node owner identity can retrieve a collection by version. |
| `TestNAC_GatesGetCollectionByVersion_NoIdentity_NotAuthorizedError` | 47-68 | NAC gates GetCollectionByVersion: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesGetCollectionByVersion_WrongIdentity_NotAuthorizedError` | 70-91 | NAC gates GetCollectionByVersion: request with wrong identity returns NotAuthorizedError. |

### `get_p2p_active_peers_test.go`

Tests that NAC gates the `GetActivePeers` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesGetActivePeers_AuthorizedIdentity_AllowAccess` | 25-56 | NAC gates GetActivePeers: authorized node owner identity can list active P2P peers. |
| `TestNAC_GatesGetActivePeers_NoIdentity_NotAuthorizedError` | 58-82 | NAC gates GetActivePeers: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesGetActivePeers_WrongIdentity_NotAuthorizedError` | 84-108 | NAC gates GetActivePeers: request with wrong identity returns NotAuthorizedError. |

### `get_p2p_peer_info_test.go`

Tests that NAC gates the `GetP2PPeerInfo` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesGetP2PPeerInfo_AuthorizedIdentity_AllowAccess` | 23-55 | NAC gates GetP2PPeerInfo: authorized node owner identity can retrieve P2P peer info. |
| `TestNAC_GatesGetP2PPeerInfo_NoIdentity_NotAuthorizedError` | 57-81 | NAC gates GetP2PPeerInfo: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesGetP2PPeerInfo_WrongIdentity_NotAuthorizedError` | 83-107 | NAC gates GetP2PPeerInfo: request with wrong identity returns NotAuthorizedError. |

### `list_all_encrypted_index_test.go`

Tests that NAC gates the `ListAllEncryptedIndex` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesListAllEncryptedIndex_AuthorizedIdentity_AllowAccess` | 25-51 | NAC gates ListAllEncryptedIndex: authorized node owner identity can list all encrypted indexes. |
| `TestNAC_GatesListAllEncryptedIndex_NoIdentity_NotAuthorizedError` | 53-71 | NAC gates ListAllEncryptedIndex: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesListAllEncryptedIndex_WrongIdentity_NotAuthorizedError` | 73-91 | NAC gates ListAllEncryptedIndex: request with wrong identity returns NotAuthorizedError. |

### `list_collection_test.go`

Tests that NAC gates the `ListCollection` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesListCollection_AuthorizedIdentity_AllowAccess` | 23-43 | NAC gates ListCollection: authorized node owner identity can list all collections. |
| `TestNAC_GatesListCollection_NoIdentity_NotAuthorizedError` | 45-65 | NAC gates ListCollection: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesListCollection_WrongIdentity_NotAuthorizedError` | 67-87 | NAC gates ListCollection: request with wrong identity returns NotAuthorizedError. |

### `list_encrypted_index_test.go`

Tests that NAC gates the `ListEncryptedIndex` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesListEncryptedIndex_AuthorizedIdentity_AllowAccess` | 26-65 | NAC gates ListEncryptedIndex: authorized node owner identity can list encrypted indexes for a collection. |
| `TestNAC_GatesListEncryptedIndex_NoIdentity_NotAuthorizedError` | 67-106 | NAC gates ListEncryptedIndex: request with no identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesListEncryptedIndex_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 108-148 | NAC gates ListEncryptedIndex: request with no identity returns NotAuthorizedError (CLI, C, HTTP clients). |
| `TestNAC_GatesListEncryptedIndex_WrongIdentity_NotAuthorizedError` | 150-189 | NAC gates ListEncryptedIndex: request with wrong identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesListEncryptedIndex_WrongIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 191-231 | NAC gates ListEncryptedIndex: request with wrong identity returns NotAuthorizedError (CLI, C, HTTP clients). |

### `list_index_test.go`

Tests that NAC gates the `ListIndex` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesListIndex_AuthorizedIdentity_AllowAccess` | 26-65 | NAC gates ListIndex: authorized node owner identity can list indexes for a collection. |
| `TestNAC_GatesListIndex_NoIdentity_NotAuthorizedError` | 67-105 | NAC gates ListIndex: request with no identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesListIndex_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 107-147 | NAC gates ListIndex: request with no identity returns NotAuthorizedError (CLI, C, HTTP clients). |
| `TestNAC_GatesListIndex_WrongIdentity_NotAuthorizedError` | 149-187 | NAC gates ListIndex: request with wrong identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesListIndex_WrongIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 189-229 | NAC gates ListIndex: request with wrong identity returns NotAuthorizedError (CLI, C, HTTP clients). |

### `list_lens_test.go`

Tests that NAC gates the `ListLens` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesListLens_AuthorizedIdentity_AllowAccess` | 22-41 | NAC gates ListLens: authorized node owner identity can list lens migrations. |
| `TestNAC_GatesListLens_NoIdentity_NotAuthorizedError` | 43-63 | NAC gates ListLens: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesListLens_WrongIdentity_NotAuthorizedError` | 65-85 | NAC gates ListLens: request with wrong identity returns NotAuthorizedError. |

### `list_p2p_collection_test.go`

Tests that NAC gates the `ListP2PCollection` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesListP2PCollection_AuthorizedIdentity_AllowAccess` | 25-77 | NAC gates ListP2PCollection: authorized node owner identity can list P2P collection subscriptions. |
| `TestNAC_GatesListP2PCollection_NoIdentity_NotAuthorizedError` | 79-131 | NAC gates ListP2PCollection: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesListP2PCollection_WrongIdentity_NotAuthorizedError` | 133-185 | NAC gates ListP2PCollection: request with wrong identity returns NotAuthorizedError. |

### `list_p2p_document_test.go`

Tests that NAC gates the `ListP2PDocument` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesListP2PDocument_AuthorizedIdentity_AllowAccess` | 25-87 | NAC gates ListP2PDocument: authorized node owner identity can list P2P document subscriptions. |
| `TestNAC_GatesListP2PDocument_NoIdentity_NotAuthorizedError` | 89-149 | NAC gates ListP2PDocument: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesListP2PDocument_WrongIdentity_NotAuthorizedError` | 151-211 | NAC gates ListP2PDocument: request with wrong identity returns NotAuthorizedError. |

### `list_p2p_replicator_test.go`

Tests that NAC gates the `ListP2PReplicator` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesListP2PReplicator_AuthorizedIdentity_AllowAccess` | 25-62 | NAC gates ListP2PReplicator: authorized node owner identity can list P2P replicators. |
| `TestNAC_GatesListP2PReplicator_NoIdentity_NotAuthorizedError` | 64-101 | NAC gates ListP2PReplicator: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesListP2PReplicator_WrongIdentity_NotAuthorizedError` | 103-140 | NAC gates ListP2PReplicator: request with wrong identity returns NotAuthorizedError. |

### `nac_restart_test.go`

Tests that NAC-enabled state persists correctly across node restarts, including when the same identity, no identity, or a different identity is specified at restart time.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_RestartNodeWithNACEnabledWithoutNACArgs_RestartsAndNACIsStillEnabled` | 24-53 | Restarting a NAC-enabled node without NAC args keeps NAC enabled with the original identity. |
| `TestNAC_RestartNodeWithNACEnabledWithExplicitlySpecifyingSameArgs_RestartsAndNACIsStillEnabled` | 55-88 | Restarting a NAC-enabled node with the same NAC args keeps NAC enabled. |
| `TestNAC_RestartNodeWithNACEnabledWithAnotherIdentity_IgnoreNewIdentityAndRestartWithExistingNACState` | 90-123 | Restarting a NAC-enabled node with a different identity ignores the new identity and preserves existing NAC state. |

### `nac_start_test.go`

Tests the NAC status at node startup, covering default (disabled) state and the various combinations of identity and NAC enable flag.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_StartWithDefaultConfig_NACStatusIsDisabled` | 21-32 | Starting a node with default config results in NAC being disabled. |
| `TestNAC_StartWithDefaultConfigWithIdentity_NACStatusIsDisabled` | 34-46 | Starting a node with default config and an identity still results in NAC being disabled. |
| `TestNAC_StartNodeWithIdentityAndWithNACEnableTrue_NACEnabledSuccessfully` | 48-74 | Starting a node with an identity and NAC enable flag set to true enables NAC successfully. |
| `TestNAC_StartNodeNoIdentityWithNACEnableTrue_ErrorAsIdentityIsNeeded` | 76-94 | Starting a node with NAC enable true but no identity returns an error. |
| `TestNAC_StartNodeWithIdentityAndWithNACEnableFalse_NACNotEnabled` | 96-117 | Starting a node with an identity but NAC enable false does not enable NAC. |

### `new_encrypted_index_test.go`

Tests that NAC gates the `NewEncryptedIndex` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesNewEncryptedIndex_AuthorizedIdentity_AllowAccess` | 25-58 | NAC gates NewEncryptedIndex: authorized node owner identity can create a new encrypted index. |
| `TestNAC_GatesNewEncryptedIndex_NoIdentity_NotAuthorizedError` | 60-94 | NAC gates NewEncryptedIndex: request with no identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesNewEncryptedIndex_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 96-131 | NAC gates NewEncryptedIndex: request with no identity returns NotAuthorizedError (CLI, C, HTTP clients). |
| `TestNAC_GatesNewEncryptedIndex_WrongIdentity_NotAuthorizedError` | 133-167 | NAC gates NewEncryptedIndex: request with wrong identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesNewEncryptedIndex_WrongIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 169-204 | NAC gates NewEncryptedIndex: request with wrong identity returns NotAuthorizedError (CLI, C, HTTP clients). |

### `new_index_test.go`

Tests that NAC gates the `NewIndex` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesNewIndex_AuthorizedIdentity_AllowAccess` | 25-64 | NAC gates NewIndex: authorized node owner identity can create a new collection index. |
| `TestNAC_GatesNewIndex_NoIdentity_NotAuthorizedError` | 66-106 | NAC gates NewIndex: request with no identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesNewIndex_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 108-149 | NAC gates NewIndex: request with no identity returns NotAuthorizedError (CLI, C, HTTP clients). |
| `TestNAC_GatesNewIndex_WrongIdentity_NotAuthorizedError` | 151-191 | NAC gates NewIndex: request with wrong identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesNewIndex_WrongIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 193-234 | NAC gates NewIndex: request with wrong identity returns NotAuthorizedError (CLI, C, HTTP clients). |

### `patch_collection_test.go`

Tests that NAC gates the `PatchCollection` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesPatchCollection_AuthorizedIdentity_AllowAccess` | 24-56 | NAC gates PatchCollection: authorized node owner identity can patch a collection schema. |
| `TestNAC_GatesPatchCollection_NoIdentity_NotAuthorizedError` | 58-100 | NAC gates PatchCollection: request with no identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesPatchCollection_NoIdentity_CLIClient_NotAuthorizedError` | 102-142 | NAC gates PatchCollection: request with no identity returns NotAuthorizedError (CLI client). |
| `TestNAC_GatesPatchCollection_WrongIdentity_NotAuthorizedError` | 144-187 | NAC gates PatchCollection: request with wrong identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesPatchCollection_WrongIdentity_CLIClient_NotAuthorizedError` | 189-229 | NAC gates PatchCollection: request with wrong identity returns NotAuthorizedError (CLI client). |

### `read_document_test.go`

Tests that NAC gates the `ReadDocument` operation across all identity cases, including via both cacheless and materialized view paths.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesReadDocument_AuthorizedIdentity_AllowAccess` | 24-58 | NAC gates ReadDocument: authorized node owner identity can read documents. |
| `TestNAC_GatesReadDocument_NoIdentity_NotAuthorizedError` | 60-95 | NAC gates ReadDocument: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesReadDocument_NoIdentity_MaterializedView_NotAuthorizedError` | 97-132 | NAC gates ReadDocument via materialized view: request with no identity returns refresh-view NotAuthorizedError. |
| `TestNAC_GatesReadDocument_WrongIdentity_NotAuthorizedError` | 134-169 | NAC gates ReadDocument: request with wrong identity returns NotAuthorizedError. |
| `TestNAC_GatesReadDocument_WrongIdentity_MaterializedView_NotAuthorizedError` | 171-206 | NAC gates ReadDocument via materialized view: request with wrong identity returns refresh-view NotAuthorizedError. |

### `re_enable_nac_test.go`

Tests the full range of NAC re-enable behaviour: preconditions, error cases (not configured, not disabled, wrong identity, already enabled), the success case, and restart persistence of re-enabled state.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_ReEnableNotConfiguredBefore_Error` | 25-36 | Re-enabling NAC when it has not been configured returns an error. |
| `TestNAC_ReEnableNotConfiguredBeforeWithIdentity_Error` | 38-50 | Re-enabling NAC with an identity when it has not been configured returns an error. |
| `TestNAC_ReEnableWithNoIdentityWhenTemporarilyDisabled_Error` | 52-76 | Re-enabling NAC without an identity when NAC is temporarily disabled returns an error. |
| `TestNAC_ReEnableWithWrongIdentityWhenTemporarilyDisabled_Error` | 78-103 | Re-enabling NAC with a wrong identity when NAC is temporarily disabled returns an error. |
| `TestNAC_ReEnableWithValidIdentityWhenTemporarilyDisabled_NACReEnabled` | 105-135 | Re-enabling NAC with the correct node owner identity when NAC is temporarily disabled succeeds. |
| `TestNAC_ReEnableWithNoIdentityWhenAlreadyEnabled_Error` | 137-159 | Re-enabling NAC without an identity when NAC is already enabled returns an error. |
| `TestNAC_ReEnableWithWrongIdentityWhenAlreadyEnabled_Error` | 161-184 | Re-enabling NAC with a wrong identity when NAC is already enabled returns an error. |
| `TestNAC_ReEnableWithValidIdentityWhenAlreadyEnabled_Error` | 186-209 | Re-enabling NAC with the correct identity when NAC is already enabled returns an error. |
| `TestNAC_ReEnableSuccessfullyThenRestartWithNoArgs_RemainsReEnabled` | 211-254 | Re-enabling NAC then restarting without args keeps NAC enabled. |
| `TestNAC_ReEnableSuccessfullyThenRestartWithStartArgs_RemainsReEnabled` | 256-302 | Re-enabling NAC then restarting with start args keeps NAC enabled. |
| `TestNAC_ReEnableTemporarilyDisabledNACAfterRestart_ReEnabledSuccessfully` | 304-350 | Re-enabling NAC that was temporarily disabled before a restart succeeds after restart. |

### `refresh_view_test.go`

Tests that NAC gates the `RefreshView` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesRefreshView_AuthorizedIdentity_AllowAccess` | 22-41 | NAC gates RefreshView: authorized node owner identity can refresh materialized views. |
| `TestNAC_GatesRefreshView_NoIdentity_NotAuthorizedError` | 43-63 | NAC gates RefreshView: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesRefreshView_WrongIdentity_NotAuthorizedError` | 65-85 | NAC gates RefreshView: request with wrong identity returns NotAuthorizedError. |

### `set_active_collection_version_test.go`

Tests that NAC gates the `SetActiveCollectionVersion` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesSetActiveCollectionVersion_AuthorizedIdentity_AllowAccess` | 24-63 | NAC gates SetActiveCollectionVersion: authorized node owner identity can set the active collection version. |
| `TestNAC_GatesSetActiveCollectionVersion_NoIdentity_NotAuthorizedError` | 65-115 | NAC gates SetActiveCollectionVersion: request with no identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesSetActiveCollectionVersion_NoIdentity_CLIClient_NotAuthorizedError` | 117-164 | NAC gates SetActiveCollectionVersion: request with no identity returns NotAuthorizedError (CLI client). |
| `TestNAC_GatesSetActiveCollectionVersion_WrongIdentity_NotAuthorizedError` | 166-216 | NAC gates SetActiveCollectionVersion: request with wrong identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesSetActiveCollectionVersion_WrongIdentity_CLIClient_NotAuthorizedError` | 218-265 | NAC gates SetActiveCollectionVersion: request with wrong identity returns NotAuthorizedError (CLI client). |

### `set_migration_test.go`

Tests that NAC gates the `SetMigration` (lens migration configuration) operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesSetMigration_AuthorizedIdentity_AllowAccess` | 28-80 | NAC gates SetMigration: authorized node owner identity can configure a lens migration. |
| `TestNAC_GatesSetMigration_NoIdentity_NotAuthorizedError` | 82-117 | NAC gates SetMigration: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesSetMigration_WrongIdentity_NotAuthorizedError` | 119-154 | NAC gates SetMigration: request with wrong identity returns NotAuthorizedError. |

### `sync_p2p_branchable_collection_test.go`

Tests that NAC gates the `SyncP2PBranchableCollection` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesSyncP2PBranchableCollection_AuthorizedIdentity_AllowAccess` | 25-79 | NAC gates SyncP2PBranchableCollection: authorized node owner identity can sync a branchable P2P collection. |
| `TestNAC_GatesSyncP2PBranchableCollection_NoIdentity_NotAuthorizedError` | 81-122 | NAC gates SyncP2PBranchableCollection: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesSyncP2PBranchableCollection_WrongIdentity_NotAuthorizedError` | 124-165 | NAC gates SyncP2PBranchableCollection: request with wrong identity returns NotAuthorizedError. |

### `sync_p2p_collection_versions_test.go`

Tests that NAC gates the `SyncP2PCollectionVersions` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesSyncP2PCollectionVersions_AuthorizedIdentity_AllowAccess` | 25-68 | NAC gates SyncP2PCollectionVersions: authorized node owner identity can sync P2P collection versions. |
| `TestNAC_GatesSyncP2PCollectionVersions_NoIdentity_NotAuthorizedError` | 70-95 | NAC gates SyncP2PCollectionVersions: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesSyncP2PCollectionVersions_WrongIdentity_NotAuthorizedError` | 97-122 | NAC gates SyncP2PCollectionVersions: request with wrong identity returns NotAuthorizedError. |

### `sync_p2p_documents_test.go`

Tests that NAC gates the `SyncP2PDocuments` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesSyncP2PDocuments_AuthorizedIdentity_AllowAccess` | 25-83 | NAC gates SyncP2PDocuments: authorized node owner identity can sync P2P documents. |
| `TestNAC_GatesSyncP2PDocuments_NoIdentity_NotAuthorizedError` | 85-127 | NAC gates SyncP2PDocuments: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesSyncP2PDocuments_WrongIdentity_NotAuthorizedError` | 129-171 | NAC gates SyncP2PDocuments: request with wrong identity returns NotAuthorizedError. |

### `truncate_collection_test.go`

Tests that NAC gates the `TruncateCollection` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesTruncateCollection_AuthorizedIdentity_AllowAccess` | 24-51 | NAC gates TruncateCollection: authorized node owner identity can truncate a collection. |
| `TestNAC_GatesTruncateCollection_NoIdentity_NotAuthorizedError` | 53-90 | NAC gates TruncateCollection: request with no identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesTruncateCollection_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 92-130 | NAC gates TruncateCollection: request with no identity returns NotAuthorizedError (CLI, C, HTTP clients). |
| `TestNAC_GatesTruncateCollection_WrongIdentity_NotAuthorizedError` | 132-168 | NAC gates TruncateCollection: request with wrong identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesTruncateCollection_WrongIdentity_CLIandHTTPClient_NotAuthorizedError` | 170-207 | NAC gates TruncateCollection: request with wrong identity returns NotAuthorizedError (CLI and HTTP clients). |

### `update_document_test.go`

Tests that NAC gates the `UpdateDocument` operation across all identity cases.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesUpdateDocument_AuthorizedIdentity_AllowAccess` | 21-68 | NAC gates UpdateDocument: authorized node owner identity can update a document. |
| `TestNAC_GatesUpdateDocument_NoIdentity_NotAuthorizedError` | 70-124 | NAC gates UpdateDocument: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesUpdateDocument_WrongIdentity_NotAuthorizedError` | 126-180 | NAC gates UpdateDocument: request with wrong identity returns NotAuthorizedError. |

### `update_document_with_filter_test.go`

Tests that NAC gates the `UpdateDocumentWithFilter` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesUpdateDocumentWithFilter_AuthorizedIdentity_AllowAccess` | 24-71 | NAC gates UpdateDocumentWithFilter: authorized node owner identity can update documents using a filter. |
| `TestNAC_GatesUpdateDocumentWithFilter_NoIdentity_NotAuthorizedError` | 73-129 | NAC gates UpdateDocumentWithFilter: request with no identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesUpdateDocumentWithFilter_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 131-188 | NAC gates UpdateDocumentWithFilter: request with no identity returns NotAuthorizedError (CLI, C, HTTP clients). |
| `TestNAC_GatesUpdateDocumentWithFilter_WrongIdentity_NotAuthorizedError` | 190-246 | NAC gates UpdateDocumentWithFilter: request with wrong identity returns NotAuthorizedError (Go client). |
| `TestNAC_GatesUpdateDocumentWithFilter_WrongIdentity_CLIandCandHTTPClient_NotAuthorizedError` | 248-305 | NAC gates UpdateDocumentWithFilter: request with wrong identity returns NotAuthorizedError (CLI, C, HTTP clients). |

### `verify_signature_test.go`

Tests that NAC gates the `VerifySignature` operation across all identity cases and client types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_GatesVerifySignature_AuthorizedIdentity_AllowAccess` | 25-72 | NAC gates VerifySignature: authorized HTTP and CLI client node owner identity can verify a block signature. |
| `TestNAC_GatesVerifySignature_GoClient_AuthorizedIdentity_AllowAccess` | 74-122 | NAC gates VerifySignature: authorized Go and C client node owner identity can verify a block signature. |
| `TestNAC_GatesVerifySignature_NoIdentity_NotAuthorizedError` | 124-155 | NAC gates VerifySignature: request with no identity returns NotAuthorizedError. |
| `TestNAC_GatesVerifySignature_WrongIdentity_NotAuthorizedError` | 157-188 | NAC gates VerifySignature: request with wrong identity returns NotAuthorizedError. |

## Subdirectories

| Directory | Summary |
|---|---|
| [`relation_admin/`](relation_admin/INDEX.md) | Tests that the NAC `admin` relation grants a delegated user permission to perform every NAC-gated node operation, including the DAC bypass, and that revoking the relation removes that access. |
