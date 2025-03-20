// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
)

type ACPType string

const (
	// NoACPType disables the ACP subsystem.
	NoACPType ACPType = "none"
	// DefaultACPType uses the default ACP implementation for this build.
	DefaultACPType ACPType = ""
)

// acpConstructors is a map of [ACPType]s to acp implementations.
//
// It is populated by the `init` functions in the implementation-specific files - this
// allows it's population to be managed by build flags.
var acpConstructors = map[ACPType]func(context.Context, *ACPOptions) (immutable.Option[acp.ACP], error){
	NoACPType: func(ctx context.Context, a *ACPOptions) (immutable.Option[acp.ACP], error) {
		return acp.NoACP, nil
	},
}

// ACPOptions contains ACP configuration values.
type ACPOptions struct {
	acpType ACPType

	// Note: An empty path will result in an in-memory ACP instance.
	//
	// This is only used for local acp.
	path string

	signer                   immutable.Option[TxSigner]
	sourceHubChainID         string
	sourceHubGRPCAddress     string
	sourceHubCometRPCAddress string
}

// TxSigner models an entity capable of providing signatures for a Tx.
//
// Effectively, it can be either a secp256k1 cosmos-sdk key or a pointer to a
// secp256k1 key in a cosmos-sdk like keyring.
type TxSigner interface {
	GetAccAddress() string
	GetPrivateKey() cryptotypes.PrivKey
}

// DefaultACPOptions returns new options with default values.
func DefaultACPOptions() *ACPOptions {
	return &ACPOptions{
		acpType: LocalACPType,
	}
}

// StoreOpt is a function for setting configuration values.
type ACPOpt func(*ACPOptions)

// WithACPType sets the ACP type.
func WithACPType(acpType ACPType) ACPOpt {
	return func(o *ACPOptions) {
		o.acpType = acpType
	}
}

// WithACPPath sets the ACP path.
//
// Note: An empty path will result in an in-memory ACP instance.
func WithACPPath(path string) ACPOpt {
	return func(o *ACPOptions) {
		o.path = path
	}
}

// WithKeyring sets the txn signer for Defra to use.
//
// It is only required when SourceHub ACP is active.
func WithTxnSigner(signer immutable.Option[TxSigner]) ACPOpt {
	return func(o *ACPOptions) {
		o.signer = signer
	}
}

// WithSourceHubChainID specifies the chainID of the SourceHub (cosmos) chain
// to use for SourceHub ACP.
func WithSourceHubChainID(sourceHubChainID string) ACPOpt {
	return func(o *ACPOptions) {
		o.sourceHubChainID = sourceHubChainID
	}
}

// WithSourceHubGRPCAddress specifies the GRPC address of the SourceHub node to use
// for ACP calls.
func WithSourceHubGRPCAddress(address string) ACPOpt {
	return func(o *ACPOptions) {
		o.sourceHubGRPCAddress = address
	}
}

// WithSourceHubCometRPCAddress specifies the Comet RPC address of the SourceHub node to use
// for ACP calls.
func WithSourceHubCometRPCAddress(address string) ACPOpt {
	return func(o *ACPOptions) {
		o.sourceHubCometRPCAddress = address
	}
}

// NewACP returns a new ACP module with the given options.
func NewACP(ctx context.Context, opts ...ACPOpt) (immutable.Option[acp.ACP], error) {
	options := DefaultACPOptions()
	for _, opt := range opts {
		opt(options)
	}
	acpConstructor, ok := acpConstructors[options.acpType]
	if ok {
		return acpConstructor(ctx, options)
	}
	return immutable.None[acp.ACP](), NewErrACPTypeNotSupported(options.acpType)
}
