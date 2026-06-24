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

type cachedTxn struct {
	txn client.Txn
	ttl time.Duration
}

type txnCache struct {
	defaultTTL time.Duration
	cache      *ttl.Cache[uint64, cachedTxn]
}

func newTxnCache(ctx context.Context, opts *options.NodeHTTPOptions) (*txnCache, error) {
	cfg := txnCacheConfig(opts)
	cache, err := ttl.NewCache(ctx, cfg.tick, cfg.buckets, func(_ uint64, item cachedTxn) {
		item.txn.Discard()
	})
	if err != nil {
		return nil, err
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

func (c *txnCache) Store(txn client.Txn, txnTTL time.Duration) error {
	if txnTTL == 0 {
		txnTTL = c.defaultTTL
	}
	return c.cache.Store(txn.ID(), cachedTxn{txn: txn, ttl: txnTTL}, txnTTL)
}

func (c *txnCache) Load(id uint64) (client.Txn, bool) {
	item, ok := c.cache.Load(id)
	if !ok {
		return nil, false
	}
	if err := c.cache.UpdateTTL(id, item.ttl); err != nil {
		log.ErrorE("failed to refresh transaction ttl", err)
	}
	return item.txn, true
}

func (c *txnCache) LoadAndDelete(id uint64) (client.Txn, bool) {
	item, ok := c.cache.LoadAndDelete(id)
	if !ok {
		return nil, false
	}
	return item.txn, true
}

func (c *txnCache) Delete(id uint64) {
	c.cache.Delete(id)
}

func (c *txnCache) Close() {
	c.cache.Stop()
}
