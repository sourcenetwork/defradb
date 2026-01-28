/*
 * DefraDB Rust FFI - Auto-generated header
 *
 * This header provides C bindings for the DefraDB Rust implementation.
 * Use with CGO to integrate with Go code.
 *
 * Usage:
 *   1. Build library: cargo build --release -p ffi
 *   2. Link with: -L target/release -lffi
 */


#ifndef DEFRA_H
#define DEFRA_H

#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

/*
 FFI result type matching Go's Result struct.

 Status codes:
 - 0: Success
 - 1: Error (message in error field)
 - 2: Subscription (ID in value field, not yet implemented)
 */
typedef struct FfiResult {
  /*
   Status code: 0=success, 1=error, 2=subscription
   */
  int status;
  /*
   Error message (null on success). Caller must free with `defra_free_string`.
   */
  char *error;
  /*
   JSON value (null on error). Caller must free with `defra_free_string`.
   */
  char *value;
} FfiResult;

/*
 FFI result for node creation, containing a node handle.

 Matches Go's NewNodeResult struct.
 */
typedef struct NewNodeResult {
  /*
   Status code: 0=success, 1=error
   */
  int status;
  /*
   Error message (null on success). Caller must free with `defra_free_string`.
   */
  char *error;
  /*
   Handle to the node (0 on error).
   */
  uintptr_t node_ptr;
} NewNodeResult;

/*
 Options for node initialization.

 Matches Go's NodeInitOptions struct.
 */
typedef struct NodeInitOptions {
  /*
   Path to store directory (null for in-memory).
   */
  const char *db_path;
  /*
   Use in-memory storage (1=true, 0=false).
   */
  int in_memory;
} NodeInitOptions;

/*
 FFI result for transaction creation, containing a transaction ID.
 */
typedef struct NewTxnResult {
  /*
   Status code: 0=success, 1=error
   */
  int status;
  /*
   Error message (null on success). Caller must free with `defra_free_string`.
   */
  char *error;
  /*
   Transaction ID (null on error). Caller must free with `defra_free_string`.
   */
  char *txn_id;
} NewTxnResult;

/*
 FFI result for subscription creation.
 */
typedef struct CreateSubscriptionResult {
  /*
   Status code: 0=success, 1=error
   */
  int status;
  /*
   Error message (null on success). Caller must free with `defra_free_string`.
   */
  char *error;
  /*
   Subscription handle (0 on error).
   */
  uintptr_t subscription_handle;
} CreateSubscriptionResult;

/*
 FFI result for polling subscriptions.

 Status codes:
 - 0: Event available (value contains JSON event data)
 - 1: Error occurred
 - 2: No event available (subscription open but no pending events)
 - 3: Subscription closed (no more events will arrive)
 */
typedef struct PollSubscriptionResult {
  /*
   Status code (see above)
   */
  int status;
  /*
   Error message (null unless status=1). Caller must free with `defra_free_string`.
   */
  char *error;
  /*
   Event data as JSON (null unless status=0). Caller must free with `defra_free_string`.
   */
  char *value;
  /*
   Number of events dropped due to buffer overflow since last poll.
   When non-zero, the client should re-fetch data to ensure consistency.
   */
  uint64_t dropped_count;
} PollSubscriptionResult;

/*
 FFI result for closing subscriptions.
 */
typedef struct CloseSubscriptionResult {
  /*
   Status code: 0=success, 1=error
   */
  int status;
  /*
   Error message (null on success). Caller must free with `defra_free_string`.
   */
  char *error;
} CloseSubscriptionResult;

#ifdef __cplusplus
extern "C" {
#endif // __cplusplus

/*
 Initialize the FFI library.

 This must be called once before any other FFI functions.
 Safe to call multiple times.
 */
void defra_init(void);

/*
 Get the library version.

 Returns a null-terminated string that must be freed with `defra_free_string`.
 */
char *defra_version(void);

/*
 Get the current NAC status.

 Returns a JSON object with NAC status information:
 ```json
 {
   "status": "enabled" | "disabled_temporarily" | "not_configured",
   "configured_enabled": true | false,
   "dev_mode": true | false,
   "owner": "did:key:..." | null
 }
 ```
 */
struct FfiResult get_nac_status(uintptr_t node_ptr);

/*
 Temporarily disable NAC.

 The requestor_did must be an admin. Returns success on completion.

 # Safety

 `requestor_did` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult disable_nac(uintptr_t node_ptr, const char *requestor_did);

/*
 Re-enable NAC after temporary disable.

 The requestor_did must be an admin. Returns success on completion.

 # Safety

 `requestor_did` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult re_enable_nac(uintptr_t node_ptr, const char *requestor_did);

/*
 Enable NAC with the given owner identity.

 This initializes NAC and sets the owner. Can only be called when NAC
 is not already enabled.

 # Safety

 `owner_did` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult enable_nac(uintptr_t node_ptr, const char *owner_did);

/*
 Add a NAC actor relationship (grant admin to target).

 The requestor must be an admin. Returns JSON with success status:
 ```json
 { "added": true }  // or false if already exists
 ```

 # Safety

 All string parameters must be valid null-terminated UTF-8 strings.
 */
struct FfiResult add_nac_actor_relationship(uintptr_t node_ptr,
                                            const char *requestor_did,
                                            const char *target_did);

/*
 Delete a NAC actor relationship (remove admin from target).

 The requestor must be an admin. The owner cannot be removed.
 Returns JSON with success status:
 ```json
 { "deleted": true }  // or false if didn't exist
 ```

 # Safety

 All string parameters must be valid null-terminated UTF-8 strings.
 */
struct FfiResult delete_nac_actor_relationship(uintptr_t node_ptr,
                                               const char *requestor_did,
                                               const char *target_did);

/*
 Add a DAC policy.

 Accepts a policy definition in YAML or JSON format.
 Returns a JSON object with the policy ID:
 ```json
 { "PolicyID": "sha256_hash_of_policy" }
 ```

 # Safety

 `policy` must be a valid null-terminated UTF-8 string containing
 the policy definition in YAML or JSON format.
 */
struct FfiResult add_dac_policy(uintptr_t node_ptr,
                                const char *identity_did,
                                const char *policy);

/*
 Get a DAC policy by ID.

 Returns a JSON object with the policy content, or null if not found.

 # Safety

 `policy_id` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult get_dac_policy(uintptr_t node_ptr,
                                const char *policy_id);

/*
 List all DAC policy IDs.

 Returns a JSON array of policy IDs.
 */
struct FfiResult list_dac_policies(uintptr_t node_ptr);

/*
 Add a DAC actor relationship (share document access with target).

 The requestor must be the document owner. Relation can be:
 - "reader" - read access
 - "updater" - read + update access
 - "deleter" - read + delete access

 Returns JSON with success status:
 ```json
 { "added": true }  // or false if already exists
 ```

 # Safety

 All string parameters must be valid null-terminated UTF-8 strings.
 */
/*
 Add a DAC policy to the node.

 Registers a policy document and returns its content-addressed ID.

 Returns JSON with the policy ID:
 ```json
 { "PolicyID": "bafyreigh..." }
 ```

 # Safety

 All string parameters must be valid null-terminated UTF-8 strings.
 */
struct FfiResult add_dac_policy(uintptr_t node_ptr,
                                const char *identity_did,
                                const char *policy);

/*
 Get a DAC policy by ID.

 Returns JSON with the policy definition.

 # Safety

 All string parameters must be valid null-terminated UTF-8 strings.
 */
struct FfiResult get_dac_policy(uintptr_t node_ptr,
                                const char *policy_id);

struct FfiResult add_dac_actor_relationship(uintptr_t node_ptr,
                                            const char *requestor_did,
                                            const char *target_did,
                                            const char *collection_id,
                                            const char *doc_id,
                                            const char *relation);

/*
 Delete a DAC actor relationship (revoke document access from target).

 The requestor must be the document owner.

 Returns JSON with success status:
 ```json
 { "deleted": true }  // or false if didn't exist
 ```

 # Safety

 All string parameters must be valid null-terminated UTF-8 strings.
 */
struct FfiResult delete_dac_actor_relationship(uintptr_t node_ptr,
                                               const char *requestor_did,
                                               const char *target_did,
                                               const char *collection_id,
                                               const char *doc_id,
                                               const char *relation);

/*
 Get the node's identity (DID).

 Returns JSON with the node identity:
 ```json
 { "did": "did:key:z6Mk..." }
 ```

 Returns an error if no node identity is configured.
 */
struct FfiResult get_node_identity(uintptr_t node_ptr);

/*
 Create a new DefraDB node.

 This creates an in-memory database instance with a query runner.
 The returned handle must be passed to `node_close` when done.

 # Safety

 The returned `node_ptr` must be freed by calling `node_close`.
 */
struct NewNodeResult new_node(struct NodeInitOptions options);

/*
 Close a DefraDB node and release resources.

 # Safety

 The `node_ptr` must be a valid handle returned by `new_node`.
 After this call, the handle is no longer valid.
 */
struct FfiResult node_close(uintptr_t node_ptr);

/*
 Execute a GraphQL query or mutation.

 Returns a JSON object with the query result in GraphQL format:
 ```json
 {
     "data": { ... },
     "errors": [ ... ]
 }
 ```

 # Arguments

 * `node_ptr` - Handle to the node
 * `request_query` - GraphQL query string (required)
 * `operation_name` - Optional operation name for multi-operation documents (null if not used)
 * `variables` - Optional JSON string of variables (null if not used)

 # Safety

 All string pointers must be either null or valid null-terminated UTF-8 strings.
 */
struct FfiResult exec_request(uintptr_t node_ptr,
                              const char *request_query,
                              const char *operation_name,
                              const char *variables);

/*
 Add a schema to the database.

 The schema should be a GraphQL SDL string defining types.

 Returns a JSON array of CollectionVersion objects on success.

 # Example SDL

 ```graphql
 type User {
     name: String
     age: Int
 }
 ```

 # Safety

 `schema_sdl` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult add_schema(uintptr_t node_ptr, const char *schema_sdl);

/*
 Get all collections from the database.

 Returns a JSON array of collection descriptions.
 */
struct FfiResult get_collections(uintptr_t node_ptr);

/*
 Get a collection by name.

 Returns a JSON object containing the collection's schema (CollectionVersion)
 if found, or an error if the collection doesn't exist.

 # Arguments

 * `node_ptr` - Handle to the node
 * `name` - The collection name

 # Safety

 `name` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult get_collection_by_name(uintptr_t node_ptr, const char *name);

/*
 Check if a collection exists by name.

 Returns "true" or "false" as the value.

 # Arguments

 * `node_ptr` - Handle to the node
 * `name` - The collection name to check

 # Safety

 `name` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult has_collection(uintptr_t node_ptr, const char *name);

/*
 Delete a collection by name.

 Deletes the collection and all its documents.

 # Arguments

 * `node_ptr` - Handle to the node
 * `name` - The collection name to delete

 # Safety

 `name` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult delete_collection(uintptr_t node_ptr, const char *name);

/*
 Find a collection by its collection ID (schema version ID).

 This is useful for P2P sync where we receive blocks with schema_version_id
 and need to find the corresponding collection.

 # Arguments

 * `node_ptr` - Handle to the node
 * `collection_id` - The collection ID (schema version ID)

 # Returns

 Value contains JSON CollectionVersion or "null" if not found.

 # Safety

 `collection_id` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult find_collection_by_id(uintptr_t node_ptr, const char *collection_id);

/*
 Set the active collection version.

 This activates the collection with the given version ID and deactivates
 any other versions of the same collection.

 # Arguments

 * `node_ptr` - Handle to the node
 * `version_id` - The version ID of the collection to activate

 # Safety

 `version_id` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult set_active_collection_version(uintptr_t node_ptr, const char *version_id);

/*
 Patch a collection's schema using JSON patch operations.

 This applies the given JSON patch to the collection's schema,
 validates the result, and updates the collection.

 # Arguments

 * `node_ptr` - Handle to the node
 * `collection_name` - The name of the collection to patch
 * `patch` - A JSON patch string (RFC 6902 format)

 # Returns

 Value contains the updated CollectionVersion as JSON.

 # Safety

 `collection_name` and `patch` must be valid null-terminated UTF-8 strings.
 */
struct FfiResult patch_collection(uintptr_t node_ptr, const char *collection_name, const char *patch);

/*
 Get a collection by its version ID.

 This searches all collections for one matching the given version ID.

 # Arguments

 * `node_ptr` - Handle to the node
 * `version_id` - The version ID to search for

 # Returns

 Value contains JSON CollectionVersion or "null" if not found.

 # Safety

 `version_id` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult get_collection_by_version_id(uintptr_t node_ptr, const char *version_id);

/*
 Add a view to the database.

 Creates a new Defra View from a GQL query and SDL schema.

 # Arguments

 * `node_ptr` - Handle to the node
 * `gql_query` - The GraphQL query defining the view
 * `sdl` - The SDL schema for the view output type
 * `transform` - Optional Lens transform configuration (JSON, null for none)

 # Returns

 Value contains JSON array of CollectionVersions.

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings or null.

 # Note

 Not yet implemented. See issue #178.
 */
struct FfiResult add_view(uintptr_t node_ptr, const char *gql_query, const char *sdl, const char *transform);

/*
 Refresh view caches.

 Refreshes the caches of all views matching the given options.

 # Arguments

 * `node_ptr` - Handle to the node
 * `options` - JSON string of CollectionFetchOptions (null for all views)

 # Safety

 `options` must be null or a valid null-terminated UTF-8 string.

 # Note

 Not yet implemented. See issue #178.
 */
struct FfiResult refresh_views(uintptr_t node_ptr, const char *options);

/*
 Set migration for collection versions.

 Sets the migration for all collections using the given source-destination
 collection version IDs.

 # Arguments

 * `node_ptr` - Handle to the node
 * `config` - JSON string of LensConfig

 # Returns

 Value contains the Lens transform ID.

 # Safety

 `config` must be a valid null-terminated UTF-8 string.

 # Note

 Not yet implemented. See issue #179.
 */
struct FfiResult set_migration(uintptr_t node_ptr, const char *config);

/*
 Begin a new transaction.

 Returns a transaction ID that can be used with `exec_request_in_txn`,
 `commit_txn`, and `rollback_txn`.

 # Arguments

 * `node_ptr` - Handle to the node
 * `readonly` - If non-zero, creates a read-only transaction

 # Returns

 A `NewTxnResult` containing the transaction ID on success.
 */
struct NewTxnResult begin_txn(uintptr_t node_ptr, int32_t readonly);

/*
 Commit a transaction.

 After commit, all operations performed within the transaction become permanent.
 The transaction ID is no longer valid after this call.

 # Arguments

 * `node_ptr` - Handle to the node
 * `txn_id` - Transaction ID from `begin_txn`

 # Safety

 `txn_id` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult commit_txn(uintptr_t node_ptr, const char *txn_id);

/*
 Rollback (discard) a transaction.

 After rollback, all operations performed within the transaction are discarded.
 The transaction ID is no longer valid after this call.

 # Arguments

 * `node_ptr` - Handle to the node
 * `txn_id` - Transaction ID from `begin_txn`

 # Safety

 `txn_id` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult rollback_txn(uintptr_t node_ptr, const char *txn_id);

/*
 Execute a GraphQL query or mutation within a transaction.

 The operation will be part of the specified transaction and will not
 be visible to other transactions until committed.

 # Arguments

 * `node_ptr` - Handle to the node
 * `txn_id` - Transaction ID from `begin_txn`
 * `request_query` - GraphQL query string (required)
 * `operation_name` - Optional operation name (null if not used)
 * `variables` - Optional JSON string of variables (null if not used)

 # Safety

 All string pointers must be either null or valid null-terminated UTF-8 strings.
 */
struct FfiResult exec_request_in_txn(uintptr_t node_ptr,
                                     const char *txn_id,
                                     const char *request_query,
                                     const char *operation_name,
                                     const char *variables);

/*
 Free a string allocated by FFI functions.

 # Safety

 The pointer must have been allocated by an FFI function in this crate.
 */
void defra_free_string(char *ptr);

/*
 Create a new index on a collection.

 # Arguments

 * `node_ptr` - Handle to the node
 * `collection_name` - Name of the collection to create the index on
 * `index_json` - JSON object describing the index to create

 # Index JSON Format

 ```json
 {
     "Name": "my_index",
     "Fields": [
         {"Name": "field1", "Descending": false},
         {"Name": "field2", "Descending": true}
     ],
     "Unique": false
 }
 ```

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings.
 */
struct FfiResult create_index(uintptr_t node_ptr,
                              const char *collection_name,
                              const char *index_json);

/*
 Drop an index from a collection.

 # Arguments

 * `node_ptr` - Handle to the node
 * `collection_name` - Name of the collection
 * `index_name` - Name of the index to drop

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings.
 */
struct FfiResult drop_index(uintptr_t node_ptr,
                            const char *collection_name,
                            const char *index_name);

/*
 Get all indexes for a collection.

 # Arguments

 * `node_ptr` - Handle to the node
 * `collection_name` - Name of the collection

 # Returns

 JSON array of index descriptions.

 # Safety

 `collection_name` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult get_indexes(uintptr_t node_ptr, const char *collection_name);

/*
 Get all indexes across all collections.

 # Arguments

 * `node_ptr` - Handle to the node

 # Returns

 JSON object mapping collection names to their index arrays.
 */
struct FfiResult get_all_indexes(uintptr_t node_ptr);

/*
 Create a subscription to database events.

 # Arguments

 * `node_ptr` - Handle to the node
 * `collection_filter` - Optional collection name to filter events (null for all)

 # Returns

 A handle that can be used with `poll_subscription` and `close_subscription`.

 # Safety

 The collection_filter must be either null or a valid null-terminated UTF-8 string.
 */
struct CreateSubscriptionResult create_subscription(uintptr_t node_ptr,
                                                    const char *collection_filter);

/*
 Poll a subscription for the next event (non-blocking).

 # Arguments

 * `subscription_handle` - Handle from `create_subscription`

 # Returns

 - status=0: Event available (value contains JSON)
 - status=1: Error occurred
 - status=2: No event available yet
 - status=3: Subscription closed
 */
struct PollSubscriptionResult poll_subscription(uintptr_t subscription_handle);

/*
 Close a subscription and release resources.

 # Arguments

 * `subscription_handle` - Handle from `create_subscription`

 # Safety

 After this call, the subscription handle is no longer valid.
 */
struct CloseSubscriptionResult close_subscription(uintptr_t subscription_handle);

// =============================================================================
// P2P Functions
// =============================================================================

/*
 Create a new DefraDB node with P2P enabled.

 # Arguments

 * `options` - Node initialization options
 * `listen_addr` - P2P multiaddr to listen on (e.g., "/ip4/127.0.0.1/tcp/9171")

 # Returns

 A NewNodeResult containing the node handle on success.

 # Safety

 `listen_addr` must be a valid null-terminated UTF-8 string.
 The returned `node_ptr` must be freed by calling `node_close`.
 */
struct NewNodeResult new_node_with_p2p(struct NodeInitOptions options,
                                       const char *listen_addr);

/*
 Get P2P peer info (local peer ID and listening addresses).

 Returns a JSON array of full multiaddrs with peer ID embedded:
 `["/ip4/127.0.0.1/tcp/9171/p2p/12D3KooW..."]`

 # Safety

 The caller must free the returned string with `defra_free_string`.
 */
struct FfiResult p2p_peer_info(uintptr_t node_ptr);

/*
 Get list of connected peers.

 Returns a JSON array of peer IDs.

 # Safety

 The caller must free the returned string with `defra_free_string`.
 */
struct FfiResult p2p_active_peers(uintptr_t node_ptr);

/*
 Connect to a peer at the given multiaddr.

 # Arguments

 * `node_ptr` - Handle to the node
 * `addr` - Full multiaddr including peer ID (e.g., "/ip4/127.0.0.1/tcp/9171/p2p/12D3KooW...")

 # Safety

 `addr` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult p2p_connect(uintptr_t node_ptr, const char *addr);

/*
 Set (add/update) a replicator for collections.

 # Arguments

 * `node_ptr` - Handle to the node
 * `peer_addr` - Full multiaddr of the peer including peer ID
 * `collections_json` - JSON array of collection names

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings.
 */
struct FfiResult p2p_set_replicator(uintptr_t node_ptr,
                                    const char *peer_addr,
                                    const char *collections_json);

/*
 Delete a replicator.

 # Arguments

 * `node_ptr` - Handle to the node
 * `peer_id_str` - Peer ID string (e.g., "12D3KooW...")

 # Safety

 `peer_id_str` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult p2p_delete_replicator(uintptr_t node_ptr, const char *peer_id_str);

/*
 Get all replicators.

 Returns a JSON array of replicator info objects.

 # Safety

 The caller must free the returned string with `defra_free_string`.
 */
struct FfiResult p2p_get_all_replicators(uintptr_t node_ptr);

/*
 Add collections to P2P replication.

 # Arguments

 * `node_ptr` - Handle to the node
 * `collections_json` - JSON array of collection names

 # Safety

 `collections_json` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult p2p_add_collections(uintptr_t node_ptr, const char *collections_json);

/*
 Remove collections from P2P replication.

 # Arguments

 * `node_ptr` - Handle to the node
 * `collections_json` - JSON array of collection names

 # Safety

 `collections_json` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult p2p_remove_collections(uintptr_t node_ptr, const char *collections_json);

/*
 Get all P2P collections.

 Returns a JSON array of collection names.

 # Safety

 The caller must free the returned string with `defra_free_string`.
 */
struct FfiResult p2p_get_all_collections(uintptr_t node_ptr);

#ifdef __cplusplus
}  // extern "C"
#endif  // __cplusplus

#endif  /* DEFRA_H */
