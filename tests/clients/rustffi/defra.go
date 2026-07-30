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
	var cOpts C.struct_NodeInitOptions

	if opts.DBPath != "" {
		cDBPath := C.CString(opts.DBPath)
		defer C.free(unsafe.Pointer(cDBPath))
		cOpts.db_path = cDBPath
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
		cKeyType := C.CString(keyType)
		defer C.free(unsafe.Pointer(cKeyType))
		cOpts.signing_key_type = cKeyType
	}

	// Pass SourceHub config if provided
	if opts.SourceHubGRPCAddress != "" {
		cGRPC := C.CString(opts.SourceHubGRPCAddress)
		defer C.free(unsafe.Pointer(cGRPC))
		cOpts.sourcehub_grpc_address = cGRPC

		cComet := C.CString(opts.SourceHubCometRPCAddress)
		defer C.free(unsafe.Pointer(cComet))
		cOpts.sourcehub_comet_rpc_address = cComet

		cChainID := C.CString(opts.SourceHubChainID)
		defer C.free(unsafe.Pointer(cChainID))
		cOpts.sourcehub_chain_id = cChainID

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

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("node_close", err)
	}

	return nil
}

// AddSchema adds a GraphQL SDL schema to the database.
// Returns the JSON response containing created collection versions.
func (n *Node) AddSchema(identityDID string, sdl string) (string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cSDL := C.CString(sdl)
	defer C.free(unsafe.Pointer(cSDL))

	result := C.add_schema(n.ptr, cIdentityDID, cSDL)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("add_schema", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// AddSchemaInTxn adds a GraphQL SDL schema within a specific transaction.
// Returns the JSON response containing created collection versions visible in that transaction.
func (n *Node) AddSchemaInTxn(txnID string, identityDID string, sdl string) (string, error) {
	cTxnID := C.CString(txnID)
	defer C.free(unsafe.Pointer(cTxnID))

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cSDL := C.CString(sdl)
	defer C.free(unsafe.Pointer(cSDL))

	result := C.add_schema_in_txn(n.ptr, cTxnID, cIdentityDID, cSDL)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("add_schema_in_txn", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// GetCollections returns all collections in the database as JSON.
func (n *Node) GetCollections(identityDID string) (string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.get_collections(n.ptr, cIdentityDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("get_collections", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// GetCollectionsInTxn returns all collection versions visible within a specific transaction.
// This reads from the transaction's systemstore, which includes uncommitted writes
// (e.g., placeholders from set_migration_in_txn).
func (n *Node) GetCollectionsInTxn(txnID string, identityDID string) (string, error) {
	cTxnID := C.CString(txnID)
	defer C.free(unsafe.Pointer(cTxnID))

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.get_collections_in_txn(n.ptr, cTxnID, cIdentityDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("get_collections_in_txn", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
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
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cQuery := C.CString(query)
	defer C.free(unsafe.Pointer(cQuery))

	var cOpName *C.char
	if operationName != "" {
		cOpName = C.CString(operationName)
		defer C.free(unsafe.Pointer(cOpName))
	}

	var cVars *C.char
	if variables != "" {
		cVars = C.CString(variables)
		defer C.free(unsafe.Pointer(cVars))
	}

	cBatchSessionID := C.CString("")
	defer C.free(unsafe.Pointer(cBatchSessionID))

	result := C.exec_request(n.ptr, cIdentityDID, cQuery, cOpName, cVars, cBatchSessionID)

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
	cSubID := C.CString(subscriptionID)
	defer C.free(unsafe.Pointer(cSubID))

	result := C.poll_graphql_subscription(cSubID)

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
	cSubID := C.CString(subscriptionID)
	defer C.free(unsafe.Pointer(cSubID))

	result := C.close_graphql_subscription(cSubID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("close_graphql_subscription", err)
	}

	return nil
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

// Mutate executes a GraphQL mutation and returns a parsed result.
func (n *Node) Mutate(mutation string) (*QueryResult, error) {
	return n.Query(mutation)
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
	cTxnID := C.CString(t.id)
	defer C.free(unsafe.Pointer(cTxnID))

	result := C.commit_txn(t.node.ptr, cTxnID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("commit_txn", err)
	}

	return nil
}

// Rollback discards all changes made in the transaction.
// After rollback, the transaction is no longer valid.
func (t *Transaction) Rollback() error {
	cTxnID := C.CString(t.id)
	defer C.free(unsafe.Pointer(cTxnID))

	result := C.rollback_txn(t.node.ptr, cTxnID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("rollback_txn", err)
	}

	return nil
}

// ExecRequest executes a GraphQL query or mutation within the transaction.
// identityDID is the DID of the caller for ACP permission checks (empty string for anonymous).
func (t *Transaction) ExecRequest(identityDID string, query string, operationName string, variables string) (string, error) {
	cTxnID := C.CString(t.id)
	defer C.free(unsafe.Pointer(cTxnID))

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cQuery := C.CString(query)
	defer C.free(unsafe.Pointer(cQuery))

	var cOpName *C.char
	if operationName != "" {
		cOpName = C.CString(operationName)
		defer C.free(unsafe.Pointer(cOpName))
	}

	var cVars *C.char
	if variables != "" {
		cVars = C.CString(variables)
		defer C.free(unsafe.Pointer(cVars))
	}

	cBatchSessionID := C.CString("")
	defer C.free(unsafe.Pointer(cBatchSessionID))

	result := C.exec_request_in_txn(t.node.ptr, cTxnID, cIdentityDID, cQuery, cOpName, cVars, cBatchSessionID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("exec_request_in_txn", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
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
	cTxnID := C.CString(t.id)
	defer C.free(unsafe.Pointer(cTxnID))

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	targetsJSON, err := json.Marshal(targets)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal collection targets: %w", err)
	}
	cTargets := C.CString(string(targetsJSON))
	defer C.free(unsafe.Pointer(cTargets))

	result := C.delete_collections_in_txn(
		t.node.ptr,
		cTxnID,
		cIdentityDID,
		cTargets,
		C.bool(activeOnly),
	)
	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("delete_collections_in_txn", err)
	}

	C.defra_free_string(result.value)
	return nil
}

// SetCollectionActive updates a collection version within the transaction.
func (t *Transaction) SetCollectionActive(identityDID string, versionID string, isActive bool) error {
	cTxnID := C.CString(t.id)
	defer C.free(unsafe.Pointer(cTxnID))

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cVersionID := C.CString(versionID)
	defer C.free(unsafe.Pointer(cVersionID))

	result := C.set_collection_active_in_txn(
		t.node.ptr,
		cTxnID,
		cIdentityDID,
		cVersionID,
		C.bool(isActive),
	)
	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("set_collection_active_in_txn", err)
	}

	C.defra_free_string(result.value)
	return nil
}

// ============================================================================
// Collection Functions
// ============================================================================

// GetCollectionByName returns a collection by its name.
// Returns the collection's schema as JSON if found.
func (n *Node) GetCollectionByName(identityDID string, name string) (string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	result := C.get_collection_by_name(n.ptr, cIdentityDID, cName)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("get_collection_by_name", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// HasCollection checks if a collection exists by name.
func (n *Node) HasCollection(identityDID string, name string) (bool, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	result := C.has_collection(n.ptr, cIdentityDID, cName)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return false, mapFFIError("has_collection", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value == "true", nil
}

// DeleteCollection deletes a collection and all its documents.
func (n *Node) DeleteCollection(identityDID string, name string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	result := C.delete_collection(n.ptr, cIdentityDID, cName)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("delete_collection", err)
	}

	C.defra_free_string(result.value)
	return nil
}

// DeleteCollectionVersions deletes multiple collection versions by their version IDs.
// Versions are deleted in topological order (children before parents).
func (n *Node) DeleteCollectionVersions(identityDID string, versionIDs []string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	idsJSON, err := json.Marshal(versionIDs)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal version IDs: %w", err)
	}

	cIDs := C.CString(string(idsJSON))
	defer C.free(unsafe.Pointer(cIDs))

	result := C.delete_collection_versions(n.ptr, cIdentityDID, cIDs)

	if result.status != 0 {
		errStr := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("delete_collection_versions", errStr)
	}

	C.defra_free_string(result.value)
	return nil
}

// TruncateCollection deletes all documents from a collection while preserving the schema.
func (n *Node) TruncateCollection(identityDID string, name string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	result := C.truncate_collection(n.ptr, cIdentityDID, cName)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("truncate_collection", err)
	}

	C.defra_free_string(result.value)
	return nil
}

// FindCollectionByID finds a collection by its collection ID (schema version ID).
// Returns the collection's schema as JSON if found, or "null" if not found.
func (n *Node) FindCollectionByID(identityDID string, collectionID string) (string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cID := C.CString(collectionID)
	defer C.free(unsafe.Pointer(cID))

	result := C.find_collection_by_id(n.ptr, cIdentityDID, cID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("find_collection_by_id", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// SetActiveCollectionVersion activates the collection with the given version ID.
func (n *Node) SetActiveCollectionVersion(identityDID string, versionID string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cVersionID := C.CString(versionID)
	defer C.free(unsafe.Pointer(cVersionID))

	result := C.set_active_collection_version(n.ptr, cIdentityDID, cVersionID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("set_active_collection_version", err)
	}

	C.defra_free_string(result.value)
	return nil
}

// PatchCollection applies a JSON patch to a collection's schema.
// Returns the updated collection schema as JSON.
func (n *Node) PatchCollection(identityDID string, collectionName string, patch string) (string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cName := C.CString(collectionName)
	defer C.free(unsafe.Pointer(cName))

	cPatch := C.CString(patch)
	defer C.free(unsafe.Pointer(cPatch))

	result := C.patch_collection(n.ptr, cIdentityDID, cName, cPatch)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("patch_collection", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// GetCollectionByVersionID returns a collection by its version ID.
// Returns the collection's schema as JSON if found, or "null" if not found.
func (n *Node) GetCollectionByVersionID(identityDID string, versionID string) (string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cVersionID := C.CString(versionID)
	defer C.free(unsafe.Pointer(cVersionID))

	result := C.get_collection_by_version_id(n.ptr, cIdentityDID, cVersionID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("get_collection_by_version_id", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// AddView creates a new Defra View from a GQL query and SDL schema.
// The transform parameter is optional (pass empty string for none).
// Note: Not yet implemented - see issue #178.
func (n *Node) AddView(identityDID string, gqlQuery string, sdl string, transform string) (string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cQuery := C.CString(gqlQuery)
	defer C.free(unsafe.Pointer(cQuery))

	cSDL := C.CString(sdl)
	defer C.free(unsafe.Pointer(cSDL))

	var cTransform *C.char
	if transform != "" {
		cTransform = C.CString(transform)
		defer C.free(unsafe.Pointer(cTransform))
	}

	result := C.add_view(n.ptr, cIdentityDID, cQuery, cSDL, cTransform)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("add_view", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// RefreshViews refreshes the caches of all views matching the given options.
// Pass empty string for options to refresh all views.
// Note: Not yet implemented - see issue #178.
func (n *Node) RefreshViews(identityDID string, options string) error {
	var cOptions *C.char
	if options != "" {
		cOptions = C.CString(options)
		defer C.free(unsafe.Pointer(cOptions))
	}

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.refresh_views(n.ptr, cIdentityDID, cOptions)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("refresh_views", err)
	}

	C.defra_free_string(result.value)
	return nil
}

// MaterializeCollection eagerly migrates and caches every known-version
// document in a collection. It returns the number of documents advanced.
func (n *Node) MaterializeCollection(identityDID string, collectionName string) (int, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cCollectionName := C.CString(collectionName)
	defer C.free(unsafe.Pointer(cCollectionName))

	result := C.materialize_collection(n.ptr, cIdentityDID, cCollectionName)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return 0, mapFFIError("materialize_collection", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	count, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("ffi: failed to parse materialized document count %q: %w", value, err)
	}
	return count, nil
}

// SetMigration sets the migration for collection versions.
// The config parameter should be a JSON string containing LensConfig.
func (n *Node) SetMigration(identityDID string, config string) (string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cConfig := C.CString(config)
	defer C.free(unsafe.Pointer(cConfig))

	result := C.set_migration(n.ptr, cIdentityDID, cConfig)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("set_migration", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// SetMigrationInTxn sets the migration for collection versions within a transaction.
// The migration will only be visible after the transaction is committed.
func (n *Node) SetMigrationInTxn(txnID string, identityDID string, config string) (string, error) {
	cTxnID := C.CString(txnID)
	defer C.free(unsafe.Pointer(cTxnID))

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cConfig := C.CString(config)
	defer C.free(unsafe.Pointer(cConfig))

	result := C.set_migration_in_txn(n.ptr, cTxnID, cIdentityDID, cConfig)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("set_migration_in_txn", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// LensAdd adds a lens transform to the database.
// The lensJSON parameter should be a JSON string matching Go's model.Lens format.
func (n *Node) LensAdd(identityDID string, lensJSON string) (string, error) {
	cLens := C.CString(lensJSON)
	defer C.free(unsafe.Pointer(cLens))

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.lens_add(n.ptr, cIdentityDID, cLens)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("lens_add", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// LensAddInTxn adds a lens transform within a transaction.
func (n *Node) LensAddInTxn(txnID string, identityDID string, lensJSON string) (string, error) {
	cTxnID := C.CString(txnID)
	defer C.free(unsafe.Pointer(cTxnID))

	cLens := C.CString(lensJSON)
	defer C.free(unsafe.Pointer(cLens))

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.lens_add_in_txn(n.ptr, cTxnID, cIdentityDID, cLens)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("lens_add_in_txn", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// LensList returns all lens transforms as a map of ID -> LensModule JSON.
func (n *Node) LensList(identityDID string) (string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.lens_list(n.ptr, cIdentityDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("lens_list", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
}

// LensListInTxn lists all lens transforms visible within a transaction.
func (n *Node) LensListInTxn(txnID string, identityDID string) (string, error) {
	cTxnID := C.CString(txnID)
	defer C.free(unsafe.Pointer(cTxnID))

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.lens_list_in_txn(n.ptr, cTxnID, cIdentityDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("lens_list_in_txn", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)
	return value, nil
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
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cCollName := C.CString(collectionName)
	defer C.free(unsafe.Pointer(cCollName))

	indexInput := IndexDescription{
		Name:   indexName,
		Fields: fields,
		Unique: unique,
	}
	indexJSON, err := json.Marshal(indexInput)
	if err != nil {
		return nil, fmt.Errorf("ffi: failed to marshal index: %w", err)
	}

	cIndexJSON := C.CString(string(indexJSON))
	defer C.free(unsafe.Pointer(cIndexJSON))

	result := C.create_index(n.ptr, cIdentityDID, cCollName, cIndexJSON)

	if result.status != 0 {
		errMsg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("create_index", errMsg)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

	var index IndexDescription
	if err := json.Unmarshal([]byte(value), &index); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse index: %w", err)
	}

	return &index, nil
}

// DropIndex drops an index from a collection.
func (n *Node) DropIndex(identityDID string, collectionName string, indexName string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cCollName := C.CString(collectionName)
	defer C.free(unsafe.Pointer(cCollName))

	cIndexName := C.CString(indexName)
	defer C.free(unsafe.Pointer(cIndexName))

	result := C.delete_index(n.ptr, cIdentityDID, cCollName, cIndexName)

	if result.status != 0 {
		errMsg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("drop_index", errMsg)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// GetIndexes returns all indexes for a collection.
func (n *Node) GetIndexes(identityDID string, collectionName string) ([]IndexResult, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cCollName := C.CString(collectionName)
	defer C.free(unsafe.Pointer(cCollName))

	result := C.get_indexes(n.ptr, cIdentityDID, cCollName)

	if result.status != 0 {
		errMsg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("get_indexes", errMsg)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

	var indexes []IndexResult
	if err := json.Unmarshal([]byte(value), &indexes); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse indexes: %w", err)
	}

	return indexes, nil
}

// GetAllIndexes returns all indexes across all collections.
func (n *Node) GetAllIndexes(identityDID string) (map[string][]IndexResult, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.list_all_indexes(n.ptr, cIdentityDID)

	if result.status != 0 {
		errMsg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("get_all_indexes", errMsg)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

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
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cCollName := C.CString(collectionName)
	defer C.free(unsafe.Pointer(cCollName))

	cFieldName := C.CString(fieldName)
	defer C.free(unsafe.Pointer(cFieldName))

	result := C.add_encrypted_index(n.ptr, cIdentityDID, cCollName, cFieldName)

	if result.status != 0 {
		errMsg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("create_encrypted_index", errMsg)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

	var encIdx EncryptedIndexDescription
	if err := json.Unmarshal([]byte(value), &encIdx); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse encrypted index: %w", err)
	}

	return &encIdx, nil
}

// DeleteEncryptedIndex deletes an encrypted index from a collection.
func (n *Node) DeleteEncryptedIndex(identityDID string, collectionName string, fieldName string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cCollName := C.CString(collectionName)
	defer C.free(unsafe.Pointer(cCollName))

	cFieldName := C.CString(fieldName)
	defer C.free(unsafe.Pointer(cFieldName))

	result := C.delete_encrypted_index(n.ptr, cIdentityDID, cCollName, cFieldName)

	if result.status != 0 {
		errMsg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("delete_encrypted_index", errMsg)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// ListEncryptedIndexes returns all encrypted indexes for a collection.
func (n *Node) ListEncryptedIndexes(identityDID string, collectionName string) ([]EncryptedIndexDescription, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cCollName := C.CString(collectionName)
	defer C.free(unsafe.Pointer(cCollName))

	result := C.list_encrypted_indexes(n.ptr, cIdentityDID, cCollName)

	if result.status != 0 {
		errMsg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("list_encrypted_indexes", errMsg)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

	var indexes []EncryptedIndexDescription
	if err := json.Unmarshal([]byte(value), &indexes); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse encrypted indexes: %w", err)
	}

	return indexes, nil
}

// ListAllEncryptedIndexes returns all encrypted indexes across all collections.
func (n *Node) ListAllEncryptedIndexes(identityDID string) (map[string][]EncryptedIndexDescription, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.list_all_encrypted_indexes(n.ptr, cIdentityDID)

	if result.status != 0 {
		errMsg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("list_all_encrypted_indexes", errMsg)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

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
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.get_nac_status(n.ptr, cIdentityDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("get_nac_status", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

	var status NACStatus
	if err := json.Unmarshal([]byte(value), &status); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse NAC status: %w", err)
	}

	return &status, nil
}

// EnableNAC enables NAC with the given owner DID.
func (n *Node) EnableNAC(ownerDID string) error {
	cOwnerDID := C.CString(ownerDID)
	defer C.free(unsafe.Pointer(cOwnerDID))

	result := C.enable_nac(n.ptr, cOwnerDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("enable_nac", err)
	}

	return nil
}

// DisableNAC temporarily disables NAC.
// The requestor must be an admin.
func (n *Node) DisableNAC(requestorDID string) error {
	cRequestorDID := C.CString(requestorDID)
	defer C.free(unsafe.Pointer(cRequestorDID))

	result := C.disable_nac(n.ptr, cRequestorDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("disable_nac", err)
	}

	return nil
}

// ReEnableNAC re-enables NAC after temporary disable.
// The requestor must be an admin.
func (n *Node) ReEnableNAC(requestorDID string) error {
	cRequestorDID := C.CString(requestorDID)
	defer C.free(unsafe.Pointer(cRequestorDID))

	result := C.re_enable_nac(n.ptr, cRequestorDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("re_enable_nac", err)
	}

	return nil
}

// AddNACActorRelationship adds a NAC relationship for the given relation.
// Returns true if added, false if already exists.
func (n *Node) AddNACActorRelationship(requestorDID, relation, targetDID string) (bool, error) {
	cRequestorDID := C.CString(requestorDID)
	defer C.free(unsafe.Pointer(cRequestorDID))

	cRelation := C.CString(relation)
	defer C.free(unsafe.Pointer(cRelation))

	cTargetDID := C.CString(targetDID)
	defer C.free(unsafe.Pointer(cTargetDID))

	result := C.add_nac_actor_relationship(n.ptr, cRequestorDID, cRelation, cTargetDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return false, mapFFIError("add_nac_actor_relationship", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

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
	cRequestorDID := C.CString(requestorDID)
	defer C.free(unsafe.Pointer(cRequestorDID))

	cRelation := C.CString(relation)
	defer C.free(unsafe.Pointer(cRelation))

	cTargetDID := C.CString(targetDID)
	defer C.free(unsafe.Pointer(cTargetDID))

	result := C.delete_nac_actor_relationship(n.ptr, cRequestorDID, cRelation, cTargetDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return false, mapFFIError("delete_nac_actor_relationship", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

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
	cIdentityDID := C.CString(identityDID)
	defer C.free(unsafe.Pointer(cIdentityDID))

	cPolicy := C.CString(policy)
	defer C.free(unsafe.Pointer(cPolicy))

	result := C.add_dac_policy(n.ptr, cIdentityDID, cPolicy)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("add_dac_policy", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

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
	cRequestorDID := C.CString(requestorDID)
	defer C.free(unsafe.Pointer(cRequestorDID))

	cTargetDID := C.CString(targetDID)
	defer C.free(unsafe.Pointer(cTargetDID))

	cCollectionID := C.CString(collectionID)
	defer C.free(unsafe.Pointer(cCollectionID))

	cDocID := C.CString(docID)
	defer C.free(unsafe.Pointer(cDocID))

	cRelation := C.CString(relation)
	defer C.free(unsafe.Pointer(cRelation))

	result := C.add_dac_actor_relationship(n.ptr, cRequestorDID, cTargetDID, cCollectionID, cDocID, cRelation)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return false, mapFFIError("add_dac_actor_relationship", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

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
	cRequestorDID := C.CString(requestorDID)
	defer C.free(unsafe.Pointer(cRequestorDID))

	cTargetDID := C.CString(targetDID)
	defer C.free(unsafe.Pointer(cTargetDID))

	cCollectionID := C.CString(collectionID)
	defer C.free(unsafe.Pointer(cCollectionID))

	cDocID := C.CString(docID)
	defer C.free(unsafe.Pointer(cDocID))

	cRelation := C.CString(relation)
	defer C.free(unsafe.Pointer(cRelation))

	result := C.delete_dac_actor_relationship(n.ptr, cRequestorDID, cTargetDID, cCollectionID, cDocID, cRelation)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return false, mapFFIError("delete_dac_actor_relationship", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

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

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return "", mapFFIError("get_node_identity", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

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
	cDid := C.CString(did)
	defer C.free(unsafe.Pointer(cDid))

	result := C.node_set_default_identity(n.ptr, cDid)
	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("node_set_default_identity", err)
	}
	if result.value != nil {
		C.defra_free_string(result.value)
	}
	return nil
}

// ============================================================================
// Block Functions
// ============================================================================

// BlockVerifySignature verifies the signature of a block identified by CID.
func (n *Node) BlockVerifySignature(keyType, publicKey, blockCid, identityDID string) error {
	cKeyType := C.CString(keyType)
	defer C.free(unsafe.Pointer(cKeyType))

	cPubKey := C.CString(publicKey)
	defer C.free(unsafe.Pointer(cPubKey))

	cBlockCid := C.CString(blockCid)
	defer C.free(unsafe.Pointer(cBlockCid))

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.block_verify_signature(n.ptr, cKeyType, cPubKey, cBlockCid, cIdentityDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("block_verify_signature", err)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
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
	cTxnID := C.CString(txnID)
	defer C.free(unsafe.Pointer(cTxnID))

	cKeyType := C.CString(keyType)
	defer C.free(unsafe.Pointer(cKeyType))

	cPubKey := C.CString(publicKey)
	defer C.free(unsafe.Pointer(cPubKey))

	cBlockCid := C.CString(blockCid)
	defer C.free(unsafe.Pointer(cBlockCid))

	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.block_verify_signature_in_txn(
		n.ptr,
		cTxnID,
		cKeyType,
		cPubKey,
		cBlockCid,
		cIdentityDID,
	)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("block_verify_signature_in_txn", err)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
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
	var cFilter *C.char
	if collectionFilter != "" {
		cFilter = C.CString(collectionFilter)
		defer C.free(unsafe.Pointer(cFilter))
	}

	result := C.create_subscription(n.ptr, cFilter)

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

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("close_subscription", err)
	}

	return nil
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
	var cOpts C.struct_NodeInitOptions

	if opts.DBPath != "" {
		cDBPath := C.CString(opts.DBPath)
		defer C.free(unsafe.Pointer(cDBPath))
		cOpts.db_path = cDBPath
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
		cKeyType := C.CString(keyType)
		defer C.free(unsafe.Pointer(cKeyType))
		cOpts.signing_key_type = cKeyType
	}

	// Pass SourceHub config if provided
	if opts.SourceHubGRPCAddress != "" {
		cGRPC := C.CString(opts.SourceHubGRPCAddress)
		defer C.free(unsafe.Pointer(cGRPC))
		cOpts.sourcehub_grpc_address = cGRPC

		cComet := C.CString(opts.SourceHubCometRPCAddress)
		defer C.free(unsafe.Pointer(cComet))
		cOpts.sourcehub_comet_rpc_address = cComet

		cChainID := C.CString(opts.SourceHubChainID)
		defer C.free(unsafe.Pointer(cChainID))
		cOpts.sourcehub_chain_id = cChainID

		if len(opts.SourceHubSignerKey) > 0 {
			cOpts.sourcehub_signer_key = (*C.uint8_t)(unsafe.Pointer(&opts.SourceHubSignerKey[0]))
			cOpts.sourcehub_signer_key_len = C.uintptr_t(len(opts.SourceHubSignerKey))
		}
	}

	cListenAddr := C.CString(listenAddr)
	defer C.free(unsafe.Pointer(cListenAddr))

	result := C.new_node_with_p2p(cOpts, cListenAddr)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("new_node_with_p2p", err)
	}

	return &Node{ptr: result.node_ptr}, nil
}

// P2PPeerInfo returns the local peer info as a list of multiaddrs with peer ID.
func (n *Node) P2PPeerInfo(identityDID string) ([]string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.p2p_peer_info(n.ptr, cIdentityDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("p2p_peer_info", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

	var addrs []string
	if err := json.Unmarshal([]byte(value), &addrs); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse peer info: %w", err)
	}

	return addrs, nil
}

// P2PActivePeers returns the list of connected peer IDs.
func (n *Node) P2PActivePeers(identityDID string) ([]string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.p2p_active_peers(n.ptr, cIdentityDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("p2p_active_peers", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

	var peers []string
	if err := json.Unmarshal([]byte(value), &peers); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse active peers: %w", err)
	}

	return peers, nil
}

// P2PConnect connects to a peer at the given multiaddr.
func (n *Node) P2PConnect(identityDID string, addr string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cAddr := C.CString(addr)
	defer C.free(unsafe.Pointer(cAddr))

	result := C.p2p_connect(n.ptr, cIdentityDID, cAddr)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_connect", err)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// P2PDisconnect disconnects from a peer at the given multiaddr.
func (n *Node) P2PDisconnect(identityDID string, addr string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cAddr := C.CString(addr)
	defer C.free(unsafe.Pointer(cAddr))

	result := C.p2p_disconnect(n.ptr, cIdentityDID, cAddr)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_disconnect", err)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// P2PSetReplicator sets a replicator for the given collections.
func (n *Node) P2PSetReplicator(identityDID string, peerAddr string, collections []string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cPeerAddr := C.CString(peerAddr)
	defer C.free(unsafe.Pointer(cPeerAddr))

	if collections == nil {
		collections = []string{}
	}
	collectionsJSON, err := json.Marshal(collections)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal collections: %w", err)
	}

	cCollections := C.CString(string(collectionsJSON))
	defer C.free(unsafe.Pointer(cCollections))

	result := C.p2p_add_replicator(n.ptr, cIdentityDID, cPeerAddr, cCollections)

	if result.status != 0 {
		errStr := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_create_replicator", errStr)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// P2PRetryReplicators re-pushes all existing documents to connected replicator peers.
// Call this after reconnecting peers following a node restart.
func (n *Node) P2PRetryReplicators() error {
	result := C.p2p_retry_replicators(n.ptr)

	if result.status != 0 {
		errStr := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_retry_replicators", errStr)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
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

	if result.status != 0 {
		errStr := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("set_se_encryption_key", errStr)
	}

	return nil
}

// P2PDeleteReplicator removes a replicator by peer ID.
// If collections is non-empty, only those collections are removed from the replicator.
// If collections is empty, the entire replicator is deleted.
func (n *Node) P2PDeleteReplicator(identityDID string, peerID string, collections []string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cPeerID := C.CString(peerID)
	defer C.free(unsafe.Pointer(cPeerID))

	collectionsJSON, err := json.Marshal(collections)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal collections: %w", err)
	}
	cCollections := C.CString(string(collectionsJSON))
	defer C.free(unsafe.Pointer(cCollections))

	result := C.p2p_delete_replicator(n.ptr, cIdentityDID, cPeerID, cCollections)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_delete_replicator", err)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// ReplicatorInfo represents replicator information returned from FFI.
type ReplicatorInfo struct {
	PeerID      string   `json:"peer_id"`
	Addresses   []string `json:"addresses"`
	Collections []string `json:"collections"`
}

// P2PGetAllReplicators returns all replicators.
func (n *Node) P2PGetAllReplicators(identityDID string) ([]ReplicatorInfo, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.p2p_list_replicators(n.ptr, cIdentityDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("p2p_list_replicators", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

	var replicators []ReplicatorInfo
	if err := json.Unmarshal([]byte(value), &replicators); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse replicators: %w", err)
	}

	return replicators, nil
}

// P2PAddCollections adds collections to P2P replication.
func (n *Node) P2PAddCollections(identityDID string, collections []string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	collectionsJSON, err := json.Marshal(collections)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal collections: %w", err)
	}

	cCollections := C.CString(string(collectionsJSON))
	defer C.free(unsafe.Pointer(cCollections))

	result := C.p2p_add_collections(n.ptr, cIdentityDID, cCollections)

	if result.status != 0 {
		errStr := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_create_collections", errStr)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// P2PRemoveCollections removes collections from P2P replication.
func (n *Node) P2PRemoveCollections(identityDID string, collections []string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	collectionsJSON, err := json.Marshal(collections)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal collections: %w", err)
	}

	cCollections := C.CString(string(collectionsJSON))
	defer C.free(unsafe.Pointer(cCollections))

	result := C.p2p_delete_collections(n.ptr, cIdentityDID, cCollections)

	if result.status != 0 {
		errStr := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_delete_collections", errStr)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// P2PGetAllCollections returns all collections configured for P2P replication.
func (n *Node) P2PGetAllCollections(identityDID string) ([]string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.p2p_list_collections(n.ptr, cIdentityDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("p2p_list_collections", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

	var collections []string
	if err := json.Unmarshal([]byte(value), &collections); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse collections: %w", err)
	}

	return collections, nil
}

// P2PAddDocuments adds documents to P2P replication by subscribing to their GossipSub topics.
func (n *Node) P2PAddDocuments(identityDID string, docIDs []string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	docIDsJSON, err := json.Marshal(docIDs)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal doc IDs: %w", err)
	}

	cDocIDs := C.CString(string(docIDsJSON))
	defer C.free(unsafe.Pointer(cDocIDs))

	result := C.p2p_add_documents(n.ptr, cIdentityDID, cDocIDs)

	if result.status != 0 {
		errStr := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_create_documents", errStr)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// P2PRemoveDocuments removes documents from P2P replication.
func (n *Node) P2PRemoveDocuments(identityDID string, docIDs []string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	docIDsJSON, err := json.Marshal(docIDs)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal doc IDs: %w", err)
	}

	cDocIDs := C.CString(string(docIDsJSON))
	defer C.free(unsafe.Pointer(cDocIDs))

	result := C.p2p_delete_documents(n.ptr, cIdentityDID, cDocIDs)

	if result.status != 0 {
		errStr := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_delete_documents", errStr)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// P2PGetAllDocuments returns all documents configured for P2P replication.
func (n *Node) P2PGetAllDocuments(identityDID string) ([]string, error) {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	result := C.p2p_list_documents(n.ptr, cIdentityDID)

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return nil, mapFFIError("p2p_list_documents", err)
	}

	value := C.GoString(result.value)
	C.defra_free_string(result.value)

	var docIDs []string
	if err := json.Unmarshal([]byte(value), &docIDs); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse documents: %w", err)
	}

	return docIDs, nil
}

// P2PSyncDocuments syncs specific documents from peers.
// This implements the DocSync pull-based protocol.
func (n *Node) P2PSyncDocuments(identityDID string, collectionName string, docIDs []string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cCollectionName := C.CString(collectionName)
	defer C.free(unsafe.Pointer(cCollectionName))

	docIDsJSON, err := json.Marshal(docIDs)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal doc IDs: %w", err)
	}
	cDocIDsJSON := C.CString(string(docIDsJSON))
	defer C.free(unsafe.Pointer(cDocIDsJSON))

	result := C.p2p_sync_documents(n.ptr, cIdentityDID, cCollectionName, cDocIDsJSON)

	if result.status != 0 {
		errMsg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_sync_documents", errMsg)
	}
	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// P2PSyncBranchableCollection syncs a branchable collection from peers.
func (n *Node) P2PSyncBranchableCollection(identityDID string, collectionID string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	cCollectionID := C.CString(collectionID)
	defer C.free(unsafe.Pointer(cCollectionID))

	result := C.p2p_sync_branchable_collection(n.ptr, cIdentityDID, cCollectionID)

	if result.status != 0 {
		errMsg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_sync_branchable_collection", errMsg)
	}
	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// P2PSyncCollectionVersions syncs collection versions by their CIDs.
func (n *Node) P2PSyncCollectionVersions(identityDID string, versionIDs []string) error {
	var cIdentityDID *C.char
	if identityDID != "" {
		cIdentityDID = C.CString(identityDID)
		defer C.free(unsafe.Pointer(cIdentityDID))
	}

	versionIDsJSON, err := json.Marshal(versionIDs)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal version IDs: %w", err)
	}

	cVersionIDsJSON := C.CString(string(versionIDsJSON))
	defer C.free(unsafe.Pointer(cVersionIDsJSON))

	result := C.p2p_sync_collection_versions(n.ptr, cIdentityDID, cVersionIDsJSON)

	if result.status != 0 {
		errMsg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("p2p_sync_collection_versions", errMsg)
	}
	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}

// BasicExportDB exports the database to a JSON file.
func (n *Node) BasicExportDB(configJSON string) error {
	cConfig := C.CString(configJSON)
	defer C.free(unsafe.Pointer(cConfig))

	result := C.basic_export(n.ptr, cConfig)
	if result.status != 0 {
		msg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("basic_export", msg)
	}
	if result.value != nil {
		C.defra_free_string(result.value)
	}
	return nil
}

// BasicImportDB imports documents from a JSON backup file.
func (n *Node) BasicImportDB(filepath string) error {
	cFilepath := C.CString(filepath)
	defer C.free(unsafe.Pointer(cFilepath))

	result := C.basic_import(n.ptr, cFilepath)
	if result.status != 0 {
		msg := C.GoString(result.error)
		C.defra_free_string(result.error)
		return mapFFIError("basic_import", msg)
	}
	if result.value != nil {
		C.defra_free_string(result.value)
	}
	return nil
}
