// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build rust_ffi

package rustffi

/*
#include <stdlib.h>
#include "defra.h"
*/
import "C"

import (
	"errors"
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
)

var _ client.TxnStore = (*Wrapper)(nil)
var _ client.P2P = (*Wrapper)(nil)

// RemoteDACConfig points the node's document access control at Vera.
type RemoteDACConfig struct {
	GRPCAddress     string
	CometRPCAddress string
	ChainID         string
	// SignerKey is the raw secp256k1 key that pays for Vera transactions.
	SignerKey []byte
}

// defaultMaxTxnRetries is Go's db default. The harness runs each action
// MaxTxnRetries times at most, so zero would run none.
const defaultMaxTxnRetries = 5

// Config is how a Rust node starts: what C.NodeInitOptions carries.
type Config struct {
	// DBPath is the store directory; empty keeps the store in memory.
	DBPath string
	// ListenAddresses are the P2P listen multiaddrs; none runs without P2P.
	ListenAddresses []string
	// ReplicatorRetryIntervals are comma-separated whole seconds.
	ReplicatorRetryIntervals string
	EnableSigning            bool
	// NodeIdentity is the node's identity, and its NAC owner.
	NodeIdentity identity.Identity
	// EnableNAC turns node access control on at startup.
	EnableNAC bool
	// MaxTxnRetries is how many times a conflicting transaction is tried;
	// zero takes Go's default.
	MaxTxnRetries           int
	SearchableEncryptionKey []byte
	StoreEncryptionKey      []byte
	RemoteDAC               *RemoteDACConfig
}

// Wrapper is a Rust node, as a test client.
type Wrapper struct {
	node          C.uintptr_t
	nodeIdentity  identity.Identity
	maxTxnRetries int
	events        *eventBridge
	goNodeCloser  func()
}

// NewWrapper starts a Rust node configured by cfg.
func NewWrapper(cfg Config) (*Wrapper, error) {
	var args cstrings
	defer args.free()

	if cfg.MaxTxnRetries == 0 {
		cfg.MaxTxnRetries = defaultMaxTxnRetries
	}

	var opts C.struct_NodeInitOptions
	opts.max_transaction_retries = C.int(cfg.MaxTxnRetries)
	if cfg.DBPath == "" {
		opts.in_memory = 1
	} else {
		opts.db_path = args.str(cfg.DBPath)
	}
	if len(cfg.ListenAddresses) == 0 {
		opts.disable_p2p = 1
	} else {
		opts.listening_addresses = args.str(strings.Join(cfg.ListenAddresses, ","))
		opts.enable_pub_sub = 1
	}
	if cfg.ReplicatorRetryIntervals != "" {
		opts.replicator_retry_intervals = args.str(cfg.ReplicatorRetryIntervals)
	}
	if !cfg.EnableSigning {
		opts.disable_signing = 1
	}
	if cfg.EnableNAC {
		opts.enable_node_acp = 1
	}
	if key := cfg.SearchableEncryptionKey; len(key) > 0 {
		opts.searchable_encryption_key = args.bytes(key)
		opts.searchable_encryption_key_len = C.int(len(key))
	}
	if key := cfg.StoreEncryptionKey; len(key) > 0 {
		opts.store_encryption_key = args.bytes(key)
		opts.store_encryption_key_len = C.int(len(key))
	}
	if dac := cfg.RemoteDAC; dac != nil {
		opts.document_acp_type = args.str("remote")
		opts.remote_dac_grpc_address = args.str(dac.GRPCAddress)
		opts.remote_dac_comet_rpc_address = args.str(dac.CometRPCAddress)
		opts.vera_chain_id = args.str(dac.ChainID)
		opts.vera_signer_key = args.bytes(dac.SignerKey)
		opts.vera_signer_key_len = C.int(len(dac.SignerKey))
	}

	// The node keeps the identity's key for itself; the handle only names it
	// for this call.
	nodeIdentity, err := identityHandle(immutable.Some(cfg.NodeIdentity))
	if err != nil {
		return nil, err
	}
	defer freeIdentity(nodeIdentity)
	opts.identity_ptr = nodeIdentity

	cResult := C.defra_node_new(opts)
	node := cResult.node_ptr
	message := C.GoString(cResult.error)
	failed := cResult.status != 0
	C.defra_new_node_result_free(&cResult)
	if failed {
		return nil, errors.New(message)
	}

	w := &Wrapper{
		node:          node,
		nodeIdentity:  cfg.NodeIdentity,
		maxTxnRetries: cfg.MaxTxnRetries,
	}
	w.events, err = newEventBridge(w)
	if err != nil {
		w.closeNode()
		return nil, err
	}
	return w, nil
}

// SetGoNodeCloser sets what closes the Go node started next to this one, so
// that a restart releases the Go store's lock as well.
func (w *Wrapper) SetGoNodeCloser(closer func()) {
	w.goNodeCloser = closer
}

// GrantNACAdmin makes ident an admin of the node, as the node identity that
// owns its access control.
func (w *Wrapper) GrantNACAdmin(ident identity.Identity) error {
	var args cstrings
	defer args.free()
	owner, err := identityHandle(immutable.Some(w.nodeIdentity))
	if err != nil {
		return err
	}
	defer freeIdentity(owner)
	return take(C.defra_acp_nac_relationship_add(
		w.node, owner, args.str("admin"), args.str(ident.DID()))).err()
}

// DisableNACAsNode turns node access control off for now, as the node
// identity that owns it.
func (w *Wrapper) DisableNACAsNode() error {
	owner, err := identityHandle(immutable.Some(w.nodeIdentity))
	if err != nil {
		return err
	}
	defer freeIdentity(owner)
	return take(C.defra_acp_nac_disable(w.node, owner)).err()
}

func (w *Wrapper) Close() {
	if w.events != nil {
		w.events.close()
		w.events = nil
	}
	w.closeNode()
	if w.goNodeCloser != nil {
		w.goNodeCloser()
		w.goNodeCloser = nil
	}
}

func (w *Wrapper) closeNode() {
	if w.node != 0 {
		_ = take(C.defra_node_close(w.node))
		w.node = 0
	}
}

func (w *Wrapper) Events() event.Bus {
	return w.events.bus
}

func (w *Wrapper) MaxTxnRetries() int {
	return w.maxTxnRetries
}
