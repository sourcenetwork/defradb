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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/lens"
)

const (
	defaultMaxTxnRetries  = 5
	updateEventBufferSize = 100
)

// Option is a funtion that sets a config value on the db.
type Option func(*db)

// WithACP enables access control. If path is empty then acp runs in-memory.
func WithACP(path string) Option {
	return func(db *db) {
		var acpLocal acp.ACPLocal
		acpLocal.Init(context.Background(), path)
		db.acp = immutable.Some[acp.ACP](&acpLocal)
	}
}

// WithACPInMemory enables access control in-memory.
func WithACPInMemory() Option { return WithACP("") }

// WithUpdateEvents enables the update events channel.
func WithUpdateEvents() Option {
	return func(db *db) {
		db.events = events.Events{
			Updates: immutable.Some(events.New[events.Update](0, updateEventBufferSize)),
		}
	}
}

// WithMaxRetries sets the maximum number of retries per transaction.
func WithMaxRetries(num int) Option {
	return func(db *db) {
		db.maxTxnRetries = immutable.Some(num)
	}
}

// WithLensOptions sets the lens registry options for the db.
func WithLensOptions(opts ...lens.Option) Option {
	return func(db *db) {
		db.lensOptions = opts
	}
}
