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
 Host callback used for non-exportable signing keys.

 The host app keeps the private key and signs the payload on demand. This is
 intended for device-bound keys such as Secure Enclave-backed P-256 keys.

 Contract:
 - `signer_handle` is an opaque host-managed context value.
 - `did` is the registered signing DID for this key.
 - `key_type` is the registered Defra key type, for example `secp256r1`.
 - `payload_ptr` / `payload_len` describe the exact message bytes to sign.
 - The callback must write the signature bytes into `out_signature_ptr`,
   set `out_signature_len`, and return `0` on success.

 Signature format:
 - For ECDSA keys (`secp256r1`, `secp256k1`), Defra expects an ASN.1 DER /
   X9.62 ECDSA signature, not raw `r || s`.
 - For Secure Enclave P-256 keys on Apple platforms, use
   `SecKeyCreateSignature(..., .ecdsaSignatureMessageX962SHA256, ...)`.
 */
typedef int (*DefraRemoteSignCallback)(uintptr_t signer_handle,
                                       const char *did,
                                       const char *key_type,
                                       const uint8_t *payload_ptr,
                                       uintptr_t payload_len,
                                       uint8_t *out_signature_ptr,
                                       uintptr_t out_signature_capacity,
                                       uintptr_t *out_signature_len);

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
  /*
   Storage backend name: "redb" (default), "fjall", "rocksdb", "memory".
   Null uses "redb" for persistent or "memory" for in-memory.
   */
  const char *datastore_backend;
  /*
   Enable block signing (1=true, 0=false).
   When enabled, the node uses a signing key for block signatures.
   If signing_private_key is provided, that key is used.
   Otherwise, a random secp256k1 key pair is generated.
   */
  int enable_signing;
  /*
   Optional: signing key type string (e.g. "secp256k1", "secp256r1", "ed25519").
   Only used when signing_private_key is provided.
   */
  const char *signing_key_type;
  /*
   Optional: raw private key bytes used for the node identity.
   Block signing remains controlled by enable_signing.
   */
  const uint8_t *signing_private_key;
  /*
   Length of signing_private_key in bytes. 0 if null.
   */
  uintptr_t signing_private_key_len;
  /*
   SourceHub gRPC/LCD address (null = use local ACP).
   */
  const char *sourcehub_grpc_address;
  /*
   SourceHub CometBFT RPC address.
   */
  const char *sourcehub_comet_rpc_address;
  /*
   SourceHub chain ID (e.g., "sourcehub-test").
   */
  const char *sourcehub_chain_id;
  /*
   SourceHub secp256k1 signer key bytes (raw 32-byte private key).
   */
  const uint8_t *sourcehub_signer_key;
  /*
   Length of sourcehub_signer_key. 0 if null.
   */
  uintptr_t sourcehub_signer_key_len;
  /*
   Optional P2P transport backend for `new_node_with_p2p`: "libp2p" (default) or "iroh".
   */
  const char *p2p_transport;
  /*
   Optional iroh relay URL.
   */
  const char *iroh_relay_url;
  /*
   Optional iroh relay mode string: "default", "disabled", or "custom".
   */
  const char *iroh_relay_mode;
  /*
   Optional JSON array of relay URLs for custom relay mode.
   */
  const char *iroh_relay_urls_json;
  /*
   Optional iroh bind address (e.g. "127.0.0.1").
   */
  const char *iroh_bind_addr;
  /*
   Optional iroh bind UDP port. 0 means OS-assigned/unspecified.
   */
  uint16_t iroh_bind_port;
  /*
   Enable iroh discovery (1=true, 0=false). Default true.
   */
  int iroh_discovery;
  /*
   Optional custom DNS origin domain for iroh discovery.
   */
  const char *iroh_discovery_origin_domain;
  /*
   Optional custom pkarr relay URL for iroh discovery publishing.
   */
  const char *iroh_pkarr_relay_url;
  /*
   Optional path to persist the iroh secret key.
   */
  const char *iroh_key_path;
  /*
   Maximum concurrent DAG fetch tasks. 0 means use default.
   */
  uintptr_t max_concurrent_dag_fetches;
  /*
   Maximum concurrent push tasks. 0 means use default.
   */
  uintptr_t max_concurrent_push_tasks;
  /*
   Maximum DocSync request document IDs. 0 means use default (1000).
   */
  uintptr_t max_doc_sync_request_doc_ids;
  /*
   Per-peer rate limit burst capacity. 0 means use default (500).
   */
  uint32_t rate_limit_burst;
  /*
   Per-peer rate limit refill rate (tokens/sec). 0 means use default (50).
   */
  double rate_limit_rate;
} NodeInitOptions;

/*
 Result type for subscription creation.
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
 Result type for polling subscriptions.

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
 Result type for closing subscriptions.
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
struct FfiResult add_dac_policy(uintptr_t node_ptr, const char *identity_did, const char *policy);

/*
 Get a DAC policy by ID.

 Returns a JSON object with the policy content, or null if not found.

 # Safety

 `policy_id` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult get_dac_policy(uintptr_t node_ptr, const char *policy_id);

/*
 List all DAC policy IDs.

 Returns a JSON array of policy IDs.

 # Safety

 No unsafe string parameters.
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
 Register an existing identity for block signing.

 This allows Go-created identities to be used for signing blocks in Rust.
 The identity's signing config is stored in the global identity store.

 # Parameters
 - `did`: The DID string (e.g., "did:key:z6Mk...")
 - `private_key_hex`: Hex-encoded private key bytes
 - `public_key_hex`: Hex-encoded public key bytes
 - `key_type`: Key type ("secp256k1" or "ed25519")

 # Returns
 Success with empty JSON object, or error on failure.
 */
struct FfiResult RegisterIdentity(const char *did,
                                  const char *private_key_hex,
                                  const char *public_key_hex,
                                  const char *key_type);

/*
 Register an identity whose private key stays outside DefraDB.

 The host provides the DID, public key, key type, and a signing callback.
 This supports device-bound keys (for example Secure Enclave P-256 keys) and
 remote identity providers such as Orbis without exporting private key bytes.

 Swift can import this symbol directly from the generated Apple package.
 The callback is required and uses a plain C function-pointer ABI.
 */
struct FfiResult register_remote_identity(const char *did,
                                          const char *public_key_hex,
                                          const char *key_type,
                                          uintptr_t signer_handle,
                                          DefraRemoteSignCallback callback);

/*
 Register an identity whose public key is already available as raw bytes.

 This is intended for Swift and Objective-C callers that already have the
 device public key as `Data` or `UnsafePointer<UInt8>`. The callback contract
 matches [`register_remote_identity`].

 Public key format:
 - `secp256r1` / `secp256k1`: uncompressed SEC1 / X9.63 bytes
 - `ed25519`: raw 32-byte public key
 */
struct FfiResult register_remote_identity_bytes(const char *did,
                                                const uint8_t *public_key_ptr,
                                                uintptr_t public_key_len,
                                                const char *key_type,
                                                uintptr_t signer_handle,
                                                DefraRemoteSignCallback callback);

/*
 Bind a host-provided bearer token to a logical DID for delegated ACP access.

 This is intentionally a passthrough: the token may be signed by a device-
 bound key that has been authorized to act on behalf of the logical DID.
 */
struct FfiResult bind_identity_bearer_token(const char *did, const char *bearer_token);

/*
 Set or clear the node's default signing identity.

 When `did` is empty or null the default identity is cleared. Otherwise the
 DID must already be registered via `RegisterIdentity` or
 `register_remote_identity` / `register_remote_identity_bytes`.
 */
struct FfiResult node_set_default_identity(uintptr_t node_ptr, const char *did);

/*
 Create a new identity (Ed25519 keypair).

 Generates a fresh Ed25519 keypair and returns the DID and private key.
 This is stateless -- no node is required.

 Returns a JSON object:
 ```json
 {
   "did": "did:key:z6Mk...",
   "privateKeyHex": "abcd...",
   "keyType": "ed25519"
 }
 ```
 */
struct FfiResult create_identity(void);

/*
 Get the current NAC status.

 Returns a JSON object with NAC status information:
 ```json
 {
   "status": "enabled" | "disabled temporarily" | "not configured",
   "configured_enabled": true | false,
   "dev_mode": true | false,
   "owner": "did:key:..." | null
 }
 ```

 This function is NAC-gated with the `NacStatus` permission.

 # Safety

 Caller must ensure all pointer arguments are valid, non-null, and point to valid C strings.
 */
struct FfiResult get_nac_status(uintptr_t node_ptr, const char *identity_did);

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
 Add a NAC actor relationship.

 The requestor must be an admin. The relation can be "admin" or a
 specific permission name (e.g., "read-document").

 Returns JSON with success status:
 ```json
 { "added": true }  // or false if already exists
 ```

 # Safety

 All string parameters must be valid null-terminated UTF-8 strings.
 */
struct FfiResult add_nac_actor_relationship(uintptr_t node_ptr,
                                            const char *requestor_did,
                                            const char *relation,
                                            const char *target_did);

/*
 Delete a NAC actor relationship.

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
                                               const char *relation,
                                               const char *target_did);

/*
 List actions that are in progress or ended with an error.

 # Safety

 `identity_did` must be null or a valid null-terminated UTF-8 string.
 */
struct FfiResult list_actions(uintptr_t node_ptr, const char *identity_did);

/*
 Export the database to a JSON file.

 The config_json parameter is a JSON string matching Go's BackupConfig:
 ```json
 {
     "filepath": "/path/to/backup.json",
     "pretty": false,
     "collections": ["User", "Address"]
 }
 ```

 If collections is empty, all collections are exported.

 # Safety

 `config_json` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult basic_export(uintptr_t node_ptr, const char *config_json);

/*
 Import documents from a JSON backup file.

 The file must be a JSON object mapping collection names to arrays of documents:
 ```json
 {
     "User": [{"_docID": "...", "_docIDNew": "...", "name": "John", "age": 30}],
     "Address": [{"_docID": "...", "_docIDNew": "...", "street": "...", "city": "..."}]
 }
 ```

 Self-referencing FK fields are stripped before creation and applied
 via update afterward, matching Go DefraDB behavior.

 # Safety

 `filepath` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult basic_import(uintptr_t node_ptr, const char *filepath);

/*
 Start (or reset) batch CID collection.

 # Arguments

 * `node_ptr` - Handle to the node
 * `identity_did` - Optional DID string (null to use node identity)
 * `session_id` - Optional unique session ID for concurrent batches.
   If null, falls back to the node's public key hex (single-session mode).

 # Safety

 String pointers must be either null or valid null-terminated UTF-8 strings.
 */
struct FfiResult batch_start(uintptr_t node_ptr, const char *identity_did, const char *session_id);

/*
 Sign all collected batch CIDs and return the batch signature as JSON.

 Returns JSON with fields: sig_type, identity, value (hex), merkle_root (hex), cid_count.

 # Arguments

 * `node_ptr` - Handle to the node
 * `identity_did` - Optional DID string (null to use node identity)
 * `session_id` - Optional unique session ID (must match the one passed to `batch_start`).
   If null, falls back to the node's public key hex.

 # Safety

 String pointers must be either null or valid null-terminated UTF-8 strings.
 */
struct FfiResult batch_sign(uintptr_t node_ptr, const char *identity_did, const char *session_id);

/*
 Verify the signature of a block.

 Loads a block from the blockstore by CID, checks that it has a signature,
 loads the signature block, and verifies the signature using the provided
 public key.

 # Arguments

 * `node_ptr` - Handle to the node
 * `key_type` - Key type string (e.g., "ed25519", "secp256k1")
 * `public_key` - Hex-encoded public key string
 * `block_cid` - CID string of the block to verify
 * `identity_did` - DID of the caller for NAC permission check

 # Safety

 All string pointers must be either null or valid null-terminated UTF-8 strings.
 */
struct FfiResult block_verify_signature(uintptr_t node_ptr,
                                        const char *key_type,
                                        const char *public_key,
                                        const char *block_cid,
                                        const char *identity_did);

/*
 Verify the signature of a block using an existing transaction's blockstore view.

 This allows signature verification to see uncommitted blocks created in the same
 transaction while preserving isolation from other transactions.

 # Safety

 All string pointers must be either null or valid null-terminated UTF-8 strings.
 */
struct FfiResult block_verify_signature_in_txn(uintptr_t node_ptr,
                                               const char *txn_id,
                                               const char *key_type,
                                               const char *public_key,
                                               const char *block_cid,
                                               const char *identity_did);

/*
 Eagerly migrate and cache every known-version document in a collection.

 This advances datastore values, stored document-version keys, and secondary
 indexes without creating or broadcasting document commits.

 # Arguments

 * `node_ptr` - Handle to the node
 * `identity_did` - Optional DID for permission checks (null for anonymous)
 * `collection_name` - Name of the collection to materialize

 # Returns

 - Status 0: Success (value contains the number of documents advanced)
 - Status 1: Error (error field contains message)

 # Safety

 `identity_did` and `collection_name` must be valid null-terminated UTF-8
 strings when non-null.
 */
struct FfiResult materialize_collection(uintptr_t node_ptr,
                                        const char *identity_did,
                                        const char *collection_name);

/*
 Set migration for collection versions.

 Sets the migration for all collections using the given source-destination
 collection version IDs.

 # Arguments

 * `node_ptr` - Handle to the node
 * `config` - JSON string of LensConfig containing:
   - `source_version_id`: Source collection version ID
   - `destination_version_id`: Destination collection version ID
   - `lens`: Lens transform configuration

 # Returns

 - Status 0: Success (value contains the Lens transform ID)
 - Status 1: Error (error field contains message)

 # Safety

 `config` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult set_migration(uintptr_t node_ptr, const char *identity_did, const char *config);

/*
 Set a migration within an existing transaction.

 This registers a lens migration configuration within the specified transaction.
 The migration will only be visible after the transaction is committed.

 # Arguments

 * `node_ptr` - Handle to the node
 * `txn_id` - Transaction ID from `begin_txn`
 * `identity_did` - Optional DID for permission checks (null for anonymous)
 * `config` - JSON string containing the lens configuration

 # Returns

 - Status 0: Success (value contains the transform ID)
 - Status 1: Error (error field contains message)

 # Safety

 `txn_id` and `config` must be valid null-terminated UTF-8 strings.
 */
struct FfiResult set_migration_in_txn(uintptr_t node_ptr,
                                      const char *txn_id,
                                      const char *identity_did,
                                      const char *config);

/*
 Delete multiple collection versions by their version IDs.

 Takes a JSON array of version ID strings. Versions are deleted in
 topological order (children before parents).

 # Arguments

 * `node_ptr` - Handle to the node
 * `version_ids_json` - JSON array of version ID strings

 # Returns

 - Status 0: Success (value is "{}")
 - Status 1: Error (error field contains message)

 # Safety

 `version_ids_json` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult delete_collection_versions(uintptr_t node_ptr,
                                            const char *identity_did,
                                            const char *version_ids_json);

/*
 Delete multiple documents by their docIDs.

 Takes a collection name and a JSON array of docID strings,
 deletes each document, and returns the count of deleted documents.

 # Arguments

 * `node_ptr` - Handle to the node
 * `identity_did` - Optional identity DID (nullable)
 * `collection_name` - The collection to delete from
 * `doc_ids_json` - JSON array of docID strings: `["bae-...", "bae-..."]`

 # Returns

 - Status 0: Success (value is `{"deleted": N}`)
 - Status 1: Error (error field contains message)

 # Safety

 `collection_name` and `doc_ids_json` must be valid null-terminated UTF-8 strings.
 */
struct FfiResult delete_documents(uintptr_t node_ptr,
                                  const char *identity_did,
                                  const char *collection_name,
                                  const char *doc_ids_json);

/*
 Get a collection by name.

 Returns a JSON object containing the collection's schema (CollectionVersion)
 if found, or an error if the collection doesn't exist.

 # Arguments

 * `node_ptr` - Handle to the node
 * `name` - The collection name

 # Returns

 - Status 0: Success (value contains JSON CollectionVersion)
 - Status 1: Error (error field contains message)

 # Safety

 `name` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult get_collection_by_name(uintptr_t node_ptr,
                                        const char *identity_did,
                                        const char *name);

/*
 Check if a collection exists by name.

 Returns a JSON boolean: `true` if the collection exists, `false` otherwise.

 # Arguments

 * `node_ptr` - Handle to the node
 * `name` - The collection name to check

 # Returns

 - Status 0: Success (value contains "true" or "false")
 - Status 1: Error (error field contains message)

 # Safety

 `name` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult has_collection(uintptr_t node_ptr, const char *identity_did, const char *name);

/*
 Find a collection by its collection ID (schema version ID).

 This is useful for P2P sync where we receive blocks with schema_version_id
 and need to find the corresponding collection.

 # Arguments

 * `node_ptr` - Handle to the node
 * `collection_id` - The collection ID (schema version ID)

 # Returns

 - Status 0: Success (value contains JSON CollectionVersion or "null" if not found)
 - Status 1: Error (error field contains message)

 # Safety

 `collection_id` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult find_collection_by_id(uintptr_t node_ptr,
                                       const char *identity_did,
                                       const char *collection_id);

/*
 Get a collection by its version ID.

 This searches all collections for one matching the given version ID.

 # Arguments

 * `node_ptr` - Handle to the node
 * `version_id` - The version ID to search for

 # Returns

 - Status 0: Success (value contains JSON CollectionVersion or "null" if not found)
 - Status 1: Error (error field contains message)

 # Safety

 `version_id` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult get_collection_by_version_id(uintptr_t node_ptr,
                                              const char *identity_did,
                                              const char *version_id);

/*
 Add a view to the database.

 Creates a new Defra View from a GQL query and SDL schema.

 # Arguments

 * `node_ptr` - Handle to the node
 * `gql_query` - The GraphQL query defining the view
 * `sdl` - The SDL schema for the view output type
 * `transform` - Optional Lens transform configuration (JSON, null for none)

 # Returns

 - Status 0: Success (value contains JSON array of CollectionVersions)
 - Status 1: Error (error field contains message)

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings or null.
 */
struct FfiResult add_view(uintptr_t node_ptr,
                          const char *identity_did,
                          const char *gql_query,
                          const char *sdl,
                          const char *transform);

/*
 Refresh view caches.

 Refreshes the caches of all materialized views matching the given options.

 # Arguments

 * `node_ptr` - Handle to the node
 * `options` - JSON string with optional "Names" field (null for all views)

 # Returns

 - Status 0: Success (value is "{}")
 - Status 1: Error (error field contains message)

 # Safety

 `options` must be null or a valid null-terminated UTF-8 string.
 */
struct FfiResult refresh_views(uintptr_t node_ptr, const char *identity_did, const char *options);

/*
 Run explicit downsample history GC.

 Applies retention policies to all downsample views matching the given options.

 # Arguments

 * `node_ptr` - Handle to the node
 * `options` - JSON string with optional "Names" field (null for all downsample views)

 # Returns

 - Status 0: Success (value is "{}")
 - Status 1: Error (error field contains message)

 # Safety

 `options` must be null or a valid null-terminated UTF-8 string.
 */
struct FfiResult gc_downsample_histories(uintptr_t node_ptr, const char *options);

/*
 Delete a collection by name.

 Deletes the collection and all its documents.

 # Arguments

 * `node_ptr` - Handle to the node
 * `name` - The collection name to delete

 # Returns

 - Status 0: Success (value is empty)
 - Status 1: Error (error field contains message)

 # Safety

 `name` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult delete_collection(uintptr_t node_ptr, const char *identity_did, const char *name);

/*
 Delete one or more collections by name.

 # Safety

 `names_json` must be a valid null-terminated UTF-8 JSON array of strings.
 */
struct FfiResult delete_collections(uintptr_t node_ptr,
                                    const char *identity_did,
                                    const char *names_json,
                                    bool active_only);

/*
 Delete collections or collection versions within an existing transaction.

 # Safety

 `txn_id` and `targets_json` must be valid null-terminated UTF-8 strings.
 */
struct FfiResult delete_collections_in_txn(uintptr_t node_ptr,
                                           const char *txn_id,
                                           const char *identity_did,
                                           const char *targets_json,
                                           bool active_only);

/*
 Set the active collection version.

 This activates the collection with the given version ID and deactivates
 any other versions of the same collection.

 # Arguments

 * `node_ptr` - Handle to the node
 * `version_id` - The version ID of the collection to activate

 # Returns

 - Status 0: Success (value is "{}")
 - Status 1: Error (error field contains message)

 # Safety

 `version_id` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult set_active_collection_version(uintptr_t node_ptr,
                                               const char *identity_did,
                                               const char *version_id);

/*
 Set a collection version's active state within an existing transaction.

 # Safety

 `txn_id` and `version_id` must be valid null-terminated UTF-8 strings.
 */
struct FfiResult set_collection_active_in_txn(uintptr_t node_ptr,
                                              const char *txn_id,
                                              const char *identity_did,
                                              const char *version_id,
                                              bool is_active);

/*
 Patch a collection's schema using JSON patch operations.

 This applies the given JSON patch to the collection's schema,
 validates the result, and updates the collection.

 # Arguments

 * `node_ptr` - Handle to the node
 * `collection_name` - The name of the collection to patch
 * `patch` - A JSON patch string (RFC 6902 format)

 # Returns

 - Status 0: Success (value contains the updated CollectionVersion as JSON)
 - Status 1: Error (error field contains message)

 # Safety

 `collection_name` and `patch` must be valid null-terminated UTF-8 strings.
 */
struct FfiResult patch_collection(uintptr_t node_ptr,
                                  const char *identity_did,
                                  const char *collection_name,
                                  const char *patch);

/*
 Truncate a collection: delete all documents while preserving the schema.

 This removes all document data, CRDT heads, blocks, and index entries
 for the collection. The collection schema remains intact.

 # Arguments

 * `node_ptr` - Handle to the node
 * `name` - The collection name to truncate

 # Returns

 - Status 0: Success (value is "{}")
 - Status 1: Error (error field contains message)

 # Safety

 `name` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult truncate_collection(uintptr_t node_ptr,
                                     const char *identity_did,
                                     const char *name);

/*
 Create document(s) in a collection.

 This function automatically detects whether the input is a single document
 (JSON object) or multiple documents (JSON array) and handles both cases.

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings.
 */
struct FfiResult collection_create(uintptr_t node_ptr,
                                   const char *identity_did,
                                   const char *collection_name,
                                   const char *json_data,
                                   const char *batch_session_id);

/*
 Check if a JSON string represents an array.

 # Safety

 `json_data` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult is_json_array(const char *json_data);

/*
 Parse a Go-style duration string into nanoseconds.

 Supports Go's duration format: "300ms", "1.5h", "2h45m30s", etc.
 Valid units: ns, us (or µs), ms, s, m, h

 # Safety

 `duration_str` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult parse_duration(const char *duration_str);

/*
 Parse a JSON string array into a vector of strings.

 This function handles both JSON arrays (e.g., `["a", "b", "c"]`) and
 comma-separated strings (e.g., `"a,b,c"`) for backwards compatibility.

 # Safety

 `input` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult parse_string_array(const char *input);

/*
 Add a new encrypted index on a collection field.

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings.
 */
struct FfiResult add_encrypted_index(uintptr_t node_ptr,
                                     const char *identity_did,
                                     const char *collection_name,
                                     const char *field_name);

/*
 Delete an encrypted index from a collection.

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings.
 */
struct FfiResult delete_encrypted_index(uintptr_t node_ptr,
                                        const char *identity_did,
                                        const char *collection_name,
                                        const char *field_name);

/*
 List encrypted indexes for a collection.

 # Safety

 `collection_name` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult list_encrypted_indexes(uintptr_t node_ptr,
                                        const char *identity_did,
                                        const char *collection_name);

/*
 List all encrypted indexes across all collections.

 # Safety

 Caller must ensure all pointer arguments are valid, non-null, and point to valid C strings.
 */
struct FfiResult list_all_encrypted_indexes(uintptr_t node_ptr, const char *identity_did);

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

 # Returns

 JSON object containing the created index description with assigned ID.

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings.
 */
struct FfiResult create_index(uintptr_t node_ptr,
                              const char *identity_did,
                              const char *collection_name,
                              const char *index_json);

/*
 Delete an index from a collection.

 # Arguments

 * `node_ptr` - Handle to the node
 * `collection_name` - Name of the collection
 * `index_name` - Name of the index to delete

 # Returns

 Empty JSON object on success, error on failure.

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings.
 */
struct FfiResult delete_index(uintptr_t node_ptr,
                              const char *identity_did,
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
struct FfiResult get_indexes(uintptr_t node_ptr,
                             const char *identity_did,
                             const char *collection_name);

/*
 Get all indexes across all collections.

 # Arguments

 * `node_ptr` - Handle to the node

 # Returns

 JSON object mapping collection names to their index arrays.

 # Safety

 Caller must ensure all pointer arguments are valid, non-null, and point to valid C strings.
 */
struct FfiResult list_all_indexes(uintptr_t node_ptr, const char *identity_did);

/*
 Add a lens transform to the database.

 Builds Go-compatible IPLD blocks (ConfigBlock -> ModuleBlock -> LensBlock),
 stores them in the blockstore for P2P Bitswap, and registers the transform
 under the ConfigBlock CID.

 # Arguments

 * `node_ptr` - Handle to the node
 * `lens_json` - JSON string containing the lens configuration

 # Returns

 - Status 0: Success (value contains the lens CID)
 - Status 1: Error (error field contains message)

 # Safety

 `lens_json` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult lens_add(uintptr_t node_ptr, const char *identity_did, const char *lens_json);

/*
 Add a lens transform within a transaction.

 The lens is only visible within the transaction until commit.

 # Safety

 Both string pointers must be valid null-terminated UTF-8 strings.
 */
struct FfiResult lens_add_in_txn(uintptr_t node_ptr,
                                 const char *txn_id,
                                 const char *identity_did,
                                 const char *lens_json);

/*
 List all lens transforms.

 Returns a JSON object mapping lens IDs to their configurations.

 # Safety

 The caller must free the returned string with `defra_free_string`.
 */
struct FfiResult lens_list(uintptr_t node_ptr, const char *identity_did);

/*
 List all lens transforms visible within a transaction.

 # Safety

 `txn_id` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult lens_list_in_txn(uintptr_t node_ptr, const char *txn_id, const char *identity_did);

/*
 Initialize the runtime for mobile embedding.
 */
struct FfiResult defra_mobile_init(void);

/*
 Open a node from a single JSON config blob.
 */
struct NewNodeResult defra_mobile_open_node(const char *config_json);

/*
 Close a node opened via the mobile wrapper.
 */
struct FfiResult defra_mobile_close_node(uintptr_t node_ptr);

/*
 Idempotently ensure an SDL schema exists on a node.
 */
struct FfiResult defra_mobile_ensure_schema(uintptr_t node_ptr, const char *schema_sdl);

/*
 Execute a GraphQL request from a single JSON payload.
 */
struct FfiResult defra_mobile_execute(uintptr_t node_ptr, const char *request_json);

/*
 Return local peer info for the configured mobile transport.
 */
struct FfiResult defra_mobile_peer_info(uintptr_t node_ptr);

/*
 Return the node's best shareable P2P address (JSON string, or JSON null
 when the transport has no dialable shareable address yet).
 */
struct FfiResult defra_mobile_shareable_address(uintptr_t node_ptr);

/*
 Return the node's live P2P sync status as JSON.
 */
struct FfiResult defra_mobile_sync_status(uintptr_t node_ptr);

/*
 Connect the node to a peer address.
 */
struct FfiResult defra_mobile_connect(uintptr_t node_ptr, const char *addr);

/*
 Disconnect the node from a peer address.
 */
struct FfiResult defra_mobile_disconnect(uintptr_t node_ptr, const char *addr);

/*
 Notify the embedded iroh transport that network conditions may have changed.
 */
struct FfiResult defra_mobile_notify_network_change(uintptr_t node_ptr);

/*
 Sync branchable collections, schema versions, or specific documents.
 */
struct FfiResult defra_mobile_sync_collection(uintptr_t node_ptr, const char *request_json);

/*
 Add a P2P replicator with optional per-collection filters for mobile embedding.

 `request_json` is camelCase and accepts `collections`, `peerAddr`, optional
 `identityDid`, and optional HTTP-shaped `filters` keyed by collection name.
 */
struct FfiResult defra_mobile_add_replicator(uintptr_t node_ptr, const char *request_json);

/*
 Create a new DefraDB node without P2P.
 */
struct NewNodeResult new_node(struct NodeInitOptions options);

/*
 Close a DefraDB node and release resources.
 */
struct FfiResult node_close(uintptr_t node_ptr);

/*
 Add P2P-enabled collections to the node.

 # Safety

 `identity_did` and `collections_json` must be valid null-terminated UTF-8 strings when
 non-null. `node_ptr` must reference a live node handle created by this library.
 */
struct FfiResult p2p_add_collections(uintptr_t node_ptr,
                                     const char *identity_did,
                                     const char *collections_json);

/*
 Remove P2P-enabled collections from the node.

 # Safety

 `identity_did` and `collections_json` must be valid null-terminated UTF-8 strings when
 non-null. `node_ptr` must reference a live node handle created by this library.
 */
struct FfiResult p2p_delete_collections(uintptr_t node_ptr,
                                        const char *identity_did,
                                        const char *collections_json);

/*
 List P2P-enabled collections for the node.

 # Safety

 `identity_did` must be a valid null-terminated UTF-8 string when non-null. `node_ptr` must
 reference a live node handle created by this library.
 */
struct FfiResult p2p_list_collections(uintptr_t node_ptr, const char *identity_did);

/*
 Add tracked P2P documents to the node.

 # Safety

 `identity_did` and `doc_ids_json` must be valid null-terminated UTF-8 strings when non-null.
 `node_ptr` must reference a live node handle created by this library.
 */
struct FfiResult p2p_add_documents(uintptr_t node_ptr,
                                   const char *identity_did,
                                   const char *doc_ids_json);

/*
 Remove tracked P2P documents from the node.

 # Safety

 `identity_did` and `doc_ids_json` must be valid null-terminated UTF-8 strings when non-null.
 `node_ptr` must reference a live node handle created by this library.
 */
struct FfiResult p2p_delete_documents(uintptr_t node_ptr,
                                      const char *identity_did,
                                      const char *doc_ids_json);

/*
 List tracked P2P documents for the node.

 # Safety

 `identity_did` must be a valid null-terminated UTF-8 string when non-null. `node_ptr` must
 reference a live node handle created by this library.
 */
struct FfiResult p2p_list_documents(uintptr_t node_ptr, const char *identity_did);

/*
 Create a new DefraDB node with P2P enabled.

 # Safety

 Any non-null C string pointers inside `options` and `listen_addr` must be valid
 null-terminated UTF-8 strings for the duration of the call.
 */
struct NewNodeResult new_node_with_p2p(struct NodeInitOptions options, const char *listen_addr);

/*
 Get local P2P peer info.

 # Safety

 `identity_did` must be a valid null-terminated UTF-8 string when non-null. `node_ptr` must
 reference a live node handle created by this library.
 */
struct FfiResult p2p_peer_info(uintptr_t node_ptr, const char *identity_did);

/*
 Get the node's best shareable P2P address as a JSON value.

 Returns the address another node can dial (a `p2p_connect` / replicator
 target) as a JSON string, or JSON `null` when the node has no P2P transport
 or the transport has no shareable address yet (e.g. an iroh endpoint before
 relay/direct-address discovery completes — an identity-only ticket is not
 shareable because a dialing peer resolves it to zero transport addresses).

 This is a stronger contract than `p2p_peer_info`: callers get the one
 address selected for remote sharing instead of guessing among the formatted
 listen addresses.

 # Safety

 `identity_did` must be a valid null-terminated UTF-8 string when non-null. `node_ptr` must
 reference a live node handle created by this library.
 */
struct FfiResult p2p_shareable_address(uintptr_t node_ptr, const char *identity_did);

/*
 Notify the active P2P transport that the network may have changed.

 # Safety

 `identity_did` must be a valid null-terminated UTF-8 string when non-null.
 `node_ptr` must reference a live node handle created by this library.
 */
struct FfiResult p2p_notify_network_change(uintptr_t node_ptr, const char *identity_did);

/*
 Get connected peers.

 # Safety

 `identity_did` must be a valid null-terminated UTF-8 string when non-null.
 `node_ptr` must reference a live node handle created by this library.
 */
struct FfiResult p2p_active_peers(uintptr_t node_ptr, const char *identity_did);

/*
 Connect to a peer address.

 # Safety

 `identity_did` and `addr` must be valid null-terminated UTF-8 strings when non-null.
 `node_ptr` must reference a live node handle created by this library.
 */
struct FfiResult p2p_connect(uintptr_t node_ptr, const char *identity_did, const char *addr);

/*
 Disconnect from a peer address.

 # Safety

 `identity_did` and `addr` must be valid null-terminated UTF-8 strings when non-null.
 `node_ptr` must reference a live node handle created by this library.
 */
struct FfiResult p2p_disconnect(uintptr_t node_ptr, const char *identity_did, const char *addr);

/*
 Retry pushing existing documents to all registered replicators, and
 regenerate/re-push their searchable-encryption artifacts.

 Mirrors Go's `RetryReplicators`: an on-demand retry pass over peerstore retry
 entries that re-pushes failed doc blocks AND the SE artifacts the replicator
 needs to answer `encrypted_<Collection>` queries (#976).

 # Safety

 `node_ptr` must be a valid node handle.
 */
struct FfiResult p2p_retry_replicators(uintptr_t node_ptr);

/*
 Set (add/update) a replicator for collections.

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings.
 */
struct FfiResult p2p_add_replicator(uintptr_t node_ptr,
                                    const char *identity_did,
                                    const char *peer_addr,
                                    const char *collections_json);

/*
 Set (add/update) a replicator for collections with per-collection filters.

 `filters_json` is a JSON object keyed by collection name, mirroring the HTTP
 `"Filters"` field. `null`, `{}`, or an empty string is treated as
 unfiltered, matching [`p2p_add_replicator`]. This symbol is additive;
 [`p2p_add_replicator`] is unchanged for ABI stability.

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings.
 `filters_json` may be null.
 */
struct FfiResult p2p_add_replicator_with_filter(uintptr_t node_ptr,
                                                const char *identity_did,
                                                const char *peer_addr,
                                                const char *collections_json,
                                                const char *filters_json);

/*
 Delete a replicator.

 # Safety

 `peer_id_str` must be a valid null-terminated UTF-8 string.
 */
struct FfiResult p2p_delete_replicator(uintptr_t node_ptr,
                                       const char *identity_did,
                                       const char *peer_id_str,
                                       const char *collections_json);

/*
 Get all replicators.

 Returns a JSON array of replicator info objects.

 # Safety

 The caller must free the returned string with `defra_free_string`.
 */
struct FfiResult p2p_list_replicators(uintptr_t node_ptr, const char *identity_did);

/*
 Return the node's live P2P sync status as JSON.

 # Safety

 `identity_did` must be a valid null-terminated UTF-8 string when non-null.
 `node_ptr` must reference a live node handle created by this library.
 */
struct FfiResult p2p_sync_status(uintptr_t node_ptr, const char *identity_did);

/*
 Sync specific documents from peers.

 # Safety

 All string pointers must be valid null-terminated UTF-8 strings.
 */
struct FfiResult p2p_sync_documents(uintptr_t node_ptr,
                                    const char *identity_did,
                                    const char *collection_name,
                                    const char *doc_ids_json);

/*
 Sync a branchable collection from connected peers.

 # Safety

 `identity_did` and `collection_id` must be valid null-terminated UTF-8 strings.
 */
struct FfiResult p2p_sync_branchable_collection(uintptr_t node_ptr,
                                                const char *identity_did,
                                                const char *collection_id);

/*
 Sync collection versions (schema definitions) from connected peers.

 # Safety

 `identity_did` and `version_ids_json` must be valid null-terminated UTF-8 strings.
 */
struct FfiResult p2p_sync_collection_versions(uintptr_t node_ptr,
                                              const char *identity_did,
                                              const char *version_ids_json);

struct FfiResult defra_profiling_start(const char *output_path);

struct FfiResult defra_profiling_stop(void);

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
 * `identity_did` - Optional DID string for ACP permission checks (null for anonymous)
 * `request_query` - GraphQL query string (required)
 * `operation_name` - Optional operation name for multi-operation documents (null if not used)
 * `variables` - Optional JSON string of variables (null if not used)
 * `batch_session_id` - Optional batch session ID for CID collection (null if not in batch mode).
   When provided, CIDs created during this request are collected under this session.

 # Safety

 All string pointers must be either null or valid null-terminated UTF-8 strings.
 */
struct FfiResult exec_request(uintptr_t node_ptr,
                              const char *identity_did,
                              const char *request_query,
                              const char *operation_name,
                              const char *variables,
                              const char *batch_session_id);

/*
 Execute a GraphQL query or mutation with a request-scoped signing override.

 `signing_override` accepts `-1` for the node default, `0` to disable signing,
 and `1` to enable signing.

 # Safety

 All string pointers must be either null or valid null-terminated UTF-8 strings.
 */
struct FfiResult exec_request_with_signing(uintptr_t node_ptr,
                                           const char *identity_did,
                                           const char *request_query,
                                           const char *operation_name,
                                           const char *variables,
                                           const char *batch_session_id,
                                           int signing_override);

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
struct FfiResult add_schema(uintptr_t node_ptr, const char *identity_did, const char *schema_sdl);

/*
 Add a schema within a specific transaction.

 The schema is only visible within the transaction until it is committed.

 # Safety

 Caller must ensure all pointer arguments are valid, non-null, and point to valid C strings.
 */
struct FfiResult add_schema_in_txn(uintptr_t node_ptr,
                                   const char *txn_id,
                                   const char *identity_did,
                                   const char *schema_sdl);

/*
 Get all collections from the database.

 Returns a JSON array of collection descriptions.

 # Safety

 Caller must ensure all pointer arguments are valid, non-null, and point to valid C strings.
 */
struct FfiResult get_collections(uintptr_t node_ptr, const char *identity_did);

/*
 Get all collection versions visible within a specific transaction.

 This reads from the transaction's systemstore, which includes uncommitted
 writes (e.g., placeholders from set_migration_in_txn).

 # Safety

 Caller must ensure all pointer arguments are valid, non-null, and point to valid C strings.
 */
struct FfiResult get_collections_in_txn(uintptr_t node_ptr,
                                        const char *txn_id,
                                        const char *identity_did);

/*
 Set the searchable encryption key for a node.

 The key is a 32-byte AES-256 key used by the SE coordinator
 for HMAC tag generation during artifact creation.

 # Safety

 * `node_ptr` must be a valid node handle
 * `key_ptr` must point to `key_len` valid bytes
 */
struct FfiResult set_se_encryption_key(uintptr_t node_ptr,
                                       const uint8_t *key_ptr,
                                       uintptr_t key_len);

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
 Create a subscription to P2P merge complete events.

 # Arguments

 * `node_ptr` - Handle to the node

 # Returns

 A handle that can be used with `poll_subscription` and `close_subscription`.
 Events will contain merge complete data (doc_id, cid, collection_id, by_peer).
 */
struct CreateSubscriptionResult create_merge_complete_subscription(uintptr_t node_ptr);

/*
 Poll a subscription for the next event (non-blocking).

 # Arguments

 * `subscription_handle` - Handle from `create_subscription`

 # Returns

 - status=0: Event available (value contains JSON)
 - status=1: Error occurred
 - status=2: No event available yet
 - status=3: Subscription closed

 # Event JSON Format

 ```json
 {
     "type": "update",
     "doc_id": "bae-...",
     "collection_id": "...",
     "block": "<base64-encoded composite block>",
     "is_relay": false
 }
 ```
 */
struct PollSubscriptionResult poll_subscription(uintptr_t subscription_handle);

/*
 Poll a GraphQL subscription for new results.

 Results have already been processed by the background task at event time,
 so this function simply checks the result buffer.
 */
struct PollSubscriptionResult poll_graphql_subscription(const char *subscription_id);

/*
 Close a subscription and release resources.

 # Arguments

 * `subscription_handle` - Handle from `create_subscription`

 # Safety

 After this call, the subscription handle is no longer valid.
 */
struct CloseSubscriptionResult close_subscription(uintptr_t subscription_handle);

/*
 Close a GraphQL subscription and release resources.

 Accepts a string subscription ID and parses it as a numeric handle.
 */
struct CloseSubscriptionResult close_graphql_subscription(const char *subscription_id);

/*
 Execute a GraphQL query or mutation within a transaction.

 The operation will be part of the specified transaction and will not
 be visible to other transactions until committed.

 # Arguments

 * `node_ptr` - Handle to the node
 * `txn_id` - Transaction ID from `begin_txn`
 * `identity_did` - Optional DID of the caller for ACP permission checks (null for anonymous)
 * `request_query` - GraphQL query string (required)
 * `operation_name` - Optional operation name (null if not used)
 * `variables` - Optional JSON string of variables (null if not used)
 * `batch_session_id` - Optional batch session ID for CID collection (null if not in batch mode)

 # Safety

 All string pointers must be either null or valid null-terminated UTF-8 strings.
 */
struct FfiResult exec_request_in_txn(uintptr_t node_ptr,
                                     const char *txn_id,
                                     const char *identity_did,
                                     const char *request_query,
                                     const char *operation_name,
                                     const char *variables,
                                     const char *batch_session_id);

/*
 Execute a GraphQL query or mutation within a transaction with a signing override.

 `signing_override` accepts `-1` for the node default, `0` to disable signing,
 and `1` to enable signing.

 # Safety

 All string pointers must be either null or valid null-terminated UTF-8 strings.
 */
struct FfiResult exec_request_in_txn_with_signing(uintptr_t node_ptr,
                                                  const char *txn_id,
                                                  const char *identity_did,
                                                  const char *request_query,
                                                  const char *operation_name,
                                                  const char *variables,
                                                  const char *batch_session_id,
                                                  int signing_override);

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
 Free a string allocated by FFI functions.

 # Safety

 The pointer must have been allocated by an FFI function in this crate.
 */
void defra_free_string(char *ptr);

#ifdef __cplusplus
}  // extern "C"
#endif  // __cplusplus

#endif  /* DEFRA_H */
