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

// NodeP2P returns P2P options with default values.
func NodeP2P() *NodeP2POptions {
	return &NodeP2POptions{
		EnablePubSub: false,
		EnableRelay:  false,
	}
}

// SetListenAddresses sets the listen addresses.
func (o *NodeP2POptions) SetListenAddresses(addresses ...string) *NodeP2POptions {
	o.ListenAddresses = addresses
	return o
}

// SetBootstrapPeers sets the bootstrap peers.
func (o *NodeP2POptions) SetBootstrapPeers(peers ...string) *NodeP2POptions {
	o.BootstrapPeers = peers
	return o
}

// SetEnablePubSub sets whether PubSub is enabled.
func (o *NodeP2POptions) SetEnablePubSub(enable bool) *NodeP2POptions {
	o.EnablePubSub = enable
	return o
}

// SetEnableRelay sets whether relay is enabled.
func (o *NodeP2POptions) SetEnableRelay(enable bool) *NodeP2POptions {
	o.EnableRelay = enable
	return o
}

// SetPrivateKey sets the private key for the P2P node.
func (o *NodeP2POptions) SetPrivateKey(key []byte) *NodeP2POptions {
	o.PrivateKey = key
	return o
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

// NodeHTTP returns HTTP options with default values.
func NodeHTTP() *NodeHTTPOptions {
	return &NodeHTTPOptions{}
}

// SetAddress sets the HTTP server address.
func (o *NodeHTTPOptions) SetAddress(address string) *NodeHTTPOptions {
	o.Address = address
	return o
}

// SetAllowedOrigins sets the allowed CORS origins.
func (o *NodeHTTPOptions) SetAllowedOrigins(origins ...string) *NodeHTTPOptions {
	o.AllowedOrigins = origins
	return o
}

// SetTLSCertPath sets the path to the TLS certificate file.
func (o *NodeHTTPOptions) SetTLSCertPath(path string) *NodeHTTPOptions {
	o.TLSCertPath = path
	return o
}

// SetTLSKeyPath sets the path to the TLS private key file.
func (o *NodeHTTPOptions) SetTLSKeyPath(path string) *NodeHTTPOptions {
	o.TLSKeyPath = path
	return o
}

// Node returns a new NodeOptions instance with default values.
// By default P2P and API are enabled.
func Node() *NodeOptions {
	return &NodeOptions{
		DisableP2P:        false,
		DisableAPI:        false,
		EnableDevelopment: false,
		Store:             *NodeStore(),
		DocumentACP:       *NodeDocumentACP(),
		NodeACP:           *NodeACP(),
		DB:                *NodeDB(),
		P2P:               *NodeP2P(),
		HTTP:              *NodeHTTP(),
	}
}

// SetDisableP2P sets the disable P2P flag.
func (o *NodeOptions) SetDisableP2P(disable bool) *NodeOptions {
	o.DisableP2P = disable
	return o
}

// SetDisableAPI sets the disable API flag.
func (o *NodeOptions) SetDisableAPI(disable bool) *NodeOptions {
	o.DisableAPI = disable
	return o
}

// SetEnableDevelopment sets the enable development mode flag.
func (o *NodeOptions) SetEnableDevelopment(enable bool) *NodeOptions {
	o.EnableDevelopment = enable
	return o
}

// SetKMS sets the KMS type.
func (o *NodeOptions) SetKMS(kmsType NodeKMSType) *NodeOptions {
	o.KMSType = immutable.Some(kmsType)
	return o
}

// SetNodeIdentity sets the identity for the node.
// This is a convenience method that sets the identity on the DB options.
func (o *NodeOptions) SetNodeIdentity(ident identity.Identity) *NodeOptions {
	o.DB.Identity = immutable.Some(ident)
	return o
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

// NodeStore returns store options with default values.
func NodeStore() *NodeStoreOptions {
	return &NodeStoreOptions{
		Store:          NodeDefaultStore,
		Path:           "",
		BadgerInMemory: false,
		BadgerFileSize: 1 << 30, // 1GB
	}
}

// SetStore sets the store type.
func (o *NodeStoreOptions) SetStore(store NodeStoreType) *NodeStoreOptions {
	o.Store = store
	return o
}

// SetPath sets the store path.
func (o *NodeStoreOptions) SetPath(path string) *NodeStoreOptions {
	o.Path = path
	return o
}

// SetBadgerFileSize sets the Badger file size.
func (o *NodeStoreOptions) SetBadgerFileSize(size int64) *NodeStoreOptions {
	o.BadgerFileSize = size
	return o
}

// SetBadgerEncryptionKey sets the Badger encryption key.
func (o *NodeStoreOptions) SetBadgerEncryptionKey(key []byte) *NodeStoreOptions {
	o.BadgerEncryptionKey = key
	return o
}

// SetBadgerInMemory sets whether Badger should run in-memory.
func (o *NodeStoreOptions) SetBadgerInMemory(inMemory bool) *NodeStoreOptions {
	o.BadgerInMemory = inMemory
	return o
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

// NodeDocumentACP returns document ACP options with default values.
func NodeDocumentACP() *NodeDocumentACPOptions {
	return &NodeDocumentACPOptions{
		DocumentACPType: NodeLocalDocumentACPType,
	}
}

// SetDocumentACPType sets the ACP type.
func (o *NodeDocumentACPOptions) SetDocumentACPType(acpType NodeDocumentACPType) *NodeDocumentACPOptions {
	o.DocumentACPType = acpType
	return o
}

// SetPath sets the document ACP system path.
func (o *NodeDocumentACPOptions) SetPath(path string) *NodeDocumentACPOptions {
	o.Path = path
	return o
}

// SetTxnSigner sets the txn signer for Defra to use.
func (o *NodeDocumentACPOptions) SetTxnSigner(signer NodeTxSigner) *NodeDocumentACPOptions {
	o.Signer = immutable.Some(signer)
	return o
}

// SetSourceHubChainID sets the chainID of the SourceHub chain.
func (o *NodeDocumentACPOptions) SetSourceHubChainID(chainID string) *NodeDocumentACPOptions {
	o.SourceHubChainID = chainID
	return o
}

// SetSourceHubGRPCAddress sets the GRPC address of the SourceHub node.
func (o *NodeDocumentACPOptions) SetSourceHubGRPCAddress(address string) *NodeDocumentACPOptions {
	o.SourceHubGRPCAddress = address
	return o
}

// SetSourceHubCometRPCAddress sets the Comet RPC address of the SourceHub node.
func (o *NodeDocumentACPOptions) SetSourceHubCometRPCAddress(address string) *NodeDocumentACPOptions {
	o.SourceHubCometRPCAddress = address
	return o
}

// NodeACPOptions contains node ACP configuration values.
type NodeACPOptions struct {
	// IsEnabled specifies whether node ACP is enabled.
	IsEnabled bool
	// Path is the filesystem path for the node ACP system.
	Path string
}

// NodeACP returns node ACP options with default values.
func NodeACP() *NodeACPOptions {
	return &NodeACPOptions{
		IsEnabled: false,
	}
}

// SetEnabled sets whether node ACP is enabled.
func (o *NodeACPOptions) SetEnabled(enabled bool) *NodeACPOptions {
	o.IsEnabled = enabled
	return o
}

// SetPath sets the node ACP system path.
func (o *NodeACPOptions) SetPath(path string) *NodeACPOptions {
	o.Path = path
	return o
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

// NodeDB returns database options with default values.
func NodeDB() *NodeDBOptions {
	return &NodeDBOptions{
		MaxTxnRetries: immutable.Some(5),
		EnableSigning: true,
		RetryIntervals: []time.Duration{
			// exponential backoff retry intervals
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
		LensPoolSize:        0, // 0 means use default
	}
}

// SetMaxTxnRetries sets the maximum number of retries per transaction.
func (o *NodeDBOptions) SetMaxTxnRetries(num int) *NodeDBOptions {
	o.MaxTxnRetries = immutable.Some(num)
	return o
}

// SetIdentity sets the identity for the node.
func (o *NodeDBOptions) SetIdentity(ident identity.Identity) *NodeDBOptions {
	o.Identity = immutable.Some(ident)
	return o
}

// SetEnableSigning sets whether block signing is enabled.
func (o *NodeDBOptions) SetEnableSigning(enable bool) *NodeDBOptions {
	o.EnableSigning = enable
	return o
}

// SetSearchableEncryptionKey sets the key used for searchable encryption.
func (o *NodeDBOptions) SetSearchableEncryptionKey(key []byte) *NodeDBOptions {
	o.SearchableEncryptionKey = key
	return o
}

// SetRetryIntervals sets the intervals between transaction retries.
func (o *NodeDBOptions) SetRetryIntervals(intervals []time.Duration) *NodeDBOptions {
	if len(intervals) > 0 {
		o.RetryIntervals = intervals
	}
	return o
}

// SetP2PBlockSyncTimeout sets the timeout duration for syncing block links.
func (o *NodeDBOptions) SetP2PBlockSyncTimeout(timeout time.Duration) *NodeDBOptions {
	o.P2PBlockSyncTimeout = timeout
	return o
}

// SetLensRuntime sets the lens runtime type.
func (o *NodeDBOptions) SetLensRuntime(runtime NodeLensRuntimeType) *NodeDBOptions {
	o.LensRuntime = runtime
	return o
}

// SetLensPoolSize sets the pool size for the lens runtime.
func (o *NodeDBOptions) SetLensPoolSize(size int) *NodeDBOptions {
	o.LensPoolSize = size
	return o
}

// SetChunkSize sets the chunk size for the blockstore.
func (o *NodeDBOptions) SetChunkSize(size int) *NodeDBOptions {
	o.ChunkSize = immutable.Some(size)
	return o
}
