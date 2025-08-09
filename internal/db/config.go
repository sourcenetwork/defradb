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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
)

const (
	defaultMaxTxnRetries  = 5
	updateEventBufferSize = 100
)

type dbOptions struct {
	maxTxnRetries  immutable.Option[int]
	identity       immutable.Option[identity.Identity]
	disableSigning bool
	searchableEncryptionKey []byte
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

// WithSearchableEncryptionKey sets the key used for searchable encryption.
// This key is used to generate search tags for encrypted fields.
func WithSearchableEncryptionKey(key []byte) Option {
	return func(opts *dbOptions) {
		opts.searchableEncryptionKey = key
	}
}
