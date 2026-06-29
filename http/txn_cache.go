// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"context"
	"errors"
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/ttl"
)

const (
	// DefaultTxnTTL is the default idle TTL for explicit HTTP transactions.
	DefaultTxnTTL = 60 * time.Second
	// DefaultTxnTTLTick is the default timer resolution for HTTP transaction TTLs.
	DefaultTxnTTLTick = time.Second
	// DefaultTxnTTLBuckets is the default timer bucket count for HTTP transaction TTLs.
	DefaultTxnTTLBuckets = 60
)

// cachedTxn pairs a cached transaction with its configured idle TTL, so the TTL
// can be reused if the transaction is re-stored (for example, restored after a
// failed commit).
type cachedTxn struct {
	txn client.Txn
	ttl time.Duration
}

// txnLease is a hold on a cached transaction that keeps it from expiring while
// a request is using it. It is released once the request completes.
type txnLease = ttl.Lease[uint64, cachedTxn]

// txnCache holds the explicit HTTP transactions keyed by their ID, expiring
// each after it has been idle for its TTL. A transaction does not expire while
// a request holds a lease on it, so long-running requests cannot have their
// transaction discarded mid-flight.
type txnCache struct {
	defaultTTL time.Duration
	cache      *ttl.Cache[uint64, cachedTxn]
}

func newTxnCache(ctx context.Context, opts *options.NodeHTTPOptions) (*txnCache, error) {
	cfg := txnCacheConfig(opts)
	cache, err := ttl.NewCache(ctx, cfg.tick, cfg.buckets, func(_ uint64, cached cachedTxn) {
		cached.txn.Discard()
	})
	if err != nil {
		return nil, err
	}
	if err := cache.ValidateTTL(cfg.defaultTTL); err != nil {
		cache.Close()
		return nil, errors.Join(ErrInvalidTTL, err)
	}
	return &txnCache{
		defaultTTL: cfg.defaultTTL,
		cache:      cache,
	}, nil
}

type txnCacheConfiguration struct {
	defaultTTL time.Duration
	tick       time.Duration
	buckets    int
}

func txnCacheConfig(opts *options.NodeHTTPOptions) txnCacheConfiguration {
	cfg := txnCacheConfiguration{
		defaultTTL: DefaultTxnTTL,
		tick:       DefaultTxnTTLTick,
		buckets:    DefaultTxnTTLBuckets,
	}
	if opts == nil {
		return cfg
	}
	if opts.TxnTTL != 0 {
		cfg.defaultTTL = opts.TxnTTL
	}
	if opts.TxnTTLTick != 0 {
		cfg.tick = opts.TxnTTLTick
	}
	if opts.TxnTTLBuckets != 0 {
		cfg.buckets = opts.TxnTTLBuckets
	}
	return cfg
}

// Store caches txn until it has been idle for the cache's default TTL.
func (c *txnCache) Store(txn client.Txn) error {
	return c.StoreFor(txn, c.defaultTTL)
}

// StoreFor caches txn until it has been idle for txnTTL.
func (c *txnCache) StoreFor(txn client.Txn, txnTTL time.Duration) error {
	return c.cache.Store(txn.ID(), cachedTxn{txn: txn, ttl: txnTTL}, txnTTL)
}

// Get returns the cached transaction with the given ID without leasing it or
// affecting its expiration.
func (c *txnCache) Get(id uint64) (client.Txn, bool) {
	cached, ok := c.cache.Load(id)
	if !ok {
		return nil, false
	}
	return cached.txn, true
}

// Acquire returns a lease on the cached transaction with the given ID,
// suspending its expiration until the lease is released.
func (c *txnCache) Acquire(id uint64) (*txnLease, bool) {
	return c.cache.Acquire(id)
}

// LoadAndDelete removes the transaction with the given ID and returns it along
// with its configured TTL, transferring ownership to the caller.
func (c *txnCache) LoadAndDelete(id uint64) (client.Txn, time.Duration, bool) {
	cached, ok := c.cache.LoadAndDelete(id)
	if !ok {
		return nil, 0, false
	}
	return cached.txn, cached.ttl, true
}

func (c *txnCache) Close() {
	c.cache.Close()
}
