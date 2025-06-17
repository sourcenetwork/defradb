# Searchable Encryption (SE)

Searchable encryption enables privacy-preserving queries on encrypted data in DefraDB. This feature allows nodes to search through encrypted fields without ever accessing the plaintext values or encryption keys, maintaining data privacy while enabling functionality.

## Overview

When a DefraDB collection has encrypted indexes defined, the system generates searchable artifacts during document operations. These artifacts are cryptographic tags that allow pattern matching without revealing the underlying data. The implementation uses a producer-consumer model where the node that creates or updates a document generates search artifacts, while peer nodes store these artifacts to enable distributed search capabilities.

## Architecture and Data Flow

The searchable encryption system is built around several key components that work together through an event-driven architecture.

The `ReplicationCoordinator` serves as the central orchestrator for SE operations. It listens to block commit events from the database layer and coordinates the generation and distribution of search artifacts. The coordinator runs several background goroutines:
- Event listeners for `event.Update`, `se.StoreArtifactsEvent`, and `se.ReplicationFailureEvent`
- A retry handler that periodically processes failed replications

The `Context` structure provides transaction-scoped configuration and state management. Each transaction that involves searchable fields gets its own SE context, which tracks the collection configuration, encrypted field definitions, and accumulates artifacts during the transaction's lifetime.

When a document is created or updated, the following flow occurs:

1. The database layer commits IPLD blocks and publishes an `event.Update` containing the block data, document ID, and collection ID
2. ReplicationCoordinator receives the event and deserializes the block to identify which fields changed
3. For each encrypted field that was modified, the coordinator:
   - Fetches the current document from the collection
   - Retrieves the field value and encodes it deterministically
   - Generates a search tag using HMAC-SHA256 with the collection ID, field name, and encoded value
   - Creates an SE artifact containing the tag and document reference
4. Artifacts are packaged into a `se.ReplicateEvent` and sent to P2P peers
5. Remote peers receive artifacts through `se.StoreArtifactsEvent` and store them under keys structured as `/se/<collectionID>/<indexID>/<searchTag>/<docID>`

## Replication and Reliability

The system includes a retry mechanism for handling replication failures. When a peer fails to process SE artifacts (network issues, node unavailability, etc.), the failure is recorded in the peerstore with retry information. 

Retry information is stored under `/se/retry/<peerID>/<collectionID>/<docID>` keys and includes:
- Document and collection IDs
- Affected field names
- Number of retry attempts
- Next retry timestamp

During retry, the system regenerates artifacts by fetching current document values, ensuring that retries always use the latest data.

## Design Characteristics

Producer nodes do not store SE artifacts locally. This reduces storage overhead on nodes that primarily write data and ensures that search operations naturally distribute load across reader nodes in the network.

The system uses HMAC-SHA256 for tag generation, providing deterministic tags without revealing patterns in the data. The same field value will always produce the same tag, enabling consistent search results across the network.
