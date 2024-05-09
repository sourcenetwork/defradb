// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"

	"github.com/lens-vm/lens/host-go/engine/module"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/internal/acp"
	"github.com/sourcenetwork/defradb/internal/db"
)

const (
	updateEventBufferSize = 100
)

// DBOpt is a funtion that sets a config value on the db.
type DBOpt func(*db.Options)

// WithACP enables access control. If path is empty then acp runs in-memory.
func WithACP(path string) DBOpt {
	return func(opt *db.Options) {
		var acpLocal acp.ACPLocal
		acpLocal.Init(context.Background(), path)
		opt.ACP = immutable.Some[acp.ACP](&acpLocal)
	}
}

// WithACPInMemory enables access control in-memory.
func WithACPInMemory() DBOpt { return WithACP("") }

// WithUpdateEvents enables the update events channel.
func WithUpdateEvents() DBOpt {
	return func(opt *db.Options) {
		opt.Events = events.Events{
			Updates: immutable.Some(events.New[events.Update](0, updateEventBufferSize)),
		}
	}
}

// WithMaxRetries sets the maximum number of retries per transaction.
func WithMaxRetries(num int) DBOpt {
	return func(opt *db.Options) {
		opt.MaxTxnRetries = immutable.Some(num)
	}
}

// WithLensPoolSize sets the maximum number of cached migrations instances to preserve per schema version.
//
// Will default to `5` if not set.
func WithLensPoolSize(size int) DBOpt {
	return func(opt *db.Options) {
		opt.LensPoolSize = immutable.Some(size)
	}
}

// WithLensRuntime returns an option that sets the lens registry runtime.
func WithLensRuntime(runtime module.Runtime) DBOpt {
	return func(opt *db.Options) {
		opt.LensRuntime = immutable.Some(runtime)
	}
}

// NewDB returns a new store with the given options.
func NewDB(ctx context.Context, rootstore datastore.RootStore, opts ...DBOpt) (client.DB, error) {
	options := &db.Options{}
	for _, opt := range opts {
		opt(options)
	}

	return db.NewDB(ctx, rootstore, options)
}
