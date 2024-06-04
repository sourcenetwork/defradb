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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/events"
)

const (
	defaultMaxTxnRetries  = 5
	updateEventBufferSize = 100
)

// Option is a funtion that sets a config value on the db.
type Option func(*db)

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
