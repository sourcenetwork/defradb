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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/keyring"
)

type ACPType uint8

const (
	NoACPType        ACPType = 0
	LocalACPType     ACPType = 1
	SourceHubACPType ACPType = 2
)

// ACPOptions contains ACP configuration values.
type ACPOptions struct {
	acpType ACPType

	// Note: An empty path will result in an in-memory ACP instance.
	path string

	keyring                  immutable.Option[keyring.Keyring]
	sourceHubKeyName         string
	sourceHubChainID         string
	sourceHubGRPCAddress     string
	sourceHubCometRPCAddress string
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

// WithKeyring sets the keyring for Defra to use.
//
// It is only required when SourceHub ACP is active.
func WithKeyring(keyring immutable.Option[keyring.Keyring]) ACPOpt {
	return func(o *ACPOptions) {
		o.keyring = keyring
	}
}

// WithSourceHubKeyName specifies the name of the key in the keyring to use to sign
// and (pay for) SourceHub transactions.
func WithSourceHubKeyName(sourceHubKeyName string) ACPOpt {
	return func(o *ACPOptions) {
		o.sourceHubKeyName = sourceHubKeyName
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

	switch options.acpType {
	case NoACPType:
		return acp.NoACP, nil

	case LocalACPType:
		acpLocal := acp.NewLocalACP()
		acpLocal.Init(ctx, options.path)
		return immutable.Some[acp.ACP](acpLocal), nil

	case SourceHubACPType:
		if !options.keyring.HasValue() {
			return acp.NoACP, ErrKeyringMissingForSourceHubACP
		}

		acpSourceHub, err := acp.NewSourceHubACP(
			options.sourceHubChainID,
			options.sourceHubGRPCAddress,
			options.sourceHubCometRPCAddress,
			options.keyring.Value(),
			options.sourceHubKeyName,
		)
		if err != nil {
			return acp.NoACP, err
		}

		return immutable.Some(acpSourceHub), nil

	default:
		acpLocal := acp.NewLocalACP()
		acpLocal.Init(ctx, options.path)
		return immutable.Some[acp.ACP](acpLocal), nil
	}
}
