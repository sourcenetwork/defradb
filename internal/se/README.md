# Searchable Encryption (SE)

Searchable encryption enables privacy-preserving queries on encrypted data in DefraDB. This feature allows nodes to search through encrypted fields without ever accessing the plaintext values or encryption keys, maintaining data privacy while enabling functionality.

## Overview

When a DefraDB collection has encrypted indexes defined, the system generates searchable artifacts during document operations. These artifacts are cryptographic tags that allow pattern matching without revealing the underlying data. The implementation uses a producer-consumer model where the node that creates or updates a document generates search artifacts, while peer nodes store these artifacts to enable distributed search capabilities.

## Architecture and Data Flow

The searchable encryption system is built around several key components that work together through direct method invocation rather than events.

The `Coordinator` serves as the central orchestrator for SE operations. It is registered as a push handler with the P2P layer and is called automatically when documents are replicated. The coordinator runs a background goroutine for retry handling of failed artifact pushes.

### Doc Creation and Update Flow

When a document is created or updated, the following flow occurs:

1. The database layer commits IPLD blocks and prepares to push updates to replicators
2. P2P layer calls all registered push handlers, including the SE Coordinator's `HandlePushToReplicators` method
3. Coordinator deserializes the block to identify which fields changed
4. For each encrypted field that was modified, the coordinator:
   - Fetches the current document from the collection
   - Retrieves the field value and encodes it deterministically
   - Generates a search tag using HMAC-SHA256 with identity, collection ID, field name, and encoded value
   - Creates an SE artifact containing the tag and document reference
5. Artifacts are sent directly to replicator peers via the P2P communication protocol
6. Remote peers receive artifacts, store them in their datastore, and publish an `SEArtifactReceived` event
7. Artifacts are stored under keys structured as `/se/<collectionID>/<indexID>/<searchTag>/<docID>`

If artifact push fails (e.g., peer offline), the failure is stored in the peerstore with retry metadata for later retry attempts.

### Query Execution

When executing queries on encrypted fields, the system follows a distributed search pattern:

1. The query planner detects filters on encrypted fields and creates an `seScanNode`
2. The scan node extracts field values from filter conditions and creates normal values
3. It calls `db.QueryDocIDsWithSETags()` which delegates to the SE Coordinator
4. The Coordinator:
   - Generates search tags for each field value using the same HMAC-SHA256 computation
   - Identifies all replicator nodes for the collection
   - Sends query requests to each replicator one by one until a successful response is received
5. Each replicator searches its local datastore for matching artifacts and returns document IDs
6. The coordinator returns the list of matching document IDs to the scan node
7. The scan node returns these IDs to the query executor

This approach ensures that encrypted data remains private while enabling efficient distributed search. The query planner handles filter extraction, the coordinator handles tag generation and P2P communication, maintaining clean separation of concerns.

## Replication and Reliability

The system includes a retry mechanism for handling replication failures. When a peer fails to process SE artifacts, the failure is recorded in the peerstore with retry information.

The retry handler runs periodically, checking for failed replications that are due for retry. It uses exponential backoff to avoid overwhelming peers. During retry, the system regenerates artifacts by fetching current document values, ensuring that retries always use the latest data.

## Design Characteristics

Producer nodes do not store SE artifacts locally. This reduces storage overhead on nodes that primarily write data and ensures that search operations naturally distribute load across reader nodes in the network.

The system uses HMAC-SHA256 for tag generation, providing deterministic tags without revealing patterns in the data. The same field value will always produce the same tag for the same identity, enabling consistent search results while isolating data per identity on shared remote nodes. The tag includes the identity ID in its domain separator to prevent cross-identity tag collisions.
