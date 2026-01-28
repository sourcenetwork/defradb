// Package rustffi provides Go bindings for the DefraDB Rust FFI.
//
// This file implements the DefraDB client.TxnStore interface for integration testing.
package rustffi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/immutable"
	lensmodel "github.com/sourcenetwork/lens/host-go/config/model"
)

// jsonToGraphQLInput converts a JSON object to GraphQL input format.
// JSON uses quoted keys: {"Age": 21, "Name": "John"}
// GraphQL uses unquoted keys: {Age: 21, Name: "John"}
func jsonToGraphQLInput(jsonStr string) string {
	// Match quoted keys followed by colon: "key":
	re := regexp.MustCompile(`"([^"]+)"\s*:`)
	return re.ReplaceAllString(jsonStr, "$1:")
}

// Verify interface compliance at compile time
var _ clients.Client = (*Wrapper)(nil)

// Wrapper wraps an FFI Node to implement the DefraDB client.TxnStore interface.
type Wrapper struct {
	node     *Node
	events   *eventBus
	txnIDGen uint64
}

// NewWrapper creates a new Rust FFI client wrapper.
// This creates a standalone Rust FFI node (not wrapping a Go node).
func NewWrapper() (*Wrapper, error) {
	Init() // Initialize FFI library

	node, err := NewNode(NodeOptions{InMemory: true})
	if err != nil {
		return nil, fmt.Errorf("failed to create FFI node: %w", err)
	}

	return &Wrapper{
		node:   node,
		events: newEventBus(),
	}, nil
}

// NewWrapperWithP2P creates a new Rust FFI client wrapper with P2P enabled.
// listenAddr should be a multiaddr like "/ip4/127.0.0.1/tcp/0"
func NewWrapperWithP2P(listenAddr string) (*Wrapper, error) {
	Init() // Initialize FFI library

	node, err := NewNodeWithP2P(NodeOptions{InMemory: true}, listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create FFI node with P2P: %w", err)
	}

	return &Wrapper{
		node:   node,
		events: newEventBus(),
	}, nil
}

// extractCollectionNameFromPatch extracts the collection name from a JSON patch path.
// Patch format: [{"op": "add", "path": "/CollectionName/Fields/-", "value": {...}}]
// All operations must target the same collection.
func extractCollectionNameFromPatch(patch string) (string, error) {
	var ops []struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(patch), &ops); err != nil {
		return "", fmt.Errorf("failed to parse patch JSON: %w", err)
	}
	if len(ops) == 0 {
		return "", fmt.Errorf("patch contains no operations")
	}

	var collectionName string
	for i, op := range ops {
		// Path format: /CollectionName/Fields/- or /CollectionName/Fields/0
		path := op.Path
		if len(path) == 0 || path[0] != '/' {
			return "", fmt.Errorf("invalid patch path in operation %d: %s", i, path)
		}

		// Remove leading slash and extract first component
		path = path[1:]
		name := path
		for j, c := range path {
			if c == '/' {
				name = path[:j]
				break
			}
		}

		if collectionName == "" {
			collectionName = name
		} else if collectionName != name {
			return "", fmt.Errorf("patch contains operations for multiple collections (%s, %s); only single-collection patches are supported", collectionName, name)
		}
	}
	return collectionName, nil
}

// ============================================================================
// clients.Client interface
// ============================================================================

func (w *Wrapper) Close() {
	if w.node != nil {
		w.node.Close()
	}
	if w.events != nil {
		w.events.Close()
	}
}

func (w *Wrapper) MaxTxnRetries() int {
	return 5 // Default value, matches Go DefraDB
}

func (w *Wrapper) Events() event.Bus {
	return w.events
}

// ============================================================================
// client.TxnStore interface
// ============================================================================

func (w *Wrapper) NewTxn(readOnly bool) (client.Txn, error) {
	txn, err := w.node.BeginTxn(readOnly)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&w.txnIDGen, 1)
	return &TxnWrapper{
		wrapper:  w,
		txn:      txn,
		id:       id,
		readOnly: readOnly,
		startTS:  time.Now(),
	}, nil
}

func (w *Wrapper) NewConcurrentTxn(readOnly bool) (client.Txn, error) {
	// Our FFI transactions are already thread-safe
	return w.NewTxn(readOnly)
}

// ============================================================================
// client.Store interface - Core methods
// ============================================================================

func (w *Wrapper) ExecRequest(
	ctx context.Context,
	request string,
	opts ...client.RequestOption,
) *client.RequestResult {
	gqlOpts := &client.GQLOptions{}
	for _, opt := range opts {
		opt(gqlOpts)
	}

	varsJSON := ""
	if gqlOpts.Variables != nil {
		varsBytes, err := json.Marshal(gqlOpts.Variables)
		if err != nil {
			return &client.RequestResult{
				GQL: client.GQLResult{
					Errors: []error{fmt.Errorf("failed to marshal variables: %w", err)},
				},
			}
		}
		varsJSON = string(varsBytes)
	}

	responseJSON, err := w.node.ExecRequest(request, gqlOpts.OperationName, varsJSON)
	if err != nil {
		return &client.RequestResult{
			GQL: client.GQLResult{
				Errors: []error{err},
			},
		}
	}

	// Use json.Decoder with UseNumber() to preserve numeric precision.
	// This ensures integers stay as json.Number rather than being converted to float64,
	// which is required for test assertions to pass.
	var gqlResult client.GQLResult
	decoder := json.NewDecoder(bytes.NewReader([]byte(responseJSON)))
	decoder.UseNumber()
	if err := decoder.Decode(&gqlResult); err != nil {
		return &client.RequestResult{
			GQL: client.GQLResult{
				Errors: []error{fmt.Errorf("failed to parse response: %w", err)},
			},
		}
	}

	return &client.RequestResult{GQL: gqlResult}
}

func (w *Wrapper) AddSchema(ctx context.Context, sdl string) ([]client.CollectionVersion, error) {
	responseJSON, err := w.node.AddSchema(sdl)
	if err != nil {
		return nil, err
	}

	var versions []client.CollectionVersion
	if err := json.Unmarshal([]byte(responseJSON), &versions); err != nil {
		return nil, fmt.Errorf("failed to parse schema response: %w", err)
	}

	return versions, nil
}

func (w *Wrapper) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	responseJSON, err := w.node.GetCollectionByName(name)
	if err != nil {
		return nil, err
	}

	var version client.CollectionVersion
	if err := json.Unmarshal([]byte(responseJSON), &version); err != nil {
		return nil, fmt.Errorf("failed to parse collection: %w", err)
	}

	return &CollectionWrapper{
		wrapper: w,
		version: version,
	}, nil
}

func (w *Wrapper) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	responseJSON, err := w.node.GetCollections()
	if err != nil {
		return nil, err
	}

	var versions []client.CollectionVersion
	if err := json.Unmarshal([]byte(responseJSON), &versions); err != nil {
		return nil, fmt.Errorf("failed to parse collections: %w", err)
	}

	// Apply filters
	var filtered []client.CollectionVersion
	for _, v := range versions {
		if options.Name.HasValue() && v.Name != options.Name.Value() {
			continue
		}
		if options.VersionID.HasValue() && v.VersionID != options.VersionID.Value() {
			continue
		}
		if options.CollectionID.HasValue() && v.CollectionID != options.CollectionID.Value() {
			continue
		}
		filtered = append(filtered, v)
	}

	collections := make([]client.Collection, len(filtered))
	for i, v := range filtered {
		collections[i] = &CollectionWrapper{
			wrapper: w,
			version: v,
		}
	}

	return collections, nil
}

func (w *Wrapper) SetActiveCollectionVersion(ctx context.Context, versionID string) error {
	return w.node.SetActiveCollectionVersion(versionID)
}

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[lensmodel.Lens],
) error {
	// Extract collection name from JSON patch path
	// Patch format: [{"op": "add", "path": "/CollectionName/Fields/-", "value": {...}}]
	collectionName, err := extractCollectionNameFromPatch(patch)
	if err != nil {
		return err
	}
	_, err = w.node.PatchCollection(collectionName, patch)
	return err
}

func (w *Wrapper) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	result, err := w.node.GetAllIndexes()
	if err != nil {
		return nil, err
	}

	// Convert from our IndexDescription to client.IndexDescription
	indexes := make(map[client.CollectionName][]client.IndexDescription)
	for name, descs := range result {
		converted := make([]client.IndexDescription, len(descs))
		for i, d := range descs {
			fields := make([]client.IndexedFieldDescription, len(d.Fields))
			for j, f := range d.Fields {
				fields[j] = client.IndexedFieldDescription{
					Name:       f.Name,
					Descending: f.Descending,
				}
			}
			converted[i] = client.IndexDescription{
				Name:   d.Name,
				ID:     d.ID,
				Fields: fields,
				Unique: d.Unique,
			}
		}
		indexes[name] = converted
	}

	return indexes, nil
}

// ============================================================================
// client.Store interface - ACP methods
// ============================================================================

func (w *Wrapper) AddDACPolicy(ctx context.Context, policy string) (client.AddPolicyResult, error) {
	identityDID := ""
	if id := identity.FromContext(ctx); id.HasValue() {
		identityDID = id.Value().DID()
	}

	policyID, err := w.node.AddDACPolicy(identityDID, policy)
	if err != nil {
		return client.AddPolicyResult{}, err
	}

	return client.AddPolicyResult{PolicyID: policyID}, nil
}

func (w *Wrapper) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	added, err := w.node.AddDACActorRelationship("", targetActor, collectionName, docID, relation)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	return client.AddActorRelationshipResult{ExistedAlready: !added}, nil
}

func (w *Wrapper) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	deleted, err := w.node.DeleteDACActorRelationship("", targetActor, collectionName, docID, relation)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return client.DeleteActorRelationshipResult{RecordFound: deleted}, nil
}

func (w *Wrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	added, err := w.node.AddNACActorRelationship("", targetActor)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	return client.AddActorRelationshipResult{ExistedAlready: !added}, nil
}

func (w *Wrapper) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	deleted, err := w.node.DeleteNACActorRelationship("", targetActor)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return client.DeleteActorRelationshipResult{RecordFound: deleted}, nil
}

func (w *Wrapper) ReEnableNAC(ctx context.Context) error {
	return w.node.ReEnableNAC("")
}

func (w *Wrapper) DisableNAC(ctx context.Context) error {
	return w.node.DisableNAC("")
}

func (w *Wrapper) GetNACStatus(ctx context.Context) (client.NACStatusResult, error) {
	status, err := w.node.GetNACStatus()
	if err != nil {
		return client.NACStatusResult{}, err
	}
	return client.NACStatusResult{
		Status: status.Status,
	}, nil
}

func (w *Wrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	did, err := w.node.GetNodeIdentity()
	if err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	if did == "" {
		return immutable.None[identity.PublicRawIdentity](), nil
	}
	return immutable.Some(identity.PublicRawIdentity{DID: did}), nil
}

func (w *Wrapper) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	return fmt.Errorf("VerifySignature not yet implemented in FFI")
}

// ============================================================================
// client.Store interface - View/Migration methods
// ============================================================================

func (w *Wrapper) AddView(
	ctx context.Context,
	gqlQuery string,
	sdl string,
	transformCID immutable.Option[string],
) ([]client.CollectionVersion, error) {
	transformStr := ""
	if transformCID.HasValue() {
		transformStr = transformCID.Value()
	}

	responseJSON, err := w.node.AddView(gqlQuery, sdl, transformStr)
	if err != nil {
		return nil, err
	}

	var versions []client.CollectionVersion
	if err := json.Unmarshal([]byte(responseJSON), &versions); err != nil {
		return nil, fmt.Errorf("failed to parse view response: %w", err)
	}

	return versions, nil
}

func (w *Wrapper) RefreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
	optsJSON := ""
	if opts.Name.HasValue() || opts.VersionID.HasValue() {
		data, err := json.Marshal(opts)
		if err != nil {
			return fmt.Errorf("failed to marshal options: %w", err)
		}
		optsJSON = string(data)
	}
	return w.node.RefreshViews(optsJSON)
}

func (w *Wrapper) SetMigration(ctx context.Context, config client.LensConfig) (string, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	return w.node.SetMigration(string(configJSON))
}

func (w *Wrapper) AddLens(ctx context.Context, lens lensmodel.Lens) (string, error) {
	return "", fmt.Errorf("AddLens not yet implemented in FFI")
}

func (w *Wrapper) ListLenses(ctx context.Context) (map[string]lensmodel.Lens, error) {
	return nil, fmt.Errorf("ListLenses not yet implemented in FFI")
}

// ============================================================================
// client.Store interface - Backup methods
// ============================================================================

func (w *Wrapper) BasicImport(ctx context.Context, filepath string) error {
	return fmt.Errorf("BasicImport not yet implemented in FFI")
}

func (w *Wrapper) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	return fmt.Errorf("BasicExport not yet implemented in FFI")
}

// ============================================================================
// client.Store interface - Utility methods
// ============================================================================

func (w *Wrapper) PrintDump(ctx context.Context) error {
	return fmt.Errorf("PrintDump not yet implemented in FFI")
}

func (w *Wrapper) ListAllEncryptedIndexes(ctx context.Context) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	return nil, fmt.Errorf("ListAllEncryptedIndexes not yet implemented in FFI")
}

// ============================================================================
// client.P2P interface
// ============================================================================

func (w *Wrapper) PeerInfo() ([]string, error) {
	return w.node.P2PPeerInfo()
}

func (w *Wrapper) ActivePeers(ctx context.Context) ([]string, error) {
	return w.node.P2PActivePeers()
}

func (w *Wrapper) Connect(ctx context.Context, addresses []string) error {
	for _, addr := range addresses {
		if err := w.node.P2PConnect(addr); err != nil {
			return err
		}
	}
	return nil
}

func (w *Wrapper) SetReplicator(ctx context.Context, addresses []string, collections ...string) error {
	if len(addresses) == 0 {
		return fmt.Errorf("at least one address is required")
	}
	// Use the first address as the peer address
	return w.node.P2PSetReplicator(addresses[0], collections)
}

func (w *Wrapper) DeleteReplicator(ctx context.Context, id string, collections ...string) error {
	return w.node.P2PDeleteReplicator(id)
}

func (w *Wrapper) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	replicators, err := w.node.P2PGetAllReplicators()
	if err != nil {
		return nil, err
	}

	// Convert to client.Replicator
	result := make([]client.Replicator, len(replicators))
	for i, r := range replicators {
		result[i] = client.Replicator{
			ID:            r.PeerID,
			Addresses:     r.Addresses,
			CollectionIDs: r.Collections,
		}
	}
	return result, nil
}

func (w *Wrapper) AddP2PCollections(ctx context.Context, collectionNames ...string) error {
	return w.node.P2PAddCollections(collectionNames)
}

func (w *Wrapper) RemoveP2PCollections(ctx context.Context, collectionNames ...string) error {
	return w.node.P2PRemoveCollections(collectionNames)
}

func (w *Wrapper) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	return w.node.P2PGetAllCollections()
}

func (w *Wrapper) AddP2PDocuments(ctx context.Context, docIDs ...string) error {
	// Document-level P2P not yet implemented in Rust FFI
	return fmt.Errorf("P2P document sync not yet implemented in FFI client")
}

func (w *Wrapper) RemoveP2PDocuments(ctx context.Context, docIDs ...string) error {
	// Document-level P2P not yet implemented in Rust FFI
	return fmt.Errorf("P2P document sync not yet implemented in FFI client")
}

func (w *Wrapper) GetAllP2PDocuments(ctx context.Context) ([]string, error) {
	// Document-level P2P not yet implemented in Rust FFI
	return nil, fmt.Errorf("P2P document sync not yet implemented in FFI client")
}

func (w *Wrapper) SyncDocuments(ctx context.Context, collectionName string, docIDs []string) error {
	// Document sync not yet implemented in Rust FFI
	return fmt.Errorf("P2P document sync not yet implemented in FFI client")
}

func (w *Wrapper) SyncCollectionVersions(ctx context.Context, versionIDs ...string) error {
	// Collection version sync not yet implemented in Rust FFI
	return fmt.Errorf("P2P collection version sync not yet implemented in FFI client")
}

func (w *Wrapper) SyncBranchableCollection(ctx context.Context, collectionID string) error {
	// Branchable collection sync not yet implemented in Rust FFI
	return fmt.Errorf("P2P branchable collection sync not yet implemented in FFI client")
}

// ============================================================================
// TxnWrapper implements client.Txn
// ============================================================================

type TxnWrapper struct {
	wrapper  *Wrapper
	txn      *Transaction
	id       uint64
	readOnly bool
	startTS  time.Time
}

var _ client.Txn = (*TxnWrapper)(nil)

func (t *TxnWrapper) ID() uint64 {
	return t.id
}

func (t *TxnWrapper) StartTS() time.Time {
	return t.startTS
}

func (t *TxnWrapper) Commit() error {
	return t.txn.Commit()
}

func (t *TxnWrapper) Discard() {
	_ = t.txn.Rollback()
}

// Delegate Store methods to the underlying transaction
func (t *TxnWrapper) ExecRequest(ctx context.Context, request string, opts ...client.RequestOption) *client.RequestResult {
	gqlOpts := &client.GQLOptions{}
	for _, opt := range opts {
		opt(gqlOpts)
	}

	varsJSON := ""
	if gqlOpts.Variables != nil {
		varsBytes, err := json.Marshal(gqlOpts.Variables)
		if err != nil {
			return &client.RequestResult{
				GQL: client.GQLResult{
					Errors: []error{fmt.Errorf("failed to marshal variables: %w", err)},
				},
			}
		}
		varsJSON = string(varsBytes)
	}

	responseJSON, err := t.txn.ExecRequest(request, gqlOpts.OperationName, varsJSON)
	if err != nil {
		return &client.RequestResult{
			GQL: client.GQLResult{
				Errors: []error{err},
			},
		}
	}

	// Use json.Decoder with UseNumber() to preserve numeric precision.
	// This ensures integers stay as json.Number rather than being converted to float64,
	// which is required for test assertions to pass.
	var gqlResult client.GQLResult
	decoder := json.NewDecoder(bytes.NewReader([]byte(responseJSON)))
	decoder.UseNumber()
	if err := decoder.Decode(&gqlResult); err != nil {
		return &client.RequestResult{
			GQL: client.GQLResult{
				Errors: []error{fmt.Errorf("failed to parse response: %w", err)},
			},
		}
	}

	return &client.RequestResult{GQL: gqlResult}
}

// Stub out other Store methods on transaction - most delegate to wrapper
func (t *TxnWrapper) AddSchema(ctx context.Context, sdl string) ([]client.CollectionVersion, error) {
	return t.wrapper.AddSchema(ctx, sdl)
}

func (t *TxnWrapper) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	return t.wrapper.GetCollectionByName(ctx, name)
}

func (t *TxnWrapper) GetCollections(ctx context.Context, options client.CollectionFetchOptions) ([]client.Collection, error) {
	return t.wrapper.GetCollections(ctx, options)
}

func (t *TxnWrapper) SetActiveCollectionVersion(ctx context.Context, versionID string) error {
	return t.wrapper.SetActiveCollectionVersion(ctx, versionID)
}

func (t *TxnWrapper) PatchCollection(ctx context.Context, patch string, migration immutable.Option[lensmodel.Lens]) error {
	return t.wrapper.PatchCollection(ctx, patch, migration)
}

func (t *TxnWrapper) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	return t.wrapper.GetAllIndexes(ctx)
}

func (t *TxnWrapper) AddDACPolicy(ctx context.Context, policy string) (client.AddPolicyResult, error) {
	return t.wrapper.AddDACPolicy(ctx, policy)
}

func (t *TxnWrapper) AddDACActorRelationship(ctx context.Context, collectionName, docID, relation, targetActor string) (client.AddActorRelationshipResult, error) {
	return t.wrapper.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
}

func (t *TxnWrapper) DeleteDACActorRelationship(ctx context.Context, collectionName, docID, relation, targetActor string) (client.DeleteActorRelationshipResult, error) {
	return t.wrapper.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
}

func (t *TxnWrapper) AddNACActorRelationship(ctx context.Context, relation, targetActor string) (client.AddActorRelationshipResult, error) {
	return t.wrapper.AddNACActorRelationship(ctx, relation, targetActor)
}

func (t *TxnWrapper) DeleteNACActorRelationship(ctx context.Context, relation, targetActor string) (client.DeleteActorRelationshipResult, error) {
	return t.wrapper.DeleteNACActorRelationship(ctx, relation, targetActor)
}

func (t *TxnWrapper) ReEnableNAC(ctx context.Context) error {
	return t.wrapper.ReEnableNAC(ctx)
}

func (t *TxnWrapper) DisableNAC(ctx context.Context) error {
	return t.wrapper.DisableNAC(ctx)
}

func (t *TxnWrapper) GetNACStatus(ctx context.Context) (client.NACStatusResult, error) {
	return t.wrapper.GetNACStatus(ctx)
}

func (t *TxnWrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	return t.wrapper.GetNodeIdentity(ctx)
}

func (t *TxnWrapper) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	return t.wrapper.VerifySignature(ctx, blockCid, pubKey)
}

func (t *TxnWrapper) AddView(ctx context.Context, gqlQuery, sdl string, transformCID immutable.Option[string]) ([]client.CollectionVersion, error) {
	return t.wrapper.AddView(ctx, gqlQuery, sdl, transformCID)
}

func (t *TxnWrapper) RefreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
	return t.wrapper.RefreshViews(ctx, opts)
}

func (t *TxnWrapper) SetMigration(ctx context.Context, config client.LensConfig) (string, error) {
	return t.wrapper.SetMigration(ctx, config)
}

func (t *TxnWrapper) AddLens(ctx context.Context, lens lensmodel.Lens) (string, error) {
	return t.wrapper.AddLens(ctx, lens)
}

func (t *TxnWrapper) ListLenses(ctx context.Context) (map[string]lensmodel.Lens, error) {
	return t.wrapper.ListLenses(ctx)
}

func (t *TxnWrapper) BasicImport(ctx context.Context, filepath string) error {
	return t.wrapper.BasicImport(ctx, filepath)
}

func (t *TxnWrapper) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	return t.wrapper.BasicExport(ctx, config)
}

func (t *TxnWrapper) PrintDump(ctx context.Context) error {
	return t.wrapper.PrintDump(ctx)
}

func (t *TxnWrapper) ListAllEncryptedIndexes(ctx context.Context) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	return t.wrapper.ListAllEncryptedIndexes(ctx)
}

// P2P methods - not available in transactions
func (t *TxnWrapper) PeerInfo() ([]string, error) { return nil, fmt.Errorf("P2P not available") }
func (t *TxnWrapper) ActivePeers(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) Connect(ctx context.Context, addresses []string) error {
	return fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) SetReplicator(ctx context.Context, addresses []string, collections ...string) error {
	return fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) DeleteReplicator(ctx context.Context, id string, collections ...string) error {
	return fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	return nil, fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) AddP2PCollections(ctx context.Context, collectionNames ...string) error {
	return fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) RemoveP2PCollections(ctx context.Context, collectionNames ...string) error {
	return fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) AddP2PDocuments(ctx context.Context, docIDs ...string) error {
	return fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) RemoveP2PDocuments(ctx context.Context, docIDs ...string) error {
	return fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) GetAllP2PDocuments(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) SyncDocuments(ctx context.Context, collectionName string, docIDs []string) error {
	return fmt.Errorf("P2P not available")
}
func (t *TxnWrapper) SyncCollectionVersions(ctx context.Context, versionIDs ...string) error {
	return fmt.Errorf("P2P not available")
}

func (t *TxnWrapper) SyncBranchableCollection(ctx context.Context, collectionID string) error {
	return fmt.Errorf("P2P not available")
}

// ============================================================================
// eventBus implements event.Bus for testing
// ============================================================================

type eventBus struct {
	closed bool
	subs   []event.Subscription
}

func newEventBus() *eventBus {
	return &eventBus{}
}

func (e *eventBus) Publish(msg event.Message) {
	// Deliver message to all matching subscribers
	for _, sub := range e.subs {
		if es, ok := sub.(*eventSubscription); ok {
			// Check if subscription wants this event
			if es.wantsEvent(msg.Name) {
				select {
				case es.ch <- msg:
				default:
					// Channel full, skip
				}
			}
		}
	}
}

func (e *eventBus) Subscribe(events ...event.Name) (event.Subscription, error) {
	sub := &eventSubscription{
		ch:     make(chan event.Message, 100),
		events: events,
	}
	e.subs = append(e.subs, sub)
	return sub, nil
}

func (e *eventBus) Unsubscribe(sub event.Subscription) {
	for i, s := range e.subs {
		if s == sub {
			e.subs = append(e.subs[:i], e.subs[i+1:]...)
			break
		}
	}
}

func (e *eventBus) Close() {
	e.closed = true
	for _, sub := range e.subs {
		if es, ok := sub.(*eventSubscription); ok {
			close(es.ch)
		}
	}
}

type eventSubscription struct {
	ch     chan event.Message
	events []event.Name
}

func (s *eventSubscription) wantsEvent(name event.Name) bool {
	if len(s.events) == 0 {
		return true // no filter, accept all
	}
	for _, e := range s.events {
		if e == name || e == event.WildCardName {
			return true
		}
	}
	return false
}

func (s *eventSubscription) Message() <-chan event.Message {
	return s.ch
}

// ============================================================================
// CollectionWrapper implements client.Collection
// ============================================================================

type CollectionWrapper struct {
	wrapper *Wrapper
	version client.CollectionVersion
}

var _ client.Collection = (*CollectionWrapper)(nil)

func (c *CollectionWrapper) Name() string {
	return c.version.Name
}

func (c *CollectionWrapper) VersionID() string {
	return c.version.VersionID
}

func (c *CollectionWrapper) CollectionID() string {
	return c.version.CollectionID
}

func (c *CollectionWrapper) Version() client.CollectionVersion {
	return c.version
}

func (c *CollectionWrapper) Create(ctx context.Context, doc *client.Document, opts ...client.DocCreateOption) error {
	// Use GraphQL mutation to create document
	docJSON, err := doc.ToJSONPatch()
	if err != nil {
		return fmt.Errorf("failed to convert document to JSON: %w", err)
	}

	// Convert JSON to GraphQL input format (unquoted keys)
	gqlInput := jsonToGraphQLInput(string(docJSON))
	mutation := fmt.Sprintf(`mutation { create_%s(input: %s) { _docID } }`, c.version.Name, gqlInput)
	result := c.wrapper.ExecRequest(ctx, mutation)
	if len(result.GQL.Errors) > 0 {
		return result.GQL.Errors[0]
	}

	// Extract docID from the response and publish update event
	if data, ok := result.GQL.Data.(map[string]any); ok {
		mutationKey := "create_" + c.version.Name
		if mutResult, ok := data[mutationKey].([]any); ok && len(mutResult) > 0 {
			if docData, ok := mutResult[0].(map[string]any); ok {
				if docID, ok := docData["_docID"].(string); ok {
					c.wrapper.events.Publish(event.NewMessage(event.UpdateName, event.Update{
						DocID:        docID,
						CollectionID: c.version.CollectionID,
					}))
				}
			}
		}
	}

	return nil
}

func (c *CollectionWrapper) CreateMany(ctx context.Context, docs []*client.Document, opts ...client.DocCreateOption) error {
	for _, doc := range docs {
		if err := c.Create(ctx, doc, opts...); err != nil {
			return err
		}
	}
	return nil
}

func (c *CollectionWrapper) Update(ctx context.Context, doc *client.Document) error {
	docJSON, err := doc.ToJSONPatch()
	if err != nil {
		return fmt.Errorf("failed to convert document to JSON: %w", err)
	}

	// Convert JSON to GraphQL input format (unquoted keys)
	gqlInput := jsonToGraphQLInput(string(docJSON))
	mutation := fmt.Sprintf(`mutation { update_%s(docID: "%s", input: %s) { _docID } }`,
		c.version.Name, doc.ID().String(), gqlInput)
	result := c.wrapper.ExecRequest(ctx, mutation)
	if len(result.GQL.Errors) > 0 {
		return result.GQL.Errors[0]
	}

	// Publish update event
	c.wrapper.events.Publish(event.NewMessage(event.UpdateName, event.Update{
		DocID:        doc.ID().String(),
		CollectionID: c.version.CollectionID,
	}))

	return nil
}

func (c *CollectionWrapper) Save(ctx context.Context, doc *client.Document, opts ...client.DocCreateOption) error {
	// Check if doc exists in the database by querying for it
	exists, err := c.Exists(ctx, doc.ID())
	if err != nil {
		// If error checking existence, try create
		return c.Create(ctx, doc, opts...)
	}
	if !exists {
		return c.Create(ctx, doc, opts...)
	}
	return c.Update(ctx, doc)
}

func (c *CollectionWrapper) Delete(ctx context.Context, docID client.DocID) (bool, error) {
	mutation := fmt.Sprintf(`mutation { delete_%s(docID: "%s") { _docID } }`, c.version.Name, docID.String())
	result := c.wrapper.ExecRequest(ctx, mutation)
	if len(result.GQL.Errors) > 0 {
		return false, result.GQL.Errors[0]
	}
	return true, nil
}

func (c *CollectionWrapper) Exists(ctx context.Context, docID client.DocID) (bool, error) {
	query := fmt.Sprintf(`{ %s(docID: "%s") { _docID } }`, c.version.Name, docID.String())
	result := c.wrapper.ExecRequest(ctx, query)
	if len(result.GQL.Errors) > 0 {
		return false, nil
	}
	if data, ok := result.GQL.Data.(map[string]any); ok {
		if docs, ok := data[c.version.Name].([]any); ok {
			return len(docs) > 0, nil
		}
	}
	return false, nil
}

func (c *CollectionWrapper) UpdateWithFilter(ctx context.Context, filter any, updater string) (*client.UpdateResult, error) {
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filter: %w", err)
	}

	// Convert JSON to GraphQL input format (unquoted keys)
	gqlFilter := jsonToGraphQLInput(string(filterJSON))
	gqlUpdater := jsonToGraphQLInput(updater)
	mutation := fmt.Sprintf(`mutation { update_%s(filter: %s, input: %s) { _docID } }`,
		c.version.Name, gqlFilter, gqlUpdater)
	result := c.wrapper.ExecRequest(ctx, mutation)
	if len(result.GQL.Errors) > 0 {
		return nil, result.GQL.Errors[0]
	}

	// Extract results
	updateResult := &client.UpdateResult{}
	if data, ok := result.GQL.Data.(map[string]any); ok {
		if docs, ok := data["update_"+c.version.Name].([]any); ok {
			updateResult.Count = int64(len(docs))
			for _, d := range docs {
				if doc, ok := d.(map[string]any); ok {
					if id, ok := doc["_docID"].(string); ok {
						updateResult.DocIDs = append(updateResult.DocIDs, id)
					}
				}
			}
		}
	}
	return updateResult, nil
}

func (c *CollectionWrapper) DeleteWithFilter(ctx context.Context, filter any) (*client.DeleteResult, error) {
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filter: %w", err)
	}

	// Convert JSON to GraphQL input format (unquoted keys)
	gqlFilter := jsonToGraphQLInput(string(filterJSON))
	mutation := fmt.Sprintf(`mutation { delete_%s(filter: %s) { _docID } }`, c.version.Name, gqlFilter)
	result := c.wrapper.ExecRequest(ctx, mutation)
	if len(result.GQL.Errors) > 0 {
		return nil, result.GQL.Errors[0]
	}

	deleteResult := &client.DeleteResult{}
	if data, ok := result.GQL.Data.(map[string]any); ok {
		if docs, ok := data["delete_"+c.version.Name].([]any); ok {
			deleteResult.Count = int64(len(docs))
			for _, d := range docs {
				if doc, ok := d.(map[string]any); ok {
					if id, ok := doc["_docID"].(string); ok {
						deleteResult.DocIDs = append(deleteResult.DocIDs, id)
					}
				}
			}
		}
	}
	return deleteResult, nil
}

func (c *CollectionWrapper) Get(ctx context.Context, docID client.DocID, showDeleted bool) (*client.Document, error) {
	// Query the document by ID
	query := fmt.Sprintf(`{ %s(docID: "%s") { _docID } }`, c.version.Name, docID.String())
	result := c.wrapper.ExecRequest(ctx, query)
	if len(result.GQL.Errors) > 0 {
		return nil, result.GQL.Errors[0]
	}

	// Parse the result to check if document exists
	data, ok := result.GQL.Data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("document not found: %s", docID.String())
	}

	docs, ok := data[c.version.Name].([]any)
	if !ok || len(docs) == 0 {
		return nil, fmt.Errorf("document not found: %s", docID.String())
	}

	docData, ok := docs[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("document not found: %s", docID.String())
	}

	// Create a new document from the retrieved data
	doc, err := client.NewDocFromMap(ctx, docData, c.version)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return doc, nil
}

func (c *CollectionWrapper) GetAllDocIDs(ctx context.Context) (<-chan client.DocIDResult, error) {
	ch := make(chan client.DocIDResult)
	go func() {
		defer close(ch)
		query := fmt.Sprintf(`{ %s { _docID } }`, c.version.Name)
		result := c.wrapper.ExecRequest(ctx, query)
		if len(result.GQL.Errors) > 0 {
			ch <- client.DocIDResult{Err: result.GQL.Errors[0]}
			return
		}
		if data, ok := result.GQL.Data.(map[string]any); ok {
			if docs, ok := data[c.version.Name].([]any); ok {
				for _, d := range docs {
					if doc, ok := d.(map[string]any); ok {
						if id, ok := doc["_docID"].(string); ok {
							docID, err := client.NewDocIDFromString(id)
							if err != nil {
								ch <- client.DocIDResult{Err: err}
								continue
							}
							ch <- client.DocIDResult{ID: docID}
						}
					}
				}
			}
		}
	}()
	return ch, nil
}

func (c *CollectionWrapper) CreateIndex(ctx context.Context, req client.IndexCreateRequest) (client.IndexDescription, error) {
	fields := make([]IndexField, len(req.Fields))
	for i, f := range req.Fields {
		fields[i] = IndexField{
			Name:       f.Name,
			Descending: f.Descending,
		}
	}

	index, err := c.wrapper.node.CreateIndex(c.version.Name, req.Name, fields, req.Unique)
	if err != nil {
		return client.IndexDescription{}, err
	}

	resultFields := make([]client.IndexedFieldDescription, len(index.Fields))
	for i, f := range index.Fields {
		resultFields[i] = client.IndexedFieldDescription{
			Name:       f.Name,
			Descending: f.Descending,
		}
	}

	return client.IndexDescription{
		Name:   index.Name,
		ID:     index.ID,
		Fields: resultFields,
		Unique: index.Unique,
	}, nil
}

func (c *CollectionWrapper) DropIndex(ctx context.Context, indexName string) error {
	return c.wrapper.node.DropIndex(c.version.Name, indexName)
}

func (c *CollectionWrapper) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	indexes, err := c.wrapper.node.GetIndexes(c.version.Name)
	if err != nil {
		return nil, err
	}

	result := make([]client.IndexDescription, len(indexes))
	for i, idx := range indexes {
		fields := make([]client.IndexedFieldDescription, len(idx.Fields))
		for j, f := range idx.Fields {
			fields[j] = client.IndexedFieldDescription{
				Name:       f.Name,
				Descending: f.Descending,
			}
		}
		result[i] = client.IndexDescription{
			Name:   idx.Name,
			ID:     idx.ID,
			Fields: fields,
			Unique: idx.Unique,
		}
	}
	return result, nil
}

func (c *CollectionWrapper) CreateEncryptedIndex(ctx context.Context, desc client.EncryptedIndexDescription) (client.EncryptedIndexDescription, error) {
	return client.EncryptedIndexDescription{}, fmt.Errorf("encrypted indexes not yet implemented in FFI")
}

func (c *CollectionWrapper) DeleteEncryptedIndex(ctx context.Context, fieldName string) error {
	return fmt.Errorf("encrypted indexes not yet implemented in FFI")
}

func (c *CollectionWrapper) ListEncryptedIndexes(ctx context.Context) ([]client.EncryptedIndexDescription, error) {
	return nil, fmt.Errorf("encrypted indexes not yet implemented in FFI")
}

func (c *CollectionWrapper) Truncate(ctx context.Context) error {
	return fmt.Errorf("Truncate not yet implemented in FFI")
}
