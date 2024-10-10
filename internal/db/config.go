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
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
)

const (
	defaultMaxTxnRetries  = 5
	updateEventBufferSize = 100
)

type dbOptions struct {
	maxTxnRetries  immutable.Option[int]
	RetryIntervals []time.Duration
	identity       immutable.Option[identity.Identity]
}

// defaultOptions returns the default db options.
func defaultOptions() *dbOptions {
	return &dbOptions{
		RetryIntervals: []time.Duration{
			// exponential backoff retry intervals
			time.Second * 30,
			time.Minute,
			time.Minute * 2,
			time.Minute * 4,
			time.Minute * 8,
			time.Minute * 16,
			time.Minute * 32,
		},
	}
}

// Option is a funtion that sets a config value on the db.
type Option func(*dbOptions)

// WithMaxRetries sets the maximum number of retries per transaction.
func WithMaxRetries(num int) Option {
	return func(opts *dbOptions) {
		opts.maxTxnRetries = immutable.Some(num)
	}
}

func WithRetryInterval(interval []time.Duration) Option {
	return func(opt *dbOptions) {
		if len(interval) > 0 {
			opt.RetryIntervals = interval
		}
	}
}

func WithNodeIdentity(ident identity.Identity) Option {
	return func(opts *dbOptions) {
		opts.identity = immutable.Some(ident)
	}
}
