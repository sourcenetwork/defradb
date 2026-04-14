# Index: `tests/integration/acp/nac/relation_admin`

## Overview

This folder contains integration tests that verify the behavior of the NAC (Node Access Control) `admin` relation in DefraDB. Each test confirms that a user granted the `admin` relation by the node owner gains permission to perform a specific node-level operation that is otherwise blocked, covering the full surface of NAC-gated operations including schema management, document CRUD, DAC interaction, P2P networking, index management, lens migrations, signature verification, and NAC lifecycle control.

## Test Index

### `add_collection_test.go`

Verifies that a user granted the NAC admin relation can add a new collection schema to the node.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanAddCollection` | 22-65 | NAC admin relation grants a user permission to add a collection. |

### `add_dac_policy_test.go`

Verifies that a user granted the NAC admin relation can add a DAC policy to the node.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanAddDACPolicy` | 22-65 | NAC admin relation grants a user permission to add a DAC policy. |

### `add_dac_relationship_test.go`

Verifies that NAC admin relation interacts correctly with DAC manager relation when adding DAC actor relationships.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_WithDACManagerRelation_CanAddDACActorRelationship` | 22-87 | NAC admin with DAC manager relation can add a DAC actor relationship. |
| `TestNAC_AdminRelation_WithoutManagerDACRelation_CanNotAddDACActorRelationship` | 89-155 | NAC admin without DAC manager relation cannot add a DAC actor relationship. |

### `add_document_test.go`

Verifies that a user granted the NAC admin relation can add a document to the node.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanAddDocument` | 21-70 | NAC admin relation grants a user permission to add a document. |

### `add_lens_test.go`

Verifies that a user granted the NAC admin relation can add a lens migration to the node.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanAddLens` | 25-80 | NAC admin relation grants a user permission to add a lens migration. |

### `add_nac_relationship_test.go`

Verifies that a user granted the NAC admin relation can add NAC actor relationships.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanAddNACRelationship` | 21-58 | NAC admin relation grants a user permission to add a NAC actor relationship. |

### `add_p2p_collection_test.go`

Verifies that a user granted the NAC admin relation can add a P2P collection subscription.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanAddP2PCollection` | 25-88 | NAC admin relation grants a user permission to add a P2P collection subscription. |

### `add_p2p_document_test.go`

Verifies that a user granted the NAC admin relation can add a P2P document subscription.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanAddP2PDocument` | 25-98 | NAC admin relation grants a user permission to add a P2P document subscription. |

### `add_p2p_replicator_test.go`

Verifies that a user granted the NAC admin relation can add a P2P replicator.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanAddP2PReplicator` | 24-72 | NAC admin relation grants a user permission to add a P2P replicator. |

### `add_view_test.go`

Verifies that a user granted the NAC admin relation can add a view to the node.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanAddView` | 22-84 | NAC admin relation grants a user permission to add a view. |

### `connect_p2p_peer_test.go`

Verifies that a user granted the NAC admin relation can connect to a P2P peer.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanConnectP2PPeer` | 24-72 | NAC admin relation grants a user permission to connect to a P2P peer. |

### `dac_bypass_with_nac_off_test.go`

Verifies that the NAC admin DAC bypass does not take effect when NAC is disabled, regardless of document ownership or visibility.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_Disabled_AdminRelation_DoesNotOwnTheDocument_CanNotAccessAndCanNotDACBypass` | 21-102 | With NAC disabled, admin relation does not grant DAC bypass for non-owned documents. |
| `TestNAC_Disabled_AdminRelation_OwnThePrivateDocument_CanAccessButNotDACBypass` | 104-193 | With NAC disabled, admin relation does not grant DAC bypass for owned private documents. |
| `TestNAC_Disabled_AdminRelation_PublicDocument_CanAccessButNotDACBypass` | 195-283 | With NAC disabled, admin relation does not grant DAC bypass for public documents. |

### `dac_bypass_with_nac_on_test.go`

Verifies that the NAC admin relation enables DAC bypass when NAC is enabled, and that revoking the relation removes bypass access.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_DoesNotOwnTheDocument_CanBypassDAC` | 24-97 | NAC admin relation allows a user to bypass DAC for documents they do not own. |
| `TestNAC_AdminRelation_DoesNotOwnTheDocument_MaterializedView_CanBypassDAC` | 99-172 | NAC admin relation allows DAC bypass for non-owned documents via materialized view. |
| `TestNAC_AdminRelation_OwnThePrivateDocument_CanBypassDAC` | 174-256 | NAC admin relation allows a user to bypass DAC for their own private document. |
| `TestNAC_AdminRelation_OwnThePrivateDocument_MaterializedView_CanBypassDAC` | 258-340 | NAC admin relation allows DAC bypass for owned private documents via materialized view. |
| `TestNAC_AdminRelation_PublicDocument_CanAccessPublicDocument` | 342-422 | NAC admin relation allows access to public documents that are not gated by DAC. |
| `TestNAC_AdminRelation_PublicDocument_MaterializedView_CanAccessPublicDocument` | 424-504 | NAC admin relation allows access to public documents via materialized view. |
| `TestNAC_AdminRelation_DACByPassRevokation_CanNotDACBypass` | 506-587 | Revoking NAC admin relation removes DAC bypass access and clears the bypass cache. |
| `TestNAC_AdminRelation_DACByPassRevokation_MaterializedView_CanNotDACBypass` | 589-670 | Revoking NAC admin relation removes DAC bypass access via materialized view. |

### `delete_dac_relationship_test.go`

Verifies that NAC admin relation interacts correctly with DAC manager relation when deleting DAC actor relationships.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_WithDACManagerRelation_CanDeleteDACActorRelationship` | 22-95 | NAC admin with DAC manager relation can delete a DAC actor relationship. |
| `TestNAC_AdminRelation_WithoutDACManagerRelation_CanNotDeleteDACActorRelationship` | 97-171 | NAC admin without DAC manager relation cannot delete a DAC actor relationship. |

### `delete_document_test.go`

Verifies that a user granted the NAC admin relation can delete a document, across client type variants.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanDeleteDocument` | 24-88 | NAC admin relation grants a user permission to delete a document. |
| `TestNAC_AdminRelation_CLIandCandHTTPClient_CanDeleteDocument` | 90-155 | NAC admin relation grants CLI, C and HTTP clients permission to delete a document. |

### `delete_encrypted_index_test.go`

Verifies that a user granted the NAC admin relation can delete an encrypted index, across client type variants.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanDeleteEncryptedIndex` | 25-83 | NAC admin relation grants a user permission to delete an encrypted index. |
| `TestNAC_AdminRelation_CLIandCandHTTPClient_CanDeleteEncryptedIndex` | 85-145 | NAC admin relation grants CLI, C and HTTP clients permission to delete an encrypted index. |

### `delete_index_test.go`

Verifies that a user granted the NAC admin relation can delete a collection index, across client type variants.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanDeleteIndex` | 25-77 | NAC admin relation grants a user permission to delete an index. |
| `TestNAC_AdminRelation_CLIandCandHTTPClient_CanDeleteIndex` | 79-133 | NAC admin relation grants CLI, C and HTTP clients permission to delete an index. |

### `delete_nac_relationship_test.go`

Verifies that a user granted the NAC admin relation can delete NAC actor relationships.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanDeleteNACRelationship` | 21-67 | NAC admin relation grants a user permission to delete a NAC actor relationship. |

### `delete_p2p_collection_test.go`

Verifies that a user granted the NAC admin relation can delete a P2P collection subscription.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanDeleteP2PCollection` | 25-93 | NAC admin relation grants a user permission to delete a P2P collection subscription. |

### `delete_p2p_document_test.go`

Verifies that a user granted the NAC admin relation can delete a P2P document subscription.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanDeleteP2PDocument` | 25-105 | NAC admin relation grants a user permission to delete a P2P document subscription. |

### `delete_p2p_replicator_test.go`

Verifies that a user granted the NAC admin relation can delete a P2P replicator.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanDeleteP2PReplicator` | 24-77 | NAC admin relation grants a user permission to delete a P2P replicator. |

### `disable_nac_test.go`

Verifies that a user granted the NAC admin relation can disable NAC.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanDisableNAC` | 22-61 | NAC admin relation grants a user permission to disable NAC. |

### `get_collection_by_id_test.go`

Verifies that a user granted the NAC admin relation can retrieve a collection by its ID.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanGetCollectionByID` | 24-60 | NAC admin relation grants a user permission to get a collection by ID. |

### `get_collection_by_name_test.go`

Verifies that a user granted the NAC admin relation can retrieve a collection by its name.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanGetCollectionByName` | 24-80 | NAC admin relation grants a user permission to get a collection by name. |

### `get_collection_by_version_test.go`

Verifies that a user granted the NAC admin relation can retrieve a collection by its version ID.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanGetCollectionByVersion` | 24-60 | NAC admin relation grants a user permission to get a collection by version ID. |

### `get_nac_status_test.go`

Verifies that a user granted the NAC admin relation can read the current NAC status.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanGetNACStatus` | 22-56 | NAC admin relation grants a user permission to get the NAC status. |

### `get_p2p_active_peers_test.go`

Verifies that a user granted the NAC admin relation can list active P2P peers.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanGetActivePeers` | 25-71 | NAC admin relation grants a user permission to get active P2P peers. |

### `get_p2p_peer_info_test.go`

Verifies that a user granted the NAC admin relation can retrieve P2P peer info.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanGetP2PPeerInfo` | 24-72 | NAC admin relation grants a user permission to get P2P peer info. |

### `list_all_encrypted_index_test.go`

Verifies that a user granted the NAC admin relation can list all encrypted indexes across all collections.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanListAllEncryptedIndex` | 22-55 | NAC admin relation grants a user permission to list all encrypted indexes. |

### `list_collection_test.go`

Verifies that a user granted the NAC admin relation can list all collections on the node.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanListCollection` | 23-57 | NAC admin relation grants a user permission to list all collections. |

### `list_encrypted_index_test.go`

Verifies that a user granted the NAC admin relation can list encrypted indexes for a specific collection, across client type variants.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanListEncryptedIndex` | 26-79 | NAC admin relation grants a user permission to list encrypted indexes for a collection. |
| `TestNAC_AdminRelation_CLIandCandHTTPClient_CanListEncryptedIndex` | 81-136 | NAC admin relation grants CLI, C and HTTP clients permission to list encrypted indexes. |

### `list_index_test.go`

Verifies that a user granted the NAC admin relation can list indexes for a specific collection, across client type variants.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanListIndex` | 26-79 | NAC admin relation grants a user permission to list indexes for a collection. |
| `TestNAC_AdminRelation_CLIandCandHTTPClient_CanListIndex` | 81-136 | NAC admin relation grants CLI, C and HTTP clients permission to list collection indexes. |

### `list_lens_test.go`

Verifies that a user granted the NAC admin relation can list lens migrations.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanListLenses` | 22-55 | NAC admin relation grants a user permission to list lens migrations. |

### `list_p2p_collection_test.go`

Verifies that a user granted the NAC admin relation can list P2P collection subscriptions.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanListP2PCollection` | 25-93 | NAC admin relation grants a user permission to list P2P collection subscriptions. |

### `list_p2p_document_test.go`

Verifies that a user granted the NAC admin relation can list P2P document subscriptions.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanListP2PDocument` | 25-102 | NAC admin relation grants a user permission to list P2P document subscriptions. |

### `list_p2p_replicator_test.go`

Verifies that a user granted the NAC admin relation can list P2P replicators.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanListP2PReplicator` | 25-77 | NAC admin relation grants a user permission to list P2P replicators. |

### `new_encrypted_index_test.go`

Verifies that a user granted the NAC admin relation can create a new encrypted index, across client type variants.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanNewEncryptedIndex` | 25-69 | NAC admin relation grants a user permission to create a new encrypted index. |
| `TestNAC_AdminRelation_CLIandCandHTTPClient_CanNewEncryptedIndex` | 71-117 | NAC admin relation grants CLI, C and HTTP clients permission to create a new encrypted index. |

### `new_index_test.go`

Verifies that a user granted the NAC admin relation can create a new collection index, across client type variants.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanMakeNewIndex` | 25-79 | NAC admin relation grants a user permission to create a new collection index. |
| `TestNAC_AdminRelation_CLIandCandHTTPClient_CanMakeNewIndex` | 81-137 | NAC admin relation grants CLI, C and HTTP clients permission to create a new index. |

### `patch_collection_test.go`

Verifies that a user granted the NAC admin relation can patch a collection schema, across client type variants.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanPatchCollection` | 24-85 | NAC admin relation grants a user permission to patch a collection schema. |
| `TestNAC_AdminRelation_CLIClient_CanPatchCollection` | 87-145 | NAC admin relation grants the CLI client permission to patch a collection schema. |

### `read_document_test.go`

Verifies that a user granted the NAC admin relation can read documents via both cacheless and materialized view types.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanReadDocument` | 24-76 | NAC admin relation grants a user permission to read documents from a collection. |
| `TestNAC_AdminRelation_MaterializedView_CanReadDocument` | 78-130 | NAC admin relation grants a user permission to read documents via a materialized view. |

### `re_enable_nac_test.go`

Verifies that a user granted the NAC admin relation can re-enable NAC after it was disabled.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanReEnableNAC` | 22-66 | NAC admin relation grants a user permission to re-enable NAC after it was disabled. |

### `refresh_view_test.go`

Verifies that a user granted the NAC admin relation can refresh materialized views.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanRefreshViews` | 22-55 | NAC admin relation grants a user permission to refresh materialized views. |

### `set_active_collection_version_test.go`

Verifies that a user granted the NAC admin relation can set the active collection version, across client type variants.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanSetActiveCollectionVersion` | 24-88 | NAC admin relation grants a user permission to set the active collection version. |
| `TestNAC_AdminRelation_CLIClient_CanSetActiveCollectionVersion` | 90-151 | NAC admin relation grants the CLI client permission to set the active collection version. |

### `set_migration_test.go`

Verifies that a user granted the NAC admin relation can configure a lens migration.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanSetMigration` | 25-88 | NAC admin relation grants a user permission to configure a lens migration. |

### `sync_p2p_branchable_collection_test.go`

Verifies that a user granted the NAC admin relation can sync a branchable P2P collection.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanSyncP2PBranchableCollection` | 25-93 | NAC admin relation grants a user permission to sync a branchable P2P collection. |

### `sync_p2p_collection_versions_test.go`

Verifies that a user granted the NAC admin relation can sync P2P collection versions.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanSyncP2PCollectionVersions` | 25-84 | NAC admin relation grants a user permission to sync P2P collection versions. |

### `sync_p2p_documents_test.go`

Verifies that a user granted the NAC admin relation can sync P2P documents across nodes.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanSyncP2PDocuments` | 25-100 | NAC admin relation grants a user permission to sync P2P documents across nodes. |

### `truncate_collection_test.go`

Verifies that a user granted the NAC admin relation can truncate a collection, across client type variants.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanTruncateCollection` | 24-75 | NAC admin relation grants a user permission to truncate a collection. |
| `TestNAC_AdminRelation_CLIandCandHTTPClient_CanTruncateCollection` | 77-129 | NAC admin relation grants CLI, C and HTTP clients permission to truncate a collection. |

### `update_document_test.go`

Verifies that a user granted the NAC admin relation can update a document.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanUpdateDocument` | 21-83 | NAC admin relation grants a user permission to update a document. |

### `verify_signature_test.go`

Verifies that a user granted the NAC admin relation can verify block signatures, across client type variants.

| Test Function | Line | Description |
|---|---|---|
| `TestNAC_AdminRelation_CanVerifySignature` | 25-88 | NAC admin relation grants HTTP and CLI clients permission to verify a block signature. |
| `TestNAC_AdminRelation_GoClient_CanVerifySignature` | 90-154 | NAC admin relation grants Go and C clients permission to verify a block signature. |
