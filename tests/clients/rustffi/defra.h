/*
 * DefraDB C API - generated header.
 *
 * The functions cover the calls of Go DefraDB's cbindings (libdefradb.h),
 * named defra_<noun>_<verb>. Names and types are prefixed and
 * NodeInitOptions adds fields after Go's, so this header is not a drop-in
 * for libdefradb.h. A call that works on a store takes a handle
 * naming a node or one of its transactions, and an identity handle naming
 * the caller (0 for none). Free every result a call returns with its free
 * function: defra_result_free, or defra_new_<kind>_result_free.
 */


#ifndef DEFRA_H
#define DEFRA_H

#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

/*
 The status an asynchronous signature completes with when the request did
 not reach the signer or its answer did not come back. Signing the same
 bytes again is safe, so DefraDB may ask again.
 */
#define DEFRA_SIGN_RETRYABLE 2

/*
 FFI result type matching Go's Result struct.

 Status codes:
 - 0: Success
 - 1: Error (message in error field)
 - 2: Subscription: `defra_query_execute` started a GraphQL subscription,
   whose ID is in the value field; read its events with
   `defra_subscription_poll` and end it with `defra_subscription_close`

 The caller owns its strings and frees them with [`defra_result_free`].
 */
typedef struct DefraResult {
  /*
   Status code: 0=success, 1=error, 2=subscription
   */
  int status;
  /*
   Error message (null on success). Freed with the result.
   */
  char *error;
  /*
   JSON value (null on error). Freed with the result.
   */
  char *value;
} DefraResult;

/*
 FFI result for identity creation, containing an identity handle.

 Matches Go's NewIdentityResult struct. The caller frees it with
 [`defra_new_identity_result_free`].
 */
typedef struct DefraNewIdentityResult {
  /*
   Status code: 0=success, 1=error
   */
  int status;
  /*
   Error message (null on success). Freed with the result.
   */
  char *error;
  /*
   Handle to the identity (0 on error).
   */
  uintptr_t identity_ptr;
} DefraNewIdentityResult;

/*
 Which collection a call acts on, and how it signs.

 Matches Go's CollectionOptions struct. Every string is optional: null or
 empty leaves that criterion out.
 */
typedef struct CollectionOptions {
  /*
   Collection version ID.
   */
  const char *version;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Collection name.
   */
  const char *name;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int get_inactive;
  /*
   Block signing for this call: 0 node default, 1 enabled, -1 disabled.
   */
  int enable_signing;
} CollectionOptions;

/*
 FFI result for node creation, containing a node handle.

 Matches Go's NewNodeResult struct. The caller frees it with
 [`defra_new_node_result_free`].
 */
typedef struct DefraNewNodeResult {
  /*
   Status code: 0=success, 1=error
   */
  int status;
  /*
   Error message (null on success). Freed with the result.
   */
  char *error;
  /*
   Handle to the node (0 on error).
   */
  uintptr_t node_ptr;
} DefraNewNodeResult;

/*
 Options for node initialization.

 Go's NodeInitOptions struct, field for field: an all-zero struct (null
 strings, zero numbers) is the default node, and every opt-out is a
 `disable_` flag. The fields after Go's are this implementation's own, so
 the struct is larger than Go's: a caller passes this header's struct,
 never one laid out by Go's `defra_structs.h`.

 Go's Badger options name the store regolith replaced, so they are named
 for the store here (`store_file_size`, `store_encryption_key`). A Go
 option this implementation does not support yet fails `defra_node_new`
 when it is set, rather than being ignored.
 */
typedef struct NodeInitOptions {
  /*
   Path to the store directory. Null or empty keeps the store in memory.
   */
  const char *db_path;
  /*
   Comma-separated P2P listen multiaddrs. Empty listens on
   `/ip4/0.0.0.0/tcp/9171`, as Go does.
   */
  const char *listening_addresses;
  /*
   Comma-separated seconds between replicator retries, the last one
   repeating. Empty keeps the default ladder.
   */
  const char *replicator_retry_intervals;
  /*
   Comma-separated multiaddrs of peers to connect to once started.
   */
  const char *peers;
  /*
   The node's identity, from `defra_identity_new` (0 for none).
   */
  uintptr_t identity_ptr;
  /*
   Keep the store in memory (1=true, 0=false).
   */
  int in_memory;
  /*
   Run without P2P (1=true, 0=false).
   */
  int disable_p2p;
  /*
   Run without the HTTP API (1=true, 0=false). An embedded node does not
   serve the HTTP API yet.
   */
  int disable_api;
  /*
   Enable node access control at startup, owned by the node identity
   (1=true, 0=false).
   */
  int enable_node_acp;
  /*
   Retries of a conflicting transaction. 0 keeps the default.
   */
  int max_transaction_retries;
  /*
   Store type: "regolith" (default) or "memory". Go's store name,
   "badger", means "regolith".
   */
  const char *store_type;
  /*
   The store's compaction output target in bytes (Go's Badger value log
   file size). 0 keeps the default.
   */
  int64_t store_file_size;
  /*
   32-byte key encrypting the store at rest, or null.
   */
  const uint8_t *store_encryption_key;
  /*
   Length of `store_encryption_key` in bytes.
   */
  int store_encryption_key_len;
  /*
   Sign blocks (1=true, 0=false), with the node identity, or with a
   generated secp256k1 key when there is none.
   */
  int enable_signing;
  /*
   32-byte searchable encryption key, or null.
   */
  const uint8_t *searchable_encryption_key;
  /*
   Length of `searchable_encryption_key` in bytes.
   */
  int searchable_encryption_key_len;
  /*
   Milliseconds a P2P block sync may take. 0 keeps the default.
   */
  int64_t p2p_block_sync_timeout_ms;
  /*
   Lens runtime pool size. Not supported yet: must be 0.
   */
  int lens_pool_size;
  /*
   Bytes per block chunk. 0 keeps the default.
   */
  int chunk_size;
  /*
   Run GossipSub (1=true, 0=false).
   */
  int enable_pub_sub;
  /*
   Run the relay client transport (1=true, 0=false).
   */
  int enable_relay;
  /*
   Clear dial backoff before retrying a replicator. Not supported yet:
   must be 0.
   */
  int enable_clear_backoff_on_retry;
  /*
   The libp2p Ed25519 private key, or null for the node's stored key.
   */
  const uint8_t *p2p_private_key;
  /*
   Length of `p2p_private_key` in bytes.
   */
  int p2p_private_key_len;
  /*
   HTTP API address. Not supported yet: must be null or empty.
   */
  const char *http_address;
  /*
   Comma-separated allowed CORS origins. Not supported yet.
   */
  const char *http_allowed_origins;
  /*
   TLS certificate path. Not supported yet.
   */
  const char *tls_cert_path;
  /*
   TLS key path. Not supported yet.
   */
  const char *tls_key_path;
  /*
   HTTP read timeout in milliseconds. Not supported yet.
   */
  int64_t http_read_timeout_ms;
  /*
   HTTP write timeout in milliseconds. Not supported yet.
   */
  int64_t http_write_timeout_ms;
  /*
   HTTP idle timeout in milliseconds. Not supported yet.
   */
  int64_t http_idle_timeout_ms;
  /*
   Document ACP: "local" (default) or "remote" (Vera, configured by the
   `remote_dac_` and `vera_` fields).
   */
  const char *document_acp_type;
  /*
   Separate document ACP store path. Not supported: document ACP lives
   in the node's store.
   */
  const char *document_acp_path;
  /*
   SourceHub log ID. Not supported: remote document ACP is Vera.
   */
  const char *remote_dac_log_id;
  /*
   Vera gRPC/LCD address.
   */
  const char *remote_dac_grpc_address;
  /*
   Vera CometBFT RPC address.
   */
  const char *remote_dac_comet_rpc_address;
  /*
   Separate node ACP store path. Not supported: node ACP lives in the
   node's store.
   */
  const char *node_acp_path;
  /*
   Vera chain ID (e.g., "vera-test").
   */
  const char *vera_chain_id;
  /*
   Vera secp256k1 signer key bytes (raw 32-byte private key).
   */
  const uint8_t *vera_signer_key;
  /*
   Length of `vera_signer_key` in bytes.
   */
  int vera_signer_key_len;
  /*
   P2P transport: "libp2p" (default) or "iroh".
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
   Run iroh without discovery (1=true, 0=false).
   */
  int iroh_disable_discovery;
  /*
   Optional custom DNS origin domain for iroh discovery.
   */
  const char *iroh_discovery_origin_domain;
  /*
   Optional custom pkarr relay URL for iroh discovery publishing.
   */
  const char *iroh_pkarr_relay_url;
  /*
   Iroh key file from before the node shared one peer key across transports.
   Imported as the node's peer key when the store has none; never written.
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
 Receives the result of an asynchronous call, exactly once.

 - `context` is the value the caller passed with the call.
 - `result` is the call's result, as its synchronous form returns it: the
   caller owns it and frees it with `defra_result_free`.

 The callback runs on a DefraDB thread that serves other work: it must
 return promptly, and a synchronous DefraDB call made from it is refused
 (an asynchronous one is fine). A call that fails before it is submitted
 calls it on the caller's own thread, before the call returns.
 */
typedef void (*DefraResultCallback)(uintptr_t context, struct DefraResult result);

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

 The callback blocks the thread that calls it, so a wave of signatures (the
 blocks of a write) calls it on a thread set aside for blocking work. A
 host whose signer answers later registers a
 [`DefraRemoteSignAsyncCallback`] instead, which holds no thread.

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
 Host callback that starts a signature and returns without waiting for it.

 Contract:
 - `signer_handle`, `did` and `key_type` are as for
   [`DefraRemoteSignCallback`].
 - `payload_ptr` / `payload_len` are the exact message bytes to sign. They,
   `did` and `key_type` are valid only until the callback returns: copy the
   payload if the signature is made later.
 - `completion` names this one request. Return `0` once the signature is
   under way, then report its result exactly once, from any thread, with
   `defra_remote_sign_complete(completion, ...)`; reporting it from inside
   the callback is fine too. Return any other value when the signature
   cannot be started; the request then fails at once and must not be
   reported.
 - The callback runs on a DefraDB thread that serves other work: it must
   return promptly and must not call another DefraDB function that waits.

 Signature format: as for [`DefraRemoteSignCallback`].
 */
typedef int (*DefraRemoteSignAsyncCallback)(uintptr_t signer_handle,
                                            const char *did,
                                            const char *key_type,
                                            const uint8_t *payload_ptr,
                                            uintptr_t payload_len,
                                            uintptr_t completion);

/*
 FFI result for transaction creation, containing a transaction handle.

 Matches Go's NewTxnResult struct. The caller frees it with
 [`defra_new_txn_result_free`].
 */
typedef struct DefraNewTxnResult {
  /*
   Status code: 0=success, 1=error
   */
  int status;
  /*
   Error message (null on success). Freed with the result.
   */
  char *error;
  /*
   Handle to the transaction (0 on error).
   */
  uintptr_t txn_ptr;
} DefraNewTxnResult;

#ifdef __cplusplus
extern "C" {
#endif // __cplusplus

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
 Adds the document access-control policy `policy` (YAML), created by the
 identity `identity_ptr` names, and answers with `{"PolicyID": ...}`.

 # Safety

 `policy` must be null or a valid null-terminated string.
 */
struct DefraResult defra_acp_dac_policy_add(uintptr_t node_ptr,
                                            uintptr_t identity_ptr,
                                            const char *policy);

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
 Adds the relation `relation` between `actor` and document `doc_id` of
 collection `collection`, as the identity `identity_ptr` names, and
 answers with the result as JSON.

 # Safety

 Every string must be null or a valid null-terminated string.
 */
struct DefraResult defra_acp_dac_relationship_add(uintptr_t node_ptr,
                                                  uintptr_t identity_ptr,
                                                  const char *collection,
                                                  const char *doc_id,
                                                  const char *relation,
                                                  const char *actor);

/*
 Delete a DAC actor relationship (revoke document access from target).

 The requestor must be the document owner.

 Returns JSON with success status:
 ```json
 { "deleted": true }  // or false if didn't exist
 ```

 # Safety

 All string parameters must be valid null-terminated UTF-8 strings.
 Deletes the relation `relation` between `actor` and document `doc_id` of
 collection `collection`, as the identity `identity_ptr` names, and
 answers with the result as JSON.

 # Safety

 Every string must be null or a valid null-terminated string.
 */
struct DefraResult defra_acp_dac_relationship_delete(uintptr_t node_ptr,
                                                     uintptr_t identity_ptr,
                                                     const char *collection,
                                                     const char *doc_id,
                                                     const char *relation,
                                                     const char *actor);

/*
 Creates an identity for the logical DID `did`, acting by the bearer token
 the host provides instead of a key of its own: the token may be signed by
 a device-bound key authorized to act on behalf of `did`. Answers with the
 identity's handle.

 # Safety

 `did` and `bearer_token` must be null or valid null-terminated strings.
 */
struct DefraNewIdentityResult defra_identity_new_from_bearer_token(const char *did,
                                                                   const char *bearer_token);

/*
 Sets the node's identity, which signs the node's blocks when signing is
 enabled, to the one `identity_ptr` names; 0 clears it.
 */
struct DefraResult defra_node_set_identity(uintptr_t node_ptr, uintptr_t identity_ptr);

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
 Answers with the node access-control status as JSON.
 */
struct DefraResult defra_acp_nac_status(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Temporarily disable NAC.

 The requestor_did must be an admin. Returns success on completion.

 # Safety

 `requestor_did` must be a valid null-terminated UTF-8 string.
 Disables node access control, as the identity `identity_ptr` names.
 */
struct DefraResult defra_acp_nac_disable(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Re-enable NAC after temporary disable.

 The requestor_did must be an admin. Returns success on completion.

 # Safety

 `requestor_did` must be a valid null-terminated UTF-8 string.
 Re-enables node access control, as the identity `identity_ptr` names.
 */
struct DefraResult defra_acp_nac_re_enable(uintptr_t node_ptr, uintptr_t identity_ptr);

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
 Adds the node access-control relation `relation` ("admin" for null or
 empty) of `actor`, as the identity `identity_ptr` names.

 # Safety

 `relation` and `actor` must be null or valid null-terminated strings.
 */
struct DefraResult defra_acp_nac_relationship_add(uintptr_t node_ptr,
                                                  uintptr_t identity_ptr,
                                                  const char *relation,
                                                  const char *actor);

/*
 Delete a NAC actor relationship.

 The requestor must be an admin. The owner cannot be removed.

 Returns JSON with success status:
 ```json
 { "deleted": true }  // or false if didn't exist
 ```

 # Safety

 All string parameters must be valid null-terminated UTF-8 strings.
 Deletes the node access-control relation `relation` ("admin" for null or
 empty) of `actor`, as the identity `identity_ptr` names.

 # Safety

 `relation` and `actor` must be null or valid null-terminated strings.
 */
struct DefraResult defra_acp_nac_relationship_delete(uintptr_t node_ptr,
                                                     uintptr_t identity_ptr,
                                                     const char *relation,
                                                     const char *actor);

/*
 List actions that are in progress or ended with an error.

 */
struct DefraResult defra_action_list(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Exports the documents of the comma-separated `collections` (every
 collection for null or empty) to the file `filepath`, as JSON (`format`
 "json", the default for null or empty), indented when `pretty` is set.

 # Safety

 Every string must be null or a valid null-terminated string.
 */
struct DefraResult defra_backup_export(uintptr_t node_ptr,
                                       const char *filepath,
                                       const char *collections,
                                       const char *format,
                                       int pretty);

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

 `filepath` must be null or a valid null-terminated string.
 */
struct DefraResult defra_backup_import(uintptr_t node_ptr, const char *filepath);

/*
 Starts (or restarts) collecting the blocks the identity `identity_ptr`
 names (0 for the node's) signs.
 */
struct DefraResult defra_batch_start(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Sign all collected batch CIDs and return the batch signature as JSON.

 Returns JSON with fields: sig_type, identity, value (hex), merkle_root (hex), cid_count.

 The identity is the one `defra_batch_start` was given.
 */
struct DefraResult defra_batch_sign(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Verifies that block `cid` was signed by `public_key`, of type `key_type`
 ("secp256k1", the default for null or empty, "secp256r1" or "ed25519"),
 on a node or in a transaction.

 # Safety

 Every string must be null or a valid null-terminated string.
 */
struct DefraResult defra_block_verify(uintptr_t node_ptr,
                                      const char *key_type,
                                      const char *public_key,
                                      const char *cid,
                                      uintptr_t identity_ptr);

/*
 Adds the collections an SDL document defines, on a node or in a
 transaction, and answers with the created collection versions as JSON.

 # Safety

 `sdl` must be null or a valid null-terminated UTF-8 string.
 */
struct DefraResult defra_collection_add(uintptr_t node_ptr,
                                        const char *sdl,
                                        uintptr_t identity_ptr);

/*
 Deletes the comma-separated collections `names`, on a node or in a
 transaction; with `active_only` set, only their active versions.

 # Safety

 `names` must be null or a valid null-terminated string.
 */
struct DefraResult defra_collection_delete(uintptr_t node_ptr,
                                           const char *names,
                                           int active_only,
                                           uintptr_t identity_ptr);

/*
 Answers with the collection versions `opts` selects as JSON, on a node or
 in a transaction.

 # Safety

 Every non-null string in `opts` must be a valid null-terminated string.
 */
struct DefraResult defra_collection_describe(uintptr_t node_ptr,
                                             struct CollectionOptions opts,
                                             uintptr_t identity_ptr);

/*
 Applies a JSON Patch (RFC 6902) to the collection definitions, with an
 optional migration lens (`{"Lenses": [...]}`, null or empty for none).
 The patch's first operation names the collection: `/Users/Fields/-`
 patches `Users`.

 # Safety

 `patch` and `lens_config` must be null or valid null-terminated strings.
 */
struct DefraResult defra_collection_patch(uintptr_t node_ptr,
                                          const char *patch,
                                          const char *lens_config,
                                          uintptr_t identity_ptr);

/*
 Makes the collection version `opts.version` names active, deactivating
 the other versions of its collection, on a node or in a transaction.

 # Safety

 Every non-null string in `opts` must be a valid null-terminated string.
 */
struct DefraResult defra_collection_set_active(uintptr_t node_ptr,
                                               struct CollectionOptions opts,
                                               uintptr_t identity_ptr);

/*
 Deletes every document in the collection `opts` selects, keeping the
 collection.

 # Safety

 Every non-null string in `opts` must be a valid null-terminated string.
 */
struct DefraResult defra_collection_truncate(uintptr_t node_ptr,
                                             struct CollectionOptions opts,
                                             uintptr_t identity_ptr);

/*
 Deletes the documents of the collection `opts` selects that match
 `filter_json`, a GraphQL filter as a JSON object, keeping the collection.

 # Safety

 `filter_json` must be null or a valid null-terminated string; every
 non-null string in `opts` must be one too.
 */
struct DefraResult defra_collection_truncate_with_filter(uintptr_t node_ptr,
                                                         struct CollectionOptions opts,
                                                         uintptr_t identity_ptr,
                                                         const char *filter_json);

/*
 Adds a JSON document, or a JSON array of documents, to the collection
 `opts` selects, and answers with the new document IDs as a JSON array.
 `is_encrypted` encrypts every field; otherwise `encrypted_fields`, a
 comma-separated list, names the fields to encrypt.

 # Safety

 `json_data` and `encrypted_fields` must be null or valid null-terminated
 strings; every non-null string in `opts` must be one too.
 */
struct DefraResult defra_document_add(uintptr_t node_ptr,
                                      const char *json_data,
                                      int is_encrypted,
                                      const char *encrypted_fields,
                                      struct CollectionOptions opts,
                                      uintptr_t identity_ptr);

/*
 Deletes the document `doc_id` of the collection `opts` selects, answering
 with no value; or, without a document ID, every document matching the
 JSON `filter`, answering with `{"Count": ..., "DocIDs": [...]}`.

 # Safety

 `doc_id` and `filter` must be null or valid null-terminated strings;
 every non-null string in `opts` must be one too.
 */
struct DefraResult defra_document_delete(uintptr_t node_ptr,
                                         const char *doc_id,
                                         const char *filter,
                                         struct CollectionOptions opts,
                                         uintptr_t identity_ptr);

/*
 Answers with the document `doc_id` of the collection `opts` selects as a
 JSON object of its fields and `_docID`; with `show_deleted` set, a
 deleted document too.

 # Safety

 `doc_id` must be null or a valid null-terminated string; every non-null
 string in `opts` must be one too.
 */
struct DefraResult defra_document_get(uintptr_t node_ptr,
                                      const char *doc_id,
                                      int show_deleted,
                                      struct CollectionOptions opts,
                                      uintptr_t identity_ptr);

/*
 Applies the JSON `updater` to the document `doc_id` of the collection
 `opts` selects, answering with no value; or, without a document ID, to
 every document matching the JSON `filter`, answering with
 `{"Count": ..., "DocIDs": [...]}`.

 # Safety

 `doc_id`, `filter` and `updater` must be null or valid null-terminated
 strings; every non-null string in `opts` must be one too.
 */
struct DefraResult defra_document_update(uintptr_t node_ptr,
                                         const char *doc_id,
                                         const char *filter,
                                         const char *updater,
                                         struct CollectionOptions opts,
                                         uintptr_t identity_ptr);

/*
 Logs every key of the node's store, with its value's size. Unwise on a
 large database.
 */
struct DefraResult defra_print_dump(uintptr_t node_ptr);

/*
 Creates an encrypted index of type `index_type` ("equality", the default
 for null or empty) on field `field_name` of collection `collection_name`,
 and answers with its description as JSON.

 # Safety

 Every string must be null or a valid null-terminated string.
 */
struct DefraResult defra_encrypted_index_new(uintptr_t node_ptr,
                                             const char *collection_name,
                                             const char *field_name,
                                             const char *index_type_name,
                                             uintptr_t identity_ptr);

/*
 Deletes the encrypted index on field `field_name` of collection
 `collection_name`.

 # Safety

 Every string must be null or a valid null-terminated string.
 */
struct DefraResult defra_encrypted_index_delete(uintptr_t node_ptr,
                                                const char *collection_name,
                                                const char *field_name,
                                                uintptr_t identity_ptr);

/*
 Answers with the encrypted indexes of collection `collection_name` as a
 JSON array, or, for null or empty, every collection's as a JSON object
 keyed by collection name.

 # Safety

 `collection_name` must be null or a valid null-terminated string.
 */
struct DefraResult defra_encrypted_index_list(uintptr_t node_ptr,
                                              const char *collection_name,
                                              uintptr_t identity_ptr);

/*
 Creates an identity with a new key of type `key_type`: "secp256k1" (the
 default, for null or empty), "secp256r1" or "ed25519".

 # Safety

 `key_type` must be null or a valid null-terminated string.
 */
struct DefraNewIdentityResult defra_identity_new(const char *key_type);

/*
 Imports an identity from a hex-encoded 32-byte secp256k1 private key as
 Go's `NewIdentityFromPrivateKey` does. Other key types import with
 [`defra_identity_new_from_typed_private_key`].

 # Safety

 `private_key_hex` must be null or a valid null-terminated string.
 */
struct DefraNewIdentityResult defra_identity_new_from_private_key(const char *private_key_hex);

/*
 Imports an identity of `key_type` (as [`defra_identity_new`] names it)
 from the hex-encoded private key [`defra_identity_export_private_key`]
 gave. Not in Go's C API, which imports secp256k1 keys only.

 # Safety

 `key_type` and `private_key_hex` must be null or valid null-terminated
 strings.
 */
struct DefraNewIdentityResult defra_identity_new_from_typed_private_key(const char *key_type,
                                                                        const char *private_key_hex);

/*
 Returns the identity's private key, hex-encoded.
 */
struct DefraResult defra_identity_export_private_key(uintptr_t identity_ptr);

/*
 Returns the identity's DID (`did:key:...`), the name ACP calls take for
 an actor. Not in Go's C API, which offers no way to read it.
 */
struct DefraResult defra_identity_did(uintptr_t identity_ptr);

/*
 Releases an identity handle and its hold on the identity's private key,
 which leaves memory once no other handle or node holds it. Releasing the
 null handle, or one already released, does nothing.
 */
void defra_identity_free(uintptr_t identity_ptr);

/*
 Returns the node's public identity as JSON, `{"PublicKey": ..., "DID":
 ...}`, for a node or a transaction on it: the one it started with, or the
 one `defra_node_set_identity` gave it since. A node without an identity
 answers with a message saying so.
 */
struct DefraResult defra_node_identity(uintptr_t node_ptr);

/*
 Creates the index `index_name` on the collection `opts.name` names, over
 `fields`: comma-separated field names, each optionally `:ASC` (the
 default) or `:DESC`. `vector_json`, a vector index description, makes it a
 vector index (null or empty for none). Answers with the index
 description as JSON.

 # Safety

 `index_name`, `fields` and `vector_json` must be null or valid
 null-terminated strings; every non-null string in `opts` must be one too.
 */
struct DefraResult defra_index_new(uintptr_t node_ptr,
                                   const char *index_name,
                                   const char *fields,
                                   int is_unique,
                                   const char *vector_json,
                                   struct CollectionOptions opts,
                                   uintptr_t identity_ptr);

/*
 Deletes the index `index_name` from the collection `opts.name` names.

 # Safety

 `index_name` must be null or a valid null-terminated string; every
 non-null string in `opts` must be one too.
 */
struct DefraResult defra_index_delete(uintptr_t node_ptr,
                                      const char *index_name,
                                      struct CollectionOptions opts,
                                      uintptr_t identity_ptr);

/*
 Answers with the indexes of the collection `opts.name` names as a JSON
 array, or, when it names none, with every collection's indexes as a JSON
 object keyed by collection name.

 # Safety

 Every non-null string in `opts` must be a valid null-terminated string.
 */
struct DefraResult defra_index_list(uintptr_t node_ptr,
                                    struct CollectionOptions opts,
                                    uintptr_t identity_ptr);

/*
 Adds a lens, on a node or in a transaction, and answers with its ID.

 # Safety

 `cfg` must be null or a valid null-terminated string.
 */
struct DefraResult defra_lens_add(uintptr_t node_ptr, uintptr_t identity_ptr, const char *cfg);

/*
 Answers with the node's lenses as JSON, on a node or in a transaction.
 */
struct DefraResult defra_lens_list(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Sets the lens `cfg` as the migration from collection version `src` to
 `dst`, on a node or in a transaction, and answers with the lens ID.

 # Safety

 `src`, `dst` and `cfg` must be null or valid null-terminated strings.
 */
struct DefraResult defra_lens_set(uintptr_t node_ptr,
                                  uintptr_t identity_ptr,
                                  const char *src,
                                  const char *dst,
                                  const char *cfg);

/*
 Creates and starts an embedded DefraDB node.

 # Safety

 Every non-null pointer in `options` must be valid for the call: strings
 null-terminated, byte buffers at least their stated length.
 */
struct DefraNewNodeResult defra_node_new(struct NodeInitOptions options);

/*
 Closes the node and releases its handle, with every transaction and
 subscription still open on it.
 */
struct DefraResult defra_node_close(uintptr_t node_ptr);

/*
 Returns the node's configuration options as JSON, with every value that
 could locate or unlock the node redacted.
 */
struct DefraResult defra_node_options(uintptr_t node_ptr);

/*
 Adds the comma-separated `collections` to the collections the node
 replicates.

 # Safety

 `collections` must be null or a valid null-terminated string.
 */
struct DefraResult defra_p2p_collection_add(uintptr_t node_ptr,
                                            const char *collections,
                                            uintptr_t identity_ptr);

/*
 Removes the comma-separated `collections` from the collections the node
 replicates.

 # Safety

 `collections` must be null or a valid null-terminated string.
 */
struct DefraResult defra_p2p_collection_delete(uintptr_t node_ptr,
                                               const char *collections,
                                               uintptr_t identity_ptr);

/*
 Answers with the collections the node replicates as a JSON array.
 */
struct DefraResult defra_p2p_collection_list(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Synchronizes the branchable collection `collection_id` from connected
 peers, within `timeout_str`, a Go duration (null or empty for none).

 # Safety

 `collection_id` and `timeout_str` must be null or valid null-terminated
 strings.
 */
struct DefraResult defra_p2p_collection_sync_branchable(uintptr_t node_ptr,
                                                        const char *collection_id,
                                                        const char *timeout_str,
                                                        uintptr_t identity_ptr);

/*
 Synchronizes the comma-separated collection versions `version_ids` from
 connected peers, within `timeout_str`, a Go duration (null or empty for
 none).

 # Safety

 `version_ids` and `timeout_str` must be null or valid null-terminated
 strings.
 */
struct DefraResult defra_p2p_collection_sync_versions(uintptr_t node_ptr,
                                                      const char *version_ids,
                                                      const char *timeout_str,
                                                      uintptr_t identity_ptr);

/*
 Adds the comma-separated document IDs `doc_ids` to the documents the
 node replicates.

 # Safety

 `doc_ids` must be null or a valid null-terminated string.
 */
struct DefraResult defra_p2p_document_add(uintptr_t node_ptr,
                                          const char *doc_ids,
                                          uintptr_t identity_ptr);

/*
 Removes the comma-separated document IDs `doc_ids` from the documents the
 node replicates.

 # Safety

 `doc_ids` must be null or a valid null-terminated string.
 */
struct DefraResult defra_p2p_document_delete(uintptr_t node_ptr,
                                             const char *doc_ids,
                                             uintptr_t identity_ptr);

/*
 Answers with the IDs of the documents the node replicates as a JSON
 array.
 */
struct DefraResult defra_p2p_document_list(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Synchronizes the comma-separated documents `doc_ids` of collection
 `collection` from connected peers, within `timeout_str`, a Go duration
 (null or empty for none).

 # Safety

 Every string must be null or a valid null-terminated string.
 */
struct DefraResult defra_p2p_document_sync(uintptr_t node_ptr,
                                           const char *collection,
                                           const char *doc_ids,
                                           const char *timeout_str,
                                           uintptr_t identity_ptr);

/*
 Answers with the best address to share with another peer so it can dial
 this node, as a JSON string, or JSON `null` while there is none yet (the
 node runs without P2P, or its transport has no dialable address).
 */
struct DefraResult defra_mobile_shareable_address(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Answers with the node's live P2P sync status as JSON: its outbound push
 backlog, the DAGs it is still fetching, and its retry and overload
 counters.
 */
struct DefraResult defra_mobile_sync_status(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Answers with everything the node knows about the peer `peer_id`, a raw
 transport peer ID, as JSON: whether it is connected, what was pushed to
 it, and its retry state.

 # Safety

 `peer_id` must be null or a valid null-terminated string.
 */
struct DefraResult defra_mobile_peer_details(uintptr_t node_ptr,
                                             const char *peer_id,
                                             uintptr_t identity_ptr);

/*
 Tells the node its network may have changed (Wi-Fi to cellular, a new
 address, waking up), so the iroh transport looks for new paths to its
 peers at once instead of waiting to notice. With libp2p it does nothing.
 */
struct DefraResult defra_mobile_notify_network_change(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Connects to the comma-separated peer multiaddrs `peer_addresses`.

 # Safety

 `peer_addresses` must be null or a valid null-terminated string.
 */
struct DefraResult defra_p2p_connect(uintptr_t node_ptr,
                                     const char *peer_addresses,
                                     uintptr_t identity_ptr);

/*
 Disconnects from the comma-separated peer multiaddrs `peer_addresses`.

 # Safety

 `peer_addresses` must be null or a valid null-terminated string.
 */
struct DefraResult defra_p2p_disconnect(uintptr_t node_ptr,
                                        const char *peer_addresses,
                                        uintptr_t identity_ptr);

/*
 Answers with the addresses other peers reach this node at, as a JSON
 array: each multiaddr with this node's peer ID appended. A node without
 P2P answers `null`, as Go's does, rather than an error.
 */
struct DefraResult defra_p2p_info(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Answers with the peers the node is connected to as a JSON array.
 */
struct DefraResult defra_p2p_peer_list(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Replicates the comma-separated `collections` to the peer the
 comma-separated multiaddrs `addresses` name.

 # Safety

 `collections` and `addresses` must be null or valid null-terminated
 strings.
 */
struct DefraResult defra_p2p_replicator_add(uintptr_t node_ptr,
                                            const char *collections,
                                            const char *addresses,
                                            uintptr_t identity_ptr);

/*
 Stops replicating the comma-separated `collections` (every collection for
 null or empty) to the replicator `id`.

 # Safety

 `collections` and `id` must be null or valid null-terminated strings.
 */
struct DefraResult defra_p2p_replicator_delete(uintptr_t node_ptr,
                                               const char *collections,
                                               const char *id,
                                               uintptr_t identity_ptr);

/*
 Answers with the node's replicators as a JSON array of Go's
 `client.Replicator`.
 */
struct DefraResult defra_p2p_replicator_list(uintptr_t node_ptr, uintptr_t identity_ptr);

struct DefraResult defra_profiling_start(const char *output_path);

struct DefraResult defra_profiling_stop(void);

/*
 Executes a GraphQL request on a node, or in a transaction, as the
 identity `identity_ptr` names (0 for none).

 A query or mutation answers with status 0 and its GraphQL result as JSON,
 `{"data": ..., "errors": [...]}`; GraphQL errors are part of that result.
 A subscription answers with status 2 and the subscription's ID, which
 `defra_subscription_poll` reads and `defra_subscription_close` ends.
 Status 1 reports a request that could not run.

 # Safety

 Every string must be null or a valid null-terminated UTF-8 string;
 `operation_name` and `variables` may be null or empty for none.
 */
struct DefraResult defra_query_execute(uintptr_t node_ptr,
                                       const char *query,
                                       uintptr_t identity_ptr,
                                       const char *operation_name,
                                       const char *variables);

/*
 [`defra_query_execute`] for a caller that is called back instead of
 waiting: returns once the request is admitted to the node's worker pool,
 and `callback` receives `context` and the result exactly once, as
 [`defra_query_execute`] would return it, on the DefraDB thread that ran
 the request. A request refused before it runs (an invalid argument, a
 denied permission) and a subscription, which answers at once with its ID,
 call it on the caller's thread before this returns.

 # Safety

 As for [`defra_query_execute`]; the strings are read only before this
 returns.
 */
void defra_query_execute_async(uintptr_t node_ptr,
                               const char *query,
                               uintptr_t identity_ptr,
                               const char *operation_name,
                               const char *variables,
                               DefraResultCallback callback,
                               uintptr_t context);

/*
 Report the result of a signature that a `DefraRemoteSignAsyncCallback`
 started.

 Call it exactly once per request, from any thread, at any time after the
 callback was called (also from inside it).

 - `completion` is the value the callback received.
 - `status` is `0` when the payload was signed, `DEFRA_SIGN_RETRYABLE`
   when the request did not reach the signer or its answer did not come
   back (DefraDB may ask again), and any other value when the signer
   refused.
 - `signature_ptr` / `signature_len` are the signature when `status` is
   `0`; the bytes are copied before this returns.

 Returns `0` when the result was delivered, and `1` when no request waits
 under `completion`: it was reported already, or the write stopped waiting
 for it (its deadline passed).
 */
int defra_remote_sign_complete(uintptr_t completion,
                               int status,
                               const uint8_t *signature_ptr,
                               uintptr_t signature_len);

/*
 Register an identity whose private key stays outside DefraDB.

 The host provides the DID, public key, key type, and a signing callback.
 This supports device-bound keys (for example Secure Enclave P-256 keys) and
 remote identity providers such as Orbis without exporting private key bytes.

 Swift can import this symbol directly from the generated Apple package.
 The callback is required and uses a plain C function-pointer ABI.
 */
struct DefraNewIdentityResult defra_identity_new_remote(const char *did,
                                                        const char *public_key_hex,
                                                        const char *key_type,
                                                        uintptr_t signer_handle,
                                                        DefraRemoteSignCallback callback);

/*
 Register an identity whose public key is already available as raw bytes.

 This is intended for Swift and Objective-C callers that already have the
 device public key as `Data` or `UnsafePointer<UInt8>`. The callback contract
 matches [`defra_identity_new_remote`].

 Public key format:
 - `secp256r1` / `secp256k1`: uncompressed SEC1 / X9.63 bytes
 - `ed25519`: raw 32-byte public key
 */
struct DefraNewIdentityResult defra_identity_new_remote_bytes(const char *did,
                                                              const uint8_t *public_key_ptr,
                                                              uintptr_t public_key_len,
                                                              const char *key_type,
                                                              uintptr_t signer_handle,
                                                              DefraRemoteSignCallback callback);

/*
 [`defra_identity_new_remote`] for a host whose signer answers later: the
 callback starts each signature and returns at once, and the host reports
 the result with `defra_remote_sign_complete`. No DefraDB thread waits
 while the host signs, so this is the form for a host that signs on its
 own queue or over the network, and for a node running a single worker.
 */
struct DefraNewIdentityResult defra_identity_new_remote_async(const char *did,
                                                              const char *public_key_hex,
                                                              const char *key_type,
                                                              uintptr_t signer_handle,
                                                              DefraRemoteSignAsyncCallback callback);

/*
 [`defra_identity_new_remote_bytes`] with an asynchronous callback, as
 [`defra_identity_new_remote_async`].
 */
struct DefraNewIdentityResult defra_identity_new_remote_bytes_async(const char *did,
                                                                    const uint8_t *public_key_ptr,
                                                                    uintptr_t public_key_len,
                                                                    const char *key_type,
                                                                    uintptr_t signer_handle,
                                                                    DefraRemoteSignAsyncCallback callback);

/*
 Returns the next pending result of subscription `id`: status 0 with the
 GraphQL result as JSON, or status 2 with no value when none is pending.
 A subscription that has ended reports status 1 and is closed.

 # Safety

 `id` must be null or a valid null-terminated string.
 */
struct DefraResult defra_subscription_poll(const char *id);

/*
 Cancels subscription `id`. Closing an ID no subscription has does
 nothing.

 # Safety

 `id` must be null or a valid null-terminated string.
 */
struct DefraResult defra_subscription_close(const char *id);

/*
 Subscribes to node `node_ptr`'s events named in `names`, comma-separated
 (`"update,merge-complete"`; `"*"` names them all), as the identity
 `identity_ptr` names, and answers with status 2 and the subscription's
 ID.

 Each `defra_subscription_poll` of the ID yields the next event as a JSON
 object whose `name` is the event's name, with that event's fields
 beside it. Events carry documents without document access control, so
 when node access control is on, only an identity allowed to bypass
 document access control may subscribe.

 # Safety

 `names` must be null or a valid null-terminated string.
 */
struct DefraResult defra_event_subscribe(uintptr_t node_ptr,
                                         const char *names,
                                         uintptr_t identity_ptr);

/*
 Creates a database transaction on node `node_ptr`.

 The returned handle stands in for the node in any call that takes a node
 or a transaction, which then runs in the transaction. It is released by
 [`defra_transaction_commit`] or [`defra_transaction_discard`].
 */
struct DefraNewTxnResult defra_transaction_create(uintptr_t node_ptr, int is_read_only);

/*
 Commits the transaction and releases its handle, whether or not the
 commit succeeds.
 */
struct DefraResult defra_transaction_commit(uintptr_t txn_ptr);

/*
 Discards the transaction and releases its handle. Discarding a handle
 that is already released does nothing.
 */
void defra_transaction_discard(uintptr_t txn_ptr);

/*
 Frees the strings of a result and nulls them, so freeing it again does
 nothing. A null `result` does nothing.

 # Safety

 `result` must be null or point to a result a call returned.
 */
void defra_result_free(struct DefraResult *result);

/*
 Frees the error of a node creation result and nulls it, so freeing it
 again does nothing. The node stays open. A null `result` does nothing.

 # Safety

 `result` must be null or point to a result a call returned.
 */
void defra_new_node_result_free(struct DefraNewNodeResult *result);

/*
 Frees the error of a transaction creation result and nulls it, so
 freeing it again does nothing. The transaction stays open. A null
 `result` does nothing.

 # Safety

 `result` must be null or point to a result a call returned.
 */
void defra_new_txn_result_free(struct DefraNewTxnResult *result);

/*
 Frees the error of an identity creation result and nulls it, so freeing
 it again does nothing. The identity stays registered. A null `result`
 does nothing.

 # Safety

 `result` must be null or point to a result a call returned.
 */
void defra_new_identity_result_free(struct DefraNewIdentityResult *result);

/*
 Returns DefraDB version information: as JSON when `flag_json` is set,
 otherwise as text, with full build information when `flag_full` is set.
 The same output as `defradb version`.
 */
struct DefraResult defra_version_get(int flag_full, int flag_json);

/*
 Adds the views an SDL document defines over the GraphQL `query`, with an
 optional lens `transform_cid` (null or empty for none), and answers with
 the created collection versions as JSON. A materialized view is
 refreshed once created.

 # Safety

 `query`, `sdl` and `transform_cid` must be null or valid null-terminated
 strings.
 */
struct DefraResult defra_view_add(uintptr_t node_ptr,
                                  const char *query,
                                  const char *sdl,
                                  const char *transform_cid,
                                  uintptr_t identity_ptr);

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
 Refreshes the materialized views `opts` selects, every one for empty
 options.

 # Safety

 Every non-null string in `opts` must be a valid null-terminated string.
 */
struct DefraResult defra_view_refresh(uintptr_t node_ptr,
                                      struct CollectionOptions opts,
                                      uintptr_t identity_ptr);

#ifdef __cplusplus
}  // extern "C"
#endif  // __cplusplus

#endif  /* DEFRA_H */
