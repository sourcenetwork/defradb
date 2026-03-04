// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package options

import (
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
)

// NodeKMSType represents different KMS service types.
type NodeKMSType string

const (
	// NodePubSubKMSType is the KMS type that uses PubSub mechanism.
	NodePubSubKMSType NodeKMSType = "pubsub"
)

// NodeStoreType represents different store implementations.
type NodeStoreType string

const (
	// NodeDefaultStore is the default store type.
	NodeDefaultStore NodeStoreType = ""
	// NodeBadgerStore specifies the badger datastore.
	NodeBadgerStore NodeStoreType = "badger"
	// NodeMemoryStore specifies the in-memory datastore.
	NodeMemoryStore NodeStoreType = "memory"
)

// NodeDocumentACPType represents different document ACP implementations.
type NodeDocumentACPType string

const (
	// NodeNoDocumentACPType disables the document ACP subsystem.
	NodeNoDocumentACPType NodeDocumentACPType = "none"
	// NodeDefaultDocumentACPType uses the default ACP implementation for this build.
	NodeDefaultDocumentACPType NodeDocumentACPType = ""
	// NodeLocalDocumentACPType uses the local ACP implementation.
	NodeLocalDocumentACPType NodeDocumentACPType = "local"
	// NodeSourceHubDocumentACPType uses the SourceHub ACP implementation.
	NodeSourceHubDocumentACPType NodeDocumentACPType = "source-hub"
)

// NodeLensRuntimeType represents the lens runtime type.
type NodeLensRuntimeType string

const (
	// NodeDefaultLensRuntime is the default lens runtime type.
	// The actual runtime type that this resolves to depends on the build target.
	NodeDefaultLensRuntime NodeLensRuntimeType = ""
	// NodeWASMLensRuntime is the WASM lens runtime.
	NodeWASMLensRuntime NodeLensRuntimeType = "wasm"
	// NodeJSLensRuntime is the JavaScript lens runtime (for JS/WASM builds).
	NodeJSLensRuntime NodeLensRuntimeType = "js"
)

// NodeTxSigner models an entity capable of providing signatures for a Tx.
type NodeTxSigner interface {
	GetAccAddress() string
	GetPrivateKey() cryptotypes.PrivKey
}

// NodeOptions is the unified configuration for a DefraDB node.
// It contains all configuration needed to create and start a node.
type NodeOptions struct {
	// DisableP2P disables the P2P networking system.
	DisableP2P bool
	// DisableAPI disables the HTTP API server.
	DisableAPI bool
	// EnableDevelopment enables development mode features.
	EnableDevelopment bool
	// KMSType specifies the key management system type.
	KMSType immutable.Option[NodeKMSType]

	// Store contains store configuration.
	Store NodeStoreOptions
	// DocumentACP contains document ACP configuration.
	DocumentACP NodeDocumentACPOptions
	// NodeACP contains node ACP configuration.
	NodeACP NodeACPOptions
	// DB contains database configuration.
	DB NodeDBOptions
	// P2P contains P2P networking configuration.
	P2P NodeP2POptions
	// HTTP contains HTTP API server configuration.
	HTTP NodeHTTPOptions
}

// NodeP2POptions contains P2P networking configuration values.
type NodeP2POptions struct {
	// ListenAddresses are the addresses to listen on for P2P connections.
	ListenAddresses []string
	// BootstrapPeers are the addresses of peers to connect to on startup.
	BootstrapPeers []string
	// EnablePubSub enables the PubSub system.
	EnablePubSub bool
	// EnableRelay enables the relay system.
	EnableRelay bool
	// EnableClearBackoffOnRetry enables clearing backoff on retry for connections.
	EnableClearBackoffOnRetry bool
	// PrivateKey is the private key for the P2P node.
	PrivateKey []byte
}

// NodeHTTPOptions contains HTTP API server configuration values.
type NodeHTTPOptions struct {
	// Address is the address to listen on for HTTP connections.
	Address string
	// AllowedOrigins are the allowed CORS origins.
	AllowedOrigins []string
	// TLSCertPath is the path to the TLS certificate file.
	TLSCertPath string
	// TLSKeyPath is the path to the TLS private key file.
	TLSKeyPath string
	// ReadTimeout is the read timeout for connections.
	ReadTimeout time.Duration
	// WriteTimeout is the write timeout for connections.
	WriteTimeout time.Duration
	// IdleTimeout is the idle timeout for connections.
	IdleTimeout time.Duration
}

// NodeStoreOptions contains store configuration values.
type NodeStoreOptions struct {
	// Store specifies the store type (badger, memory, etc.).
	Store NodeStoreType
	// Path is the filesystem path for the store.
	Path string
	// BadgerFileSize is the maximum file size for Badger.
	BadgerFileSize int64
	// BadgerEncryptionKey is the encryption key for Badger.
	BadgerEncryptionKey []byte
	// BadgerInMemory specifies whether to run Badger in-memory.
	BadgerInMemory bool
}

// NodeDocumentACPOptions contains document ACP configuration values.
type NodeDocumentACPOptions struct {
	// DocumentACPType specifies the document ACP implementation to use.
	DocumentACPType NodeDocumentACPType
	// Path is the filesystem path for the document ACP system.
	Path string
	// Signer is the transaction signer for SourceHub ACP.
	Signer immutable.Option[NodeTxSigner]
	// SourceHubChainID is the chain ID for SourceHub.
	SourceHubChainID string
	// SourceHubGRPCAddress is the gRPC address for SourceHub.
	SourceHubGRPCAddress string
	// SourceHubCometRPCAddress is the Comet RPC address for SourceHub.
	SourceHubCometRPCAddress string
}

// NodeACPOptions contains node ACP configuration values.
type NodeACPOptions struct {
	// IsEnabled specifies whether node ACP is enabled.
	IsEnabled bool
	// Path is the filesystem path for the node ACP system.
	Path string
}

// NodeDBOptions contains database configuration values.
type NodeDBOptions struct {
	// MaxTxnRetries is the maximum number of retries per transaction.
	MaxTxnRetries immutable.Option[int]
	// Identity is the identity to use for the node.
	Identity immutable.Option[identity.Identity]
	// EnableSigning enables block signing.
	EnableSigning bool
	// SearchableEncryptionKey is the key used for searchable encryption.
	SearchableEncryptionKey []byte
	// RetryIntervals are the intervals between transaction retries.
	RetryIntervals []time.Duration
	// P2PBlockSyncTimeout is the timeout duration for syncing block links.
	P2PBlockSyncTimeout time.Duration
	// LensRuntime specifies the lens runtime type.
	LensRuntime NodeLensRuntimeType
	// LensPoolSize is the pool size for the lens runtime.
	LensPoolSize int
	// ChunkSize is the chunk size for the blockstore.
	ChunkSize immutable.Option[int]
}

// nodeSubBuilder provides parent linkage, forwarding, and Node() navigation
// for sub-builders. Embed in sub-builders alongside nothing else — it already
// embeds enumerableBuilder[T].
type nodeSubBuilder[T any] struct {
	enumerableBuilder[T]
	parent  *NodeOptionsBuilder // nil when standalone
	project func(*NodeOptions) *T
}

// append records the option locally and, when linked to a parent, forwards it.
func (l *nodeSubBuilder[T]) append(fn func(*T)) {
	l.enumerableBuilder.append(fn)
	if l.parent != nil {
		l.parent.append(func(o *NodeOptions) {
			fn(l.project(o))
		})
	}
}

// mustParent panics with a descriptive message if the parent is nil.
func (l *nodeSubBuilder[T]) mustParent() {
	if l.parent == nil {
		panic("nodeSubBuilder: parent is nil; use Node() constructor or a parent sub-builder")
	}
}

// Node returns the parent builder.
func (l *nodeSubBuilder[T]) Node() *NodeOptionsBuilder {
	l.mustParent()
	return l.parent
}

// Store navigates to a Store sub-builder on the parent.
func (l *nodeSubBuilder[T]) Store() *NodeStoreOptionsBuilder {
	l.mustParent()
	return l.parent.Store()
}

// DB navigates to a DB sub-builder on the parent.
func (l *nodeSubBuilder[T]) DB() *NodeDBOptionsBuilder {
	l.mustParent()
	return l.parent.DB()
}

// P2P navigates to a P2P sub-builder on the parent.
func (l *nodeSubBuilder[T]) P2P() *NodeP2POptionsBuilder {
	l.mustParent()
	return l.parent.P2P()
}

// HTTP navigates to a HTTP sub-builder on the parent.
func (l *nodeSubBuilder[T]) HTTP() *NodeHTTPOptionsBuilder {
	l.mustParent()
	return l.parent.HTTP()
}

// DocumentACP navigates to a DocumentACP sub-builder on the parent.
func (l *nodeSubBuilder[T]) DocumentACP() *NodeDocumentACPOptionsBuilder {
	l.mustParent()
	return l.parent.DocumentACP()
}

// NodeACP navigates to a NodeACP sub-builder on the parent.
func (l *nodeSubBuilder[T]) NodeACP() *NodeACPOptionsBuilder {
	l.mustParent()
	return l.parent.NodeACP()
}

// NodeOptionsBuilder is a builder for NodeOptions.
type NodeOptionsBuilder struct {
	enumerableBuilder[NodeOptions]
}

// Node creates a new NodeOptionsBuilder instance.
func Node() *NodeOptionsBuilder {
	return &NodeOptionsBuilder{}
}

// SetDisableP2P sets the disable P2P flag.
func (b *NodeOptionsBuilder) SetDisableP2P(disable bool) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DisableP2P = disable
	})
	return b
}

// SetDisableAPI sets the disable API flag.
func (b *NodeOptionsBuilder) SetDisableAPI(disable bool) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DisableAPI = disable
	})
	return b
}

// SetEnableDevelopment sets the enable development mode flag.
func (b *NodeOptionsBuilder) SetEnableDevelopment(enable bool) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.EnableDevelopment = enable
	})
	return b
}

// SetKMS sets the KMS type.
func (b *NodeOptionsBuilder) SetKMS(kmsType NodeKMSType) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.KMSType = immutable.Some(kmsType)
	})
	return b
}

// Store returns a linked NodeStoreOptionsBuilder for scoped chaining.
func (b *NodeOptionsBuilder) Store() *NodeStoreOptionsBuilder {
	return &NodeStoreOptionsBuilder{nodeSubBuilder[NodeStoreOptions]{parent: b,
		project: func(o *NodeOptions) *NodeStoreOptions { return &o.Store }}}
}

// DB returns a linked NodeDBOptionsBuilder for scoped chaining.
func (b *NodeOptionsBuilder) DB() *NodeDBOptionsBuilder {
	return &NodeDBOptionsBuilder{nodeSubBuilder[NodeDBOptions]{parent: b,
		project: func(o *NodeOptions) *NodeDBOptions { return &o.DB }}}
}

// P2P returns a linked NodeP2POptionsBuilder for scoped chaining.
func (b *NodeOptionsBuilder) P2P() *NodeP2POptionsBuilder {
	return &NodeP2POptionsBuilder{nodeSubBuilder[NodeP2POptions]{parent: b,
		project: func(o *NodeOptions) *NodeP2POptions { return &o.P2P }}}
}

// HTTP returns a linked NodeHTTPOptionsBuilder for scoped chaining.
func (b *NodeOptionsBuilder) HTTP() *NodeHTTPOptionsBuilder {
	return &NodeHTTPOptionsBuilder{nodeSubBuilder[NodeHTTPOptions]{parent: b,
		project: func(o *NodeOptions) *NodeHTTPOptions { return &o.HTTP }}}
}

// DocumentACP returns a linked NodeDocumentACPOptionsBuilder for scoped chaining.
func (b *NodeOptionsBuilder) DocumentACP() *NodeDocumentACPOptionsBuilder {
	return &NodeDocumentACPOptionsBuilder{nodeSubBuilder[NodeDocumentACPOptions]{parent: b,
		project: func(o *NodeOptions) *NodeDocumentACPOptions { return &o.DocumentACP }}}
}

// NodeACP returns a linked NodeACPOptionsBuilder for scoped chaining.
func (b *NodeOptionsBuilder) NodeACP() *NodeACPOptionsBuilder {
	return &NodeACPOptionsBuilder{nodeSubBuilder[NodeACPOptions]{parent: b,
		project: func(o *NodeOptions) *NodeACPOptions { return &o.NodeACP }}}
}

// NodeStoreOptionsBuilder is a builder for NodeStoreOptions.
// It can be used standalone or as part of a NodeOptionsBuilder chain.
type NodeStoreOptionsBuilder struct {
	nodeSubBuilder[NodeStoreOptions]
}

// NodeStore creates a standalone NodeStoreOptionsBuilder.
func NodeStore() *NodeStoreOptionsBuilder {
	return &NodeStoreOptionsBuilder{}
}

// SetType sets the store type.
func (sb *NodeStoreOptionsBuilder) SetType(store NodeStoreType) *NodeStoreOptionsBuilder {
	sb.append(func(opts *NodeStoreOptions) { opts.Store = store })
	return sb
}

// SetPath sets the store path.
func (sb *NodeStoreOptionsBuilder) SetPath(path string) *NodeStoreOptionsBuilder {
	sb.append(func(opts *NodeStoreOptions) { opts.Path = path })
	return sb
}

// SetBadgerFileSize sets the Badger file size.
func (sb *NodeStoreOptionsBuilder) SetBadgerFileSize(size int64) *NodeStoreOptionsBuilder {
	sb.append(func(opts *NodeStoreOptions) { opts.BadgerFileSize = size })
	return sb
}

// SetBadgerEncryptionKey sets the Badger encryption key.
func (sb *NodeStoreOptionsBuilder) SetBadgerEncryptionKey(key []byte) *NodeStoreOptionsBuilder {
	sb.append(func(opts *NodeStoreOptions) { opts.BadgerEncryptionKey = key })
	return sb
}

// SetBadgerInMemory sets whether Badger should run in-memory.
func (sb *NodeStoreOptionsBuilder) SetBadgerInMemory(inMemory bool) *NodeStoreOptionsBuilder {
	sb.append(func(opts *NodeStoreOptions) { opts.BadgerInMemory = inMemory })
	return sb
}

// SetAll sets all store options from a plain data struct.
func (sb *NodeStoreOptionsBuilder) SetAll(storeOpts NodeStoreOptions) *NodeStoreOptionsBuilder {
	sb.append(func(opts *NodeStoreOptions) { *opts = storeOpts })
	return sb
}

// NodeDBOptionsBuilder is a builder for NodeDBOptions.
// It can be used standalone or as part of a NodeOptionsBuilder chain.
type NodeDBOptionsBuilder struct {
	nodeSubBuilder[NodeDBOptions]
}

// NodeDB creates a standalone NodeDBOptionsBuilder.
func NodeDB() *NodeDBOptionsBuilder {
	return &NodeDBOptionsBuilder{}
}

// SetMaxTxnRetries sets the maximum number of retries per transaction.
func (sb *NodeDBOptionsBuilder) SetMaxTxnRetries(num int) *NodeDBOptionsBuilder {
	sb.append(func(opts *NodeDBOptions) { opts.MaxTxnRetries = immutable.Some(num) })
	return sb
}

// SetNodeIdentity sets the identity for the node.
func (sb *NodeDBOptionsBuilder) SetNodeIdentity(ident identity.Identity) *NodeDBOptionsBuilder {
	sb.append(func(opts *NodeDBOptions) { opts.Identity = immutable.Some(ident) })
	return sb
}

// SetEnableSigning sets whether block signing is enabled.
func (sb *NodeDBOptionsBuilder) SetEnableSigning(enable bool) *NodeDBOptionsBuilder {
	sb.append(func(opts *NodeDBOptions) { opts.EnableSigning = enable })
	return sb
}

// SetSearchableEncryptionKey sets the key used for searchable encryption.
func (sb *NodeDBOptionsBuilder) SetSearchableEncryptionKey(key []byte) *NodeDBOptionsBuilder {
	sb.append(func(opts *NodeDBOptions) { opts.SearchableEncryptionKey = key })
	return sb
}

// SetRetryIntervals sets the intervals between transaction retries.
func (sb *NodeDBOptionsBuilder) SetRetryIntervals(intervals []time.Duration) *NodeDBOptionsBuilder {
	sb.append(func(opts *NodeDBOptions) {
		if len(intervals) > 0 {
			opts.RetryIntervals = intervals
		}
	})
	return sb
}

// SetP2PBlockSyncTimeout sets the timeout duration for syncing block links.
func (sb *NodeDBOptionsBuilder) SetP2PBlockSyncTimeout(timeout time.Duration) *NodeDBOptionsBuilder {
	sb.append(func(opts *NodeDBOptions) { opts.P2PBlockSyncTimeout = timeout })
	return sb
}

// SetLensRuntime sets the lens runtime type.
func (sb *NodeDBOptionsBuilder) SetLensRuntime(runtime NodeLensRuntimeType) *NodeDBOptionsBuilder {
	sb.append(func(opts *NodeDBOptions) { opts.LensRuntime = runtime })
	return sb
}

// SetLensPoolSize sets the pool size for the lens runtime.
func (sb *NodeDBOptionsBuilder) SetLensPoolSize(size int) *NodeDBOptionsBuilder {
	sb.append(func(opts *NodeDBOptions) { opts.LensPoolSize = size })
	return sb
}

// SetChunkSize sets the chunk size for the blockstore.
func (sb *NodeDBOptionsBuilder) SetChunkSize(size int) *NodeDBOptionsBuilder {
	sb.append(func(opts *NodeDBOptions) { opts.ChunkSize = immutable.Some(size) })
	return sb
}

// SetAll sets all DB options from a plain data struct.
func (sb *NodeDBOptionsBuilder) SetAll(dbOpts NodeDBOptions) *NodeDBOptionsBuilder {
	sb.append(func(opts *NodeDBOptions) { *opts = dbOpts })
	return sb
}

// NodeP2POptionsBuilder is a builder for NodeP2POptions.
// It can be used standalone or as part of a NodeOptionsBuilder chain.
type NodeP2POptionsBuilder struct {
	nodeSubBuilder[NodeP2POptions]
}

// NodeP2P creates a standalone NodeP2POptionsBuilder.
func NodeP2P() *NodeP2POptionsBuilder {
	return &NodeP2POptionsBuilder{}
}

// SetListenAddresses sets the listen addresses.
func (sb *NodeP2POptionsBuilder) SetListenAddresses(addresses ...string) *NodeP2POptionsBuilder {
	sb.append(func(opts *NodeP2POptions) { opts.ListenAddresses = addresses })
	return sb
}

// SetBootstrapPeers sets the bootstrap peers.
func (sb *NodeP2POptionsBuilder) SetBootstrapPeers(peers ...string) *NodeP2POptionsBuilder {
	sb.append(func(opts *NodeP2POptions) { opts.BootstrapPeers = peers })
	return sb
}

// SetEnablePubSub sets whether PubSub is enabled.
func (sb *NodeP2POptionsBuilder) SetEnablePubSub(enable bool) *NodeP2POptionsBuilder {
	sb.append(func(opts *NodeP2POptions) { opts.EnablePubSub = enable })
	return sb
}

// SetEnableRelay sets whether relay is enabled.
func (sb *NodeP2POptionsBuilder) SetEnableRelay(enable bool) *NodeP2POptionsBuilder {
	sb.append(func(opts *NodeP2POptions) { opts.EnableRelay = enable })
	return sb
}

// SetEnableClearBackoffOnRetry sets whether to clear backoff on retry.
func (sb *NodeP2POptionsBuilder) SetEnableClearBackoffOnRetry(enable bool) *NodeP2POptionsBuilder {
	sb.append(func(opts *NodeP2POptions) { opts.EnableClearBackoffOnRetry = enable })
	return sb
}

// SetPrivateKey sets the private key for the P2P node.
func (sb *NodeP2POptionsBuilder) SetPrivateKey(key []byte) *NodeP2POptionsBuilder {
	sb.append(func(opts *NodeP2POptions) { opts.PrivateKey = key })
	return sb
}

// SetAll sets all P2P options from a plain data struct.
func (sb *NodeP2POptionsBuilder) SetAll(p2pOpts NodeP2POptions) *NodeP2POptionsBuilder {
	sb.append(func(opts *NodeP2POptions) { *opts = p2pOpts })
	return sb
}

// NodeHTTPOptionsBuilder is a builder for NodeHTTPOptions.
// It can be used standalone or as part of a NodeOptionsBuilder chain.
type NodeHTTPOptionsBuilder struct {
	nodeSubBuilder[NodeHTTPOptions]
}

// NodeHTTP creates a standalone NodeHTTPOptionsBuilder.
func NodeHTTP() *NodeHTTPOptionsBuilder {
	return &NodeHTTPOptionsBuilder{}
}

// SetAddress sets the HTTP server address.
func (sb *NodeHTTPOptionsBuilder) SetAddress(address string) *NodeHTTPOptionsBuilder {
	sb.append(func(opts *NodeHTTPOptions) { opts.Address = address })
	return sb
}

// SetAllowedOrigins sets the allowed CORS origins.
func (sb *NodeHTTPOptionsBuilder) SetAllowedOrigins(origins ...string) *NodeHTTPOptionsBuilder {
	sb.append(func(opts *NodeHTTPOptions) { opts.AllowedOrigins = origins })
	return sb
}

// SetCertPath sets the path to the TLS certificate file.
func (sb *NodeHTTPOptionsBuilder) SetCertPath(path string) *NodeHTTPOptionsBuilder {
	sb.append(func(opts *NodeHTTPOptions) { opts.TLSCertPath = path })
	return sb
}

// SetKeyPath sets the path to the TLS private key file.
func (sb *NodeHTTPOptionsBuilder) SetKeyPath(path string) *NodeHTTPOptionsBuilder {
	sb.append(func(opts *NodeHTTPOptions) { opts.TLSKeyPath = path })
	return sb
}

// SetReadTimeout sets the server read timeout.
func (sb *NodeHTTPOptionsBuilder) SetReadTimeout(timeout time.Duration) *NodeHTTPOptionsBuilder {
	sb.append(func(opts *NodeHTTPOptions) { opts.ReadTimeout = timeout })
	return sb
}

// SetWriteTimeout sets the server write timeout.
func (sb *NodeHTTPOptionsBuilder) SetWriteTimeout(timeout time.Duration) *NodeHTTPOptionsBuilder {
	sb.append(func(opts *NodeHTTPOptions) { opts.WriteTimeout = timeout })
	return sb
}

// SetIdleTimeout sets the server idle timeout.
func (sb *NodeHTTPOptionsBuilder) SetIdleTimeout(timeout time.Duration) *NodeHTTPOptionsBuilder {
	sb.append(func(opts *NodeHTTPOptions) { opts.IdleTimeout = timeout })
	return sb
}

// SetAll sets all HTTP options from a plain data struct.
func (sb *NodeHTTPOptionsBuilder) SetAll(httpOpts NodeHTTPOptions) *NodeHTTPOptionsBuilder {
	sb.append(func(opts *NodeHTTPOptions) { *opts = httpOpts })
	return sb
}

// NodeDocumentACPOptionsBuilder is a builder for NodeDocumentACPOptions.
// It can be used standalone or as part of a NodeOptionsBuilder chain.
type NodeDocumentACPOptionsBuilder struct {
	nodeSubBuilder[NodeDocumentACPOptions]
}

// NodeDocumentACP creates a standalone NodeDocumentACPOptionsBuilder.
func NodeDocumentACP() *NodeDocumentACPOptionsBuilder {
	return &NodeDocumentACPOptionsBuilder{}
}

// SetType sets the document ACP type.
func (sb *NodeDocumentACPOptionsBuilder) SetType(acpType NodeDocumentACPType) *NodeDocumentACPOptionsBuilder {
	sb.append(func(opts *NodeDocumentACPOptions) { opts.DocumentACPType = acpType })
	return sb
}

// SetPath sets the document ACP system path.
func (sb *NodeDocumentACPOptionsBuilder) SetPath(path string) *NodeDocumentACPOptionsBuilder {
	sb.append(func(opts *NodeDocumentACPOptions) { opts.Path = path })
	return sb
}

// SetTxnSigner sets the txn signer for Defra to use.
func (sb *NodeDocumentACPOptionsBuilder) SetTxnSigner(signer NodeTxSigner) *NodeDocumentACPOptionsBuilder {
	sb.append(func(opts *NodeDocumentACPOptions) { opts.Signer = immutable.Some(signer) })
	return sb
}

// SetChainID sets the chainID of the SourceHub chain.
func (sb *NodeDocumentACPOptionsBuilder) SetChainID(chainID string) *NodeDocumentACPOptionsBuilder {
	sb.append(func(opts *NodeDocumentACPOptions) { opts.SourceHubChainID = chainID })
	return sb
}

// SetGRPCAddress sets the GRPC address of the SourceHub node.
func (sb *NodeDocumentACPOptionsBuilder) SetGRPCAddress(address string) *NodeDocumentACPOptionsBuilder {
	sb.append(func(opts *NodeDocumentACPOptions) { opts.SourceHubGRPCAddress = address })
	return sb
}

// SetCometRPCAddress sets the Comet RPC address of the SourceHub node.
func (sb *NodeDocumentACPOptionsBuilder) SetCometRPCAddress(address string) *NodeDocumentACPOptionsBuilder {
	sb.append(func(opts *NodeDocumentACPOptions) { opts.SourceHubCometRPCAddress = address })
	return sb
}

// SetAll sets all document ACP options from a plain data struct.
func (sb *NodeDocumentACPOptionsBuilder) SetAll(dacOpts NodeDocumentACPOptions) *NodeDocumentACPOptionsBuilder {
	sb.append(func(opts *NodeDocumentACPOptions) { *opts = dacOpts })
	return sb
}

// NodeACPOptionsBuilder is a builder for NodeACPOptions.
// It can be used standalone or as part of a NodeOptionsBuilder chain.
type NodeACPOptionsBuilder struct {
	nodeSubBuilder[NodeACPOptions]
}

// NodeACP creates a standalone NodeACPOptionsBuilder.
func NodeACP() *NodeACPOptionsBuilder {
	return &NodeACPOptionsBuilder{}
}

// SetEnabled sets whether node ACP is enabled.
func (sb *NodeACPOptionsBuilder) SetEnabled(enabled bool) *NodeACPOptionsBuilder {
	sb.append(func(opts *NodeACPOptions) { opts.IsEnabled = enabled })
	return sb
}

// SetPath sets the node ACP system path.
func (sb *NodeACPOptionsBuilder) SetPath(path string) *NodeACPOptionsBuilder {
	sb.append(func(opts *NodeACPOptions) { opts.Path = path })
	return sb
}

// SetAll sets all node ACP options from a plain data struct.
func (sb *NodeACPOptionsBuilder) SetAll(nacOpts NodeACPOptions) *NodeACPOptionsBuilder {
	sb.append(func(opts *NodeACPOptions) { *opts = nacOpts })
	return sb
}
