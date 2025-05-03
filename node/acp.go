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

	"github.com/sourcenetwork/defradb/acp/dac"
)

type DocumentACPType string

const (
	// NoDocumentACPType disables the document ACP subsystem.
	NoDocumentACPType DocumentACPType = "none"
	// DefaultDocumentACPType uses the default ACP implementation for this build.
	DefaultDocumentACPType DocumentACPType = ""
)

// documentACPConstructors is a map of [DocumentACPType]s to acp implementations.
//
// It is populated by the `init` functions in the implementation-specific files - this
// allows it's population to be managed by build flags.
var documentACPConstructors = map[DocumentACPType]func(
	context.Context,
	*DocumentACPOptions,
) (immutable.Option[dac.DocumentACP], error){
	NoDocumentACPType: func(
		ctx context.Context,
		a *DocumentACPOptions,
	) (immutable.Option[dac.DocumentACP], error) {
		return dac.NoDocumentACP, nil
	},
}

// DocumentACPOptions contains ACP configuration values.
type DocumentACPOptions struct {
	documentACPType DocumentACPType

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
func DefaultACPOptions() *DocumentACPOptions {
	return &DocumentACPOptions{
		documentACPType: LocalDocumentACPType,
	}
}

// DocumentACPOpt is a function for setting document ACP configuration values.
type DocumentACPOpt func(*DocumentACPOptions)

// WithDocumentACPType sets the ACP type.
func WithDocumentACPType(acpType DocumentACPType) DocumentACPOpt {
	return func(o *DocumentACPOptions) {
		o.documentACPType = acpType
	}
}

// WithDocumentACPPath sets the document ACP system path.
//
// Note: An empty path will result in an in-memory document ACP instance.
func WithDocumentACPPath(path string) DocumentACPOpt {
	return func(o *DocumentACPOptions) {
		o.path = path
	}
}

// WithKeyring sets the txn signer for Defra to use.
//
// It is only required when SourceHub ACP is active.
func WithTxnSigner(signer immutable.Option[TxSigner]) DocumentACPOpt {
	return func(o *DocumentACPOptions) {
		o.signer = signer
	}
}

// WithSourceHubChainID specifies the chainID of the SourceHub (cosmos) chain
// to use for SourceHub ACP.
func WithSourceHubChainID(sourceHubChainID string) DocumentACPOpt {
	return func(o *DocumentACPOptions) {
		o.sourceHubChainID = sourceHubChainID
	}
}

// WithSourceHubGRPCAddress specifies the GRPC address of the SourceHub node to use
// for ACP calls.
func WithSourceHubGRPCAddress(address string) DocumentACPOpt {
	return func(o *DocumentACPOptions) {
		o.sourceHubGRPCAddress = address
	}
}

// WithSourceHubCometRPCAddress specifies the Comet RPC address of the SourceHub node to use
// for ACP calls.
func WithSourceHubCometRPCAddress(address string) DocumentACPOpt {
	return func(o *DocumentACPOptions) {
		o.sourceHubCometRPCAddress = address
	}
}

// NewDocumentACP returns a new ACP module with the given options.
func NewDocumentACP(ctx context.Context, opts ...DocumentACPOpt) (immutable.Option[dac.DocumentACP], error) {
	options := DefaultACPOptions()
	for _, opt := range opts {
		opt(options)
	}
	acpConstructor, ok := documentACPConstructors[options.documentACPType]
	if ok {
		return acpConstructor(ctx, options)
	}
	return immutable.None[dac.DocumentACP](), NewErrACPTypeNotSupported(options.documentACPType)
}
