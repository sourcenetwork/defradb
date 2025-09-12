// Copyright 2025 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/client"
)

const (
	defaultMaxTxnRetries  = 5
	updateEventBufferSize = 100
)

type dbOptions struct {
	maxTxnRetries  immutable.Option[int]
	identity       immutable.Option[identity.Identity]
	disableSigning bool
	p2p            immutable.Option[client.Host]
	retryIntervals []time.Duration
	// timeout duration for syncing block links.
	p2pBlockSyncTimeout time.Duration
}

func defaultDBOptions() *dbOptions {
	return &dbOptions{
		maxTxnRetries: immutable.Some(defaultMaxTxnRetries),
		retryIntervals: []time.Duration{
			// exponential backoff retry intervals
			time.Second * 30,
			time.Minute,
			time.Minute * 2,
			time.Minute * 4,
			time.Minute * 8,
			time.Minute * 16,
			time.Minute * 32,
		},
		p2pBlockSyncTimeout: time.Second * 5,
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

func WithNodeIdentity(ident identity.Identity) Option {
	return func(opts *dbOptions) {
		opts.identity = immutable.Some(ident)
	}
}

// WithEnabledSigning sets the signing algorithm to use for DAG blocks.
// If false, block signing is disabled. By default, block signing is enabled.
func WithEnabledSigning(value bool) Option {
	return func(opts *dbOptions) {
		opts.disableSigning = !value
	}
}

func WithRetryInterval(interval []time.Duration) Option {
	return func(opt *dbOptions) {
		if len(interval) > 0 {
			opt.retryIntervals = interval
		}
	}
}

func WithP2P(host client.Host) Option {
	return func(opts *dbOptions) {
		opts.p2p = immutable.Some(host)
	}
}

func WithP2PBlockSyncTimeout(timeout time.Duration) Option {
	return func(opt *dbOptions) {
		opt.p2pBlockSyncTimeout = timeout
	}
}
