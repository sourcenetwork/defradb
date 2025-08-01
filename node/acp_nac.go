// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/internal/db"
)

// NodeACPOpt is a function for setting node ACP configuration values.
type NodeACPOpt func(*NodeACPOptions)

// NodeACPOptions contains ACP configuration values.
type NodeACPOptions struct {
	// isEnabled is true if node acp is enabled, and false otherwise.
	isEnabled bool

	// Note: An empty path will result in an in-memory ACP instance.
	path string
}

// DefaultNodeACPOptions returns the default node acp options.
func DefaultNodeACPOptions() *NodeACPOptions {
	return &NodeACPOptions{
		isEnabled: false,
	}
}

// WithNodeACPPath sets the node ACP system path.
//
// Note: An empty path will result in an in-memory node ACP instance.
func WithNodeACPPath(path string) NodeACPOpt {
	return func(o *NodeACPOptions) {
		o.path = path
	}
}

// WithEnableNodeACP enables node acp.
func WithEnableNodeACP(enable bool) NodeACPOpt {
	return func(o *NodeACPOptions) {
		o.isEnabled = enable
	}
}

func NewNodeACP(ctx context.Context, opts ...NodeACPOpt) (db.NACInfo, error) {
	options := DefaultNodeACPOptions()
	for _, opt := range opts {
		opt(options)
	}

	return db.NewNACInfo(ctx, options.path, options.isEnabled)
}
