// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build rust_ffi

// Package rustffi provides Go bindings for the DefraDB Rust FFI.
//
// This file implements the DefraDB client.TxnStore interface for integration testing.
package rustffi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	gocid "github.com/ipfs/go-cid"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/utils"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/immutable"
	lensmodel "github.com/sourcenetwork/lens/host-go/config/model"
)

const (
	// nullStr is the string representation of a null value.
	nullStr = "null"

	// copyOp is the JSON Patch "copy" operation name.
	copyOp = "copy"
)

// jsonToGraphQLInput converts a JSON object to GraphQL input format.
// jsonToGraphQLInput converts JSON object syntax to GraphQL input syntax.
// JSON uses quoted keys: {"Age": 21, "Name": "John"}
// GraphQL uses unquoted keys: {Age: 21, Name: "John"}
// This function properly handles nested string values containing JSON.
func jsonToGraphQLInput(jsonStr string) string {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// If parsing fails, return as-is (shouldn't happen for valid JSON)
		return jsonStr
	}
	return valueToGQLInput(data)
}

// valueToGQLInput converts a Go value to GraphQL input syntax.
func valueToGQLInput(v any) string {
	switch val := v.(type) {
	case nil:
		return nullStr
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		// Check if it's actually an integer
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%v", val)
	case json.Number:
		return val.String()
	case string:
		// Properly escape the string for GraphQL
		escaped := strings.ReplaceAll(val, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		escaped = strings.ReplaceAll(escaped, "\r", `\r`)
		escaped = strings.ReplaceAll(escaped, "\t", `\t`)
		return `"` + escaped + `"`
	case []any:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = valueToGQLInput(item)
		}
		return "[" + strings.Join(parts, ",") + "]"
	case map[string]any:
		parts := make([]string, 0, len(val))
		for k, item := range val {
			parts = append(parts, k+":"+valueToGQLInput(item))
		}
		return "{" + strings.Join(parts, ",") + "}"
	default:
		// Fallback: use JSON marshaling
		b, _ := json.Marshal(val)
		return string(b)
	}
}

// identityDIDFromOption extracts the DID string from an identity option.
// Returns empty string if no identity is set.
func identityDIDFromOption(ident immutable.Option[identity.Identity]) string {
	if ident.HasValue() {
		return ident.Value().DID()
	}
	return ""
}

// execRequestWithIdentity builds ExecRequest options forwarding the given identity.
func execRequestWithIdentity(ident immutable.Option[identity.Identity]) options.Enumerable[options.ExecRequestOptions] {
	b := options.ExecRequest()
	if ident.HasValue() {
		b.SetIdentity(ident.Value())
	}
	return b
}

// Verify interface compliance at compile time
var _ clients.Client = (*Wrapper)(nil)

// Wrapper wraps an FFI Node to implement the DefraDB client.TxnStore interface.
type Wrapper struct {
	node            *Node
	events          *eventBus
	txnIDGen        uint64
	stopMergePoller chan struct{}
	stopSEForwarder chan struct{}
	goNodeCloser    func()                      // Called during Close() to release Go node resources (e.g. badger lock)
	nodeIdentityRaw *identity.PublicRawIdentity // Cached node identity (with PublicKey) for GetNodeIdentity
}

type rustFFIP2PRegistryEntry struct {
	w           *Wrapper
	addrs       []string
	identityDID string
}

type rustFFIP2PRegistry struct {
	mu      sync.Mutex
	entries map[string]rustFFIP2PRegistryEntry
	parent  map[string]string
}

var globalRustFFIP2PRegistry = &rustFFIP2PRegistry{
	entries: make(map[string]rustFFIP2PRegistryEntry),
	parent:  make(map[string]string),
}

func peerIDFromP2PAddr(addr string) (string, bool) {
	parts := strings.Split(addr, "/p2p/")
	if len(parts) < 2 {
		return "", false
	}
	peerID := parts[len(parts)-1]
	if index := strings.Index(peerID, "/"); index >= 0 {
		peerID = peerID[:index]
	}
	return peerID, peerID != ""
}

func addrsByPeerID(addrs []string) map[string][]string {
	result := make(map[string][]string)
	for _, addr := range addrs {
		peerID, ok := peerIDFromP2PAddr(addr)
		if !ok {
			continue
		}
		result[peerID] = append(result[peerID], addr)
	}
	return result
}

func mergeAddrs(existing []string, additions []string) []string {
	seen := make(map[string]struct{}, len(existing)+len(additions))
	result := make([]string, 0, len(existing)+len(additions))
	for _, addr := range existing {
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		result = append(result, addr)
	}
	for _, addr := range additions {
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		result = append(result, addr)
	}
	return result
}

func (r *rustFFIP2PRegistry) find(peerID string) string {
	parent, ok := r.parent[peerID]
	if !ok {
		r.parent[peerID] = peerID
		return peerID
	}
	if parent == peerID {
		return peerID
	}
	root := r.find(parent)
	r.parent[peerID] = root
	return root
}

func (r *rustFFIP2PRegistry) union(left string, right string) {
	leftRoot := r.find(left)
	rightRoot := r.find(right)
	if leftRoot != rightRoot {
		r.parent[rightRoot] = leftRoot
	}
}

func (r *rustFFIP2PRegistry) registerLocked(w *Wrapper, addrs []string) []string {
	peerIDs := make([]string, 0, 1)
	for peerID, peerAddrs := range addrsByPeerID(addrs) {
		entry := r.entries[peerID]
		if w != nil {
			entry.w = w
			if w.nodeIdentityRaw != nil {
				entry.identityDID = w.nodeIdentityRaw.DID
			} else {
				entry.identityDID = ""
			}
		}
		entry.addrs = mergeAddrs(entry.addrs, peerAddrs)
		r.entries[peerID] = entry
		r.find(peerID)
		peerIDs = append(peerIDs, peerID)
	}
	sort.Strings(peerIDs)
	return peerIDs
}

func (r *rustFFIP2PRegistry) register(w *Wrapper, addrs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registerLocked(w, addrs)
}

func (r *rustFFIP2PRegistry) unregister(w *Wrapper) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for peerID, entry := range r.entries {
		if entry.w == w {
			delete(r.entries, peerID)
			delete(r.parent, peerID)
		}
	}
	for peerID, parent := range r.parent {
		if _, ok := r.entries[peerID]; !ok {
			delete(r.parent, peerID)
			continue
		}
		if _, ok := r.entries[parent]; !ok {
			r.parent[peerID] = peerID
		}
	}
}

func (r *rustFFIP2PRegistry) meshComponent(w *Wrapper, sourceAddrs []string, targetAddrs []string) []rustFFIP2PRegistryEntry {
	r.mu.Lock()
	defer r.mu.Unlock()

	sourcePeerIDs := r.registerLocked(w, sourceAddrs)
	targetPeerIDs := r.registerLocked(nil, targetAddrs)
	if len(sourcePeerIDs) == 0 {
		return nil
	}

	sourcePeerID := sourcePeerIDs[0]
	for _, targetPeerID := range targetPeerIDs {
		r.union(sourcePeerID, targetPeerID)
	}

	sourceRoot := r.find(sourcePeerID)
	component := make([]rustFFIP2PRegistryEntry, 0)
	for peerID, entry := range r.entries {
		if entry.w == nil || len(entry.addrs) == 0 {
			continue
		}
		if r.find(peerID) == sourceRoot {
			component = append(component, rustFFIP2PRegistryEntry{
				w:           entry.w,
				addrs:       append([]string(nil), entry.addrs...),
				identityDID: entry.identityDID,
			})
		}
	}
	return component
}

// SourceHubConfig holds SourceHub connection info for Rust FFI nodes.
type SourceHubConfig struct {
	GRPCAddress     string
	CometRPCAddress string
	ChainID         string
	SignerKey       []byte
}

// NewWrapper creates a new Rust FFI client wrapper.
// This creates a standalone Rust FFI node (not wrapping a Go node).
// If nodeIdentity is provided and enableSigning is true, the identity's private key
// will be passed to the Rust FFI for block signing.
// If dbPath is non-empty, the node uses file-based (redb) storage at that path.
func NewWrapper(
	enableSigning bool,
	nodeIdentity identity.Identity,
	dbPath string,
	shConfig *SourceHubConfig,
) (*Wrapper, error) {
	Init() // Initialize FFI library

	opts := NodeOptions{EnableSigning: enableSigning}
	if dbPath != "" {
		opts.DBPath = dbPath
		opts.InMemory = false
	} else {
		opts.InMemory = true
	}

	// If signing is enabled and we have a full identity with a private key, use it
	if enableSigning && nodeIdentity != nil {
		if fullIdent, ok := nodeIdentity.(identity.FullIdentity); ok {
			privKey := fullIdent.PrivateKey()
			opts.SigningKeyType = string(privKey.Type())
			opts.SigningPrivateKey = privKey.Raw()
		}
	}

	// Pass SourceHub config if provided
	if shConfig != nil {
		opts.SourceHubGRPCAddress = shConfig.GRPCAddress
		opts.SourceHubCometRPCAddress = shConfig.CometRPCAddress
		opts.SourceHubChainID = shConfig.ChainID
		opts.SourceHubSignerKey = shConfig.SignerKey
	}

	node, err := NewNode(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create FFI node: %w", err)
	}

	// Register and set the node's default identity so GetNodeIdentity works.
	if nodeIdentity != nil {
		if err := RegisterIdentityWithRust(nodeIdentity); err != nil {
			_ = node.Close()
			return nil, fmt.Errorf("failed to register node identity: %w", err)
		}
		did := nodeIdentity.DID()
		if did != "" {
			if err := node.SetDefaultIdentity(did); err != nil {
				_ = node.Close()
				return nil, fmt.Errorf("failed to set default identity: %w", err)
			}
		}
	}

	eb := newEventBus()
	stopCh := make(chan struct{})

	// Start event poller that bridges Rust events to Go eventBus
	mergeSub, err := node.SubscribeMergeComplete()
	if err != nil {
		_ = node.Close()
		return nil, fmt.Errorf("failed to create event subscription: %w", err)
	}

	go func() {
		defer func() { _ = mergeSub.Close() }()
		for {
			select {
			case <-stopCh:
				return
			default:
			}

			result, err := mergeSub.Poll()
			if err != nil {
				continue
			}
			if result.IsClosed {
				return
			}
			if result.HasEvent && result.Event != nil {
				if result.Event.Type == "update" {
					cidObj, cidErr := gocid.Decode(result.Event.CID)
					if cidErr != nil {
						continue
					}
					eb.Publish(event.NewMessage(event.UpdateName, event.Update{
						DocID:        result.Event.DocID,
						Cid:          cidObj,
						CollectionID: result.Event.CollectionID,
						IsRelay:      result.Event.IsRelay,
					}))
					continue
				}
				if result.Event.Type == "merge_complete" {
					cidObj, cidErr := gocid.Decode(result.Event.CID)
					if cidErr != nil {
						continue
					}
					mc := event.MergeComplete{
						Merge: event.Merge{
							DocID:        result.Event.DocID,
							Cid:          cidObj,
							CollectionID: result.Event.CollectionID,
							ByPeer:       result.Event.ByPeer,
						},
					}
					eb.Publish(event.NewMessage(event.MergeCompleteName, mc))
					continue
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	w := &Wrapper{
		node:            node,
		events:          eb,
		stopMergePoller: stopCh,
	}
	if nodeIdentity != nil {
		raw := nodeIdentity.ToPublicRawIdentity()
		w.nodeIdentityRaw = &raw
	}
	return w, nil
}

// NewWrapperWithP2P creates a new Rust FFI client wrapper with P2P enabled.
// listenAddr should be a multiaddr like "/ip4/127.0.0.1/tcp/0"
// If nodeIdentity is provided and enableSigning is true, the identity's private key
// will be passed to the Rust FFI for block signing.
// If dbPath is non-empty, the node uses file-based (redb) storage at that path.
func NewWrapperWithP2P(
	listenAddr string,
	enableSigning bool,
	nodeIdentity identity.Identity,
	dbPath string,
	shConfig *SourceHubConfig,
) (*Wrapper, error) {
	Init() // Initialize FFI library

	opts := NodeOptions{EnableSigning: enableSigning}
	if dbPath != "" {
		opts.DBPath = dbPath
		opts.InMemory = false
	} else {
		opts.InMemory = true
	}

	// If signing is enabled and we have a full identity with a private key, use it
	if enableSigning && nodeIdentity != nil {
		if fullIdent, ok := nodeIdentity.(identity.FullIdentity); ok {
			privKey := fullIdent.PrivateKey()
			opts.SigningKeyType = string(privKey.Type())
			opts.SigningPrivateKey = privKey.Raw()
		}
	}

	// Pass SourceHub config if provided
	if shConfig != nil {
		opts.SourceHubGRPCAddress = shConfig.GRPCAddress
		opts.SourceHubCometRPCAddress = shConfig.CometRPCAddress
		opts.SourceHubChainID = shConfig.ChainID
		opts.SourceHubSignerKey = shConfig.SignerKey
	}

	node, err := NewNodeWithP2P(opts, listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create FFI node with P2P: %w", err)
	}

	// Register and set the node's default identity so GetNodeIdentity works.
	if nodeIdentity != nil {
		if err := RegisterIdentityWithRust(nodeIdentity); err != nil {
			_ = node.Close()
			return nil, fmt.Errorf("failed to register node identity: %w", err)
		}
		did := nodeIdentity.DID()
		if did != "" {
			if err := node.SetDefaultIdentity(did); err != nil {
				_ = node.Close()
				return nil, fmt.Errorf("failed to set default identity: %w", err)
			}
		}
	}

	eb := newEventBus()
	stopCh := make(chan struct{})

	// Start merge complete event poller that bridges Rust events to Go eventBus
	mergeSub, err := node.SubscribeMergeComplete()
	if err != nil {
		_ = node.Close()
		return nil, fmt.Errorf("failed to create merge complete subscription: %w", err)
	}

	go func() {
		defer func() { _ = mergeSub.Close() }()
		for {
			select {
			case <-stopCh:
				return
			default:
			}

			result, err := mergeSub.Poll()
			if err != nil {
				continue
			}
			if result.IsClosed {
				return
			}
			if result.HasEvent && result.Event != nil {
				if result.Event.Type == "update" {
					cidObj, cidErr := gocid.Decode(result.Event.CID)
					if cidErr != nil {
						continue
					}
					eb.Publish(event.NewMessage(event.UpdateName, event.Update{
						DocID:        result.Event.DocID,
						Cid:          cidObj,
						CollectionID: result.Event.CollectionID,
						IsRelay:      result.Event.IsRelay,
					}))
					continue
				}
				if result.Event.Type == "merge_complete" {
					cidObj, cidErr := gocid.Decode(result.Event.CID)
					if cidErr != nil {
						continue
					}
					mc := event.MergeComplete{
						Merge: event.Merge{
							DocID:        result.Event.DocID,
							Cid:          cidObj,
							CollectionID: result.Event.CollectionID,
							ByPeer:       result.Event.ByPeer,
						},
					}
					eb.Publish(event.NewMessage(event.MergeCompleteName, mc))
					continue
				}
				if result.Event.Type == "replicator_completed" {
					eb.Publish(event.NewMessage(event.ReplicatorCompletedName, nil))
					continue
				}
				if result.Event.Type == "topic_peer_event" {
					eb.Publish(event.NewMessage(event.TopicPeerEventName, event.TopicPeerEvent{
						PeerID:    result.Event.PeerID,
						Topic:     result.Event.Topic,
						EventType: result.Event.EventType,
					}))
					continue
				}
				if result.Event.Type == "se_artifact_received" {
					eb.Publish(event.NewMessage(event.SEArtifactReceivedName, event.SEArtifactReceived{
						DocID: result.Event.DocID,
					}))
					continue
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	w := &Wrapper{
		node:            node,
		events:          eb,
		stopMergePoller: stopCh,
	}
	if nodeIdentity != nil {
		raw := nodeIdentity.ToPublicRawIdentity()
		w.nodeIdentityRaw = &raw
	}
	return w, nil
}

// EnableNACForInit enables NAC on the underlying Rust FFI node with the given
// owner DID. This mirrors Go's initializeNodeACP() and should only be called
// during test setup to ensure the Rust FFI node has the same NAC state as the
// Go node it's being tested alongside.
func (w *Wrapper) EnableNACForInit(ownerDID string) error {
	return w.node.EnableNAC(ownerDID)
}

// AddNACAdminForInit grants admin access to the target DID on the Rust FFI
// node. The requestor must be the current NAC owner. This is used during test
// setup to mirror Go's nodeIdentity shortcut (checkNodeAccess grants automatic
// access to the node's own identity).
func (w *Wrapper) AddNACAdminForInit(ownerDID, targetDID string) error {
	_, err := w.node.AddNACActorRelationship(ownerDID, "admin", targetDID)
	return err
}

// groupPatchByCollection groups a JSON patch into per-collection patches.
// Patch format: [{"op": "add", "path": "/CollectionName/Fields/-", "value": {...}}]
// Returns an ordered list of (collection_name, patch_json) pairs preserving
// the first-seen order of collection names.
func groupPatchByCollection(patch string) ([]struct{ Name, Patch string }, error) {
	var ops []json.RawMessage
	if err := json.Unmarshal([]byte(patch), &ops); err != nil {
		return nil, fmt.Errorf("failed to parse patch JSON: %w", err)
	}
	if len(ops) == 0 {
		return nil, fmt.Errorf("patch contains no operations")
	}

	// Group raw operations by collection name, preserving order
	groups := map[string][]json.RawMessage{}
	var order []string
	for i, raw := range ops {
		var op struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(raw, &op); err != nil {
			return nil, fmt.Errorf("failed to parse patch operation %d: %w", i, err)
		}
		path := op.Path
		if len(path) == 0 {
			return nil, fmt.Errorf("invalid patch path in operation %d: empty path", i)
		}
		// Go DefraDB accepts paths with or without leading '/'.
		// Strip leading '/' if present for consistent collection name extraction.
		if path[0] == '/' {
			path = path[1:]
		}
		name := path
		for j, c := range path {
			if c == '/' {
				name = path[:j]
				break
			}
		}
		if _, exists := groups[name]; !exists {
			order = append(order, name)
		}
		groups[name] = append(groups[name], raw)
	}

	result := make([]struct{ Name, Patch string }, 0, len(order))
	for _, name := range order {
		patchBytes, err := json.Marshal(groups[name])
		if err != nil {
			return nil, fmt.Errorf("failed to marshal patch for %s: %w", name, err)
		}
		result = append(result, struct{ Name, Patch string }{name, string(patchBytes)})
	}
	return result, nil
}

// ============================================================================
// clients.Client interface
// ============================================================================

// RetryReplicators re-pushes all existing documents to connected replicator peers.
func (w *Wrapper) RetryReplicators() error {
	if w.node == nil {
		return nil
	}
	return w.node.P2PRetryReplicators()
}

// SetSEEncryptionKey passes the searchable encryption key to the Rust FFI node.
func (w *Wrapper) SetSEEncryptionKey(key []byte) error {
	if w.node == nil {
		return fmt.Errorf("node not initialized")
	}
	return w.node.SetSEEncryptionKey(key)
}

// SetGoNodeCloser sets a callback to close the Go node during wrapper Close().
// This ensures the Go node's badger lock is released on restart.
func (w *Wrapper) SetGoNodeCloser(closer func()) {
	w.goNodeCloser = closer
}

func (w *Wrapper) Close() {
	if w.stopSEForwarder != nil {
		close(w.stopSEForwarder)
		w.stopSEForwarder = nil
	}
	if w.stopMergePoller != nil {
		close(w.stopMergePoller)
		w.stopMergePoller = nil
	}
	if w.node != nil {
		globalRustFFIP2PRegistry.unregister(w)
		_ = w.node.Close()
		w.node = nil
	}
	if w.events != nil {
		w.events.Close()
		w.events = nil
	}
	if w.goNodeCloser != nil {
		w.goNodeCloser()
		w.goNodeCloser = nil
	}
}

// ForwardSEEvents subscribes to Go's event bus for SE artifact received events
// and forwards them to the wrapper's custom event bus. This bridges Go's SE
// coordinator events to the Rust FFI wrapper so WaitForSESync works.
func (w *Wrapper) ForwardSEEvents(goBus event.Bus) error {
	sub, err := goBus.Subscribe(event.SEArtifactReceivedName)
	if err != nil {
		return fmt.Errorf("failed to subscribe to SE events on Go bus: %w", err)
	}

	stopCh := make(chan struct{})
	w.stopSEForwarder = stopCh

	go func() {
		defer goBus.Unsubscribe(sub)
		for {
			select {
			case <-stopCh:
				return
			case msg, ok := <-sub.Message():
				if !ok {
					return
				}
				w.events.Publish(msg)
			}
		}
	}()

	return nil
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
	opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	// If the context carries a transaction (set by db.InitContext), delegate
	// to the transaction-scoped ExecRequest so the Rust FFI executes within
	// that transaction's snapshot.
	if clientTxn, ok := datastore.CtxTryGetClientTxn(ctx); ok && clientTxn != nil {
		if txnW, ok := clientTxn.(*TxnWrapper); ok {
			return txnW.ExecRequest(ctx, request, opts...)
		}
	}

	opt := utils.NewOptions(opts...)

	varsJSON := ""
	if opt.Variables != nil {
		varsBytes, err := json.Marshal(opt.Variables)
		if err != nil {
			return &client.RequestResult{
				GQL: client.GQLResult{
					Errors: []error{fmt.Errorf("failed to marshal variables: %w", err)},
				},
			}
		}
		varsJSON = string(varsBytes)
	}

	identityDID := identityDIDFromOption(opt.GetIdentity())

	operationName := ""
	if opt.OperationName.HasValue() {
		operationName = opt.OperationName.Value()
	}
	execResult, err := w.node.ExecRequestFull(identityDID, request, operationName, varsJSON)
	if err != nil {
		return &client.RequestResult{
			GQL: client.GQLResult{
				Errors: []error{err},
			},
		}
	}

	// Handle subscription results
	if execResult.IsSubscription {
		subChan := make(chan client.GQLResult)
		go w.pollGraphQLSubscription(ctx, execResult.SubscriptionID, subChan)
		return &client.RequestResult{
			Subscription: subChan,
		}
	}

	// Use json.Decoder with UseNumber() to preserve numeric precision.
	// This ensures integers stay as json.Number rather than being converted to float64,
	// which is required for test assertions to pass.
	var gqlResult client.GQLResult
	decoder := json.NewDecoder(bytes.NewReader([]byte(execResult.Response)))
	decoder.UseNumber()
	if err := decoder.Decode(&gqlResult); err != nil {
		return &client.RequestResult{
			GQL: client.GQLResult{
				Errors: []error{fmt.Errorf("failed to parse response: %w", err)},
			},
		}
	}

	// Post-process: convert DateTime strings to time.Time objects so test
	// assertions can compare them with MustParseTime/CurrentTimestamp values.
	if gqlResult.Data != nil {
		gqlResult.Data = convertDateTimeStrings(gqlResult.Data)
		gqlResult.Data = decodeBase64Deltas(gqlResult.Data)
		if isExplainResult(gqlResult.Data) {
			gqlResult.Data = normalizeExplainTypes(gqlResult.Data)
		}
	}

	// Rust's merge handler emits update events via the event bus, which the
	// bridge goroutine forwards to Go's event bus. No need to emit manually
	// here — doing so would produce duplicate events that desynchronize the
	// test framework's waitForUpdateEvents channel.

	return &client.RequestResult{GQL: gqlResult}
}

// emitMutationEvents parses GQL mutation results and publishes update events
// for each affected document.
func (w *Wrapper) emitMutationEvents(ctx context.Context, data any) {
	dataMap, ok := data.(map[string]any)
	if !ok {
		return
	}
	for key, value := range dataMap {
		var collectionName string
		if strings.HasPrefix(key, "create_") {
			collectionName = strings.TrimPrefix(key, "create_")
		} else if strings.HasPrefix(key, "update_") {
			collectionName = strings.TrimPrefix(key, "update_")
		} else if strings.HasPrefix(key, "delete_") {
			collectionName = strings.TrimPrefix(key, "delete_")
		} else {
			continue
		}

		col, err := w.GetCollectionByName(ctx, collectionName)
		if err != nil {
			continue
		}
		cw, ok := col.(*CollectionWrapper)
		if !ok {
			continue
		}

		docs, ok := value.([]any)
		if !ok {
			continue
		}
		for _, d := range docs {
			doc, ok := d.(map[string]any)
			if !ok {
				continue
			}
			docID, _ := doc["_docID"].(string)
			if docID == "" {
				continue
			}
			compositeCid := cw.getLatestCompositeCID(ctx, docID)
			w.events.Publish(event.NewMessage(event.UpdateName, event.Update{
				DocID:        docID,
				CollectionID: cw.version.CollectionID,
				Cid:          compositeCid,
			}))

			if cw.version.IsBranchable {
				w.events.Publish(event.NewMessage(event.UpdateName, event.Update{
					DocID:        "",
					CollectionID: cw.version.CollectionID,
					Cid:          compositeCid,
				}))
			}
		}
	}
}

// pollGraphQLSubscription polls a GraphQL subscription and sends results to the channel.
func (w *Wrapper) pollGraphQLSubscription(ctx context.Context, subscriptionID string, ch chan client.GQLResult) {
	defer close(ch)
	defer func() { _ = CloseGraphQLSubscription(subscriptionID) }()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result, err := PollGraphQLSubscription(subscriptionID)
		if err != nil {
			select {
			case ch <- client.GQLResult{Errors: []error{err}}:
			case <-ctx.Done():
			}
			return
		}

		if result.IsClosed {
			return
		}

		if result.HasResult {
			var gqlResult client.GQLResult
			decoder := json.NewDecoder(bytes.NewReader([]byte(result.Result)))
			decoder.UseNumber()
			if err := decoder.Decode(&gqlResult); err != nil {
				select {
				case ch <- client.GQLResult{Errors: []error{fmt.Errorf("failed to parse subscription result: %w", err)}}:
				case <-ctx.Done():
				}
				continue
			}

			// Post-process: convert DateTime strings and base64 deltas
			if gqlResult.Data != nil {
				gqlResult.Data = convertDateTimeStrings(gqlResult.Data)
				gqlResult.Data = decodeBase64Deltas(gqlResult.Data)
			}

			select {
			case ch <- gqlResult:
			case <-ctx.Done():
				return
			}
		}

		// Small sleep to avoid busy-waiting
		time.Sleep(10 * time.Millisecond)
	}
}

func (w *Wrapper) AddCollection(ctx context.Context, sdl string, opts ...options.Enumerable[options.AddCollectionOptions]) ([]client.CollectionVersion, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	responseJSON, err := w.node.AddSchema(identityDID, sdl)
	if err != nil {
		return nil, err
	}

	var versions []client.CollectionVersion
	if err := json.Unmarshal([]byte(responseJSON), &versions); err != nil {
		return nil, fmt.Errorf("failed to parse schema response: %w", err)
	}

	return versions, nil
}

func (w *Wrapper) GetCollectionByName(ctx context.Context, name client.CollectionName, opts ...options.Enumerable[options.GetCollectionByNameOptions]) (client.Collection, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	responseJSON, err := w.node.GetCollectionByName(identityDID, name)
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
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	// If VersionID is specified, use the dedicated FFI function which returns
	// "key not found" for missing versions (matches Go behavior).
	if opt.VersionID.HasValue() {
		versionJSON, err := w.node.GetCollectionByVersionID(identityDID, opt.VersionID.Value())
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "key not found") {
				return nil, fmt.Errorf("collection not found")
			}
			return nil, err
		}
		if versionJSON == nullStr || versionJSON == "" {
			return nil, fmt.Errorf("collection not found")
		}
		var version client.CollectionVersion
		if err := json.Unmarshal([]byte(versionJSON), &version); err != nil {
			return nil, fmt.Errorf("failed to parse collection version: %w", err)
		}
		return []client.Collection{&CollectionWrapper{wrapper: w, version: version}}, nil
	}

	// If the context carries a transaction, use the transaction-aware function
	// so we can see uncommitted writes (e.g., placeholders from set_migration_in_txn).
	var responseJSON string
	var err error
	if clientTxn, ok := datastore.CtxTryGetClientTxn(ctx); ok && clientTxn != nil {
		if txnW, ok := clientTxn.(*TxnWrapper); ok {
			responseJSON, err = w.node.GetCollectionsInTxn(txnW.txn.id, identityDID)
		} else {
			responseJSON, err = w.node.GetCollections(identityDID)
		}
	} else {
		responseJSON, err = w.node.GetCollections(identityDID)
	}
	if err != nil {
		return nil, err
	}

	var versions []client.CollectionVersion
	if err := json.Unmarshal([]byte(responseJSON), &versions); err != nil {
		return nil, fmt.Errorf("failed to parse collections: %w", err)
	}

	// Apply filters
	includeInactive := opt.GetInactive.HasValue() && opt.GetInactive.Value()
	var filtered []client.CollectionVersion
	for _, v := range versions {
		if !includeInactive && !v.IsActive {
			continue
		}
		if opt.CollectionName.HasValue() && v.Name != opt.CollectionName.Value() {
			continue
		}
		if opt.CollectionID.HasValue() && v.CollectionID != opt.CollectionID.Value() {
			continue
		}
		if opt.CollectionSetID.HasValue() {
			if !v.CollectionSet.HasValue() || v.CollectionSet.Value().CollectionSetID != opt.CollectionSetID.Value() {
				continue
			}
		}
		filtered = append(filtered, v)
	}

	// When filtering by CollectionID, Go returns the root version (whose VersionID == CollectionID)
	// first, then remaining versions in KV order. Rust returns all versions in KV key order.
	// Reorder to match Go's GetCollectionVersionIDs behavior.
	if opt.CollectionID.HasValue() {
		rootID := opt.CollectionID.Value()
		var root []client.CollectionVersion
		var rest []client.CollectionVersion
		for _, v := range filtered {
			if v.VersionID == rootID {
				root = append(root, v)
			} else {
				rest = append(rest, v)
			}
		}
		filtered = append(root, rest...)
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

func (w *Wrapper) SetActiveCollectionVersion(ctx context.Context, versionID string, opts ...options.Enumerable[options.SetActiveCollectionVersionOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	return w.setActiveCollectionVersion(identityDID, versionID)
}

func (w *Wrapper) setActiveCollectionVersion(identityDID string, versionID string) error {
	return w.node.SetActiveCollectionVersion(identityDID, versionID)
}

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[lensmodel.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	return w.patchCollection(identityDID, patch, migration)
}

func (w *Wrapper) DeleteCollection(
	ctx context.Context,
	names []string,
	opts ...options.Enumerable[options.DeleteCollectionOptions],
) error {
	opt := utils.NewOptions(opts...)
	ident := opt.GetIdentity()

	patch, err := buildDeleteCollectionPatch(
		names,
		opt.ActiveOnly,
		func(name string) (client.Collection, error) {
			getOpts := options.GetCollectionByName()
			if ident.HasValue() {
				getOpts.SetIdentity(ident.Value())
			}
			return w.GetCollectionByName(ctx, client.CollectionName(name), getOpts)
		},
		func(collectionID string) ([]client.Collection, error) {
			getOpts := options.GetCollections().
				SetCollectionID(collectionID).
				SetGetInactive(true)
			if ident.HasValue() {
				getOpts.SetIdentity(ident.Value())
			}
			return w.GetCollections(ctx, getOpts)
		},
	)
	if err != nil {
		return err
	}

	return w.patchCollection(identityDIDFromOption(ident), patch, immutable.None[lensmodel.Lens]())
}

func (w *Wrapper) patchCollection(
	identityDID string,
	patch string,
	migration immutable.Option[lensmodel.Lens],
) error {
	// Parse all operations to classify into collection-level removes vs modifications.
	// Go's patchCollection operates on a global dict of all versions atomically.
	// The FFI wrapper must coordinate these since Rust operates per-collection.
	var allOps []json.RawMessage
	if err := json.Unmarshal([]byte(patch), &allOps); err != nil {
		return fmt.Errorf("failed to parse patch JSON: %w", err)
	}

	type patchOp struct {
		Op   string `json:"op"`
		Path string `json:"path"`
		From string `json:"from,omitempty"`
	}

	var removeTargets []string // collection names or version IDs to delete
	var modOps []json.RawMessage
	var hasCopy bool

	for _, raw := range allOps {
		var op patchOp
		if err := json.Unmarshal(raw, &op); err != nil {
			return fmt.Errorf("failed to parse patch operation: %w", err)
		}

		if op.Op == "remove" {
			// Collection-level: path has no subpath (e.g., "/Users", "/bafyrei...")
			// Field-level: path has subpath (e.g., "/Users/Name", "/Users/Fields/0")
			target := strings.TrimPrefix(op.Path, "/")
			if strings.Contains(target, "/") {
				modOps = append(modOps, raw)
			} else {
				removeTargets = append(removeTargets, target)
			}
		} else {
			modOps = append(modOps, raw)
			if op.Op == copyOp {
				hasCopy = true
			}
		}
	}

	// No collection-level removes — use the original grouping logic
	if len(removeTargets) == 0 {
		return w.applyPatchGroups(identityDID, patch, migration)
	}

	// Handle copy+add+remove pattern: rewrite as add-field on source, no deletion.
	// In Go, copy creates a dict entry, add modifies it, remove deletes the source.
	// The net effect is identical to add-field on the source collection.
	if hasCopy && len(modOps) > 0 {
		var copyPatchOp patchOp
		var addOps []json.RawMessage
		for _, raw := range modOps {
			var op patchOp
			if err := json.Unmarshal(raw, &op); err != nil {
				continue
			}
			if op.Op == copyOp {
				copyPatchOp = op
			} else {
				addOps = append(addOps, raw)
			}
		}

		if copyPatchOp.Op == copyOp {
			copyFrom := strings.TrimPrefix(copyPatchOp.From, "/")
			if idx := strings.Index(copyFrom, "/"); idx != -1 {
				copyFrom = copyFrom[:idx]
			}
			copyTo := strings.TrimPrefix(copyPatchOp.Path, "/")
			if idx := strings.Index(copyTo, "/"); idx != -1 {
				copyTo = copyTo[:idx]
			}

			isSourceRemoved := false
			for _, target := range removeTargets {
				if target == copyFrom {
					isSourceRemoved = true
					break
				}
			}

			if isSourceRemoved && len(addOps) > 0 {
				// Rewrite add ops targeting the copy destination to target the source
				var rewrittenOps []json.RawMessage
				for _, raw := range addOps {
					rawStr := string(raw)
					rawStr = strings.ReplaceAll(rawStr, "/"+copyTo+"/", "/"+copyFrom+"/")
					rawStr = strings.ReplaceAll(rawStr, "\""+copyTo+"\"", "\""+copyFrom+"\"")
					rewrittenOps = append(rewrittenOps, json.RawMessage(rawStr))
				}
				rewrittenPatch, err := json.Marshal(rewrittenOps)
				if err != nil {
					return fmt.Errorf("failed to marshal rewritten patch: %w", err)
				}
				// Apply as add-field. The old version becomes inactive automatically.
				_, err = w.node.PatchCollection(identityDID, copyFrom, string(rewrittenPatch))
				return err
			}
		}
	}

	// Resolve remove targets to version IDs BEFORE applying any mods.
	// This captures the current state so we know which versions to delete.
	removeVersionIDs, err := w.resolveRemoveTargets(identityDID, removeTargets)
	if err != nil {
		return err
	}
	if err := w.validateCollectionRemovals(identityDID, removeVersionIDs); err != nil {
		return err
	}

	// Build a set of version IDs being deleted, so we can skip mods targeting them.
	// In Go's global dict model, if an entry is both modified and removed, the remove wins.
	removeSet := make(map[string]bool, len(removeVersionIDs))
	for _, vid := range removeVersionIDs {
		removeSet[vid] = true
	}

	// Filter mod groups: skip any whose target collection's version_id is being deleted.
	var filteredModOps []json.RawMessage
	if len(modOps) > 0 {
		for _, raw := range modOps {
			var op patchOp
			if err := json.Unmarshal(raw, &op); err != nil {
				filteredModOps = append(filteredModOps, raw)
				continue
			}
			// Extract collection name from the path
			target := strings.TrimPrefix(op.Path, "/")
			if idx := strings.Index(target, "/"); idx != -1 {
				target = target[:idx]
			}
			// Resolve target to version_id
			targetVID := w.resolveToVersionID(identityDID, target)
			if removeSet[targetVID] {
				continue // Skip: this collection is being deleted
			}
			filteredModOps = append(filteredModOps, raw)
		}
	}

	// Execute batch delete FIRST (before mods).
	// This is needed for cases like remove+reactivate (Test 4) where the
	// reactivation can only succeed after the conflicting active version is deleted.
	if len(removeVersionIDs) > 0 {
		if err := w.node.DeleteCollectionVersions(identityDID, removeVersionIDs); err != nil {
			return err
		}
	}

	// Apply remaining modification ops
	if len(filteredModOps) > 0 {
		modPatch, err := json.Marshal(filteredModOps)
		if err != nil {
			return fmt.Errorf("failed to marshal modification ops: %w", err)
		}
		return w.applyPatchGroups(identityDID, string(modPatch), migration)
	}

	return nil
}

// applyPatchGroups groups a patch by collection and applies each group.
func (w *Wrapper) applyPatchGroups(
	identityDID string,
	patch string,
	migration immutable.Option[lensmodel.Lens],
) error {
	groups, err := groupPatchByCollection(patch)
	if err != nil {
		return err
	}
	for _, g := range groups {
		var oldVersionID string
		if migration.HasValue() {
			oldCollJSON, err := w.node.GetCollectionByName(identityDID, g.Name)
			if err == nil {
				var oldVersion client.CollectionVersion
				if err := json.Unmarshal([]byte(oldCollJSON), &oldVersion); err == nil {
					oldVersionID = oldVersion.VersionID
				}
			}
		}

		newSchemaJSON, err := w.node.PatchCollection(identityDID, g.Name, g.Patch)
		if err != nil {
			return err
		}

		if migration.HasValue() && oldVersionID != "" {
			var newVersion client.CollectionVersion
			if err := json.Unmarshal([]byte(newSchemaJSON), &newVersion); err != nil {
				return fmt.Errorf("failed to parse new collection version: %w", err)
			}
			if newVersion.VersionID != "" {
				config := client.LensConfig{
					SourceCollectionVersionID:      oldVersionID,
					DestinationCollectionVersionID: newVersion.VersionID,
					Lens:                           migration.Value(),
				}
				configJSON, err := json.Marshal(config)
				if err != nil {
					return fmt.Errorf("failed to marshal migration config: %w", err)
				}
				if _, err := w.node.SetMigration(identityDID, string(configJSON)); err != nil {
					return fmt.Errorf("failed to set migration after patch: %w", err)
				}
			}
		}
	}
	return nil
}

func buildDeleteCollectionPatch(
	names []string,
	activeOnly bool,
	getByName func(string) (client.Collection, error),
	getByCollectionID func(string) ([]client.Collection, error),
) (string, error) {
	if len(names) == 0 {
		return "", client.ErrCollectionNameRequired
	}

	seen := make(map[string]struct{}, len(names))
	ops := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		col, err := getByName(name)
		if err != nil {
			return "", err
		}

		if activeOnly {
			ops = append(ops, deleteCollectionPatchOp(col.VersionID()))
			continue
		}

		versions, err := getByCollectionID(col.CollectionID())
		if err != nil {
			return "", err
		}
		for _, version := range versions {
			ops = append(ops, deleteCollectionPatchOp(version.VersionID()))
		}
	}

	return "[" + strings.Join(ops, ",") + "]", nil
}

func deleteCollectionPatchOp(versionID string) string {
	return fmt.Sprintf(`{"op": "remove", "path": "/%s"}`, versionID)
}

func (w *Wrapper) collectionVersions(identityDID string) ([]client.CollectionVersion, error) {
	collectionsJSON, err := w.node.GetCollections(identityDID)
	if err != nil {
		return nil, err
	}

	var versions []client.CollectionVersion
	if err := json.Unmarshal([]byte(collectionsJSON), &versions); err != nil {
		return nil, fmt.Errorf("failed to parse collections: %w", err)
	}
	return versions, nil
}

func (w *Wrapper) validateCollectionRemovals(identityDID string, removeVersionIDs []string) error {
	if len(removeVersionIDs) == 0 {
		return nil
	}

	collections, err := w.collectionVersions(identityDID)
	if err != nil {
		return err
	}

	removed := make(map[string]struct{}, len(removeVersionIDs))
	for _, versionID := range removeVersionIDs {
		removed[versionID] = struct{}{}
	}

	remaining := make([]client.CollectionVersion, 0, len(collections))
	for _, collection := range collections {
		if _, ok := removed[collection.VersionID]; ok {
			continue
		}
		remaining = append(remaining, collection)
	}

	oldState := newCollectionDefinitionState(collections)
	newState := newCollectionDefinitionState(remaining)
	for _, collection := range remaining {
		for _, field := range collection.Fields {
			if !field.Kind.IsObject() {
				continue
			}
			if _, ok := newState.getCollection(collection, field.Kind); ok {
				continue
			}
			if _, ok := oldState.getCollection(collection, field.Kind); ok {
				return fmt.Errorf("cannot remove a collection while another field references it")
			}
		}
	}

	return nil
}

type collectionDefinitionState struct {
	collections              []client.CollectionVersion
	collectionsByID          map[string]client.CollectionVersion
	activeCollectionsByName  map[string]client.CollectionVersion
	activeCollectionsByColID map[string]client.CollectionVersion
}

func newCollectionDefinitionState(collections []client.CollectionVersion) collectionDefinitionState {
	state := collectionDefinitionState{
		collections:              collections,
		collectionsByID:          map[string]client.CollectionVersion{},
		activeCollectionsByName:  map[string]client.CollectionVersion{},
		activeCollectionsByColID: map[string]client.CollectionVersion{},
	}

	for _, collection := range collections {
		if collection.IsActive {
			state.activeCollectionsByName[collection.Name] = collection
			state.activeCollectionsByColID[collection.CollectionID] = collection
		}
		if collection.VersionID != "" {
			state.collectionsByID[collection.VersionID] = collection
		}
	}

	return state
}

func (s collectionDefinitionState) getCollection(
	host client.CollectionVersion,
	kind client.FieldKind,
) (client.CollectionVersion, bool) {
	switch typedKind := kind.(type) {
	case *client.NamedKind:
		if collection, ok := s.activeCollectionsByName[typedKind.Name]; ok {
			return collection, true
		}
		for _, collection := range s.collections {
			if collection.Name == typedKind.Name {
				return collection, true
			}
		}
	case *client.CollectionKind:
		if collection, ok := s.activeCollectionsByColID[typedKind.CollectionID]; ok {
			return collection, true
		}
		collection, ok := s.collectionsByID[typedKind.CollectionID]
		return collection, ok
	case *client.SelfKind:
		if typedKind.RelativeID == "" {
			return host, true
		}
		if !host.CollectionSet.HasValue() {
			return client.CollectionVersion{}, false
		}
		hostSet := host.CollectionSet.Value()
		for _, collection := range s.collections {
			if !collection.CollectionSet.HasValue() {
				continue
			}
			collectionSet := collection.CollectionSet.Value()
			if collectionSet.CollectionSetID != hostSet.CollectionSetID {
				continue
			}
			if fmt.Sprint(collectionSet.RelativeID) == typedKind.RelativeID {
				return collection, true
			}
		}
	}

	return client.CollectionVersion{}, false
}

// resolveRemoveTargets resolves collection names or version IDs to version IDs.
// If a target is neither a known collection name nor a version ID (e.g. "EncryptedIndexes"),
// returns the same error Go produces for removing a nonexistent key from the global dict.
func (w *Wrapper) resolveRemoveTargets(identityDID string, targets []string) ([]string, error) {
	var versionIDs []string
	for _, target := range targets {
		collJSON, err := w.node.GetCollectionByName(identityDID, target)
		if err == nil {
			var version client.CollectionVersion
			if err := json.Unmarshal([]byte(collJSON), &version); err == nil {
				versionIDs = append(versionIDs, version.VersionID)
				continue
			}
		}
		// Not a collection name — check if it looks like a version ID (CID).
		if strings.HasPrefix(target, "bafyrei") || strings.HasPrefix(target, "bafkrei") {
			if mapped, ok := w.resolveMissingViewVersionTarget(identityDID, target); ok {
				versionIDs = append(versionIDs, mapped)
				continue
			}
			versionIDs = append(versionIDs, target)
			continue
		}
		// Neither a collection name nor a version ID — Go returns this error.
		return nil, fmt.Errorf("unable to remove nonexistent key")
	}
	return versionIDs, nil
}

func (w *Wrapper) resolveMissingViewVersionTarget(identityDID string, target string) (string, bool) {
	collections, err := w.collectionVersions(identityDID)
	if err != nil {
		return "", false
	}

	for _, collection := range collections {
		if collection.VersionID == target {
			return target, true
		}
	}

	candidates := make([]string, 0, 1)
	for _, collection := range collections {
		if collection.IsActive && collection.Query.HasValue() && collection.VersionID != "" {
			candidates = append(candidates, collection.VersionID)
		}
	}
	if len(candidates) != 1 {
		return "", false
	}
	return candidates[0], true
}

// resolveToVersionID resolves a collection name to its active version ID,
// or returns the input unchanged if it's already a version ID.
func (w *Wrapper) resolveToVersionID(identityDID string, nameOrID string) string {
	collJSON, err := w.node.GetCollectionByName(identityDID, nameOrID)
	if err == nil {
		var version client.CollectionVersion
		if err := json.Unmarshal([]byte(collJSON), &version); err == nil {
			return version.VersionID
		}
	}
	return nameOrID
}

func (w *Wrapper) ListIndexes(ctx context.Context, opts ...options.Enumerable[options.ListIndexesOptions]) (map[client.CollectionName][]client.IndexDescription, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	result, err := w.node.GetAllIndexes(identityDID)
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

func (w *Wrapper) AddDACPolicy(ctx context.Context, policy string, opts ...options.Enumerable[options.AddDACPolicyOptions]) (client.AddPolicyResult, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

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
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	opt := utils.NewOptions(opts...)
	requestorDID := identityDIDFromOption(opt.GetIdentity())
	added, err := w.node.AddDACActorRelationship(requestorDID, targetActor, collectionName, docID, relation)
	if err != nil {
		return client.AddActorRelationshipResult{},
			fmt.Errorf("failed to add document actor relationship with acp: %w", err)
	}

	// Emit update event when relationship is newly added (matches Go behavior)
	if added {
		w.publishRelationshipEvent(ctx, requestorDID, collectionName, docID)
	}

	return client.AddActorRelationshipResult{ExistedAlready: !added}, nil
}

func (w *Wrapper) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	opt := utils.NewOptions(opts...)
	requestorDID := identityDIDFromOption(opt.GetIdentity())
	deleted, err := w.node.DeleteDACActorRelationship(requestorDID, targetActor, collectionName, docID, relation)
	if err != nil {
		return client.DeleteActorRelationshipResult{},
			fmt.Errorf("failed to delete document actor relationship with acp: %w", err)
	}

	// Emit update event when relationship is deleted (matches Go behavior)
	if deleted {
		w.publishRelationshipEvent(ctx, requestorDID, collectionName, docID)
	}

	return client.DeleteActorRelationshipResult{RecordFound: deleted}, nil
}

// publishRelationshipEvent emits an update event after a DAC relationship change.
// This matches Go DefraDB behavior where relationship add/delete triggers an update event
// so the test framework's waitForUpdateEvents can synchronize.
// The requestorDID is needed to authenticate the collection lookup when NAC is enabled.
func (w *Wrapper) publishRelationshipEvent(ctx context.Context, requestorDID string, collectionName string, docID string) {
	responseJSON, err := w.node.GetCollectionByName(requestorDID, collectionName)
	if err != nil {
		return
	}

	var version client.CollectionVersion
	if err := json.Unmarshal([]byte(responseJSON), &version); err != nil {
		return
	}

	cw := &CollectionWrapper{
		wrapper: w,
		version: version,
	}

	compositeCid := cw.getLatestCompositeCID(ctx, docID)
	w.events.Publish(event.NewMessage(event.UpdateName, event.Update{
		DocID:        docID,
		CollectionID: cw.version.CollectionID,
		Cid:          compositeCid,
	}))
}

func (w *Wrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	opt := utils.NewOptions(opts...)
	requestorDID := identityDIDFromOption(opt.GetIdentity())
	added, err := w.node.AddNACActorRelationship(requestorDID, relation, targetActor)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	return client.AddActorRelationshipResult{ExistedAlready: !added}, nil
}

func (w *Wrapper) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	opt := utils.NewOptions(opts...)
	requestorDID := identityDIDFromOption(opt.GetIdentity())
	deleted, err := w.node.DeleteNACActorRelationship(requestorDID, relation, targetActor)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return client.DeleteActorRelationshipResult{RecordFound: deleted}, nil
}

func (w *Wrapper) ReEnableNAC(ctx context.Context, opts ...options.Enumerable[options.ReEnableNACOptions]) error {
	opt := utils.NewOptions(opts...)
	requestorDID := identityDIDFromOption(opt.GetIdentity())
	return w.node.ReEnableNAC(requestorDID)
}

func (w *Wrapper) DisableNAC(ctx context.Context, opts ...options.Enumerable[options.DisableNACOptions]) error {
	opt := utils.NewOptions(opts...)
	requestorDID := identityDIDFromOption(opt.GetIdentity())
	return w.node.DisableNAC(requestorDID)
}

func (w *Wrapper) GetNACStatus(ctx context.Context, opts ...options.Enumerable[options.GetNACStatusOptions]) (client.NACStatusResult, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	status, err := w.node.GetNACStatus(identityDID)
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
	// Return the cached identity which includes the PublicKey (hex-encoded).
	// The Rust FFI only returns the DID, but the test framework expects both.
	if w.nodeIdentityRaw != nil && w.nodeIdentityRaw.DID == did {
		return immutable.Some(*w.nodeIdentityRaw), nil
	}
	return immutable.Some(identity.PublicRawIdentity{DID: did}), nil
}

func (w *Wrapper) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey, opts ...options.Enumerable[options.VerifySignatureOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	return w.node.BlockVerifySignature(string(pubKey.Type()), pubKey.String(), blockCid, identityDID)
}

// ============================================================================
// client.Store interface - View/Migration methods
// ============================================================================

func (w *Wrapper) AddView(
	ctx context.Context,
	gqlQuery string,
	sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	opt := utils.NewOptions(opts...)
	transformStr := ""
	if opt.TransformCID.HasValue() {
		transformStr = opt.TransformCID.Value()
	}

	identityDID := identityDIDFromOption(opt.GetIdentity())

	responseJSON, err := w.node.AddView(identityDID, gqlQuery, sdl, transformStr)
	if err != nil {
		return nil, err
	}

	var versions []client.CollectionVersion
	if err := json.Unmarshal([]byte(responseJSON), &versions); err != nil {
		return nil, fmt.Errorf("failed to parse view response: %w", err)
	}

	return versions, nil
}

func (w *Wrapper) RefreshViews(ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	optsJSON := ""
	if opt.CollectionName.HasValue() || opt.VersionID.HasValue() {
		data, err := json.Marshal(opt)
		if err != nil {
			return fmt.Errorf("failed to marshal options: %w", err)
		}
		optsJSON = string(data)
	}
	return w.node.RefreshViews(identityDID, optsJSON)
}

func (w *Wrapper) SetMigration(ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions]) (string, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	// If the context carries a transaction, use the transaction-aware function
	// so the migration is part of that transaction and only visible after commit.
	if clientTxn, ok := datastore.CtxTryGetClientTxn(ctx); ok && clientTxn != nil {
		if txnW, ok := clientTxn.(*TxnWrapper); ok {
			return w.node.SetMigrationInTxn(txnW.txn.id, identityDID, string(configJSON))
		}
	}

	return w.node.SetMigration(identityDID, string(configJSON))
}

func (w *Wrapper) AddLens(ctx context.Context, lens lensmodel.Lens, opts ...options.Enumerable[options.AddLensOptions]) (string, error) {
	lensJSON, err := json.Marshal(lens)
	if err != nil {
		return "", fmt.Errorf("failed to marshal lens: %w", err)
	}
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	return w.node.LensAdd(identityDID, string(lensJSON))
}

func (w *Wrapper) ListLenses(ctx context.Context, opts ...options.Enumerable[options.ListLensesOptions]) (map[string]lensmodel.Lens, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	raw, err := w.node.LensList(identityDID)
	if err != nil {
		return nil, err
	}

	// Rust returns map[string]LensModule (single module per ID).
	// Go expects map[string]model.Lens (which wraps []LensModule).
	var rustMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &rustMap); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse lens list: %w", err)
	}

	result := make(map[string]lensmodel.Lens, len(rustMap))
	for id, moduleJSON := range rustMap {
		var module lensmodel.LensModule
		if err := json.Unmarshal(moduleJSON, &module); err != nil {
			return nil, fmt.Errorf("ffi: failed to parse lens module %s: %w", id, err)
		}
		result[id] = lensmodel.Lens{
			Lenses: []lensmodel.LensModule{module},
		}
	}
	return result, nil
}

// ============================================================================
// client.Store interface - Backup methods
// ============================================================================

func (w *Wrapper) BasicImport(ctx context.Context, filepath string) error {
	return w.node.BasicImportDB(filepath)
}

func (w *Wrapper) BasicExport(ctx context.Context, filepath string, opts ...options.Enumerable[options.BasicExportOptions]) error {
	opt := utils.NewOptions(opts...)
	config := &client.BackupConfig{
		Filepath:    filepath,
		Format:      opt.Format,
		Pretty:      opt.Pretty,
		Collections: opt.Collections,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("ffi: failed to marshal backup config: %w", err)
	}
	return w.node.BasicExportDB(string(configJSON))
}

// ============================================================================
// client.Store interface - Utility methods
// ============================================================================

func (w *Wrapper) PrintDump(ctx context.Context) error {
	return fmt.Errorf("printDump not yet implemented in FFI")
}

func (w *Wrapper) ListAllEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListAllEncryptedIndexesOptions],
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	ffiIndexes, err := w.node.ListAllEncryptedIndexes(identityDID)
	if err != nil {
		return nil, err
	}

	result := make(map[client.CollectionName][]client.EncryptedIndexDescription)
	for collName, indexes := range ffiIndexes {
		clientIndexes := make([]client.EncryptedIndexDescription, len(indexes))
		for i, idx := range indexes {
			clientIndexes[i] = client.EncryptedIndexDescription{
				FieldName: idx.FieldName,
				Type:      client.EncryptedIndexType(idx.Type),
			}
		}
		result[collName] = clientIndexes
	}
	return result, nil
}

// ============================================================================
// client.P2P interface
// ============================================================================

func (w *Wrapper) PeerInfo(ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions]) ([]string, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	addrs, err := w.node.P2PPeerInfo(identityDID)
	if err != nil {
		// Propagate authorization errors
		if strings.Contains(err.Error(), "not authorized") {
			return nil, err
		}
		// Return empty addresses when P2P is not enabled
		return []string{}, nil
	}
	globalRustFFIP2PRegistry.register(w, addrs)
	return addrs, nil
}

func (w *Wrapper) ActivePeers(ctx context.Context, opts ...options.Enumerable[options.ActivePeersOptions]) ([]string, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	return w.node.P2PActivePeers(identityDID)
}

func (w *Wrapper) Connect(ctx context.Context, addresses []string, opts ...options.Enumerable[options.ConnectOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	for _, addr := range addresses {
		if err := w.node.P2PConnect(identityDID, addr); err != nil {
			return err
		}
	}

	if sourceAddrs, err := w.node.P2PPeerInfo(identityDID); err == nil {
		globalRustFFIP2PRegistry.register(w, sourceAddrs)
		component := globalRustFFIP2PRegistry.meshComponent(w, sourceAddrs, addresses)
		for _, source := range component {
			if source.w == nil || source.w.node == nil {
				continue
			}
			for _, target := range component {
				if source.w == target.w {
					continue
				}
				for _, addr := range target.addrs {
					_ = source.w.node.P2PConnect(source.identityDID, addr)
				}
			}
		}
	}

	// Allow GossipSub mesh formation after connecting.
	// Without this delay, simultaneous updates from multiple peers may not
	// propagate because the mesh hasn't stabilized yet.
	// GossipSub requires multiple heartbeat cycles (1s each) to graft mesh
	// links, especially in 3+ node topologies where transitive discovery
	// adds latency. Three seconds covers three heartbeat cycles which is
	// sufficient for full mesh stabilization in CRDT convergence tests.
	time.Sleep(3 * time.Second)

	return nil
}

func (w *Wrapper) AddReplicator(ctx context.Context, addresses []string, opts ...options.Enumerable[options.AddReplicatorOptions]) error {
	if len(addresses) == 0 {
		return fmt.Errorf("at least one address is required")
	}
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	collections := opt.CollectionNames

	for _, addr := range addresses {
		if err := w.node.P2PSetReplicator(identityDID, addr, collections); err != nil {
			return err
		}
	}

	// Allow GossipSub mesh formation after adding replicator.
	// P2PSetReplicator implicitly connects to the target peer; the mesh
	// requires multiple heartbeat cycles to stabilize.
	time.Sleep(3 * time.Second)

	return nil
}

func (w *Wrapper) DeleteReplicator(ctx context.Context, id string, opts ...options.Enumerable[options.DeleteReplicatorOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	collections := opt.CollectionNames

	return w.node.P2PDeleteReplicator(identityDID, id, collections)
}

func (w *Wrapper) ListReplicators(ctx context.Context, opts ...options.Enumerable[options.ListReplicatorsOptions]) ([]client.Replicator, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	replicators, err := w.node.P2PGetAllReplicators(identityDID)
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

func (w *Wrapper) AddP2PCollections(ctx context.Context, collectionNames []string, opts ...options.Enumerable[options.AddP2PCollectionsOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	return w.node.P2PAddCollections(identityDID, collectionNames)
}

func (w *Wrapper) DeleteP2PCollections(ctx context.Context, collectionNames []string, opts ...options.Enumerable[options.DeleteP2PCollectionsOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	return w.node.P2PRemoveCollections(identityDID, collectionNames)
}

func (w *Wrapper) ListP2PCollections(ctx context.Context, opts ...options.Enumerable[options.ListP2PCollectionsOptions]) ([]string, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	collectionIDs, err := w.node.P2PGetAllCollections(identityDID)
	if err != nil {
		return nil, err
	}

	// Rust returns schema root CIDs; Go expects collection names.
	// Resolve CIDs back to names using the collections list.
	if len(collectionIDs) > 0 {
		collectionsJSON, err := w.node.GetCollections(identityDID)
		if err != nil {
			return collectionIDs, nil // fall back to raw IDs on error
		}
		var versions []client.CollectionVersion
		if err := json.Unmarshal([]byte(collectionsJSON), &versions); err != nil {
			return collectionIDs, nil
		}
		cidToName := make(map[string]string, len(versions))
		for _, v := range versions {
			cidToName[v.CollectionID] = v.Name
		}
		// Sort collection IDs lexicographically to match Go's datastore key ordering,
		// then map to names in that order.
		sort.Strings(collectionIDs)
		names := make([]string, 0, len(collectionIDs))
		for _, cid := range collectionIDs {
			if name, ok := cidToName[cid]; ok {
				names = append(names, name)
			} else {
				names = append(names, cid)
			}
		}
		return names, nil
	}

	return collectionIDs, nil
}

func (w *Wrapper) AddP2PDocuments(ctx context.Context, docIDs []string, opts ...options.Enumerable[options.AddP2PDocumentsOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	return w.node.P2PAddDocuments(identityDID, docIDs)
}

func (w *Wrapper) DeleteP2PDocuments(ctx context.Context, docIDs []string, opts ...options.Enumerable[options.DeleteP2PDocumentsOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	return w.node.P2PRemoveDocuments(identityDID, docIDs)
}

func (w *Wrapper) ListP2PDocuments(ctx context.Context, opts ...options.Enumerable[options.ListP2PDocumentsOptions]) ([]string, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	return w.node.P2PGetAllDocuments(identityDID)
}

func (w *Wrapper) SyncDocuments(ctx context.Context, collectionName string, docIDs []string, opts ...options.Enumerable[options.SyncDocumentsOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	return w.node.P2PSyncDocuments(identityDID, collectionName, docIDs)
}

func (w *Wrapper) SyncCollectionVersions(ctx context.Context, versionIDs []string, opts ...options.Enumerable[options.SyncCollectionVersionsOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	return w.node.P2PSyncCollectionVersions(identityDID, versionIDs)
}

func (w *Wrapper) SyncBranchableCollection(ctx context.Context, collectionID string, opts ...options.Enumerable[options.SyncBranchableCollectionOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	return w.node.P2PSyncBranchableCollection(identityDID, collectionID)
}

// ============================================================================
// TxnWrapper implements client.Txn
// ============================================================================

type TxnWrapper struct {
	wrapper                 *Wrapper
	txn                     *Transaction
	id                      uint64
	readOnly                bool
	startTS                 time.Time
	stagedPatchCollections  []stagedPatchCollection
	stagedSetActiveVersions []stagedSetActiveVersion
	stagedViews             []stagedViewAdd
	stagedRefreshes         []stagedRefreshViews
	stagedIndexOps          []stagedIndexOp
	stagedEncryptedIndexOps []stagedEncryptedIndexOp
	stagedInactiveVersions  map[string]struct{}
}

type stagedPatchCollection struct {
	identityDID string
	patch       string
	migration   immutable.Option[lensmodel.Lens]
}

type stagedSetActiveVersion struct {
	identityDID string
	versionID   string
}

type stagedViewAdd struct {
	identityDID string
	gqlQuery    string
	sdl         string
	transform   string
}

type stagedRefreshViews struct {
	identityDID string
	optionsJSON string
}

type stagedIndexOp struct {
	identityDID    string
	collectionName string
	create         *client.NewIndexRequest
	deleteName     string
}

type stagedEncryptedIndexOp struct {
	identityDID    string
	collectionName string
	createField    string
	deleteField    string
}

var _ client.Txn = (*TxnWrapper)(nil)

func (t *TxnWrapper) ID() uint64 {
	return t.id
}

func (t *TxnWrapper) StartTS() time.Time {
	return t.startTS
}

func (t *TxnWrapper) Commit() error {
	if err := t.txn.Commit(); err != nil {
		return err
	}

	for _, op := range t.stagedPatchCollections {
		if err := t.wrapper.patchCollection(op.identityDID, op.patch, op.migration); err != nil {
			return err
		}
	}

	for _, op := range t.stagedSetActiveVersions {
		if err := t.wrapper.setActiveCollectionVersion(op.identityDID, op.versionID); err != nil {
			return err
		}
	}

	for _, op := range t.stagedViews {
		if _, err := t.wrapper.node.AddView(op.identityDID, op.gqlQuery, op.sdl, op.transform); err != nil {
			return err
		}
	}

	for _, op := range t.stagedRefreshes {
		if err := t.wrapper.node.RefreshViews(op.identityDID, op.optionsJSON); err != nil {
			return err
		}
	}

	for _, op := range t.stagedIndexOps {
		if op.create != nil {
			fields := make([]IndexField, len(op.create.Fields))
			for i, f := range op.create.Fields {
				fields[i] = IndexField{Name: f.Name, Descending: f.Descending}
			}
			if _, err := t.wrapper.node.CreateIndex(op.identityDID, op.collectionName, op.create.Name, fields, op.create.Unique); err != nil {
				return err
			}
		} else if err := t.wrapper.node.DropIndex(op.identityDID, op.collectionName, op.deleteName); err != nil {
			return err
		}
	}

	for _, op := range t.stagedEncryptedIndexOps {
		if op.createField != "" {
			if _, err := t.wrapper.node.CreateEncryptedIndex(op.identityDID, op.collectionName, op.createField); err != nil {
				return err
			}
		} else if err := t.wrapper.node.DeleteEncryptedIndex(op.identityDID, op.collectionName, op.deleteField); err != nil {
			return err
		}
	}

	return nil
}

func (t *TxnWrapper) Discard() {
	_ = t.txn.Rollback()
}

// Delegate Store methods to the underlying transaction
func (t *TxnWrapper) ExecRequest(
	ctx context.Context,
	request string,
	opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	opt := utils.NewOptions(opts...)

	varsJSON := ""
	if opt.Variables != nil {
		varsBytes, err := json.Marshal(opt.Variables)
		if err != nil {
			return &client.RequestResult{
				GQL: client.GQLResult{
					Errors: []error{fmt.Errorf("failed to marshal variables: %w", err)},
				},
			}
		}
		varsJSON = string(varsBytes)
	}

	identityDID := identityDIDFromOption(opt.GetIdentity())

	operationName := ""
	if opt.OperationName.HasValue() {
		operationName = opt.OperationName.Value()
	}
	responseJSON, err := t.txn.ExecRequest(identityDID, request, operationName, varsJSON)
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

	if gqlResult.Data != nil {
		gqlResult.Data = convertDateTimeStrings(gqlResult.Data)
		if isExplainResult(gqlResult.Data) {
			gqlResult.Data = normalizeExplainTypes(gqlResult.Data)
		}
	}

	return &client.RequestResult{GQL: gqlResult}
}

// Stub out other Store methods on transaction - most delegate to wrapper
func (t *TxnWrapper) AddCollection(ctx context.Context, sdl string, opts ...options.Enumerable[options.AddCollectionOptions]) ([]client.CollectionVersion, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	responseJSON, err := t.wrapper.node.AddSchemaInTxn(t.txn.id, identityDID, sdl)
	if err != nil {
		return nil, err
	}

	var versions []client.CollectionVersion
	if err := json.Unmarshal([]byte(responseJSON), &versions); err != nil {
		return nil, fmt.Errorf("failed to parse schema response: %w", err)
	}

	return versions, nil
}

func (t *TxnWrapper) GetCollectionByName(ctx context.Context, name client.CollectionName, opts ...options.Enumerable[options.GetCollectionByNameOptions]) (client.Collection, error) {
	collections, err := t.GetCollections(ctx, options.GetCollections().SetCollectionName(string(name)))
	if err != nil {
		return nil, err
	}
	if len(collections) == 0 {
		return nil, fmt.Errorf("collection not found: %s", name)
	}
	return collections[0], nil
}

func (t *TxnWrapper) GetCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	responseJSON, err := t.wrapper.node.GetCollectionsInTxn(t.txn.id, identityDID)
	if err != nil {
		return nil, err
	}

	var versions []client.CollectionVersion
	if err := json.Unmarshal([]byte(responseJSON), &versions); err != nil {
		return nil, fmt.Errorf("failed to parse collections: %w", err)
	}
	t.applyStagedCollectionState(versions)

	// Apply filters
	includeInactive := opt.GetInactive.HasValue() && opt.GetInactive.Value()
	var filtered []client.CollectionVersion
	for _, v := range versions {
		if !includeInactive && !v.IsActive {
			continue
		}
		if opt.CollectionName.HasValue() && v.Name != opt.CollectionName.Value() {
			continue
		}
		if opt.CollectionID.HasValue() && v.CollectionID != opt.CollectionID.Value() {
			continue
		}
		filtered = append(filtered, v)
	}

	collections := make([]client.Collection, len(filtered))
	for i, v := range filtered {
		collections[i] = &CollectionWrapper{
			wrapper: t.wrapper,
			version: v,
			txn:     t,
		}
	}

	return collections, nil
}

func (t *TxnWrapper) SetActiveCollectionVersion(ctx context.Context, versionID string, opts ...options.Enumerable[options.SetActiveCollectionVersionOptions]) error {
	opt := utils.NewOptions(opts...)
	t.stagedSetActiveVersions = append(t.stagedSetActiveVersions, stagedSetActiveVersion{
		identityDID: identityDIDFromOption(opt.GetIdentity()),
		versionID:   versionID,
	})
	return nil
}

func (t *TxnWrapper) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[lensmodel.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	if err := t.validateImmediatePatchCollection(ctx, identityDID, patch); err != nil {
		return err
	}
	t.recordStagedCollectionState(identityDID, patch)
	t.stagedPatchCollections = append(t.stagedPatchCollections, stagedPatchCollection{
		identityDID: identityDID,
		patch:       patch,
		migration:   migration,
	})
	return nil
}

func (t *TxnWrapper) DeleteCollection(
	ctx context.Context,
	names []string,
	opts ...options.Enumerable[options.DeleteCollectionOptions],
) error {
	opt := utils.NewOptions(opts...)
	ident := opt.GetIdentity()

	patch, err := buildDeleteCollectionPatch(
		names,
		opt.ActiveOnly,
		func(name string) (client.Collection, error) {
			getOpts := options.GetCollectionByName()
			if ident.HasValue() {
				getOpts.SetIdentity(ident.Value())
			}
			return t.GetCollectionByName(ctx, client.CollectionName(name), getOpts)
		},
		func(collectionID string) ([]client.Collection, error) {
			getOpts := options.GetCollections().
				SetCollectionID(collectionID).
				SetGetInactive(true)
			if ident.HasValue() {
				getOpts.SetIdentity(ident.Value())
			}
			return t.GetCollections(ctx, getOpts)
		},
	)
	if err != nil {
		return err
	}

	t.stagedPatchCollections = append(t.stagedPatchCollections, stagedPatchCollection{
		identityDID: identityDIDFromOption(ident),
		patch:       patch,
		migration:   immutable.None[lensmodel.Lens](),
	})
	return nil
}

func (t *TxnWrapper) rawCollectionsInTxn(identityDID string) ([]client.CollectionVersion, error) {
	responseJSON, err := t.wrapper.node.GetCollectionsInTxn(t.txn.id, identityDID)
	if err != nil {
		return nil, err
	}

	var versions []client.CollectionVersion
	if err := json.Unmarshal([]byte(responseJSON), &versions); err != nil {
		return nil, fmt.Errorf("failed to parse collections: %w", err)
	}
	return versions, nil
}

func (t *TxnWrapper) applyStagedCollectionState(versions []client.CollectionVersion) {
	if len(t.stagedInactiveVersions) == 0 {
		return
	}
	for i := range versions {
		if _, ok := t.stagedInactiveVersions[versions[i].VersionID]; ok {
			versions[i].IsActive = false
		}
	}
}

func (t *TxnWrapper) recordStagedCollectionState(identityDID string, patch string) {
	type patchOp struct {
		Op    string          `json:"op"`
		Path  string          `json:"path"`
		Value json.RawMessage `json:"value"`
	}

	var ops []patchOp
	if err := json.Unmarshal([]byte(patch), &ops); err != nil {
		return
	}

	var txnVersions []client.CollectionVersion
	for _, op := range ops {
		if op.Op != "replace" || !strings.HasSuffix(op.Path, "/IsActive") {
			continue
		}
		var isActive bool
		if err := json.Unmarshal(op.Value, &isActive); err != nil || isActive {
			continue
		}
		if txnVersions == nil {
			var err error
			txnVersions, err = t.rawCollectionsInTxn(identityDID)
			if err != nil {
				return
			}
		}
		target := strings.TrimPrefix(strings.TrimSuffix(op.Path, "/IsActive"), "/")
		version, ok := findCollectionVersion(txnVersions, target)
		if !ok {
			continue
		}
		if t.stagedInactiveVersions == nil {
			t.stagedInactiveVersions = make(map[string]struct{})
		}
		t.stagedInactiveVersions[version.VersionID] = struct{}{}
	}
}

func (t *TxnWrapper) validateImmediatePatchCollection(
	ctx context.Context,
	identityDID string,
	patch string,
) error {
	removeTargets, err := collectionLevelRemoveTargets(patch)
	if err != nil || len(removeTargets) == 0 {
		return err
	}

	txnVersions, err := t.rawCollectionsInTxn(identityDID)
	if err != nil {
		return err
	}

	for _, target := range removeTargets {
		version, ok := findCollectionVersion(txnVersions, target)
		if !ok {
			continue
		}
		hasDocs, err := t.collectionHasDocumentsInTxn(ctx, version.Name)
		if err != nil {
			continue
		}
		if hasDocs {
			return fmt.Errorf("cannot delete a collection that has documents")
		}
	}

	return nil
}

func (t *TxnWrapper) collectionHasDocumentsInTxn(ctx context.Context, collectionName string) (bool, error) {
	result := t.ExecRequest(ctx, fmt.Sprintf("query { %s { _docID } }", collectionName))
	if len(result.GQL.Errors) > 0 {
		return false, result.GQL.Errors[0]
	}

	data, ok := result.GQL.Data.(map[string]any)
	if !ok {
		return false, nil
	}
	items, ok := data[collectionName].([]any)
	return ok && len(items) > 0, nil
}

func collectionLevelRemoveTargets(patch string) ([]string, error) {
	type patchOp struct {
		Op   string `json:"op"`
		Path string `json:"path"`
	}

	var ops []patchOp
	if err := json.Unmarshal([]byte(patch), &ops); err != nil {
		return nil, err
	}

	targets := make([]string, 0)
	for _, op := range ops {
		if op.Op != "remove" {
			continue
		}
		target := strings.TrimPrefix(op.Path, "/")
		if target == "" || strings.Contains(target, "/") {
			continue
		}
		targets = append(targets, target)
	}
	return targets, nil
}

func findCollectionVersion(
	versions []client.CollectionVersion,
	nameOrVersionID string,
) (client.CollectionVersion, bool) {
	for _, version := range versions {
		if version.VersionID == nameOrVersionID {
			return version, true
		}
	}
	for _, version := range versions {
		if version.Name == nameOrVersionID && version.IsActive {
			return version, true
		}
	}
	return client.CollectionVersion{}, false
}

func (t *TxnWrapper) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListIndexesOptions],
) (map[client.CollectionName][]client.IndexDescription, error) {
	return t.wrapper.ListIndexes(ctx, opts...)
}

func (t *TxnWrapper) AddDACPolicy(ctx context.Context, policy string, opts ...options.Enumerable[options.AddDACPolicyOptions]) (client.AddPolicyResult, error) {
	return t.wrapper.AddDACPolicy(ctx, policy, opts...)
}

func (t *TxnWrapper) AddDACActorRelationship(
	ctx context.Context,
	collectionName, docID, relation, targetActor string,
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	return t.wrapper.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opts...)
}

func (t *TxnWrapper) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName, docID, relation, targetActor string,
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	return t.wrapper.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opts...)
}

func (t *TxnWrapper) AddNACActorRelationship(
	ctx context.Context,
	relation, targetActor string,
	opts ...options.Enumerable[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	return t.wrapper.AddNACActorRelationship(ctx, relation, targetActor, opts...)
}

func (t *TxnWrapper) DeleteNACActorRelationship(
	ctx context.Context,
	relation, targetActor string,
	opts ...options.Enumerable[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	return t.wrapper.DeleteNACActorRelationship(ctx, relation, targetActor, opts...)
}

func (t *TxnWrapper) ReEnableNAC(ctx context.Context, opts ...options.Enumerable[options.ReEnableNACOptions]) error {
	return t.wrapper.ReEnableNAC(ctx, opts...)
}

func (t *TxnWrapper) DisableNAC(ctx context.Context, opts ...options.Enumerable[options.DisableNACOptions]) error {
	return t.wrapper.DisableNAC(ctx, opts...)
}

func (t *TxnWrapper) GetNACStatus(ctx context.Context, opts ...options.Enumerable[options.GetNACStatusOptions]) (client.NACStatusResult, error) {
	return t.wrapper.GetNACStatus(ctx, opts...)
}

func (t *TxnWrapper) GetNodeIdentity(
	ctx context.Context,
) (immutable.Option[identity.PublicRawIdentity], error) {
	return t.wrapper.GetNodeIdentity(ctx)
}

func (t *TxnWrapper) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey, opts ...options.Enumerable[options.VerifySignatureOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	return t.wrapper.node.BlockVerifySignatureInTxn(
		t.txn.id,
		string(pubKey.Type()),
		pubKey.String(),
		blockCid,
		identityDID,
	)
}

func (t *TxnWrapper) AddView(
	ctx context.Context,
	gqlQuery, sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	opt := utils.NewOptions(opts...)
	transformStr := ""
	if opt.TransformCID.HasValue() {
		transformStr = opt.TransformCID.Value()
	}
	t.stagedViews = append(t.stagedViews, stagedViewAdd{
		identityDID: identityDIDFromOption(opt.GetIdentity()),
		gqlQuery:    gqlQuery,
		sdl:         sdl,
		transform:   transformStr,
	})
	return []client.CollectionVersion{}, nil
}

func (t *TxnWrapper) RefreshViews(ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions]) error {
	opt := utils.NewOptions(opts...)
	optsJSON := ""
	if opt.CollectionName.HasValue() || opt.VersionID.HasValue() {
		data, err := json.Marshal(opt)
		if err != nil {
			return fmt.Errorf("failed to marshal options: %w", err)
		}
		optsJSON = string(data)
	}
	t.stagedRefreshes = append(t.stagedRefreshes, stagedRefreshViews{
		identityDID: identityDIDFromOption(opt.GetIdentity()),
		optionsJSON: optsJSON,
	})
	return nil
}

func (t *TxnWrapper) SetMigration(ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions]) (string, error) {
	return t.wrapper.SetMigration(ctx, config, opts...)
}

func (t *TxnWrapper) AddLens(ctx context.Context, lens lensmodel.Lens, opts ...options.Enumerable[options.AddLensOptions]) (string, error) {
	lensJSON, err := json.Marshal(lens)
	if err != nil {
		return "", fmt.Errorf("failed to marshal lens: %w", err)
	}
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())
	return t.wrapper.node.LensAddInTxn(t.txn.id, identityDID, string(lensJSON))
}

func (t *TxnWrapper) ListLenses(ctx context.Context, opts ...options.Enumerable[options.ListLensesOptions]) (map[string]lensmodel.Lens, error) {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	raw, err := t.wrapper.node.LensListInTxn(t.txn.id, identityDID)
	if err != nil {
		return nil, err
	}

	var rustMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &rustMap); err != nil {
		return nil, fmt.Errorf("ffi: failed to parse lens list: %w", err)
	}

	result := make(map[string]lensmodel.Lens, len(rustMap))
	for id, moduleJSON := range rustMap {
		var module lensmodel.LensModule
		if err := json.Unmarshal(moduleJSON, &module); err == nil {
			result[id] = lensmodel.Lens{Lenses: []lensmodel.LensModule{module}}
			continue
		}

		var lensObj lensmodel.Lens
		if err := json.Unmarshal(moduleJSON, &lensObj); err != nil {
			return nil, fmt.Errorf("ffi: failed to parse lens %s: %w", id, err)
		}
		result[id] = lensObj
	}

	return result, nil
}

func (t *TxnWrapper) BasicImport(ctx context.Context, filepath string) error {
	return t.wrapper.BasicImport(ctx, filepath)
}

func (t *TxnWrapper) BasicExport(ctx context.Context, filepath string, opts ...options.Enumerable[options.BasicExportOptions]) error {
	return t.wrapper.BasicExport(ctx, filepath, opts...)
}

func (t *TxnWrapper) PrintDump(ctx context.Context) error {
	return t.wrapper.PrintDump(ctx)
}

func (t *TxnWrapper) ListAllEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListAllEncryptedIndexesOptions],
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	return t.wrapper.ListAllEncryptedIndexes(ctx, opts...)
}

// P2P methods - not available in transactions
func (t *TxnWrapper) PeerInfo(ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions]) ([]string, error) {
	return nil, fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) ActivePeers(ctx context.Context, opts ...options.Enumerable[options.ActivePeersOptions]) ([]string, error) {
	return nil, fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) Connect(ctx context.Context, addresses []string, opts ...options.Enumerable[options.ConnectOptions]) error {
	return fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) AddReplicator(ctx context.Context, addresses []string, opts ...options.Enumerable[options.AddReplicatorOptions]) error {
	return fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) DeleteReplicator(ctx context.Context, id string, opts ...options.Enumerable[options.DeleteReplicatorOptions]) error {
	return fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) ListReplicators(ctx context.Context, opts ...options.Enumerable[options.ListReplicatorsOptions]) ([]client.Replicator, error) {
	return nil, fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) AddP2PCollections(ctx context.Context, collectionNames []string, opts ...options.Enumerable[options.AddP2PCollectionsOptions]) error {
	return fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) DeleteP2PCollections(ctx context.Context, collectionNames []string, opts ...options.Enumerable[options.DeleteP2PCollectionsOptions]) error {
	return fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) ListP2PCollections(ctx context.Context, opts ...options.Enumerable[options.ListP2PCollectionsOptions]) ([]string, error) {
	return nil, fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) AddP2PDocuments(ctx context.Context, docIDs []string, opts ...options.Enumerable[options.AddP2PDocumentsOptions]) error {
	return fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) DeleteP2PDocuments(ctx context.Context, docIDs []string, opts ...options.Enumerable[options.DeleteP2PDocumentsOptions]) error {
	return fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) ListP2PDocuments(ctx context.Context, opts ...options.Enumerable[options.ListP2PDocumentsOptions]) ([]string, error) {
	return nil, fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) SyncDocuments(ctx context.Context, collectionName string, docIDs []string, opts ...options.Enumerable[options.SyncDocumentsOptions]) error {
	return fmt.Errorf("p2p not available")
}
func (t *TxnWrapper) SyncCollectionVersions(ctx context.Context, versionIDs []string, opts ...options.Enumerable[options.SyncCollectionVersionsOptions]) error {
	return fmt.Errorf("p2p not available")
}

func (t *TxnWrapper) SyncBranchableCollection(ctx context.Context, collectionID string, opts ...options.Enumerable[options.SyncBranchableCollectionOptions]) error {
	return fmt.Errorf("p2p not available")
}

// ============================================================================
// eventBus implements event.Bus for testing
// ============================================================================

type eventBus struct {
	mu     sync.RWMutex
	closed bool
	subs   []event.Subscription
}

func newEventBus() *eventBus {
	return &eventBus{}
}

func (e *eventBus) Publish(msg event.Message) {
	e.mu.RLock()
	if e.closed {
		e.mu.RUnlock()
		return
	}
	subs := make([]event.Subscription, len(e.subs))
	copy(subs, e.subs)
	e.mu.RUnlock()

	for _, sub := range subs {
		if es, ok := sub.(*eventSubscription); ok {
			if es.wantsEvent(msg.Name) {
				// Recover from send-on-closed-channel if Close() races with Publish()
				func() {
					defer func() {
						if r := recover(); r != nil { //nolint:staticcheck
						}
					}()
					select {
					case es.ch <- msg:
					default:
					}
				}()
			}
		}
	}
}

func (e *eventBus) Subscribe(events ...event.Name) (event.Subscription, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	sub := &eventSubscription{
		ch:     make(chan event.Message, 100),
		events: events,
	}
	e.subs = append(e.subs, sub)
	return sub, nil
}

func (e *eventBus) Unsubscribe(sub event.Subscription) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, s := range e.subs {
		if s == sub {
			e.subs = append(e.subs[:i], e.subs[i+1:]...)
			break
		}
	}
}

func (e *eventBus) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return
	}
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
	txn     *TxnWrapper
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

func (c *CollectionWrapper) execRequest(
	ctx context.Context,
	request string,
	opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	if c.txn != nil {
		return c.txn.ExecRequest(ctx, request, opts...)
	}
	return c.wrapper.ExecRequest(ctx, request, opts...)
}

func (c *CollectionWrapper) AddDocument(ctx context.Context, doc *client.Document, opts ...options.Enumerable[options.AddDocumentOptions]) error {
	// Use GraphQL mutation to create document
	docJSON, err := doc.ToJSONPatch()
	if err != nil {
		return fmt.Errorf("failed to convert document to JSON: %w", err)
	}

	// Extract encryption options
	opt := utils.NewOptions(opts...)

	// Convert JSON to GraphQL input format (unquoted keys)
	gqlInput := jsonToGraphQLInput(string(docJSON))
	params := fmt.Sprintf("input: %s", gqlInput)
	if opt.EncryptDoc {
		params += ", encrypt: true"
	}
	if len(opt.EncryptedFields) > 0 {
		quoted := make([]string, len(opt.EncryptedFields))
		for i, f := range opt.EncryptedFields {
			quoted[i] = `"` + f + `"`
		}
		params += ", encryptFields: [" + strings.Join(quoted, ", ") + "]"
	}
	mutation := fmt.Sprintf(`mutation { create_%s(%s) { _docID } }`, c.version.Name, params)
	result := c.execRequest(ctx, mutation, execRequestWithIdentity(opt.GetIdentity()))
	if len(result.GQL.Errors) > 0 {
		return result.GQL.Errors[0]
	}

	// Extract docID from the response and publish update events
	if data, ok := result.GQL.Data.(map[string]any); ok {
		mutationKey := "create_" + c.version.Name
		if mutResult, ok := data[mutationKey].([]any); ok && len(mutResult) > 0 {
			if docData, ok := mutResult[0].(map[string]any); ok {
				if docID, ok := docData["_docID"].(string); ok {
					// Update the Go-side document's ID to match Rust's computed ID.
					// This is needed when Rust applies defaults (e.g. UTC_NOW) that
					// change the document content and therefore its content-addressed ID.
					newDocID, err := client.NewDocIDFromString(docID)
					if err == nil && doc.ID().String() != docID {
						doc.SetDocID(newDocID)
					}
				}
			}
		}
	}

	return nil
}

func (c *CollectionWrapper) AddManyDocuments(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	for _, doc := range docs {
		if err := c.AddDocument(ctx, doc, opts...); err != nil {
			return err
		}
	}
	return nil
}

func (c *CollectionWrapper) UpdateDocument(ctx context.Context, doc *client.Document, opts ...options.Enumerable[options.UpdateDocumentOptions]) error {
	opt := utils.NewOptions(opts...)
	docJSON, err := doc.ToJSONPatch()
	if err != nil {
		return fmt.Errorf("failed to convert document to JSON: %w", err)
	}

	// Convert JSON to GraphQL input format (unquoted keys)
	gqlInput := jsonToGraphQLInput(string(docJSON))
	mutation := fmt.Sprintf(`mutation { update_%s(docID: "%s", input: %s) { _docID } }`,
		c.version.Name, doc.ID().String(), gqlInput)
	result := c.execRequest(ctx, mutation, execRequestWithIdentity(opt.GetIdentity()))
	if len(result.GQL.Errors) > 0 {
		return result.GQL.Errors[0]
	}

	return nil
}

func (c *CollectionWrapper) SaveDocument(ctx context.Context, doc *client.Document, opts ...options.Enumerable[options.SaveDocumentOptions]) error {
	opt := utils.NewOptions(opts...)
	// Check if doc exists in the database by querying for it.
	// Pass identity so ACP-protected documents are visible to the owner.
	existsOpt := options.ExistsDocument()
	if opt.GetIdentity().HasValue() {
		existsOpt.SetIdentity(opt.GetIdentity().Value())
	}
	exists, err := c.ExistsDocument(ctx, doc.ID(), existsOpt)
	if err != nil {
		// If error checking existence, check if deleted before creating
		if c.isDocumentDeleted(ctx, doc.ID()) {
			return fmt.Errorf("a document with the given ID has been deleted")
		}
		addOpt := options.AddDocument().
			SetEncryptDoc(opt.EncryptDoc).
			SetEncryptedFields(opt.EncryptedFields)
		if opt.GetIdentity().HasValue() {
			addOpt.SetIdentity(opt.GetIdentity().Value())
		}
		return c.AddDocument(ctx, doc, addOpt)
	}
	if !exists {
		// Document doesn't exist - check if it was deleted
		if c.isDocumentDeleted(ctx, doc.ID()) {
			return fmt.Errorf("a document with the given ID has been deleted")
		}
		addOpt := options.AddDocument().
			SetEncryptDoc(opt.EncryptDoc).
			SetEncryptedFields(opt.EncryptedFields)
		if opt.GetIdentity().HasValue() {
			addOpt.SetIdentity(opt.GetIdentity().Value())
		}
		return c.AddDocument(ctx, doc, addOpt)
	}
	updateOpt := options.UpdateDocument()
	if opt.GetIdentity().HasValue() {
		updateOpt.SetIdentity(opt.GetIdentity().Value())
	}
	return c.UpdateDocument(ctx, doc, updateOpt)
}

// getLatestCompositeCID queries _commits to get the latest composite CID for a document.
// This is needed for update/delete events so the test framework can resolve CID template variables.
func (c *CollectionWrapper) getLatestCompositeCID(ctx context.Context, docID string) gocid.Cid {
	query := fmt.Sprintf(
		`{ _commits(docID: "%s", filter: {fieldName: {_eq: "_C"}}, order: {height: DESC}, limit: 1) { cid } }`,
		docID,
	)
	result := c.wrapper.ExecRequest(ctx, query)
	if c.txn != nil {
		result = c.txn.ExecRequest(ctx, query)
	}
	if len(result.GQL.Errors) > 0 {
		return gocid.Undef
	}
	if data, ok := result.GQL.Data.(map[string]any); ok {
		if commits, ok := data["_commits"].([]any); ok && len(commits) > 0 {
			if commit, ok := commits[0].(map[string]any); ok {
				if cidStr, ok := commit["cid"].(string); ok {
					parsed, err := gocid.Decode(cidStr)
					if err == nil {
						return parsed
					}
				}
			}
		}
	}
	return gocid.Undef
}

// isDocumentDeleted checks if a document is marked as deleted using showDeleted query.
func (c *CollectionWrapper) isDocumentDeleted(ctx context.Context, docID client.DocID) bool {
	query := fmt.Sprintf(`{ %s(docID: "%s", showDeleted: true) { _docID _deleted } }`, c.version.Name, docID.String())
	result := c.execRequest(ctx, query)
	if len(result.GQL.Errors) > 0 {
		return false
	}
	if data, ok := result.GQL.Data.(map[string]any); ok {
		if docs, ok := data[c.version.Name].([]any); ok && len(docs) > 0 {
			if docData, ok := docs[0].(map[string]any); ok {
				if deleted, ok := docData["_deleted"].(bool); ok {
					return deleted
				}
			}
		}
	}
	return false
}

func (c *CollectionWrapper) DeleteDocument(ctx context.Context, docID client.DocID, opts ...options.Enumerable[options.DeleteDocumentOptions]) (bool, error) {
	opt := utils.NewOptions(opts...)
	mutation := fmt.Sprintf(`mutation { delete_%s(docID: "%s") { _docID } }`, c.version.Name, docID.String())
	result := c.execRequest(ctx, mutation, execRequestWithIdentity(opt.GetIdentity()))
	if len(result.GQL.Errors) > 0 {
		return false, result.GQL.Errors[0]
	}

	// When ACP is active and the mutation returned empty results, the document was
	// invisible or unauthorized. Return the same error as Go's collection.Delete().
	if c.version.Policy.HasValue() {
		if data, ok := result.GQL.Data.(map[string]any); ok {
			key := fmt.Sprintf("delete_%s", c.version.Name)
			if docs, ok := data[key].([]any); ok && len(docs) == 0 {
				return false, fmt.Errorf("document not found or not authorized to access")
			}
		}
	}

	return true, nil
}

func (c *CollectionWrapper) ExistsDocument(ctx context.Context, docID client.DocID, opts ...options.Enumerable[options.ExistsDocumentOptions]) (bool, error) {
	opt := utils.NewOptions(opts...)
	query := fmt.Sprintf(`{ %s(docID: "%s") { _docID } }`, c.version.Name, docID.String())
	result := c.execRequest(ctx, query, execRequestWithIdentity(opt.GetIdentity()))
	if len(result.GQL.Errors) > 0 {
		return false, result.GQL.Errors[0]
	}
	if data, ok := result.GQL.Data.(map[string]any); ok {
		if docs, ok := data[c.version.Name].([]any); ok {
			return len(docs) > 0, nil
		}
	}
	return false, nil
}

func (c *CollectionWrapper) UpdateDocumentsWithFilter(
	ctx context.Context,
	filter any,
	updater string,
	opts ...options.Enumerable[options.UpdateDocumentsWithFilterOptions],
) (*client.UpdateResult, error) {
	// Validate filter (mirrors Go's collection_update.go makeSelectionPlan validation)
	var gqlFilter string
	switch f := filter.(type) {
	case string:
		if f == "" {
			return nil, fmt.Errorf("filter cannot be empty")
		}
		// String filters may be in relaxed JSON/GQL format (unquoted keys).
		// Try parsing as JSON first; if that fails, use as-is (GQL format).
		var parsed map[string]any
		if err := json.Unmarshal([]byte(f), &parsed); err != nil {
			// Not valid JSON - could be relaxed GQL format like {name: {_eq: "John"}}.
			// Validate by attempting to parse with fastjson-compatible check.
			if !isRelaxedJSONObject(f) {
				return nil, fmt.Errorf("cannot parse JSON: cannot parse object")
			}
			gqlFilter = f
		} else {
			filterJSON, _ := json.Marshal(parsed)
			gqlFilter = jsonToGraphQLInput(string(filterJSON))
		}
	case map[string]any:
		filterJSON, err := json.Marshal(f)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filter: %w", err)
		}
		gqlFilter = jsonToGraphQLInput(string(filterJSON))
	default:
		return nil, fmt.Errorf("invalid filter")
	}

	// Validate updater (mirrors Go's collection_update.go fastjson.Parse validation)
	// Go first parses the updater as JSON, then checks the type.
	var parsedUpdater any
	if err := json.Unmarshal([]byte(updater), &parsedUpdater); err != nil {
		return nil, fmt.Errorf("cannot parse JSON: cannot parse object")
	}
	switch parsedUpdater.(type) {
	case []any:
		// JSON Patch (array) - Go's implementation has a "todo" for patch support,
		// so it effectively does nothing. Return empty result.
		return &client.UpdateResult{DocIDs: make([]string, 0)}, nil
	case map[string]any:
		// Merge Patch (object) - this is the supported case
	default:
		return nil, fmt.Errorf("the updater of a document is of invalid type")
	}

	opt := utils.NewOptions(opts...)
	gqlUpdater := jsonToGraphQLInput(updater)
	mutation := fmt.Sprintf(`mutation { update_%s(filter: %s, input: %s) { _docID } }`,
		c.version.Name, gqlFilter, gqlUpdater)
	result := c.execRequest(ctx, mutation, execRequestWithIdentity(opt.GetIdentity()))
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

// isRelaxedJSONObject checks if a string looks like a JSON/GQL object (starts with { and ends with }).
func isRelaxedJSONObject(s string) bool {
	s = trimJSONWhitespace(s)
	return len(s) >= 2 && s[0] == '{' && s[len(s)-1] == '}'
}

// trimJSONWhitespace trims leading/trailing whitespace from a string.
func trimJSONWhitespace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// convertDateTimeStrings recursively walks a decoded JSON response and converts:
// - string values that look like RFC3339 datetimes into time.Time objects
// - json.Number values into int64 or float64 as appropriate
// This is needed because the Rust FFI returns datetimes as JSON strings and numbers
// as json.Number, while Go's native implementation returns time.Time and int64/float64.
// convertDateTimeStrings recursively converts RFC 3339 date-time strings to time.Time
// and json.Number values to their Go numeric types. Note: this converts ANY string that
// parses as RFC 3339, not just schema-declared DateTime fields. If a plain string field
// happens to contain a date-like value, it will be coerced to time.Time.
func convertDateTimeStrings(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, v := range val {
			val[k] = convertDateTimeStrings(v)
		}
		return val
	case []any:
		for i, v := range val {
			val[i] = convertDateTimeStrings(v)
		}
		return val
	case string:
		if t, err := time.Parse(time.RFC3339Nano, val); err == nil {
			return t
		}
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t
		}
		return val
	case json.Number:
		// Try int64 first, then float64
		if i, err := val.Int64(); err == nil {
			return i
		}
		if f, err := val.Float64(); err == nil {
			return f
		}
		return val
	default:
		return val
	}
}

// normalizeExplainTypes fixes type mismatches in explain results returned by the Rust FFI.
// Go's native query planner produces uint64 for execution metrics and typed slices
// ([]string, []map[string]any) for arrays. JSON decoding via convertDateTimeStrings
// produces int64 for numbers and []interface{} for arrays. This function converts
// int64 -> uint64 and []any of maps -> []map[string]any to match Go's native types.
//
// We also rename Rust-specific explain keys to match Go's naming (e.g. "create" -> "add").
func normalizeExplainTypes(v any) any {
	switch val := v.(type) {
	case map[string]any:
		// Rename Rust-specific keys to match Go's explain naming.
		if createVal, ok := val["create"]; ok {
			if _, hasAdd := val["add"]; !hasAdd {
				val["add"] = createVal
				delete(val, "create")
			}
		}
		for k, child := range val {
			val[k] = normalizeExplainTypes(child)
		}
		return val
	case []any:
		// Recurse into each element to normalize nested types.
		for i, item := range val {
			val[i] = normalizeExplainTypes(item)
		}
		// Convert []any of strings to []string (Go planner produces typed []string
		// for prefixes, docID fields).
		allStrings := len(val) > 0
		for _, item := range val {
			if _, ok := item.(string); !ok {
				allStrings = false
				break
			}
		}
		if allStrings {
			result := make([]string, len(val))
			for i, item := range val {
				result[i] = item.(string)
			}
			return result
		}
		// Keep []any as-is (don't convert to []map[string]any). The test
		// framework's areResultArraysEqual handles []map[string]any (expected)
		// vs []any (actual) by type-asserting actual.([]any) at the call site,
		// then recursing element-by-element through areResultsEqual.
		return val
	case int64:
		// Go's explain metrics are uint64. JSON decoding via convertDateTimeStrings
		// produces int64 from json.Number. Convert non-negative int64 to uint64
		// to match Go's native explain output.
		if val >= 0 {
			return uint64(val)
		}
		return val
	default:
		return val
	}
}

// isExplainResult returns true if the GQL result data contains an explain response.
func isExplainResult(data any) bool {
	m, ok := data.(map[string]any)
	if !ok {
		return false
	}
	_, hasExplain := m["explain"]
	return hasExplain
}

// decodeBase64Deltas walks a GQL result tree and converts any "delta" field
// whose value is a base64 string into a []byte. Rust's JSON serialization
// encodes byte slices as base64, while Go's native client returns raw []byte.
// The test framework's assertions compare against []byte, so this translation
// is necessary for parity.
func decodeBase64Deltas(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			if k == "delta" {
				if s, ok := child.(string); ok {
					if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
						val[k] = decoded
					}
				}
			} else {
				val[k] = decodeBase64Deltas(child)
			}
		}
		return val
	case []any:
		for i, child := range val {
			val[i] = decodeBase64Deltas(child)
		}
		return val
	default:
		return val
	}
}

func (c *CollectionWrapper) DeleteDocumentsWithFilter(ctx context.Context, filter any, opts ...options.Enumerable[options.DeleteDocumentsWithFilterOptions]) (*client.DeleteResult, error) {
	opt := utils.NewOptions(opts...)
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filter: %w", err)
	}

	// Convert JSON to GraphQL input format (unquoted keys)
	gqlFilter := jsonToGraphQLInput(string(filterJSON))
	mutation := fmt.Sprintf(`mutation { delete_%s(filter: %s) { _docID } }`, c.version.Name, gqlFilter)
	result := c.execRequest(ctx, mutation, execRequestWithIdentity(opt.GetIdentity()))
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

func (c *CollectionWrapper) GetDocument(ctx context.Context, docID client.DocID, opts ...options.Enumerable[options.GetDocumentOptions]) (*client.Document, error) {
	opt := utils.NewOptions(opts...)
	// Query the document by ID - must request ALL fields so SetWithJSON can properly
	// track which fields are dirty (modified) vs unchanged
	var fieldNames []string
	for _, field := range c.version.Fields {
		// Skip ALL relation fields (both single objects and arrays of objects)
		// Relations can't be directly queried as scalar values
		if field.Kind.IsObject() {
			continue
		}
		fieldNames = append(fieldNames, field.Name)
	}

	query := fmt.Sprintf(`{ %s(docID: "%s") { %s } }`, c.version.Name, docID.String(), strings.Join(fieldNames, " "))
	result := c.execRequest(ctx, query, execRequestWithIdentity(opt.GetIdentity()))
	if len(result.GQL.Errors) > 0 {
		return nil, result.GQL.Errors[0]
	}

	// Parse the result to check if document exists
	data, ok := result.GQL.Data.(map[string]any)
	if !ok {
		return nil, c.checkIfDocumentDeleted(ctx, docID)
	}

	docs, ok := data[c.version.Name].([]any)
	if !ok || len(docs) == 0 {
		return nil, c.checkIfDocumentDeleted(ctx, docID)
	}

	docData, ok := docs[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("document not found: %s", docID.String())
	}

	// Convert JSON types (json.Number -> int64/float64, datetime strings -> time.Time)
	converted, ok := convertDateTimeStrings(docData).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected type after date conversion")
	}
	docData = converted

	// Filter out secondary FK fields from docData before creating the document.
	// Go's native Get() returns a Document directly without Set() validation,
	// but our JSON round-trip triggers Set() which rejects secondary FK fields.
	for _, field := range c.version.Fields {
		if field.Kind.IsObject() && !field.IsPrimary {
			fkName := "_" + field.Name + "ID"
			delete(docData, fkName)
		}
	}

	// Create a new document from the retrieved data
	doc, err := client.NewDocFromMap(ctx, docData, c.version)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Clean the document so only subsequent Set() calls mark fields as dirty.
	// Without this, all fields loaded from the query are "dirty" and ToJSONPatch()
	// would include unchanged fields in update mutations.
	doc.Clean()

	return doc, nil
}

// checkIfDocumentDeleted checks if a document is deleted and returns the appropriate error
func (c *CollectionWrapper) checkIfDocumentDeleted(ctx context.Context, docID client.DocID) error {
	deletedQuery := fmt.Sprintf(
		`{ %s(docID: "%s", showDeleted: true) { _docID _deleted } }`,
		c.version.Name, docID.String(),
	)
	deletedResult := c.wrapper.ExecRequest(ctx, deletedQuery)
	if len(deletedResult.GQL.Errors) == 0 {
		if deletedData, ok := deletedResult.GQL.Data.(map[string]any); ok {
			if deletedDocs, ok := deletedData[c.version.Name].([]any); ok && len(deletedDocs) > 0 {
				if docData, ok := deletedDocs[0].(map[string]any); ok {
					if deleted, ok := docData["_deleted"].(bool); ok && deleted {
						return fmt.Errorf("a document with the given ID has been deleted")
					}
				}
			}
		}
	}
	// When the collection has an ACP policy, use the generic message to avoid
	// revealing whether the document exists (matches Go DefraDB behavior).
	if c.version.Policy.HasValue() {
		return fmt.Errorf("document not found or not authorized to access")
	}
	return fmt.Errorf("document not found: %s", docID.String())
}

func (c *CollectionWrapper) NewIndex(
	ctx context.Context,
	req client.NewIndexRequest,
	opts ...options.Enumerable[options.NewCollectionIndexOptions],
) (client.IndexDescription, error) {
	if c.txn != nil {
		opt := utils.NewOptions(opts...)
		copiedReq := req
		c.txn.stagedIndexOps = append(c.txn.stagedIndexOps, stagedIndexOp{
			identityDID:    identityDIDFromOption(opt.GetIdentity()),
			collectionName: c.version.Name,
			create:         &copiedReq,
		})

		id := uint32(len(c.version.Indexes) + 1)
		for _, op := range c.txn.stagedIndexOps {
			if op.collectionName == c.version.Name && op.create != nil && op.create.Name != req.Name {
				id++
			}
		}

		return client.IndexDescription{
			Name:   req.Name,
			ID:     id,
			Fields: req.Fields,
			Unique: req.Unique,
		}, nil
	}

	fields := make([]IndexField, len(req.Fields))
	for i, f := range req.Fields {
		fields[i] = IndexField{
			Name:       f.Name,
			Descending: f.Descending,
		}
	}

	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	index, err := c.wrapper.node.CreateIndex(identityDID, c.version.Name, req.Name, fields, req.Unique)
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

func (c *CollectionWrapper) DeleteIndex(ctx context.Context, indexName string, opts ...options.Enumerable[options.DeleteCollectionIndexOptions]) error {
	if c.txn != nil {
		opt := utils.NewOptions(opts...)
		c.txn.stagedIndexOps = append(c.txn.stagedIndexOps, stagedIndexOp{
			identityDID:    identityDIDFromOption(opt.GetIdentity()),
			collectionName: c.version.Name,
			deleteName:     indexName,
		})
		return nil
	}

	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	return c.wrapper.node.DropIndex(identityDID, c.version.Name, indexName)
}

func (c *CollectionWrapper) ListIndexes(ctx context.Context, opts ...options.Enumerable[options.ListCollectionIndexesOptions]) ([]client.IndexDescription, error) {
	if c.txn != nil {
		indexes := append([]client.IndexDescription(nil), c.version.Indexes...)
		for _, op := range c.txn.stagedIndexOps {
			if op.collectionName != c.version.Name {
				continue
			}
			if op.create != nil {
				indexes = append(indexes, client.IndexDescription{
					Name:   op.create.Name,
					ID:     uint32(len(indexes) + 1),
					Fields: op.create.Fields,
					Unique: op.create.Unique,
				})
				continue
			}
			filtered := indexes[:0]
			for _, idx := range indexes {
				if idx.Name != op.deleteName {
					filtered = append(filtered, idx)
				}
			}
			indexes = filtered
		}
		return indexes, nil
	}

	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	indexes, err := c.wrapper.node.GetIndexes(identityDID, c.version.Name)
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

func (c *CollectionWrapper) NewEncryptedIndex(
	ctx context.Context,
	desc client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.NewEncryptedIndexOptions],
) (client.EncryptedIndexDescription, error) {
	if c.txn != nil {
		opt := utils.NewOptions(opts...)
		c.txn.stagedEncryptedIndexOps = append(c.txn.stagedEncryptedIndexOps, stagedEncryptedIndexOp{
			identityDID:    identityDIDFromOption(opt.GetIdentity()),
			collectionName: c.version.Name,
			createField:    desc.FieldName,
		})
		return desc, nil
	}

	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	ffiIdx, err := c.wrapper.node.CreateEncryptedIndex(identityDID, c.version.Name, desc.FieldName)
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}

	return client.EncryptedIndexDescription{
		FieldName: ffiIdx.FieldName,
		Type:      client.EncryptedIndexType(ffiIdx.Type),
	}, nil
}

func (c *CollectionWrapper) DeleteEncryptedIndex(ctx context.Context, fieldName string, opts ...options.Enumerable[options.DeleteEncryptedIndexOptions]) error {
	if c.txn != nil {
		opt := utils.NewOptions(opts...)
		c.txn.stagedEncryptedIndexOps = append(c.txn.stagedEncryptedIndexOps, stagedEncryptedIndexOp{
			identityDID:    identityDIDFromOption(opt.GetIdentity()),
			collectionName: c.version.Name,
			deleteField:    fieldName,
		})
		return nil
	}

	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	return c.wrapper.node.DeleteEncryptedIndex(identityDID, c.version.Name, fieldName)
}

func (c *CollectionWrapper) ListEncryptedIndexes(ctx context.Context, opts ...options.Enumerable[options.ListCollectionEncryptedIndexesOptions]) ([]client.EncryptedIndexDescription, error) {
	if c.txn != nil {
		indexes := append([]client.EncryptedIndexDescription(nil), c.version.EncryptedIndexes...)
		for _, op := range c.txn.stagedEncryptedIndexOps {
			if op.collectionName != c.version.Name {
				continue
			}
			if op.createField != "" {
				indexes = append(indexes, client.EncryptedIndexDescription{
					FieldName: op.createField,
					Type:      client.EncryptedIndexTypeEquality,
				})
				continue
			}
			filtered := indexes[:0]
			for _, idx := range indexes {
				if idx.FieldName != op.deleteField {
					filtered = append(filtered, idx)
				}
			}
			indexes = filtered
		}
		return indexes, nil
	}

	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	ffiIndexes, err := c.wrapper.node.ListEncryptedIndexes(identityDID, c.version.Name)
	if err != nil {
		return nil, err
	}

	result := make([]client.EncryptedIndexDescription, len(ffiIndexes))
	for i, idx := range ffiIndexes {
		result[i] = client.EncryptedIndexDescription{
			FieldName: idx.FieldName,
			Type:      client.EncryptedIndexType(idx.Type),
		}
	}
	return result, nil
}

func (c *CollectionWrapper) Truncate(ctx context.Context, opts ...options.Enumerable[options.TruncateCollectionOptions]) error {
	opt := utils.NewOptions(opts...)
	identityDID := identityDIDFromOption(opt.GetIdentity())

	return c.wrapper.node.TruncateCollection(identityDID, c.version.Name)
}
