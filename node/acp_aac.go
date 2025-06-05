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

// AdminACPOpt is a function for setting admin ACP configuration values.
type AdminACPOpt func(*AdminACPOptions)

// AdminACPOptions contains ACP configuration values.
type AdminACPOptions struct {
	// isEnabled is true if admin acp is enabled, and false otherwise.
	isEnabled bool

	// Note: An empty path will result in an in-memory ACP instance.
	path string
}

// DefaultACPOptions returns the default admin acp options.
func DefaultAdminACPOptions() *AdminACPOptions {
	return &AdminACPOptions{
		isEnabled: false,
	}
}

// WithAdminACPPath sets the admin ACP system path.
//
// Note: An empty path will result in an in-memory admin ACP instance.
func WithAdminACPPath(path string) AdminACPOpt {
	return func(o *AdminACPOptions) {
		o.path = path
	}
}

// WithEnableAdminACP enables admin acp.
func WithEnableAdminACP(enable bool) AdminACPOpt {
	return func(o *AdminACPOptions) {
		o.isEnabled = enable
	}
}

// adminACPConstructors is a map of [bool] to indicate admin acp implementations.
var adminACPConstructors = map[bool]func(
	context.Context,
	*AdminACPOptions,
) (db.AdminInfo, error){
	// We keep AAC started (in both cases) to have access control ability even when admin acp
	// is disabled temporarily, as we want to only allow authorized user to re-enable admin acp.
	// Note: To free resources the caller must still call [adminInfo.AdminACP.Close()] when done.
	false: func(ctx context.Context, options *AdminACPOptions) (db.AdminInfo, error) {
		return db.NewAdminInfoWithAACDisabled(ctx, options.path)
	},
	true: func(ctx context.Context, options *AdminACPOptions) (db.AdminInfo, error) {
		return db.NewAdminInfoWithAACEnabled(ctx, options.path)
	},
}

func NewAdminACP(ctx context.Context, opts ...AdminACPOpt) (db.AdminInfo, error) {
	options := DefaultAdminACPOptions()
	for _, opt := range opts {
		opt(options)
	}
	acpConstructor, ok := adminACPConstructors[options.isEnabled]
	if ok {
		return acpConstructor(ctx, options)
	}
	return db.AdminInfo{}, ErrAdminACPTypeNotSupported
}
