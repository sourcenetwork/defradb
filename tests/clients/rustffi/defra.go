//go:build rust_ffi

// Package rustffi provides Go bindings for the DefraDB Rust FFI.
//
// This package wraps the C FFI exposed by the Rust `ffi` crate,
// providing a Go-native interface for integration testing.
//
// Build requirements:
//   - Rust library must be built first: cargo build --release -p ffi
//   - CGO must be enabled: CGO_ENABLED=1
//   - CGO_LDFLAGS must point to the Rust library: -L<path>/target/release -lffi -ldl -lpthread -lm
//   - Build tag: -tags rust_ffi
//
// Example:
//
//	CGO_ENABLED=1 \
//	CGO_LDFLAGS="-L/path/to/defradb.rs/target/release -lffi -ldl -lpthread -lm" \
//	DEFRA_CLIENT_RUST_FFI=true \
//	go test -tags rust_ffi ./tests/integration/query/simple/...
package rustffi

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR} -ldefra_ffi -Wl,-rpath,${SRCDIR}

#include "defra.h"
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/sourcenetwork/defradb/client"
)

var (
	// initOnce ensures Init is only called once
	initOnce sync.Once
)

// cargs is a scratch arena for the C strings passed to a single FFI call.
// Every allocation is released by one `defer a.free()` at function exit,
// which is where the per-allocation defers it replaces used to run.
type cargs struct {
	p []unsafe.Pointer
}

// s allocates a C string that lives until free is called.
func (a *cargs) s(v string) *C.char {
	c := C.CString(v)
	a.p = append(a.p, unsafe.Pointer(c))
	return c
}

// opt is s, except that an empty string yields a nil pointer rather than an
// allocated empty string. The FFI reads nil as "argument not supplied".
func (a *cargs) opt(v string) *C.char {
	if v == "" {
		return nil
	}
	return a.s(v)
}

// free releases every allocation made through the arena.
func (a *cargs) free() {
	for _, p := range a.p {
		C.free(p)
	}
}

// unwrap turns an FFI result's status/error/value triple into (value, error),
// releasing whichever C string the call populated. On failure the value is
// null and only the error is freed; on success the reverse. Callers with no
// value to read pass nil, which defra_free_string ignores.
func unwrap(status C.int, cError *C.char, cValue *C.char, op string) (string, error) {
	if status != 0 {
		err := C.GoString(cError)
		C.defra_free_string(cError)
		return "", mapFFIError(op, err)
	}

	value := C.GoString(cValue)
	C.defra_free_string(cValue)
	return value, nil
}

// mapFFIError maps raw FFI error strings to proper Go error types.
//
// FFI errors are flat strings like "not authorized to perform operation. Permission: xyz".
// Go tests use errors.Is() against sentinel errors like client.ErrNotAuthorizedToPerformOperation.
// This function wraps known error patterns with the correct Go sentinel so errors.Is() matches.
//
// Call as: mapFFIError("op_name", rawErrString)
// Instead of: mapFFIError("op_name", rawErrString)
func mapFFIError(ffiOp string, rawErr string) error {
	switch {
	case strings.Contains(rawErr, "could not find block:"):
		cid := strings.TrimSpace(strings.TrimPrefix(rawErr, "could not find block:"))
		return fmt.Errorf("ipld: could not find %s", cid)

	case strings.Contains(rawErr, "not authorized to perform operation"):
		// Wrap with sentinel so errors.Is() works, preserve the full message
		return fmt.Errorf("%w. %s",
			client.ErrNotAuthorizedToPerformOperation,
			extractPermissionSuffix(rawErr))

	case strings.Contains(rawErr, "operation requires ACP, but ACP not available"):
		return client.ErrACPOperationButACPNotAvailable

	case strings.Contains(rawErr, "document not found or not authorized"):
		return client.ErrDocumentNotFoundOrNotAuthorized

	case strings.Contains(rawErr, "collection not found"),
		isCollectionNotFoundError(rawErr):
		return fmt.Errorf("%s: %w", rawErr, client.ErrCollectionNotFound)

	default:
		return fmt.Errorf("ffi: %s failed: %s", ffiOp, rawErr)
	}
}

// isCollectionNotFoundError checks for Rust FFI error patterns like
// "collection 'X' not found" or "collection 'X' not found - add schema..."
// where the collection name is embedded between "collection '" and "' not found".
func isCollectionNotFoundError(rawErr string) bool {
	idx := strings.Index(rawErr, "collection '")
	if idx < 0 {
		return false
	}
	return strings.Contains(rawErr[idx:], "' not found")
}

// extractPermissionSuffix extracts "Permission: xyz" from an FFI error string.
func extractPermissionSuffix(rawErr string) string {
	idx := strings.Index(rawErr, "Permission:")
	if idx >= 0 {
		return strings.TrimSpace(rawErr[idx:])
	}
	return ""
}

// Init initializes the FFI library.
// Must be called before any other FFI functions.
// Safe to call multiple times.
func Init() {
	initOnce.Do(func() {
		C.defra_init()
	})
}

// Version returns the library version string.
func Version() string {
	cstr := C.defra_version()
	if cstr == nil {
		return ""
	}
	defer C.defra_free_string(cstr)
	return C.GoString(cstr)
}

// Node represents a DefraDB node handle.
type Node struct {
	ptr C.uintptr_t
}

// NodeOptions configures node creation.
type NodeOptions struct {
	// DBPath is the path to the database directory.
	// If empty, an in-memory database is used.
	DBPath string

	// InMemory forces in-memory storage even if DBPath is set.
	InMemory bool

	// EnableSigning enables block signing on the node.
	// When true, the node uses a signing key for block signatures.
	// If SigningPrivateKey is provided, that key is used.
	// Otherwise, a random secp256k1 key pair is generated.
	EnableSigning bool

	// SigningKeyType specifies the type of signing key ("secp256k1" or "ed25519").
	// Only used when SigningPrivateKey is provided. Defaults to "secp256k1".
	SigningKeyType string

	// SigningPrivateKey is the raw private key bytes for signing.
	// If nil, the node will auto-generate a key when EnableSigning is true.
	SigningPrivateKey []byte

	// SourceHubGRPCAddress is the gRPC/LCD address of the SourceHub node.
	// When set, the node uses SourceHub for document ACP instead of local ACP.
	SourceHubGRPCAddress string

	// SourceHubCometRPCAddress is the CometBFT RPC address of the SourceHub node.
	SourceHubCometRPCAddress string

	// SourceHubChainID is the chain ID (e.g., "sourcehub-test").
	SourceHubChainID string

	// SourceHubSignerKey is the raw secp256k1 private key bytes for SourceHub transactions.
	SourceHubSignerKey []byte
}

// NewNode creates a new DefraDB node.
// The node must be closed with Close() when done.
func NewNode(opts NodeOptions) (*Node, error) {
	var a cargs
	defer a.free()

	var cOpts C.struct_NodeInitOptions

	if opts.DBPath != "" {
		cOpts.db_path = a.s(opts.DBPath)
	}

	if opts.InMemory || opts.DBPath == "" {
		cOpts.in_memory = 1
	} else {
		cOpts.in_memory = 0
	}

	if opts.EnableSigning {
		cOpts.enable_signing = 1
	}

	// Pass signing key if provided
	if len(opts.SigningPrivateKey) > 0 {
		cOpts.signing_private_key = (*C.uint8_t)(unsafe.Pointer(&opts.SigningPrivateKey[0]))
		cOpts.signing_private_key_len = C.uintptr_t(len(opts.SigningPrivateKey))

		keyType := opts.SigningKeyType
		if keyType == "" {
			keyType = "secp256k1"
		}
		cOpts.signing_key_type = a.s(keyType)
	}

	// Pass SourceHub config if provided
	if opts.SourceHubGRPCAddress != "" {
		cOpts.sourcehub_grpc_address = a.s(opts.SourceHubGRPCAddress)
		cOpts.sourcehub_comet_rpc_address = a.s(opts.SourceHubCometRPCAddress)
		cOpts.sourcehub_chain_id = a.s(opts.SourceHubChainID)

		if len(opts.SourceHubSignerKey) > 0 {
			cOpts.sourcehub_signer_key = (*C.uint8_t)(unsafe.Pointer(&opts.SourceHubSignerKey[0]))
			cOpts.sourcehub_signer_key_len = C.uintptr_t(len(opts.SourceHubSignerKey))
		}
	}

	result := C.new_node(cOpts)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("new_node", err)
	}

	return &Node{ptr: result.node_ptr}, nil
}

// Close closes the node and releases resources.
// After calling Close, the node handle is no longer valid.
func (n *Node) Close() error {
	result := C.node_close(n.ptr)

	_, err := unwrap(result.status, result.error, nil, "node_close")
	return err
}

// AddSchema adds a GraphQL SDL schema to the database.
// Returns the JSON response containing created collection versions.
func (n *Node) AddSchema(identityDID string, sdl string) (string, error) {
	var a cargs
	defer a.free()

	result := C.add_schema(n.ptr, a.opt(identityDID), a.s(sdl))

	return unwrap(result.status, result.error, result.value, "add_schema")
}

// AddSchemaInTxn adds a GraphQL SDL schema within a specific transaction.
// Returns the JSON response containing created collection versions visible in that transaction.
func (n *Node) AddSchemaInTxn(txnID string, identityDID string, sdl string) (string, error) {
	var a cargs
	defer a.free()

	result := C.add_schema_in_txn(n.ptr, a.s(txnID), a.opt(identityDID), a.s(sdl))

	return unwrap(result.status, result.error, result.value, "add_schema_in_txn")
}

// GetCollections returns all collections in the database as JSON.
func (n *Node) GetCollections(identityDID string) (string, error) {
	var a cargs
	defer a.free()

	result := C.get_collections(n.ptr, a.opt(identityDID))

	return unwrap(result.status, result.error, result.value, "get_collections")
}

// GetCollectionsInTxn returns all collection versions visible within a specific transaction.
// This reads from the transaction's systemstore, which includes uncommitted writes
// (e.g., placeholders from set_migration_in_txn).
func (n *Node) GetCollectionsInTxn(txnID string, identityDID string) (string, error) {
	var a cargs
	defer a.free()

	result := C.get_collections_in_txn(n.ptr, a.s(txnID), a.opt(identityDID))

	return unwrap(result.status, result.error, result.value, "get_collections_in_txn")
}

// QueryResult represents a GraphQL query response.
type QueryResult struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []QueryError    `json:"errors,omitempty"`
}

// QueryError represents a GraphQL error.
type QueryError struct {
	Message   string          `json:"message"`
	Locations []ErrorLocation `json:"locations,omitempty"`
	Path      []interface{}   `json:"path,omitempty"`
}

// ErrorLocation indicates where an error occurred in the query.
type ErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// ExecRequestResult contains the result of a GraphQL request.
// For subscriptions, SubscriptionID will be set instead of Response.
type ExecRequestResult struct {
	// Response contains the JSON response for queries/mutations.
	Response string
	// SubscriptionID is set when the request is a subscription.
	// Use PollGraphQLSubscription and CloseGraphQLSubscription to manage it.
	SubscriptionID string
	// IsSubscription indicates whether this is a subscription result.
	IsSubscription bool
}

// ExecRequest executes a GraphQL query or mutation.
// Returns the raw JSON response string.
// identityDID is the DID of the caller for ACP permission checks (empty string for anonymous).
func (n *Node) ExecRequest(identityDID string, query string, operationName string, variables string) (string, error) {
	result, err := n.ExecRequestFull(identityDID, query, operationName, variables)
	if err != nil {
		return "", err
	}
	if result.IsSubscription {
		return "", fmt.Errorf("ffi: exec_request returned subscription, use ExecRequestFull for subscription support")
	}
	return result.Response, nil
}

// ExecRequestFull executes a GraphQL query, mutation, or subscription.
// Returns an ExecRequestResult that indicates whether the result is a subscription.
func (n *Node) ExecRequestFull(identityDID string, query string, operationName string, variables string) (*ExecRequestResult, error) {
	var a cargs
	defer a.free()

	result := C.exec_request(n.ptr, a.opt(identityDID), a.s(query), a.opt(operationName), a.opt(variables), a.s(""))

	switch result.status {
	case 0: // Success - query/mutation result
		value := C.GoString(result.value)
		C.defra_free_string(result.value)
		return &ExecRequestResult{Response: value}, nil

	case 1: // Error
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("exec_request", err)

	case 2: // Subscription
		subID := C.GoString(result.value)
		C.defra_free_string(result.value)
		return &ExecRequestResult{
			SubscriptionID: subID,
			IsSubscription: true,
		}, nil

	default:
		return nil, fmt.Errorf("ffi: exec_request returned unknown status: %d", result.status)
	}
}

// GraphQLSubscriptionResult represents a result from polling a GraphQL subscription.
type GraphQLSubscriptionResult struct {
	// HasResult indicates if there's a new result available.
	HasResult bool
	// Result contains the JSON result when HasResult is true.
	Result string
	// IsClosed indicates if the subscription has been closed.
	IsClosed bool
}

// PollGraphQLSubscription polls a GraphQL subscription for new results.
func PollGraphQLSubscription(subscriptionID string) (*GraphQLSubscriptionResult, error) {
	var a cargs
	defer a.free()

	result := C.poll_graphql_subscription(a.s(subscriptionID))

	switch result.status {
	case 0: // Result available
		value := C.GoString(result.value)
		C.defra_free_string(result.value)
		return &GraphQLSubscriptionResult{HasResult: true, Result: value}, nil

	case 1: // Error
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("poll_graphql_subscription", err)

	case 2: // No result available
		return &GraphQLSubscriptionResult{HasResult: false}, nil

	case 3: // Subscription closed
		return &GraphQLSubscriptionResult{IsClosed: true}, nil

	default:
		return nil, fmt.Errorf("ffi: poll_graphql_subscription returned unknown status: %d", result.status)
	}
}

// CloseGraphQLSubscription closes a GraphQL subscription and releases resources.
func CloseGraphQLSubscription(subscriptionID string) error {
	var a cargs
	defer a.free()

	result := C.close_graphql_subscription(a.s(subscriptionID))

	_, err := unwrap(result.status, result.error, nil, "close_graphql_subscription")
	return err
}

// Query executes a GraphQL query and returns a parsed result.
func (n *Node) Query(query string) (*QueryResult, error) {
	return n.QueryWithVars(query, "", nil)
}

// QueryWithVars executes a GraphQL query with variables.
func (n *Node) QueryWithVars(query string, operationName string, variables map[string]interface{}) (*QueryResult, error) {
	var varsJSON string
	if variables != nil {
		varsBytes, err := json.Marshal(variables)
		if err != nil {
			return nil, fmt.Errorf("ffi: failed to marshal variables: %w", err)
		}
		varsJSON = string(varsBytes)
	}

	responseJSON, err := n.ExecRequest("", query, operationName, varsJSON)
	if err != nil {
		return nil, err
	}

	var result QueryResult
	if err := json.Unmarshal([]byte(responseJSON), &result); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse response: %w", err)
	}

	return &result, nil
}

// Transaction represents an active database transaction.
type Transaction struct {
	node *Node
	id   string
}

// BeginTxn starts a new transaction.
// If readonly is true, the transaction cannot perform write operations.
// The transaction must be committed or rolled back when done.
func (n *Node) BeginTxn(readonly bool) (*Transaction, error) {
	var readonlyInt C.int32_t
	if readonly {
		readonlyInt = 1
	}

	result := C.begin_txn(n.ptr, readonlyInt)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("begin_txn", err)
	}

	txnID := C.GoString(result.txn_id)
	C.defra_free_string(result.txn_id)

	return &Transaction{node: n, id: txnID}, nil
}

// ID returns the transaction ID.
func (t *Transaction) ID() string {
	return t.id
}

// Commit commits the transaction, making all changes permanent.
// After commit, the transaction is no longer valid.
func (t *Transaction) Commit() error {
	var a cargs
	defer a.free()

	result := C.commit_txn(t.node.ptr, a.s(t.id))

	_, err := unwrap(result.status, result.error, nil, "commit_txn")
	return err
}

// Rollback discards all changes made in the transaction.
// After rollback, the transaction is no longer valid.
func (t *Transaction) Rollback() error {
	var a cargs
	defer a.free()

	result := C.rollback_txn(t.node.ptr, a.s(t.id))

	_, err := unwrap(result.status, result.error, nil, "rollback_txn")
	return err
}

// ExecRequest executes a GraphQL query or mutation within the transaction.
// identityDID is the DID of the caller for ACP permission checks (empty string for anonymous).
func (t *Transaction) ExecRequest(identityDID string, query string, operationName string, variables string) (string, error) {
	var a cargs
	defer a.free()

	result := C.exec_request_in_txn(t.node.ptr, a.s(t.id), a.opt(identityDID), a.s(query), a.opt(operationName), a.opt(variables), a.s(""))

	return unwrap(result.status, result.error, result.value, "exec_request_in_txn")
}

// Query executes a GraphQL query within the transaction.
func (t *Transaction) Query(query string) (*QueryResult, error) {
	responseJSON, err := t.ExecRequest("", query, "", "")
	if err != nil {
		return nil, err
	}

	var result QueryResult
	if err := json.Unmarshal([]byte(responseJSON), &result); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse response: %w", err)
	}

	return &result, nil
}

// Mutate executes a GraphQL mutation within the transaction.
func (t *Transaction) Mutate(mutation string) (*QueryResult, error) {
	return t.Query(mutation)
}

// DeleteCollections deletes collections within the transaction.
func (t *Transaction) DeleteCollections(identityDID string, targets []string, activeOnly bool) error {
	var a cargs
	defer a.free()

	targetsJSON, err := json.Marshal(targets)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal collection targets: %w", err)
	}

	result := C.delete_collections_in_txn(
		t.node.ptr,
		a.s(t.id),
		a.opt(identityDID),
		a.s(string(targetsJSON)),
		C.bool(activeOnly),
	)
	_, err = unwrap(result.status, result.error, result.value, "delete_collections_in_txn")
	return err
}

// SetCollectionActive updates a collection version within the transaction.
func (t *Transaction) SetCollectionActive(identityDID string, versionID string, isActive bool) error {
	var a cargs
	defer a.free()

	result := C.set_collection_active_in_txn(
		t.node.ptr,
		a.s(t.id),
		a.opt(identityDID),
		a.s(versionID),
		C.bool(isActive),
	)
	_, err := unwrap(result.status, result.error, result.value, "set_collection_active_in_txn")
	return err
}

// ============================================================================
// Collection Functions
// ============================================================================

// GetCollectionByName returns a collection by its name.
// Returns the collection's schema as JSON if found.
func (n *Node) GetCollectionByName(identityDID string, name string) (string, error) {
	var a cargs
	defer a.free()

	result := C.get_collection_by_name(n.ptr, a.opt(identityDID), a.s(name))

	return unwrap(result.status, result.error, result.value, "get_collection_by_name")
}

// DeleteCollectionVersions deletes multiple collection versions by their version IDs.
// Versions are deleted in topological order (children before parents).
func (n *Node) DeleteCollectionVersions(identityDID string, versionIDs []string) error {
	var a cargs
	defer a.free()

	idsJSON, err := json.Marshal(versionIDs)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal version IDs: %w", err)
	}

	result := C.delete_collection_versions(n.ptr, a.opt(identityDID), a.s(string(idsJSON)))

	_, err = unwrap(result.status, result.error, result.value, "delete_collection_versions")
	return err
}

// TruncateCollection deletes all documents from a collection while preserving the schema.
func (n *Node) TruncateCollection(identityDID string, name string) error {
	var a cargs
	defer a.free()

	result := C.truncate_collection(n.ptr, a.opt(identityDID), a.s(name))

	_, err := unwrap(result.status, result.error, result.value, "truncate_collection")
	return err
}

// SetActiveCollectionVersion activates the collection with the given version ID.
func (n *Node) SetActiveCollectionVersion(identityDID string, versionID string) error {
	var a cargs
	defer a.free()

	result := C.set_active_collection_version(n.ptr, a.opt(identityDID), a.s(versionID))

	_, err := unwrap(result.status, result.error, result.value, "set_active_collection_version")
	return err
}

// PatchCollection applies a JSON patch to a collection's schema.
// Returns the updated collection schema as JSON.
func (n *Node) PatchCollection(identityDID string, collectionName string, patch string) (string, error) {
	var a cargs
	defer a.free()

	result := C.patch_collection(n.ptr, a.opt(identityDID), a.s(collectionName), a.s(patch))

	return unwrap(result.status, result.error, result.value, "patch_collection")
}

// GetCollectionByVersionID returns a collection by its version ID.
// Returns the collection's schema as JSON if found, or "null" if not found.
func (n *Node) GetCollectionByVersionID(identityDID string, versionID string) (string, error) {
	var a cargs
	defer a.free()

	result := C.get_collection_by_version_id(n.ptr, a.opt(identityDID), a.s(versionID))

	return unwrap(result.status, result.error, result.value, "get_collection_by_version_id")
}

// AddView creates a new Defra View from a GQL query and SDL schema.
// The transform parameter is optional (pass empty string for none).
// Note: Not yet implemented - see issue #178.
func (n *Node) AddView(identityDID string, gqlQuery string, sdl string, transform string) (string, error) {
	var a cargs
	defer a.free()

	result := C.add_view(n.ptr, a.opt(identityDID), a.s(gqlQuery), a.s(sdl), a.opt(transform))

	return unwrap(result.status, result.error, result.value, "add_view")
}

// RefreshViews refreshes the caches of all views matching the given options.
// Pass empty string for options to refresh all views.
// Note: Not yet implemented - see issue #178.
func (n *Node) RefreshViews(identityDID string, options string) error {
	var a cargs
	defer a.free()

	result := C.refresh_views(n.ptr, a.opt(identityDID), a.opt(options))

	_, err := unwrap(result.status, result.error, result.value, "refresh_views")
	return err
}

// MaterializeCollection eagerly migrates and caches every known-version
// document in a collection. It returns the number of documents advanced.
func (n *Node) MaterializeCollection(identityDID string, collectionName string) (int, error) {
	var a cargs
	defer a.free()

	result := C.materialize_collection(n.ptr, a.opt(identityDID), a.s(collectionName))

	value, err := unwrap(result.status, result.error, result.value, "materialize_collection")
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("ffi: failed to parse materialized document count %q: %w", value, err)
	}
	return count, nil
}

// SetMigration sets the migration for collection versions.
// The config parameter should be a JSON string containing LensConfig.
func (n *Node) SetMigration(identityDID string, config string) (string, error) {
	var a cargs
	defer a.free()

	result := C.set_migration(n.ptr, a.opt(identityDID), a.s(config))

	return unwrap(result.status, result.error, result.value, "set_migration")
}

// SetMigrationInTxn sets the migration for collection versions within a transaction.
// The migration will only be visible after the transaction is committed.
func (n *Node) SetMigrationInTxn(txnID string, identityDID string, config string) (string, error) {
	var a cargs
	defer a.free()

	result := C.set_migration_in_txn(n.ptr, a.s(txnID), a.opt(identityDID), a.s(config))

	return unwrap(result.status, result.error, result.value, "set_migration_in_txn")
}

// LensAdd adds a lens transform to the database.
// The lensJSON parameter should be a JSON string matching Go's model.Lens format.
func (n *Node) LensAdd(identityDID string, lensJSON string) (string, error) {
	var a cargs
	defer a.free()

	result := C.lens_add(n.ptr, a.opt(identityDID), a.s(lensJSON))

	return unwrap(result.status, result.error, result.value, "lens_add")
}

// LensAddInTxn adds a lens transform within a transaction.
func (n *Node) LensAddInTxn(txnID string, identityDID string, lensJSON string) (string, error) {
	var a cargs
	defer a.free()

	result := C.lens_add_in_txn(n.ptr, a.s(txnID), a.opt(identityDID), a.s(lensJSON))

	return unwrap(result.status, result.error, result.value, "lens_add_in_txn")
}

// LensList returns all lens transforms as a map of ID -> LensModule JSON.
func (n *Node) LensList(identityDID string) (string, error) {
	var a cargs
	defer a.free()

	result := C.lens_list(n.ptr, a.opt(identityDID))

	return unwrap(result.status, result.error, result.value, "lens_list")
}

// LensListInTxn lists all lens transforms visible within a transaction.
func (n *Node) LensListInTxn(txnID string, identityDID string) (string, error) {
	var a cargs
	defer a.free()

	result := C.lens_list_in_txn(n.ptr, a.s(txnID), a.opt(identityDID))

	return unwrap(result.status, result.error, result.value, "lens_list_in_txn")
}

// ============================================================================
// Index Functions
// ============================================================================

// IndexField describes a field within an index.
type IndexField struct {
	Name       string `json:"Name"`
	Descending bool   `json:"Descending,omitempty"`
}

// IndexDescription describes a secondary index on a collection.
type IndexDescription struct {
	Name   string       `json:"Name"`
	ID     uint32       `json:"ID,omitempty"`
	Fields []IndexField `json:"Fields"`
	Unique bool         `json:"Unique,omitempty"`
}

type IndexResult struct {
	IndexDescription
	CollectionName string                 `json:"CollectionName"`
	Execution      client.ActionExecution `json:"Execution"`
}

// CreateIndex creates a new index on a collection.
// Returns the created index description with assigned ID.
func (n *Node) CreateIndex(identityDID string, collectionName string, indexName string, fields []IndexField, unique bool) (*IndexDescription, error) {
	var a cargs
	defer a.free()

	indexInput := IndexDescription{
		Name:   indexName,
		Fields: fields,
		Unique: unique,
	}
	indexJSON, err := json.Marshal(indexInput)
	if err != nil {
		return nil, fmt.Errorf("ffi: failed to marshal index: %w", err)
	}

	result := C.create_index(n.ptr, a.opt(identityDID), a.s(collectionName), a.s(string(indexJSON)))

	value, err := unwrap(result.status, result.error, result.value, "create_index")
	if err != nil {
		return nil, err
	}

	var index IndexDescription
	if err := json.Unmarshal([]byte(value), &index); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse index: %w", err)
	}

	return &index, nil
}

// DropIndex drops an index from a collection.
func (n *Node) DropIndex(identityDID string, collectionName string, indexName string) error {
	var a cargs
	defer a.free()

	result := C.delete_index(n.ptr, a.opt(identityDID), a.s(collectionName), a.s(indexName))

	_, err := unwrap(result.status, result.error, result.value, "drop_index")
	return err
}

// GetIndexes returns all indexes for a collection.
func (n *Node) GetIndexes(identityDID string, collectionName string) ([]IndexResult, error) {
	var a cargs
	defer a.free()

	result := C.get_indexes(n.ptr, a.opt(identityDID), a.s(collectionName))

	value, err := unwrap(result.status, result.error, result.value, "get_indexes")
	if err != nil {
		return nil, err
	}

	var indexes []IndexResult
	if err := json.Unmarshal([]byte(value), &indexes); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse indexes: %w", err)
	}

	return indexes, nil
}

// GetAllIndexes returns all indexes across all collections.
func (n *Node) GetAllIndexes(identityDID string) (map[string][]IndexResult, error) {
	var a cargs
	defer a.free()

	result := C.list_all_indexes(n.ptr, a.opt(identityDID))

	value, err := unwrap(result.status, result.error, result.value, "get_all_indexes")
	if err != nil {
		return nil, err
	}

	var indexes map[string][]IndexResult
	if err := json.Unmarshal([]byte(value), &indexes); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse indexes: %w", err)
	}

	return indexes, nil
}

// ============================================================================
// Encrypted Index Functions (Searchable Encryption)
// ============================================================================

// EncryptedIndexDescription represents an encrypted index for searchable encryption.
type EncryptedIndexDescription struct {
	FieldName string `json:"FieldName"`
	Type      string `json:"Type"`
}

// CreateEncryptedIndex creates an encrypted index on a collection field.
func (n *Node) CreateEncryptedIndex(identityDID string, collectionName string, fieldName string) (*EncryptedIndexDescription, error) {
	var a cargs
	defer a.free()

	result := C.add_encrypted_index(n.ptr, a.opt(identityDID), a.s(collectionName), a.s(fieldName))

	value, err := unwrap(result.status, result.error, result.value, "create_encrypted_index")
	if err != nil {
		return nil, err
	}

	var encIdx EncryptedIndexDescription
	if err := json.Unmarshal([]byte(value), &encIdx); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse encrypted index: %w", err)
	}

	return &encIdx, nil
}

// DeleteEncryptedIndex deletes an encrypted index from a collection.
func (n *Node) DeleteEncryptedIndex(identityDID string, collectionName string, fieldName string) error {
	var a cargs
	defer a.free()

	result := C.delete_encrypted_index(n.ptr, a.opt(identityDID), a.s(collectionName), a.s(fieldName))

	_, err := unwrap(result.status, result.error, result.value, "delete_encrypted_index")
	return err
}

// ListEncryptedIndexes returns all encrypted indexes for a collection.
func (n *Node) ListEncryptedIndexes(identityDID string, collectionName string) ([]EncryptedIndexDescription, error) {
	var a cargs
	defer a.free()

	result := C.list_encrypted_indexes(n.ptr, a.opt(identityDID), a.s(collectionName))

	value, err := unwrap(result.status, result.error, result.value, "list_encrypted_indexes")
	if err != nil {
		return nil, err
	}

	var indexes []EncryptedIndexDescription
	if err := json.Unmarshal([]byte(value), &indexes); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse encrypted indexes: %w", err)
	}

	return indexes, nil
}

// ListAllEncryptedIndexes returns all encrypted indexes across all collections.
func (n *Node) ListAllEncryptedIndexes(identityDID string) (map[string][]EncryptedIndexDescription, error) {
	var a cargs
	defer a.free()

	result := C.list_all_encrypted_indexes(n.ptr, a.opt(identityDID))

	value, err := unwrap(result.status, result.error, result.value, "list_all_encrypted_indexes")
	if err != nil {
		return nil, err
	}

	var indexes map[string][]EncryptedIndexDescription
	if err := json.Unmarshal([]byte(value), &indexes); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse encrypted indexes: %w", err)
	}

	return indexes, nil
}

// ============================================================================
// NAC (Node Access Control) Functions
// ============================================================================

// NACStatus represents the NAC status response.
type NACStatus struct {
	Status            string  `json:"status"`
	ConfiguredEnabled bool    `json:"configured_enabled"`
	DevMode           bool    `json:"dev_mode"`
	Owner             *string `json:"owner,omitempty"`
}

// GetNACStatus returns the current NAC status.
func (n *Node) GetNACStatus(identityDID string) (*NACStatus, error) {
	var a cargs
	defer a.free()

	result := C.get_nac_status(n.ptr, a.opt(identityDID))

	value, err := unwrap(result.status, result.error, result.value, "get_nac_status")
	if err != nil {
		return nil, err
	}

	var status NACStatus
	if err := json.Unmarshal([]byte(value), &status); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse NAC status: %w", err)
	}

	return &status, nil
}

// EnableNAC enables NAC with the given owner DID.
func (n *Node) EnableNAC(ownerDID string) error {
	var a cargs
	defer a.free()

	result := C.enable_nac(n.ptr, a.s(ownerDID))

	_, err := unwrap(result.status, result.error, nil, "enable_nac")
	return err
}

// DisableNAC temporarily disables NAC.
// The requestor must be an admin.
func (n *Node) DisableNAC(requestorDID string) error {
	var a cargs
	defer a.free()

	result := C.disable_nac(n.ptr, a.s(requestorDID))

	_, err := unwrap(result.status, result.error, nil, "disable_nac")
	return err
}

// ReEnableNAC re-enables NAC after temporary disable.
// The requestor must be an admin.
func (n *Node) ReEnableNAC(requestorDID string) error {
	var a cargs
	defer a.free()

	result := C.re_enable_nac(n.ptr, a.s(requestorDID))

	_, err := unwrap(result.status, result.error, nil, "re_enable_nac")
	return err
}

// AddNACActorRelationship adds a NAC relationship for the given relation.
// Returns true if added, false if already exists.
func (n *Node) AddNACActorRelationship(requestorDID, relation, targetDID string) (bool, error) {
	var a cargs
	defer a.free()

	result := C.add_nac_actor_relationship(n.ptr, a.s(requestorDID), a.s(relation), a.s(targetDID))

	value, err := unwrap(result.status, result.error, result.value, "add_nac_actor_relationship")
	if err != nil {
		return false, err
	}

	var response struct {
		Added bool `json:"added"`
	}
	if err := json.Unmarshal([]byte(value), &response); err != nil {
		return false, fmt.Errorf("ffi: failed to parse response: %w", err)
	}

	return response.Added, nil
}

// DeleteNACActorRelationship removes a NAC relationship for the given relation.
// Returns true if deleted, false if didn't exist.
func (n *Node) DeleteNACActorRelationship(requestorDID, relation, targetDID string) (bool, error) {
	var a cargs
	defer a.free()

	result := C.delete_nac_actor_relationship(n.ptr, a.s(requestorDID), a.s(relation), a.s(targetDID))

	value, err := unwrap(result.status, result.error, result.value, "delete_nac_actor_relationship")
	if err != nil {
		return false, err
	}

	var response struct {
		Deleted bool `json:"deleted"`
	}
	if err := json.Unmarshal([]byte(value), &response); err != nil {
		return false, fmt.Errorf("ffi: failed to parse response: %w", err)
	}

	return response.Deleted, nil
}

// ============================================================================
// DAC (Document Access Control) Functions
// ============================================================================

// AddDACPolicy registers a DAC policy and returns its content-addressed ID.
func (n *Node) AddDACPolicy(identityDID, policy string) (string, error) {
	var a cargs
	defer a.free()

	result := C.add_dac_policy(n.ptr, a.s(identityDID), a.s(policy))

	value, err := unwrap(result.status, result.error, result.value, "add_dac_policy")
	if err != nil {
		return "", err
	}

	var response struct {
		PolicyID string `json:"PolicyID"`
	}
	if err := json.Unmarshal([]byte(value), &response); err != nil {
		return "", fmt.Errorf("ffi: failed to parse response: %w", err)
	}

	return response.PolicyID, nil
}

// AddDACActorRelationship shares document access with target.
// Relation can be "reader", "updater", or "deleter".
// Returns true if added, false if already exists.
func (n *Node) AddDACActorRelationship(requestorDID, targetDID, collectionID, docID, relation string) (bool, error) {
	var a cargs
	defer a.free()

	result := C.add_dac_actor_relationship(n.ptr, a.s(requestorDID), a.s(targetDID), a.s(collectionID), a.s(docID), a.s(relation))

	value, err := unwrap(result.status, result.error, result.value, "add_dac_actor_relationship")
	if err != nil {
		return false, err
	}

	var response struct {
		Added bool `json:"added"`
	}
	if err := json.Unmarshal([]byte(value), &response); err != nil {
		return false, fmt.Errorf("ffi: failed to parse response: %w", err)
	}

	return response.Added, nil
}

// DeleteDACActorRelationship revokes document access from target.
// Returns true if deleted, false if didn't exist.
func (n *Node) DeleteDACActorRelationship(requestorDID, targetDID, collectionID, docID, relation string) (bool, error) {
	var a cargs
	defer a.free()

	result := C.delete_dac_actor_relationship(n.ptr, a.s(requestorDID), a.s(targetDID), a.s(collectionID), a.s(docID), a.s(relation))

	value, err := unwrap(result.status, result.error, result.value, "delete_dac_actor_relationship")
	if err != nil {
		return false, err
	}

	var response struct {
		Deleted bool `json:"deleted"`
	}
	if err := json.Unmarshal([]byte(value), &response); err != nil {
		return false, fmt.Errorf("ffi: failed to parse response: %w", err)
	}

	return response.Deleted, nil
}

// ============================================================================
// Identity Functions
// ============================================================================

// GetNodeIdentity returns the node's DID if configured.
func (n *Node) GetNodeIdentity() (string, error) {
	result := C.get_node_identity(n.ptr)

	value, err := unwrap(result.status, result.error, result.value, "get_node_identity")
	if err != nil {
		return "", err
	}

	var response struct {
		DID string `json:"did"`
	}
	if err := json.Unmarshal([]byte(value), &response); err != nil {
		return "", fmt.Errorf("ffi: failed to parse response: %w", err)
	}

	return response.DID, nil
}

// SetDefaultIdentity sets the node's default signing identity DID.
// The DID must already be registered via RegisterIdentityWithRust.
func (n *Node) SetDefaultIdentity(did string) error {
	var a cargs
	defer a.free()

	result := C.node_set_default_identity(n.ptr, a.s(did))
	_, err := unwrap(result.status, result.error, result.value, "node_set_default_identity")
	return err
}

// ============================================================================
// Block Functions
// ============================================================================

// BlockVerifySignature verifies the signature of a block identified by CID.
func (n *Node) BlockVerifySignature(keyType, publicKey, blockCid, identityDID string) error {
	var a cargs
	defer a.free()

	result := C.block_verify_signature(n.ptr, a.s(keyType), a.s(publicKey), a.s(blockCid), a.opt(identityDID))

	_, err := unwrap(result.status, result.error, result.value, "block_verify_signature")
	return err
}

// BlockVerifySignatureInTxn verifies the signature of a block identified by CID
// using the transaction's blockstore view so uncommitted blocks are visible.
func (n *Node) BlockVerifySignatureInTxn(
	txnID string,
	keyType string,
	publicKey string,
	blockCid string,
	identityDID string,
) error {
	var a cargs
	defer a.free()

	result := C.block_verify_signature_in_txn(
		n.ptr,
		a.s(txnID),
		a.s(keyType),
		a.s(publicKey),
		a.s(blockCid),
		a.opt(identityDID),
	)

	_, err := unwrap(result.status, result.error, result.value, "block_verify_signature_in_txn")
	return err
}

// ============================================================================
// Subscription Functions
// ============================================================================

// Subscription represents an active subscription to database events.
type Subscription struct {
	node   *Node
	handle C.uintptr_t
}

// SubscriptionEvent represents an event received from a subscription.
type SubscriptionEvent struct {
	Type         string `json:"type"`
	DocID        string `json:"doc_id,omitempty"`
	CID          string `json:"cid,omitempty"`
	CollectionID string `json:"collection_id,omitempty"`
	Block        string `json:"block,omitempty"`
	ByPeer       string `json:"by_peer,omitempty"`
	IsRetry      bool   `json:"is_retry,omitempty"`
	IsRelay      bool   `json:"is_relay,omitempty"`
	PeerID       string `json:"peer_id,omitempty"`
	Topic        string `json:"topic,omitempty"`
	EventType    string `json:"event_type,omitempty"`
}

// PollResult represents the result of polling a subscription.
type PollResult struct {
	// HasEvent is true if an event was received.
	HasEvent bool
	// Event contains the event data when HasEvent is true.
	Event *SubscriptionEvent
	// DroppedCount indicates how many events were dropped due to buffer overflow.
	DroppedCount uint64
	// IsClosed is true if the subscription has been closed.
	IsClosed bool
}

// ErrSubscriptionClosed is returned when polling a closed subscription.
var ErrSubscriptionClosed = errors.New("ffi: subscription closed")

// ErrNoEvent is returned when polling returns no event.
var ErrNoEvent = errors.New("ffi: no event available")

// Subscribe creates a subscription to database events.
// The optional collectionFilter limits events to a specific collection.
// The subscription must be closed with Close() when done.
func (n *Node) Subscribe(collectionFilter string) (*Subscription, error) {
	var a cargs
	defer a.free()

	result := C.create_subscription(n.ptr, a.opt(collectionFilter))

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("create_subscription", err)
	}

	return &Subscription{
		node:   n,
		handle: result.subscription_handle,
	}, nil
}

// Poll checks for the next event without blocking.
// Returns a PollResult indicating the status and any event data.
func (s *Subscription) Poll() (*PollResult, error) {
	result := C.poll_subscription(s.handle)

	switch result.status {
	case 0: // Event available
		value := C.GoString(result.value)
		C.defra_free_string(result.value)

		var event SubscriptionEvent
		if err := json.Unmarshal([]byte(value), &event); err != nil {
			return nil, fmt.Errorf("ffi: failed to parse event: %w", err)
		}

		return &PollResult{
			HasEvent:     true,
			Event:        &event,
			DroppedCount: uint64(result.dropped_count),
		}, nil

	case 1: // Error
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("poll_subscription", err)

	case 2: // No event available
		return &PollResult{
			HasEvent:     false,
			DroppedCount: uint64(result.dropped_count),
		}, nil

	case 3: // Subscription closed
		return &PollResult{
			HasEvent: false,
			IsClosed: true,
		}, nil

	default:
		return nil, fmt.Errorf("ffi: unknown poll status: %d", result.status)
	}
}

// Close closes the subscription and releases resources.
// After calling Close, the subscription handle is no longer valid.
func (s *Subscription) Close() error {
	result := C.close_subscription(s.handle)

	_, err := unwrap(result.status, result.error, nil, "close_subscription")
	return err
}

// SubscribeMergeComplete creates a subscription to P2P merge complete events.
// The subscription must be closed with Close() when done.
func (n *Node) SubscribeMergeComplete() (*Subscription, error) {
	result := C.create_merge_complete_subscription(n.ptr)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("create_merge_complete_subscription", err)
	}

	return &Subscription{
		node:   n,
		handle: result.subscription_handle,
	}, nil
}

// ============================================================================
// P2P Functions
// ============================================================================

// NewNodeWithP2P creates a new DefraDB node with P2P enabled.
// The node must be closed with Close() when done.
func NewNodeWithP2P(opts NodeOptions, listenAddr string) (*Node, error) {
	var a cargs
	defer a.free()

	var cOpts C.struct_NodeInitOptions

	if opts.DBPath != "" {
		cOpts.db_path = a.s(opts.DBPath)
	}

	if opts.InMemory || opts.DBPath == "" {
		cOpts.in_memory = 1
	} else {
		cOpts.in_memory = 0
	}

	if opts.EnableSigning {
		cOpts.enable_signing = 1
	}

	// Pass signing key if provided
	if len(opts.SigningPrivateKey) > 0 {
		cOpts.signing_private_key = (*C.uint8_t)(unsafe.Pointer(&opts.SigningPrivateKey[0]))
		cOpts.signing_private_key_len = C.uintptr_t(len(opts.SigningPrivateKey))

		keyType := opts.SigningKeyType
		if keyType == "" {
			keyType = "secp256k1"
		}
		cOpts.signing_key_type = a.s(keyType)
	}

	// Pass SourceHub config if provided
	if opts.SourceHubGRPCAddress != "" {
		cOpts.sourcehub_grpc_address = a.s(opts.SourceHubGRPCAddress)
		cOpts.sourcehub_comet_rpc_address = a.s(opts.SourceHubCometRPCAddress)
		cOpts.sourcehub_chain_id = a.s(opts.SourceHubChainID)

		if len(opts.SourceHubSignerKey) > 0 {
			cOpts.sourcehub_signer_key = (*C.uint8_t)(unsafe.Pointer(&opts.SourceHubSignerKey[0]))
			cOpts.sourcehub_signer_key_len = C.uintptr_t(len(opts.SourceHubSignerKey))
		}
	}

	result := C.new_node_with_p2p(cOpts, a.s(listenAddr))

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("new_node_with_p2p", err)
	}

	return &Node{ptr: result.node_ptr}, nil
}

// P2PPeerInfo returns the local peer info as a list of multiaddrs with peer ID.
func (n *Node) P2PPeerInfo(identityDID string) ([]string, error) {
	var a cargs
	defer a.free()

	result := C.p2p_peer_info(n.ptr, a.opt(identityDID))

	value, err := unwrap(result.status, result.error, result.value, "p2p_peer_info")
	if err != nil {
		return nil, err
	}

	var addrs []string
	if err := json.Unmarshal([]byte(value), &addrs); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse peer info: %w", err)
	}

	return addrs, nil
}

// P2PActivePeers returns the list of connected peer IDs.
func (n *Node) P2PActivePeers(identityDID string) ([]string, error) {
	var a cargs
	defer a.free()

	result := C.p2p_active_peers(n.ptr, a.opt(identityDID))

	value, err := unwrap(result.status, result.error, result.value, "p2p_active_peers")
	if err != nil {
		return nil, err
	}

	var peers []string
	if err := json.Unmarshal([]byte(value), &peers); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse active peers: %w", err)
	}

	return peers, nil
}

// P2PConnect connects to a peer at the given multiaddr.
func (n *Node) P2PConnect(identityDID string, addr string) error {
	var a cargs
	defer a.free()

	result := C.p2p_connect(n.ptr, a.opt(identityDID), a.s(addr))

	_, err := unwrap(result.status, result.error, result.value, "p2p_connect")
	return err
}

// P2PDisconnect disconnects from a peer at the given multiaddr.
func (n *Node) P2PDisconnect(identityDID string, addr string) error {
	var a cargs
	defer a.free()

	result := C.p2p_disconnect(n.ptr, a.opt(identityDID), a.s(addr))

	_, err := unwrap(result.status, result.error, result.value, "p2p_disconnect")
	return err
}

// P2PSetReplicator sets a replicator for the given collections.
func (n *Node) P2PSetReplicator(identityDID string, peerAddr string, collections []string) error {
	var a cargs
	defer a.free()

	if collections == nil {
		collections = []string{}
	}
	collectionsJSON, err := json.Marshal(collections)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal collections: %w", err)
	}

	result := C.p2p_add_replicator(n.ptr, a.opt(identityDID), a.s(peerAddr), a.s(string(collectionsJSON)))

	_, err = unwrap(result.status, result.error, result.value, "p2p_create_replicator")
	return err
}

// P2PRetryReplicators re-pushes all existing documents to connected replicator peers.
// Call this after reconnecting peers following a node restart.
func (n *Node) P2PRetryReplicators() error {
	result := C.p2p_retry_replicators(n.ptr)

	_, err := unwrap(result.status, result.error, result.value, "p2p_retry_replicators")
	return err
}

// SetSEEncryptionKey sets the searchable encryption key for the node.
// The key must be exactly 32 bytes (AES-256).
func (n *Node) SetSEEncryptionKey(key []byte) error {
	if len(key) == 0 {
		return fmt.Errorf("ffi: SE encryption key is empty")
	}

	result := C.set_se_encryption_key(
		n.ptr,
		(*C.uint8_t)(unsafe.Pointer(&key[0])),
		C.uintptr_t(len(key)),
	)

	_, err := unwrap(result.status, result.error, nil, "set_se_encryption_key")
	return err
}

// P2PDeleteReplicator removes a replicator by peer ID.
// If collections is non-empty, only those collections are removed from the replicator.
// If collections is empty, the entire replicator is deleted.
func (n *Node) P2PDeleteReplicator(identityDID string, peerID string, collections []string) error {
	var a cargs
	defer a.free()

	collectionsJSON, err := json.Marshal(collections)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal collections: %w", err)
	}

	result := C.p2p_delete_replicator(n.ptr, a.opt(identityDID), a.s(peerID), a.s(string(collectionsJSON)))

	_, err = unwrap(result.status, result.error, result.value, "p2p_delete_replicator")
	return err
}

// ReplicatorInfo represents replicator information returned from FFI.
type ReplicatorInfo struct {
	PeerID      string   `json:"peer_id"`
	Addresses   []string `json:"addresses"`
	Collections []string `json:"collections"`
}

// P2PGetAllReplicators returns all replicators.
func (n *Node) P2PGetAllReplicators(identityDID string) ([]ReplicatorInfo, error) {
	var a cargs
	defer a.free()

	result := C.p2p_list_replicators(n.ptr, a.opt(identityDID))

	value, err := unwrap(result.status, result.error, result.value, "p2p_list_replicators")
	if err != nil {
		return nil, err
	}

	var replicators []ReplicatorInfo
	if err := json.Unmarshal([]byte(value), &replicators); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse replicators: %w", err)
	}

	return replicators, nil
}

// P2PAddCollections adds collections to P2P replication.
func (n *Node) P2PAddCollections(identityDID string, collections []string) error {
	var a cargs
	defer a.free()

	collectionsJSON, err := json.Marshal(collections)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal collections: %w", err)
	}

	result := C.p2p_add_collections(n.ptr, a.opt(identityDID), a.s(string(collectionsJSON)))

	_, err = unwrap(result.status, result.error, result.value, "p2p_create_collections")
	return err
}

// P2PRemoveCollections removes collections from P2P replication.
func (n *Node) P2PRemoveCollections(identityDID string, collections []string) error {
	var a cargs
	defer a.free()

	collectionsJSON, err := json.Marshal(collections)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal collections: %w", err)
	}

	result := C.p2p_delete_collections(n.ptr, a.opt(identityDID), a.s(string(collectionsJSON)))

	_, err = unwrap(result.status, result.error, result.value, "p2p_delete_collections")
	return err
}

// P2PGetAllCollections returns all collections configured for P2P replication.
func (n *Node) P2PGetAllCollections(identityDID string) ([]string, error) {
	var a cargs
	defer a.free()

	result := C.p2p_list_collections(n.ptr, a.opt(identityDID))

	value, err := unwrap(result.status, result.error, result.value, "p2p_list_collections")
	if err != nil {
		return nil, err
	}

	var collections []string
	if err := json.Unmarshal([]byte(value), &collections); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse collections: %w", err)
	}

	return collections, nil
}

// P2PAddDocuments adds documents to P2P replication by subscribing to their GossipSub topics.
func (n *Node) P2PAddDocuments(identityDID string, docIDs []string) error {
	var a cargs
	defer a.free()

	docIDsJSON, err := json.Marshal(docIDs)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal doc IDs: %w", err)
	}

	result := C.p2p_add_documents(n.ptr, a.opt(identityDID), a.s(string(docIDsJSON)))

	_, err = unwrap(result.status, result.error, result.value, "p2p_create_documents")
	return err
}

// P2PRemoveDocuments removes documents from P2P replication.
func (n *Node) P2PRemoveDocuments(identityDID string, docIDs []string) error {
	var a cargs
	defer a.free()

	docIDsJSON, err := json.Marshal(docIDs)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal doc IDs: %w", err)
	}

	result := C.p2p_delete_documents(n.ptr, a.opt(identityDID), a.s(string(docIDsJSON)))

	_, err = unwrap(result.status, result.error, result.value, "p2p_delete_documents")
	return err
}

// P2PGetAllDocuments returns all documents configured for P2P replication.
func (n *Node) P2PGetAllDocuments(identityDID string) ([]string, error) {
	var a cargs
	defer a.free()

	result := C.p2p_list_documents(n.ptr, a.opt(identityDID))

	value, err := unwrap(result.status, result.error, result.value, "p2p_list_documents")
	if err != nil {
		return nil, err
	}

	var docIDs []string
	if err := json.Unmarshal([]byte(value), &docIDs); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse documents: %w", err)
	}

	return docIDs, nil
}

// P2PSyncDocuments syncs specific documents from peers.
// This implements the DocSync pull-based protocol.
func (n *Node) P2PSyncDocuments(identityDID string, collectionName string, docIDs []string) error {
	var a cargs
	defer a.free()

	docIDsJSON, err := json.Marshal(docIDs)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal doc IDs: %w", err)
	}

	result := C.p2p_sync_documents(n.ptr, a.opt(identityDID), a.s(collectionName), a.s(string(docIDsJSON)))

	_, err = unwrap(result.status, result.error, result.value, "p2p_sync_documents")
	return err
}

// P2PSyncBranchableCollection syncs a branchable collection from peers.
func (n *Node) P2PSyncBranchableCollection(identityDID string, collectionID string) error {
	var a cargs
	defer a.free()

	result := C.p2p_sync_branchable_collection(n.ptr, a.opt(identityDID), a.s(collectionID))

	_, err := unwrap(result.status, result.error, result.value, "p2p_sync_branchable_collection")
	return err
}

// P2PSyncCollectionVersions syncs collection versions by their CIDs.
func (n *Node) P2PSyncCollectionVersions(identityDID string, versionIDs []string) error {
	var a cargs
	defer a.free()

	versionIDsJSON, err := json.Marshal(versionIDs)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal version IDs: %w", err)
	}

	result := C.p2p_sync_collection_versions(n.ptr, a.opt(identityDID), a.s(string(versionIDsJSON)))

	_, err = unwrap(result.status, result.error, result.value, "p2p_sync_collection_versions")
	return err
}

// BasicExportDB exports the database to a JSON file.
func (n *Node) BasicExportDB(configJSON string) error {
	var a cargs
	defer a.free()

	result := C.basic_export(n.ptr, a.s(configJSON))
	_, err := unwrap(result.status, result.error, result.value, "basic_export")
	return err
}

// BasicImportDB imports documents from a JSON backup file.
func (n *Node) BasicImportDB(filepath string) error {
	var a cargs
	defer a.free()

	result := C.basic_import(n.ptr, a.s(filepath))
	_, err := unwrap(result.status, result.error, result.value, "basic_import")
	return err
}
