/*
 * DefraDB C API - generated header.
 *
 * The functions cover the calls of Go DefraDB's cbindings (libdefradb.h),
 * named defra_<noun>_<verb>. Names and types are prefixed and
 * DefraNodeInitOptions adds fields after Go's, so this header is not a
 * drop-in for libdefradb.h. A call that works on a store takes a handle
 * naming a node or one of its transactions. Each call that can grow takes
 * its options as a struct, by value, whose zero value is every default; the
 * caller's identity (0 for none) is one of its fields. Fields are only ever
 * appended, so code built against an older header keeps working. Free every
 * result a call returns with its free function: defra_result_free, or
 * defra_new_<kind>_result_free.
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

 Both strings are always set, as Go's `returnC` sets them: the one a result
 does not use is empty, never null, so a host may read both. The caller
 owns them and frees them with [`defra_result_free`].
 */
typedef struct DefraResult {
  /*
   Status code: 0=success, 1=error, 2=subscription
   */
  int status;
  /*
   Error message; empty, never null, on success. Freed with the result.
   */
  char *error;
  /*
   JSON value; empty, never null, on error or when there is none. Freed
   with the result.
   */
  char *value;
} DefraResult;

/*
 Options of `defra_acp_dac_policy_add`.
 */
typedef struct DefraDacPolicyAddOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraDacPolicyAddOptions;

/*
 Options of `defra_acp_dac_relationship_add` and
 `defra_acp_dac_relationship_delete`.
 */
typedef struct DefraDacRelationshipOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraDacRelationshipOptions;

/*
 Options of the node access control calls status, disable, re-enable and
 purge.
 */
typedef struct DefraNacOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraNacOptions;

/*
 Options of `defra_acp_nac_relationship_add` and
 `defra_acp_nac_relationship_delete`.
 */
typedef struct DefraNacRelationshipOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraNacRelationshipOptions;

/*
 Options of `defra_action_list`.
 */
typedef struct DefraActionListOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraActionListOptions;

/*
 Options of `defra_batch_start` and `defra_batch_sign`.
 */
typedef struct DefraBatchOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraBatchOptions;

/*
 Options of `defra_block_verify`.
 */
typedef struct DefraBlockVerifyOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraBlockVerifyOptions;

/*
 Options of `defra_collection_add`.
 */
typedef struct DefraCollectionAddOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraCollectionAddOptions;

/*
 Options of `defra_collection_delete`.
 */
typedef struct DefraCollectionDeleteOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Delete only the active versions (1=true, 0=false).
   */
  int active_only;
} DefraCollectionDeleteOptions;

/*
 Options of `defra_collection_describe`. With no collection chosen,
 every active collection is described.
 */
typedef struct DefraCollectionDescribeOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
} DefraCollectionDescribeOptions;

/*
 Options of `defra_collection_patch`.
 */
typedef struct DefraCollectionPatchOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   The migration for the versions the patch adds, a lens configuration
   (null or empty for none).
   */
  const char *lens_config;
} DefraCollectionPatchOptions;

/*
 Options of `defra_collection_set_active`.
 */
typedef struct DefraCollectionSetActiveOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraCollectionSetActiveOptions;

/*
 Options of `defra_collection_truncate`.
 */
typedef struct DefraCollectionTruncateOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
  /*
   A GraphQL filter as a JSON object; null removes every document.
   */
  const char *filter_json;
} DefraCollectionTruncateOptions;

/*
 Options of `defra_document_add` and `defra_document_save_many`.
 */
typedef struct DefraDocumentAddOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
  /*
   Encrypt every field (1=true, 0=false).
   */
  int encrypt;
  /*
   Comma-separated fields to encrypt (null or empty for none).
   */
  const char *encrypted_fields;
  /*
   Block signing: 0 node default, 1 sign, -1 do not sign.
   */
  int signing;
} DefraDocumentAddOptions;

/*
 Options of `defra_document_get`.
 */
typedef struct DefraDocumentGetOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
  /*
   Answer a deleted document too (1=true, 0=false).
   */
  int show_deleted;
} DefraDocumentGetOptions;

/*
 Options of `defra_document_update`: the documents are `doc_id`, or those
 `filter_json` matches.
 */
typedef struct DefraDocumentUpdateOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
  /*
   The document to update.
   */
  const char *doc_id;
  /*
   A GraphQL filter choosing the documents to update, as JSON.
   */
  const char *filter_json;
  /*
   Block signing: 0 node default, 1 sign, -1 do not sign.
   */
  int signing;
} DefraDocumentUpdateOptions;

/*
 Options of `defra_document_delete`: the documents are `doc_id`, or those
 `filter_json` matches.
 */
typedef struct DefraDocumentDeleteOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
  /*
   The document to delete.
   */
  const char *doc_id;
  /*
   A GraphQL filter choosing the documents to delete, as JSON.
   */
  const char *filter_json;
  /*
   Block signing: 0 node default, 1 sign, -1 do not sign.
   */
  int signing;
} DefraDocumentDeleteOptions;

/*
 Options of `defra_document_purge`.
 */
typedef struct DefraDocumentPurgeOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
} DefraDocumentPurgeOptions;

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
   Error message; empty, never null, on success. Freed with the result.
   */
  char *error;
  /*
   Handle to the identity (0 on error).
   */
  uintptr_t identity_ptr;
} DefraNewIdentityResult;

/*
 Options of `defra_identity_new` and `defra_identity_new_from_private_key`.
 */
typedef struct DefraIdentityNewOptions {
  /*
   "secp256k1" (the default, for null or empty), "secp256r1" or "ed25519".
   */
  const char *key_type;
} DefraIdentityNewOptions;

/*
 Options of `defra_identity_new_from_bearer_token`. Nothing is optional
 yet; the struct is where options will go.
 */
typedef struct DefraIdentityBearerOptions {
  /*
   Reserved: leave zero.
   */
  int32_t reserved;
} DefraIdentityBearerOptions;

/*
 Options of `defra_identity_new_remote`: the public key as hex or as
 bytes, and exactly one of the two signing callbacks.
 */
typedef struct DefraRemoteIdentityOptions {
  /*
   The key's type (null or empty for "secp256r1").
   */
  const char *key_type;
  /*
   The public key, hex-encoded.
   */
  const char *public_key_hex;
  /*
   The public key's bytes, used when `public_key_hex` is null.
   */
  const uint8_t *public_key;
  /*
   Length of `public_key` in bytes.
   */
  uintptr_t public_key_len;
  /*
   Passed back to the callback on every signing request.
   */
  uintptr_t signer_handle;
  /*
   A signer that answers before it returns: a `DefraRemoteSignCallback`,
   spelled out so C sees a nullable function pointer.
   */
  int (*sign)(uintptr_t signer_handle,
              const char *did,
              const char *key_type,
              const uint8_t *payload_ptr,
              uintptr_t payload_len,
              uint8_t *out_signature_ptr,
              uintptr_t out_signature_capacity,
              uintptr_t *out_signature_len);
  /*
   A signer that answers later, through `defra_remote_sign_complete`: a
   `DefraRemoteSignAsyncCallback`, spelled out as `sign` is.
   */
  int (*sign_async)(uintptr_t signer_handle,
                    const char *did,
                    const char *key_type,
                    const uint8_t *payload_ptr,
                    uintptr_t payload_len,
                    uintptr_t completion);
} DefraRemoteIdentityOptions;

/*
 Options of `defra_index_new`.
 */
typedef struct DefraIndexNewOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
  /*
   The index name (null or empty to generate one).
   */
  const char *name;
  /*
   A unique index (1=true, 0=false).
   */
  int unique;
  /*
   A vector index description as JSON (null or empty for an ordered index).
   */
  const char *vector_json;
} DefraIndexNewOptions;

/*
 Options of `defra_index_delete`.
 */
typedef struct DefraIndexDeleteOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
} DefraIndexDeleteOptions;

/*
 Options of `defra_index_list`. With no collection chosen, every
 collection's indexes are listed.
 */
typedef struct DefraIndexListOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
} DefraIndexListOptions;

/*
 Options of `defra_encrypted_index_new`. Encrypted indexes choose their
 collection by name.
 */
typedef struct DefraEncryptedIndexNewOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
  /*
   The index type (null or empty for "equality").
   */
  const char *index_type;
} DefraEncryptedIndexNewOptions;

/*
 Options of `defra_encrypted_index_delete`.
 */
typedef struct DefraEncryptedIndexDeleteOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
} DefraEncryptedIndexDeleteOptions;

/*
 Options of `defra_encrypted_index_list`. With no collection named, every
 collection's encrypted indexes are listed.
 */
typedef struct DefraEncryptedIndexListOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
} DefraEncryptedIndexListOptions;

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
   Error message; empty, never null, on success. Freed with the result.
   */
  char *error;
  /*
   Handle to the node (0 on error).
   */
  uintptr_t node_ptr;
} DefraNewNodeResult;

/*
 Options for node initialization.

 Go's NodeInitOptions struct, field for field except `disable_api`: this
 library serves no HTTP API yet, so it has no API to disable. An all-zero
 struct (null strings, zero numbers) is the default node, and every opt-out
 is a `disable_` flag. The fields after Go's are this implementation's own,
 so the layout differs from Go's: a caller passes this header's struct,
 never one laid out by Go's `defra_structs.h`.

 Go's Badger options name the store regolith replaced, so they are named
 for the store here (`store_file_size`, `store_encryption_key`). A Go
 option this implementation does not support yet fails `defra_node_new`
 when it is set, rather than being ignored.
 */
typedef struct DefraNodeInitOptions {
  /*
   Path to the store directory. Null or empty keeps the store in memory.
   */
  const char *db_path;
  /*
   Comma-separated P2P listen multiaddrs. Empty listens on
   `/ip4/0.0.0.0/tcp/9171`, as Go does; `disable_p2p_listen` listens on
   nothing.
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
   Run without signing blocks (1=true, 0=false). Signing is on by
   default, as in Go: with the node identity, or with a generated
   secp256k1 key when there is none.
   */
  int disable_signing;
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
   The Ed25519 private key the node's peer identity uses for this start
   (its 32-byte seed, or seed and public key), or null for the node's
   stored key. As Go's `P2PPrivateKey`, it applies to this start only: it
   is never checked against or recorded in the store, so a later start
   without it, or with another key, works and uses that start's key.
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
  /*
   Optional replication filter, asked whether to keep each replicated
   document before its commit stores anything; a
   `DefraReplicationFilterCallback`, spelled out so C sees a nullable
   function pointer. Null keeps everything.
   */
  int (*replication_filter)(uintptr_t context,
                            const char *collection_id,
                            const char *collection_name,
                            const char *doc_id,
                            const char *fields_json);
  /*
   Passed back to `replication_filter` on every call.
   */
  uintptr_t replication_filter_context;
  /*
   libp2p resource manager memory budget in MiB; zero keeps the default.
   */
  uint64_t p2p_resource_memory_mib;
  /*
   libp2p resource manager file-descriptor budget, used only together
   with `p2p_resource_memory_mib`; zero keeps the default.
   */
  uint64_t p2p_resource_file_descriptors;
  /*
   Streams one peer may hold open; zero keeps the default.
   */
  uintptr_t p2p_max_streams_per_peer;
  /*
   Run libp2p without listening (1=true, 0=false): a dial-only host that
   connects out to peers and replicators but accepts no connections, as a
   mobile embedder runs it. Several such nodes run in one process without
   competing for a port. Refused together with `listening_addresses`.
   */
  int disable_p2p_listen;
} DefraNodeInitOptions;

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
   Error message; empty, never null, on success. Freed with the result.
   */
  char *error;
  /*
   Handle to the transaction (0 on error).
   */
  uintptr_t txn_ptr;
} DefraNewTxnResult;

/*
 Options of `defra_transaction_create`.
 */
typedef struct DefraTransactionOptions {
  /*
   A transaction that only reads (1=true, 0=false).
   */
  int read_only;
} DefraTransactionOptions;

/*
 Options of `defra_version_get`.
 */
typedef struct DefraVersionOptions {
  /*
   Include build details (1=true, 0=false).
   */
  int full;
  /*
   Answer as JSON (1=true, 0=false).
   */
  int json;
} DefraVersionOptions;

/*
 Options of every P2P and mobile call that needs nothing but the caller.
 */
typedef struct DefraP2POptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraP2POptions;

/*
 Options of the three sync calls.
 */
typedef struct DefraP2PSyncOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   How long the sync may take, in milliseconds (0 for no limit).
   */
  int64_t timeout_ms;
} DefraP2PSyncOptions;

/*
 Options of `defra_p2p_replicator_add`.
 */
typedef struct DefraReplicatorAddOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Replication filters as JSON (null or empty to replicate every
   document).
   */
  const char *filters_json;
} DefraReplicatorAddOptions;

/*
 Options of `defra_p2p_replicator_delete`.
 */
typedef struct DefraReplicatorDeleteOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Comma-separated collections to stop replicating (null or empty for
   every one).
   */
  const char *collections;
} DefraReplicatorDeleteOptions;

/*
 Options of `defra_query_execute` and `defra_query_execute_async`.
 */
typedef struct DefraQueryOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   The operation to run, for a request with several (null or empty for
   the only one).
   */
  const char *operation_name;
  /*
   The request's variables as a JSON object (null or empty for none).
   */
  const char *variables;
} DefraQueryOptions;

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
 Options of `defra_event_subscribe`.
 */
typedef struct DefraEventSubscribeOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraEventSubscribeOptions;

/*
 Options of `defra_view_add`.
 */
typedef struct DefraViewAddOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   The CID of a lens to apply to the view's rows (null or empty for none).
   */
  const char *transform_cid;
} DefraViewAddOptions;

/*
 Options of `defra_view_refresh`. With no collection chosen, every view is
 refreshed.
 */
typedef struct DefraViewRefreshOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Collection name.
   */
  const char *collection_name;
  /*
   Collection version ID.
   */
  const char *collection_version_id;
  /*
   Collection ID.
   */
  const char *collection_id;
  /*
   Include inactive collection versions (1=true, 0=false).
   */
  int collection_get_inactive;
} DefraViewRefreshOptions;

/*
 Options of the three lens calls.
 */
typedef struct DefraLensOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraLensOptions;

/*
 Options of `defra_backup_export`.
 */
typedef struct DefraBackupExportOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
  /*
   Comma-separated collections to export (null or empty for every one).
   */
  const char *collections;
  /*
   The format (null or empty for "json", the only one).
   */
  const char *format;
  /*
   Indent the JSON (1=true, 0=false).
   */
  int pretty;
} DefraBackupExportOptions;

/*
 Options of `defra_backup_import`.
 */
typedef struct DefraBackupImportOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraBackupImportOptions;

/*
 Options of `defra_print_dump`.
 */
typedef struct DefraPrintDumpOptions {
  /*
   The caller's identity (0 for none).
   */
  uintptr_t identity;
} DefraPrintDumpOptions;

/*
 Host callback that starts a signature and returns without waiting for it.

 Contract:
 - `signer_handle`, `did` and `key_type` are as for
   [`DefraRemoteSignCallback`], and so is how long `signer_handle` and the
   callback must stay valid: until every node using the DID is closed or
   given another identity, not just until the handle is freed.
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
 Host callback used for non-exportable signing keys.

 The host app keeps the private key and signs the payload on demand. This is
 intended for device-bound keys such as Secure Enclave-backed P-256 keys.

 Contract:
 - `signer_handle` is an opaque host-managed context value. It must stay
   valid, and the callback callable, until every node using the DID is
   closed or given another identity, not just until the identity's handle
   is freed: a node keeps its hold on the identity, and signs through the
   callback, after `defra_identity_free`. Freeing the context together with
   the handle leaves such a node calling into freed memory.
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
 Whether to keep a replicated document: called with
 `replication_filter_context`, the collection ID, the collection name (the
 key a Go filter matches on; the ID stays the same across a rename), the
 DocID and the values the document's commit writes as a JSON object that
 keeps each value's kind: an integer as a JSON integer (any `i64` or
 `u64`), a float always with a fraction or exponent (`3.0`, never `3`), bytes
 as base64. An encrypted field, one that has not arrived, or one JSON cannot
 carry, such as a NaN, is absent. Return non-zero to keep it, 0 to refuse it.

 It is called on the node's database workers and on its network threads,
 from several threads at once, so it must be thread-safe, and it must not
 block or call back into DefraDB. One document may be asked about more
 than once (when it arrives, when it merges, and on a merge retry), so it
 must decide from its arguments alone. An ID or name holding a NUL, which
 no valid collection or DocID does, is refused without calling it.
 */
typedef int (*DefraReplicationFilterCallback)(uintptr_t context,
                                              const char *collection_id,
                                              const char *collection_name,
                                              const char *doc_id,
                                              const char *fields_json);

#ifdef __cplusplus
extern "C" {
#endif // __cplusplus

/*
 Sets the node's identity, which signs the node's blocks when signing is
 enabled, to the one `identity_ptr` names; 0 clears it.
 */
struct DefraResult defra_node_set_identity(uintptr_t node_ptr, uintptr_t identity_ptr);

/*
 Adds a document access control policy, YAML or JSON.

 # Safety

 Every string must be null or a valid null-terminated string; `opts` must
 be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_acp_dac_policy_add_sized(uintptr_t target,
                                                  const char *policy,
                                                  const struct DefraDacPolicyAddOptions *opts,
                                                  uint32_t opts_size);

/*
 Gives `actor` `relation` on document `doc_id` of `collection`.

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_acp_dac_relationship_add_sized(uintptr_t target,
                                                        const char *collection,
                                                        const char *doc_id,
                                                        const char *relation,
                                                        const char *actor,
                                                        const struct DefraDacRelationshipOptions *opts,
                                                        uint32_t opts_size);

/*
 Takes `relation` on document `doc_id` of `collection` from `actor`.

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_acp_dac_relationship_delete_sized(uintptr_t target,
                                                           const char *collection,
                                                           const char *doc_id,
                                                           const char *relation,
                                                           const char *actor,
                                                           const struct DefraDacRelationshipOptions *opts,
                                                           uint32_t opts_size);

/*
 Answers node access control's status.

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_acp_nac_status_sized(uintptr_t target,
                                              const struct DefraNacOptions *opts,
                                              uint32_t opts_size);

/*
 Turns node access control off for now.

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_acp_nac_disable_sized(uintptr_t target,
                                               const struct DefraNacOptions *opts,
                                               uint32_t opts_size);

/*
 Turns node access control back on.

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_acp_nac_re_enable_sized(uintptr_t target,
                                                 const struct DefraNacOptions *opts,
                                                 uint32_t opts_size);

/*
 Removes every node access control relationship (development mode only).

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_acp_nac_purge_sized(uintptr_t target,
                                             const struct DefraNacOptions *opts,
                                             uint32_t opts_size);

/*
 Gives `actor` (a DID, or `*` for everyone) `relation` on the node.

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_acp_nac_relationship_add_sized(uintptr_t target,
                                                        const char *relation,
                                                        const char *actor,
                                                        const struct DefraNacRelationshipOptions *opts,
                                                        uint32_t opts_size);

/*
 Takes `relation` on the node from `actor`.

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_acp_nac_relationship_delete_sized(uintptr_t target,
                                                           const char *relation,
                                                           const char *actor,
                                                           const struct DefraNacRelationshipOptions *opts,
                                                           uint32_t opts_size);

/*
 Lists the actions running or ended with an error.

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_action_list_sized(uintptr_t target,
                                           const struct DefraActionListOptions *opts,
                                           uint32_t opts_size);

/*
 Starts collecting the blocks the caller's writes create, to sign them
 together.

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_batch_start_sized(uintptr_t target,
                                           const struct DefraBatchOptions *opts,
                                           uint32_t opts_size);

/*
 Signs the blocks collected since `defra_batch_start`.

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_batch_sign_sized(uintptr_t target,
                                          const struct DefraBatchOptions *opts,
                                          uint32_t opts_size);

/*
 Checks that block `cid` was signed by `public_key`, of `key_type`.

 # Safety

 As [`defra_acp_dac_policy_add_sized`].
 */
struct DefraResult defra_block_verify_sized(uintptr_t target,
                                            const char *cid,
                                            const char *key_type,
                                            const char *public_key,
                                            const struct DefraBlockVerifyOptions *opts,
                                            uint32_t opts_size);

/*
 Adds the collections an SDL document defines. `defra_collection_add` in
 `defra.h` passes the size.

 # Safety

 `sdl` must be null or a valid null-terminated string; `opts` must be null
 or point to `opts_size` readable bytes whose strings are valid.
 */
struct DefraResult defra_collection_add_sized(uintptr_t target,
                                              const char *sdl,
                                              const struct DefraCollectionAddOptions *opts,
                                              uint32_t opts_size);

/*
 Deletes the collections named in `names`, comma-separated.

 # Safety

 As [`defra_collection_add_sized`].
 */
struct DefraResult defra_collection_delete_sized(uintptr_t target,
                                                 const char *names,
                                                 const struct DefraCollectionDeleteOptions *opts,
                                                 uint32_t opts_size);

/*
 Describes the collection versions the options choose.

 # Safety

 As [`defra_collection_add_sized`].
 */
struct DefraResult defra_collection_describe_sized(uintptr_t target,
                                                   const struct DefraCollectionDescribeOptions *opts,
                                                   uint32_t opts_size);

/*
 Applies a JSON Patch to the collection definitions.

 # Safety

 As [`defra_collection_add_sized`].
 */
struct DefraResult defra_collection_patch_sized(uintptr_t target,
                                                const char *patch,
                                                const struct DefraCollectionPatchOptions *opts,
                                                uint32_t opts_size);

/*
 Makes collection version `version_id` its collection's active one.

 # Safety

 As [`defra_collection_add_sized`].
 */
struct DefraResult defra_collection_set_active_sized(uintptr_t target,
                                                     const char *version_id,
                                                     const struct DefraCollectionSetActiveOptions *opts,
                                                     uint32_t opts_size);

/*
 Deletes the documents of the collection the options choose: every one,
 or those `filter_json` matches.

 # Safety

 As [`defra_collection_add_sized`].
 */
struct DefraResult defra_collection_truncate_sized(uintptr_t target,
                                                   const struct DefraCollectionTruncateOptions *opts,
                                                   uint32_t opts_size);

/*
 Adds the document, or the JSON array of documents, in `json`.

 # Safety

 `json` must be null or a valid null-terminated string; `opts` must be
 null or point to `opts_size` readable bytes whose strings are valid.
 */
struct DefraResult defra_document_add_sized(uintptr_t target,
                                            const char *json,
                                            const struct DefraDocumentAddOptions *opts,
                                            uint32_t opts_size);

/*
 Saves each document in the JSON array `json`: added if new, updated if
 it exists.

 # Safety

 As [`defra_document_add_sized`].
 */
struct DefraResult defra_document_save_many_sized(uintptr_t target,
                                                  const char *json,
                                                  const struct DefraDocumentAddOptions *opts,
                                                  uint32_t opts_size);

/*
 Answers document `doc_id`.

 # Safety

 As [`defra_document_add_sized`].
 */
struct DefraResult defra_document_get_sized(uintptr_t target,
                                            const char *doc_id,
                                            const struct DefraDocumentGetOptions *opts,
                                            uint32_t opts_size);

/*
 Applies the JSON `updater` to the documents the options choose.

 # Safety

 As [`defra_document_add_sized`].
 */
struct DefraResult defra_document_update_sized(uintptr_t target,
                                               const char *updater,
                                               const struct DefraDocumentUpdateOptions *opts,
                                               uint32_t opts_size);

/*
 Deletes the documents the options choose.

 # Safety

 As [`defra_document_add_sized`].
 */
struct DefraResult defra_document_delete_sized(uintptr_t target,
                                               const struct DefraDocumentDeleteOptions *opts,
                                               uint32_t opts_size);

/*
 Removes the documents in `doc_ids`, comma-separated, and their history
 from this node.

 # Safety

 As [`defra_document_add_sized`].
 */
struct DefraResult defra_document_purge_sized(uintptr_t target,
                                              const char *doc_ids,
                                              const struct DefraDocumentPurgeOptions *opts,
                                              uint32_t opts_size);

/*
 Creates an identity with a new key.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes whose strings
 are valid.
 */
struct DefraNewIdentityResult defra_identity_new_sized(const struct DefraIdentityNewOptions *opts,
                                                       uint32_t opts_size);

/*
 Imports an identity from the hex-encoded private key
 `defra_identity_export_private_key` gave.

 # Safety

 `private_key_hex` must be null or a valid null-terminated string; `opts`
 as [`defra_identity_new_sized`].
 */
struct DefraNewIdentityResult defra_identity_new_from_private_key_sized(const char *private_key_hex,
                                                                        const struct DefraIdentityNewOptions *opts,
                                                                        uint32_t opts_size);

/*
 An identity for `did` that signs with `bearer_token`.

 # Safety

 `did` and `bearer_token` must be null or valid null-terminated strings;
 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraNewIdentityResult defra_identity_new_from_bearer_token_sized(const char *did,
                                                                         const char *bearer_token,
                                                                         const struct DefraIdentityBearerOptions *opts,
                                                                         uint32_t opts_size);

/*
 Registers identity `did`, whose private key the host keeps and signs
 with through the callback the options give.

 # Safety

 `did` must be null or a valid null-terminated string; `opts` must be null
 or point to `opts_size` readable bytes, whose strings are valid and whose
 `public_key` is null or `public_key_len` readable bytes.
 */
struct DefraNewIdentityResult defra_identity_new_remote_sized(const char *did,
                                                              const struct DefraRemoteIdentityOptions *opts,
                                                              uint32_t opts_size);

/*
 Creates an index over `fields`: comma-separated field names, each
 optionally `:ASC` (the default) or `:DESC`.

 # Safety

 `fields` must be null or a valid null-terminated string; `opts` must be
 null or point to `opts_size` readable bytes whose strings are valid.
 */
struct DefraResult defra_index_new_sized(uintptr_t target,
                                         const char *fields,
                                         const struct DefraIndexNewOptions *opts,
                                         uint32_t opts_size);

/*
 Deletes index `name`.

 # Safety

 As [`defra_index_new_sized`].
 */
struct DefraResult defra_index_delete_sized(uintptr_t target,
                                            const char *name,
                                            const struct DefraIndexDeleteOptions *opts,
                                            uint32_t opts_size);

/*
 Lists the indexes of the collection the options choose, or of every
 collection.

 # Safety

 As [`defra_index_new_sized`].
 */
struct DefraResult defra_index_list_sized(uintptr_t target,
                                          const struct DefraIndexListOptions *opts,
                                          uint32_t opts_size);

/*
 Creates an encrypted index on `field`.

 # Safety

 As [`defra_index_new_sized`].
 */
struct DefraResult defra_encrypted_index_new_sized(uintptr_t target,
                                                   const char *field,
                                                   const struct DefraEncryptedIndexNewOptions *opts,
                                                   uint32_t opts_size);

/*
 Deletes the encrypted index on `field`.

 # Safety

 As [`defra_index_new_sized`].
 */
struct DefraResult defra_encrypted_index_delete_sized(uintptr_t target,
                                                      const char *field,
                                                      const struct DefraEncryptedIndexDeleteOptions *opts,
                                                      uint32_t opts_size);

/*
 Lists the encrypted indexes of the collection named, or of every
 collection.

 # Safety

 As [`defra_index_new_sized`].
 */
struct DefraResult defra_encrypted_index_list_sized(uintptr_t target,
                                                    const struct DefraEncryptedIndexListOptions *opts,
                                                    uint32_t opts_size);

/*
 Starts a node. `defra_node_new` in `defra.h` passes the size.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes, whose strings
 and byte buffers are valid as `NodeInitOptions` describes.
 */
struct DefraNewNodeResult defra_node_new_sized(const struct DefraNodeInitOptions *opts,
                                               uint32_t opts_size);

/*
 Creates a transaction on node `node`. Its handle stands in for the node
 in any call that takes a target.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraNewTxnResult defra_transaction_create_sized(uintptr_t node,
                                                        const struct DefraTransactionOptions *opts,
                                                        uint32_t opts_size);

/*
 Answers the library's version.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_version_get_sized(const struct DefraVersionOptions *opts,
                                           uint32_t opts_size);

/*
 Connects to the peers at `peer_addresses`, comma-separated.

 # Safety

 `peer_addresses` must be null or a valid null-terminated string; `opts` must be
 null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_connect_sized(uintptr_t target,
                                           const char *peer_addresses,
                                           const struct DefraP2POptions *opts,
                                           uint32_t opts_size);

/*
 Disconnects from the peers at `peer_addresses`, comma-separated.

 # Safety

 `peer_addresses` must be null or a valid null-terminated string; `opts` must be
 null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_disconnect_sized(uintptr_t target,
                                              const char *peer_addresses,
                                              const struct DefraP2POptions *opts,
                                              uint32_t opts_size);

/*
 Answers this node's peer information.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_info_sized(uintptr_t target,
                                        const struct DefraP2POptions *opts,
                                        uint32_t opts_size);

/*
 Lists the connected peers.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_peer_list_sized(uintptr_t target,
                                             const struct DefraP2POptions *opts,
                                             uint32_t opts_size);

/*
 Starts replicating the collections in `collections`, comma-separated.

 # Safety

 `collections` must be null or a valid null-terminated string; `opts` must be
 null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_collection_add_sized(uintptr_t target,
                                                  const char *collections,
                                                  const struct DefraP2POptions *opts,
                                                  uint32_t opts_size);

/*
 Stops replicating the collections in `collections`, comma-separated.

 # Safety

 `collections` must be null or a valid null-terminated string; `opts` must be
 null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_collection_delete_sized(uintptr_t target,
                                                     const char *collections,
                                                     const struct DefraP2POptions *opts,
                                                     uint32_t opts_size);

/*
 Lists the collections this node replicates.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_collection_list_sized(uintptr_t target,
                                                   const struct DefraP2POptions *opts,
                                                   uint32_t opts_size);

/*
 Starts replicating the documents in `doc_ids`, comma-separated.

 # Safety

 `doc_ids` must be null or a valid null-terminated string; `opts` must be
 null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_document_add_sized(uintptr_t target,
                                                const char *doc_ids,
                                                const struct DefraP2POptions *opts,
                                                uint32_t opts_size);

/*
 Stops replicating the documents in `doc_ids`, comma-separated.

 # Safety

 `doc_ids` must be null or a valid null-terminated string; `opts` must be
 null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_document_delete_sized(uintptr_t target,
                                                   const char *doc_ids,
                                                   const struct DefraP2POptions *opts,
                                                   uint32_t opts_size);

/*
 Lists the documents this node replicates.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_document_list_sized(uintptr_t target,
                                                 const struct DefraP2POptions *opts,
                                                 uint32_t opts_size);

/*
 Retries every replicator that failed to receive updates.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_replicator_retry_sized(uintptr_t target,
                                                    const struct DefraP2POptions *opts,
                                                    uint32_t opts_size);

/*
 Lists the replicators.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_replicator_list_sized(uintptr_t target,
                                                   const struct DefraP2POptions *opts,
                                                   uint32_t opts_size);

/*
 Answers an address another device can connect to this node at.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_mobile_shareable_address_sized(uintptr_t target,
                                                        const struct DefraP2POptions *opts,
                                                        uint32_t opts_size);

/*
 Answers how far this node's sync has got.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_mobile_sync_status_sized(uintptr_t target,
                                                  const struct DefraP2POptions *opts,
                                                  uint32_t opts_size);

/*
 Tells the node its network changed, so it reconnects.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_mobile_notify_network_change_sized(uintptr_t target,
                                                            const struct DefraP2POptions *opts,
                                                            uint32_t opts_size);

/*
 Answers what this node knows of peer `peer_id`.

 # Safety

 `peer_id` must be null or a valid null-terminated string; `opts` must be
 null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_mobile_peer_details_sized(uintptr_t target,
                                                   const char *peer_id,
                                                   const struct DefraP2POptions *opts,
                                                   uint32_t opts_size);

/*
 Syncs branchable collection `collection_id` from connected peers.

 # Safety

 `collection_id` must be null or a valid null-terminated string; `opts`
 must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_p2p_collection_sync_branchable_sized(uintptr_t target,
                                                              const char *collection_id,
                                                              const struct DefraP2PSyncOptions *opts,
                                                              uint32_t opts_size);

/*
 Syncs the collection versions in `version_ids`, comma-separated, from
 connected peers.

 # Safety

 As [`defra_p2p_collection_sync_branchable_sized`].
 */
struct DefraResult defra_p2p_collection_sync_versions_sized(uintptr_t target,
                                                            const char *version_ids,
                                                            const struct DefraP2PSyncOptions *opts,
                                                            uint32_t opts_size);

/*
 Syncs the documents in `doc_ids`, comma-separated, of `collection` from
 connected peers.

 # Safety

 As [`defra_p2p_collection_sync_branchable_sized`].
 */
struct DefraResult defra_p2p_document_sync_sized(uintptr_t target,
                                                 const char *collection,
                                                 const char *doc_ids,
                                                 const struct DefraP2PSyncOptions *opts,
                                                 uint32_t opts_size);

/*
 Adds a replicator at `addresses` for `collections`, both comma-separated.

 # Safety

 Every string must be null or a valid null-terminated string; `opts` must
 be null or point to `opts_size` readable bytes whose strings are valid.
 */
struct DefraResult defra_p2p_replicator_add_sized(uintptr_t target,
                                                  const char *collections,
                                                  const char *addresses,
                                                  const struct DefraReplicatorAddOptions *opts,
                                                  uint32_t opts_size);

/*
 Deletes replicator `id`, or stops it replicating the collections named.

 # Safety

 As [`defra_p2p_replicator_add_sized`].
 */
struct DefraResult defra_p2p_replicator_delete_sized(uintptr_t target,
                                                     const char *id,
                                                     const struct DefraReplicatorDeleteOptions *opts,
                                                     uint32_t opts_size);

/*
 Runs the GraphQL request `query` and answers with its JSON response.

 # Safety

 `query` must be null or a valid null-terminated UTF-8 string; `opts` must
 be null or point to `opts_size` readable bytes whose strings are valid.
 */
struct DefraResult defra_query_execute_sized(uintptr_t target,
                                             const char *query,
                                             const struct DefraQueryOptions *opts,
                                             uint32_t opts_size);

/*
 `defra_query_execute` for a caller called back instead of waiting:
 `callback` receives `context` and the result exactly once. Options it
 cannot read are reported to `callback` before this returns.

 # Safety

 As [`defra_query_execute_sized`]; the strings are read only before this
 returns.
 */
void defra_query_execute_async_sized(uintptr_t target,
                                     const char *query,
                                     DefraResultCallback callback,
                                     uintptr_t context,
                                     const struct DefraQueryOptions *opts,
                                     uint32_t opts_size);

/*
 Subscribes to the node events named in `names`, comma-separated, and
 answers with the subscription's ID for `defra_subscription_poll`.

 # Safety

 `names` must be null or a valid null-terminated string; `opts` must be
 null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_event_subscribe_sized(uintptr_t target,
                                               const char *names,
                                               const struct DefraEventSubscribeOptions *opts,
                                               uint32_t opts_size);

/*
 Adds a view: the collection `sdl` defines, whose documents `query`
 answers.

 # Safety

 Every string must be null or a valid null-terminated string; `opts` must
 be null or point to `opts_size` readable bytes whose strings are valid.
 */
struct DefraResult defra_view_add_sized(uintptr_t target,
                                        const char *query,
                                        const char *sdl,
                                        const struct DefraViewAddOptions *opts,
                                        uint32_t opts_size);

/*
 Refreshes the materialized views the options choose.

 # Safety

 As [`defra_view_add_sized`].
 */
struct DefraResult defra_view_refresh_sized(uintptr_t target,
                                            const struct DefraViewRefreshOptions *opts,
                                            uint32_t opts_size);

/*
 Adds the lens `config` describes, and answers with its CID.

 # Safety

 As [`defra_view_add_sized`].
 */
struct DefraResult defra_lens_add_sized(uintptr_t target,
                                        const char *config,
                                        const struct DefraLensOptions *opts,
                                        uint32_t opts_size);

/*
 Lists the lenses.

 # Safety

 As [`defra_view_add_sized`].
 */
struct DefraResult defra_lens_list_sized(uintptr_t target,
                                         const struct DefraLensOptions *opts,
                                         uint32_t opts_size);

/*
 Sets the migration from collection version `src` to `dst` to the lens
 `config` describes.

 # Safety

 As [`defra_view_add_sized`].
 */
struct DefraResult defra_lens_set_sized(uintptr_t target,
                                        const char *src,
                                        const char *dst,
                                        const char *config,
                                        const struct DefraLensOptions *opts,
                                        uint32_t opts_size);

/*
 Exports the documents the caller can read to the JSON file `filepath`.

 # Safety

 As [`defra_view_add_sized`].
 */
struct DefraResult defra_backup_export_sized(uintptr_t target,
                                             const char *filepath,
                                             const struct DefraBackupExportOptions *opts,
                                             uint32_t opts_size);

/*
 Imports the documents in the JSON backup file `filepath`.

 # Safety

 As [`defra_view_add_sized`].
 */
struct DefraResult defra_backup_import_sized(uintptr_t target,
                                             const char *filepath,
                                             const struct DefraBackupImportOptions *opts,
                                             uint32_t opts_size);

/*
 Logs every key of the node's store, with its value's size. Unwise on a
 large database.

 # Safety

 `opts` must be null or point to `opts_size` readable bytes.
 */
struct DefraResult defra_print_dump_sized(uintptr_t target,
                                          const struct DefraPrintDumpOptions *opts,
                                          uint32_t opts_size);

/*
 Returns the identity's private key, hex-encoded. The key leaves no unwiped
 copy behind, and [`defra_result_free`](crate::types::defra_result_free)
 wipes it before freeing it.
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
 Closes the node and releases its handle, with every transaction and
 subscription still open on it.
 */
struct DefraResult defra_node_close(uintptr_t node_ptr);

/*
 Returns the node's configuration options as JSON, with every value that
 could locate or unlock the node redacted.
 */
struct DefraResult defra_node_options(uintptr_t node_ptr);

struct DefraResult defra_profiling_start(const char *output_path);

struct DefraResult defra_profiling_stop(void);

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
 Commits the transaction and releases its handle, whether or not the
 commit succeeds. It runs after every call made on the transaction before
 it, asynchronous queries still queued included.
 */
struct DefraResult defra_transaction_commit(uintptr_t txn_ptr);

/*
 Discards the transaction and releases its handle, after every call made
 on it before. Discarding a handle that is already released does nothing.
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

#ifdef __cplusplus
}  // extern "C"
#endif  // __cplusplus

#endif  /* DEFRA_H */

static inline struct DefraResult defra_acp_dac_policy_add(uintptr_t target, const char *policy, DefraDacPolicyAddOptions opts) {
    return defra_acp_dac_policy_add_sized(target, policy, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_acp_dac_relationship_add(uintptr_t target, const char *collection, const char *doc_id, const char *relation, const char *actor, DefraDacRelationshipOptions opts) {
    return defra_acp_dac_relationship_add_sized(target, collection, doc_id, relation, actor, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_acp_dac_relationship_delete(uintptr_t target, const char *collection, const char *doc_id, const char *relation, const char *actor, DefraDacRelationshipOptions opts) {
    return defra_acp_dac_relationship_delete_sized(target, collection, doc_id, relation, actor, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_acp_nac_status(uintptr_t target, DefraNacOptions opts) {
    return defra_acp_nac_status_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_acp_nac_disable(uintptr_t target, DefraNacOptions opts) {
    return defra_acp_nac_disable_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_acp_nac_re_enable(uintptr_t target, DefraNacOptions opts) {
    return defra_acp_nac_re_enable_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_acp_nac_purge(uintptr_t target, DefraNacOptions opts) {
    return defra_acp_nac_purge_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_acp_nac_relationship_add(uintptr_t target, const char *relation, const char *actor, DefraNacRelationshipOptions opts) {
    return defra_acp_nac_relationship_add_sized(target, relation, actor, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_acp_nac_relationship_delete(uintptr_t target, const char *relation, const char *actor, DefraNacRelationshipOptions opts) {
    return defra_acp_nac_relationship_delete_sized(target, relation, actor, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_action_list(uintptr_t target, DefraActionListOptions opts) {
    return defra_action_list_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_batch_start(uintptr_t target, DefraBatchOptions opts) {
    return defra_batch_start_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_batch_sign(uintptr_t target, DefraBatchOptions opts) {
    return defra_batch_sign_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_block_verify(uintptr_t target, const char *cid, const char *key_type, const char *public_key, DefraBlockVerifyOptions opts) {
    return defra_block_verify_sized(target, cid, key_type, public_key, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_collection_add(uintptr_t target, const char *sdl, DefraCollectionAddOptions opts) {
    return defra_collection_add_sized(target, sdl, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_collection_delete(uintptr_t target, const char *names, DefraCollectionDeleteOptions opts) {
    return defra_collection_delete_sized(target, names, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_collection_describe(uintptr_t target, DefraCollectionDescribeOptions opts) {
    return defra_collection_describe_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_collection_patch(uintptr_t target, const char *patch, DefraCollectionPatchOptions opts) {
    return defra_collection_patch_sized(target, patch, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_collection_set_active(uintptr_t target, const char *version_id, DefraCollectionSetActiveOptions opts) {
    return defra_collection_set_active_sized(target, version_id, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_collection_truncate(uintptr_t target, DefraCollectionTruncateOptions opts) {
    return defra_collection_truncate_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_document_add(uintptr_t target, const char *json, DefraDocumentAddOptions opts) {
    return defra_document_add_sized(target, json, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_document_save_many(uintptr_t target, const char *json, DefraDocumentAddOptions opts) {
    return defra_document_save_many_sized(target, json, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_document_get(uintptr_t target, const char *doc_id, DefraDocumentGetOptions opts) {
    return defra_document_get_sized(target, doc_id, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_document_update(uintptr_t target, const char *updater, DefraDocumentUpdateOptions opts) {
    return defra_document_update_sized(target, updater, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_document_delete(uintptr_t target, DefraDocumentDeleteOptions opts) {
    return defra_document_delete_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_document_purge(uintptr_t target, const char *doc_ids, DefraDocumentPurgeOptions opts) {
    return defra_document_purge_sized(target, doc_ids, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraNewIdentityResult defra_identity_new(DefraIdentityNewOptions opts) {
    return defra_identity_new_sized(&opts, (uint32_t)sizeof opts);
}

static inline struct DefraNewIdentityResult defra_identity_new_from_private_key(const char *private_key_hex, DefraIdentityNewOptions opts) {
    return defra_identity_new_from_private_key_sized(private_key_hex, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraNewIdentityResult defra_identity_new_from_bearer_token(const char *did, const char *bearer_token, DefraIdentityBearerOptions opts) {
    return defra_identity_new_from_bearer_token_sized(did, bearer_token, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraNewIdentityResult defra_identity_new_remote(const char *did, DefraRemoteIdentityOptions opts) {
    return defra_identity_new_remote_sized(did, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_index_new(uintptr_t target, const char *fields, DefraIndexNewOptions opts) {
    return defra_index_new_sized(target, fields, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_index_delete(uintptr_t target, const char *name, DefraIndexDeleteOptions opts) {
    return defra_index_delete_sized(target, name, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_index_list(uintptr_t target, DefraIndexListOptions opts) {
    return defra_index_list_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_encrypted_index_new(uintptr_t target, const char *field, DefraEncryptedIndexNewOptions opts) {
    return defra_encrypted_index_new_sized(target, field, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_encrypted_index_delete(uintptr_t target, const char *field, DefraEncryptedIndexDeleteOptions opts) {
    return defra_encrypted_index_delete_sized(target, field, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_encrypted_index_list(uintptr_t target, DefraEncryptedIndexListOptions opts) {
    return defra_encrypted_index_list_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraNewNodeResult defra_node_new(DefraNodeInitOptions opts) {
    return defra_node_new_sized(&opts, (uint32_t)sizeof opts);
}

static inline struct DefraNewTxnResult defra_transaction_create(uintptr_t node, DefraTransactionOptions opts) {
    return defra_transaction_create_sized(node, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_version_get(DefraVersionOptions opts) {
    return defra_version_get_sized(&opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_connect(uintptr_t target, const char *peer_addresses, DefraP2POptions opts) {
    return defra_p2p_connect_sized(target, peer_addresses, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_disconnect(uintptr_t target, const char *peer_addresses, DefraP2POptions opts) {
    return defra_p2p_disconnect_sized(target, peer_addresses, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_info(uintptr_t target, DefraP2POptions opts) {
    return defra_p2p_info_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_peer_list(uintptr_t target, DefraP2POptions opts) {
    return defra_p2p_peer_list_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_collection_add(uintptr_t target, const char *collections, DefraP2POptions opts) {
    return defra_p2p_collection_add_sized(target, collections, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_collection_delete(uintptr_t target, const char *collections, DefraP2POptions opts) {
    return defra_p2p_collection_delete_sized(target, collections, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_collection_list(uintptr_t target, DefraP2POptions opts) {
    return defra_p2p_collection_list_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_document_add(uintptr_t target, const char *doc_ids, DefraP2POptions opts) {
    return defra_p2p_document_add_sized(target, doc_ids, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_document_delete(uintptr_t target, const char *doc_ids, DefraP2POptions opts) {
    return defra_p2p_document_delete_sized(target, doc_ids, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_document_list(uintptr_t target, DefraP2POptions opts) {
    return defra_p2p_document_list_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_replicator_retry(uintptr_t target, DefraP2POptions opts) {
    return defra_p2p_replicator_retry_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_replicator_list(uintptr_t target, DefraP2POptions opts) {
    return defra_p2p_replicator_list_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_mobile_shareable_address(uintptr_t target, DefraP2POptions opts) {
    return defra_mobile_shareable_address_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_mobile_sync_status(uintptr_t target, DefraP2POptions opts) {
    return defra_mobile_sync_status_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_mobile_notify_network_change(uintptr_t target, DefraP2POptions opts) {
    return defra_mobile_notify_network_change_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_mobile_peer_details(uintptr_t target, const char *peer_id, DefraP2POptions opts) {
    return defra_mobile_peer_details_sized(target, peer_id, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_collection_sync_branchable(uintptr_t target, const char *collection_id, DefraP2PSyncOptions opts) {
    return defra_p2p_collection_sync_branchable_sized(target, collection_id, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_collection_sync_versions(uintptr_t target, const char *version_ids, DefraP2PSyncOptions opts) {
    return defra_p2p_collection_sync_versions_sized(target, version_ids, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_document_sync(uintptr_t target, const char *collection, const char *doc_ids, DefraP2PSyncOptions opts) {
    return defra_p2p_document_sync_sized(target, collection, doc_ids, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_replicator_add(uintptr_t target, const char *collections, const char *addresses, DefraReplicatorAddOptions opts) {
    return defra_p2p_replicator_add_sized(target, collections, addresses, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_p2p_replicator_delete(uintptr_t target, const char *id, DefraReplicatorDeleteOptions opts) {
    return defra_p2p_replicator_delete_sized(target, id, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_query_execute(uintptr_t target, const char *query, DefraQueryOptions opts) {
    return defra_query_execute_sized(target, query, &opts, (uint32_t)sizeof opts);
}

static inline void defra_query_execute_async(uintptr_t target, const char *query, DefraResultCallback callback, uintptr_t context, DefraQueryOptions opts) {
    defra_query_execute_async_sized(target, query, callback, context, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_event_subscribe(uintptr_t target, const char *names, DefraEventSubscribeOptions opts) {
    return defra_event_subscribe_sized(target, names, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_view_add(uintptr_t target, const char *query, const char *sdl, DefraViewAddOptions opts) {
    return defra_view_add_sized(target, query, sdl, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_view_refresh(uintptr_t target, DefraViewRefreshOptions opts) {
    return defra_view_refresh_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_lens_add(uintptr_t target, const char *config, DefraLensOptions opts) {
    return defra_lens_add_sized(target, config, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_lens_list(uintptr_t target, DefraLensOptions opts) {
    return defra_lens_list_sized(target, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_lens_set(uintptr_t target, const char *src, const char *dst, const char *config, DefraLensOptions opts) {
    return defra_lens_set_sized(target, src, dst, config, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_backup_export(uintptr_t target, const char *filepath, DefraBackupExportOptions opts) {
    return defra_backup_export_sized(target, filepath, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_backup_import(uintptr_t target, const char *filepath, DefraBackupImportOptions opts) {
    return defra_backup_import_sized(target, filepath, &opts, (uint32_t)sizeof opts);
}

static inline struct DefraResult defra_print_dump(uintptr_t target, DefraPrintDumpOptions opts) {
    return defra_print_dump_sized(target, &opts, (uint32_t)sizeof opts);
}
