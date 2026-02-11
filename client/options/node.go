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
	// EnableSigning enables block signing. Defaults to true.
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

// DefaultNodeOptions returns default NodeOptions values.
func DefaultNodeOptions() NodeOptions {
	return NodeOptions{
		DisableP2P:        false,
		DisableAPI:        false,
		EnableDevelopment: false,
		Store: NodeStoreOptions{
			Store:          NodeDefaultStore,
			BadgerInMemory: false,
			BadgerFileSize: 1 << 30, // 1GB
		},
		DocumentACP: NodeDocumentACPOptions{
			DocumentACPType: NodeLocalDocumentACPType,
		},
		NodeACP: NodeACPOptions{
			IsEnabled: false,
		},
		DB: NodeDBOptions{
			MaxTxnRetries: immutable.Some(5),
			EnableSigning: true,
			RetryIntervals: []time.Duration{
				time.Second * 30,
				time.Minute,
				time.Minute * 2,
				time.Minute * 4,
				time.Minute * 8,
				time.Minute * 16,
				time.Minute * 32,
			},
			P2PBlockSyncTimeout: time.Second * 5,
			LensRuntime:         NodeDefaultLensRuntime,
		},
		P2P:  NodeP2POptions{},
		HTTP: NodeHTTPOptions{},
	}
}

// NodeOptionsBuilder is a builder for NodeOptions.
type NodeOptionsBuilder struct {
	enumerableBuilder[NodeOptions]
}

// Node creates a new NodeOptionsBuilder instance with default values.
func Node() *NodeOptionsBuilder {
	b := &NodeOptionsBuilder{}
	b.append(func(opts *NodeOptions) {
		defaults := DefaultNodeOptions()
		*opts = defaults
	})
	return b
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

// SetNodeIdentity sets the identity for the node.
func (b *NodeOptionsBuilder) SetNodeIdentity(ident identity.Identity) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DB.Identity = immutable.Some(ident)
	})
	return b
}

// --- Store setters ---

// SetStoreType sets the store type.
func (b *NodeOptionsBuilder) SetStoreType(store NodeStoreType) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.Store.Store = store
	})
	return b
}

// SetStorePath sets the store path.
func (b *NodeOptionsBuilder) SetStorePath(path string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.Store.Path = path
	})
	return b
}

// SetBadgerFileSize sets the Badger file size.
func (b *NodeOptionsBuilder) SetBadgerFileSize(size int64) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.Store.BadgerFileSize = size
	})
	return b
}

// SetBadgerEncryptionKey sets the Badger encryption key.
func (b *NodeOptionsBuilder) SetBadgerEncryptionKey(key []byte) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.Store.BadgerEncryptionKey = key
	})
	return b
}

// SetBadgerInMemory sets whether Badger should run in-memory.
func (b *NodeOptionsBuilder) SetBadgerInMemory(inMemory bool) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.Store.BadgerInMemory = inMemory
	})
	return b
}

// --- DocumentACP setters ---

// SetDocumentACPType sets the document ACP type.
func (b *NodeOptionsBuilder) SetDocumentACPType(acpType NodeDocumentACPType) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DocumentACP.DocumentACPType = acpType
	})
	return b
}

// SetDocumentACPPath sets the document ACP system path.
func (b *NodeOptionsBuilder) SetDocumentACPPath(path string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DocumentACP.Path = path
	})
	return b
}

// SetTxnSigner sets the txn signer for Defra to use.
func (b *NodeOptionsBuilder) SetTxnSigner(signer NodeTxSigner) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DocumentACP.Signer = immutable.Some(signer)
	})
	return b
}

// SetSourceHubChainID sets the chainID of the SourceHub chain.
func (b *NodeOptionsBuilder) SetSourceHubChainID(chainID string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DocumentACP.SourceHubChainID = chainID
	})
	return b
}

// SetSourceHubGRPCAddress sets the GRPC address of the SourceHub node.
func (b *NodeOptionsBuilder) SetSourceHubGRPCAddress(address string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DocumentACP.SourceHubGRPCAddress = address
	})
	return b
}

// SetSourceHubCometRPCAddress sets the Comet RPC address of the SourceHub node.
func (b *NodeOptionsBuilder) SetSourceHubCometRPCAddress(address string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DocumentACP.SourceHubCometRPCAddress = address
	})
	return b
}

// --- NodeACP setters ---

// SetNodeACPEnabled sets whether node ACP is enabled.
func (b *NodeOptionsBuilder) SetNodeACPEnabled(enabled bool) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.NodeACP.IsEnabled = enabled
	})
	return b
}

// SetNodeACPPath sets the node ACP system path.
func (b *NodeOptionsBuilder) SetNodeACPPath(path string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.NodeACP.Path = path
	})
	return b
}

// --- DB setters ---

// SetMaxTxnRetries sets the maximum number of retries per transaction.
func (b *NodeOptionsBuilder) SetMaxTxnRetries(num int) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DB.MaxTxnRetries = immutable.Some(num)
	})
	return b
}

// SetEnableSigning sets whether block signing is enabled.
func (b *NodeOptionsBuilder) SetEnableSigning(enable bool) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DB.EnableSigning = enable
	})
	return b
}

// SetSearchableEncryptionKey sets the key used for searchable encryption.
func (b *NodeOptionsBuilder) SetSearchableEncryptionKey(key []byte) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DB.SearchableEncryptionKey = key
	})
	return b
}

// SetRetryIntervals sets the intervals between transaction retries.
func (b *NodeOptionsBuilder) SetRetryIntervals(intervals []time.Duration) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		if len(intervals) > 0 {
			opts.DB.RetryIntervals = intervals
		}
	})
	return b
}

// SetP2PBlockSyncTimeout sets the timeout duration for syncing block links.
func (b *NodeOptionsBuilder) SetP2PBlockSyncTimeout(timeout time.Duration) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DB.P2PBlockSyncTimeout = timeout
	})
	return b
}

// SetLensRuntime sets the lens runtime type.
func (b *NodeOptionsBuilder) SetLensRuntime(runtime NodeLensRuntimeType) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DB.LensRuntime = runtime
	})
	return b
}

// SetLensPoolSize sets the pool size for the lens runtime.
func (b *NodeOptionsBuilder) SetLensPoolSize(size int) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DB.LensPoolSize = size
	})
	return b
}

// SetChunkSize sets the chunk size for the blockstore.
func (b *NodeOptionsBuilder) SetChunkSize(size int) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DB.ChunkSize = immutable.Some(size)
	})
	return b
}

// --- P2P setters ---

// SetListenAddresses sets the listen addresses.
func (b *NodeOptionsBuilder) SetListenAddresses(addresses ...string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.P2P.ListenAddresses = addresses
	})
	return b
}

// SetBootstrapPeers sets the bootstrap peers.
func (b *NodeOptionsBuilder) SetBootstrapPeers(peers ...string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.P2P.BootstrapPeers = peers
	})
	return b
}

// SetEnablePubSub sets whether PubSub is enabled.
func (b *NodeOptionsBuilder) SetEnablePubSub(enable bool) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.P2P.EnablePubSub = enable
	})
	return b
}

// SetEnableRelay sets whether relay is enabled.
func (b *NodeOptionsBuilder) SetEnableRelay(enable bool) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.P2P.EnableRelay = enable
	})
	return b
}

// SetEnableClearBackoffOnRetry sets whether to clear backoff on retry.
func (b *NodeOptionsBuilder) SetEnableClearBackoffOnRetry(enable bool) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.P2P.EnableClearBackoffOnRetry = enable
	})
	return b
}

// SetP2PPrivateKey sets the private key for the P2P node.
func (b *NodeOptionsBuilder) SetP2PPrivateKey(key []byte) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.P2P.PrivateKey = key
	})
	return b
}

// --- HTTP setters ---

// SetHTTPAddress sets the HTTP server address.
func (b *NodeOptionsBuilder) SetHTTPAddress(address string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.HTTP.Address = address
	})
	return b
}

// SetAllowedOrigins sets the allowed CORS origins.
func (b *NodeOptionsBuilder) SetAllowedOrigins(origins ...string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.HTTP.AllowedOrigins = origins
	})
	return b
}

// SetTLSCertPath sets the path to the TLS certificate file.
func (b *NodeOptionsBuilder) SetTLSCertPath(path string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.HTTP.TLSCertPath = path
	})
	return b
}

// SetTLSKeyPath sets the path to the TLS private key file.
func (b *NodeOptionsBuilder) SetTLSKeyPath(path string) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.HTTP.TLSKeyPath = path
	})
	return b
}

// --- Bulk setters for sub-option structs ---

// SetP2POptions sets the P2P options from a plain data struct.
func (b *NodeOptionsBuilder) SetP2POptions(p2pOpts NodeP2POptions) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.P2P = p2pOpts
	})
	return b
}

// SetStoreOptions sets the store options from a plain data struct.
func (b *NodeOptionsBuilder) SetStoreOptions(storeOpts NodeStoreOptions) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.Store = storeOpts
	})
	return b
}

// SetDocumentACPOptions sets the document ACP options from a plain data struct.
func (b *NodeOptionsBuilder) SetDocumentACPOptions(dacOpts NodeDocumentACPOptions) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DocumentACP = dacOpts
	})
	return b
}

// SetNodeACPOptions sets the node ACP options from a plain data struct.
func (b *NodeOptionsBuilder) SetNodeACPOptions(nacOpts NodeACPOptions) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.NodeACP = nacOpts
	})
	return b
}

// SetDBOptions sets the DB options from a plain data struct.
func (b *NodeOptionsBuilder) SetDBOptions(dbOpts NodeDBOptions) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.DB = dbOpts
	})
	return b
}

// SetHTTPOptions sets the HTTP options from a plain data struct.
func (b *NodeOptionsBuilder) SetHTTPOptions(httpOpts NodeHTTPOptions) *NodeOptionsBuilder {
	b.append(func(opts *NodeOptions) {
		opts.HTTP = httpOpts
	})
	return b
}
