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
)

type ACPType uint8

const (
	NoACPType    ACPType = 0
	LocalACPType ACPType = 1
)

// ACPOptions contains ACP configuration values.
type ACPOptions struct {
	acpType ACPType

	// Note: An empty path will result in an in-memory ACP instance.
	path string
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

	default:
		acpLocal := acp.NewLocalACP()
		acpLocal.Init(ctx, options.path)
		return immutable.Some[acp.ACP](acpLocal), nil
	}
}
