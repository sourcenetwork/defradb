// Copyright 2024 Democratized Data Foundation
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
	"errors"
	"sync"
	"time"

	badgerds "github.com/dgraph-io/badger/v4"
	badgeropts "github.com/dgraph-io/badger/v4/options"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/badger"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client/options"
)

func init() {
	constructor := func(ctx context.Context, opts *options.NodeStoreOptions) (corekv.TxnStore, error) {
		var path string
		if !opts.BadgerInMemory {
			// Badger will error if we give it a path and set `InMemory` to true
			path = opts.Path
		}

		badgerOpts := badgerds.DefaultOptions(path)
		badgerOpts.InMemory = opts.BadgerInMemory
		badgerOpts.ValueLogFileSize = opts.BadgerFileSize
		badgerOpts.EncryptionKey = opts.BadgerEncryptionKey
		badgerOpts.ValueThreshold = 1 << 8
		badgerOpts.Compression = badgeropts.ZSTD
		badgerOpts.ZSTDCompressionLevel = 1

		if len(opts.BadgerEncryptionKey) > 0 {
			// Having a cache improves the performance.
			// Otherwise, your reads would be very slow while encryption is enabled.
			// https://dgraph.io/docs/badger/get-started/#encryption-mode
			badgerOpts.IndexCacheSize = 100 << 20
		}

		// Value log GC is unsupported for in-memory stores, so only persistent
		// stores get the GC wrapper.
		if opts.BadgerInMemory {
			return badger.NewDatastore(path, badgerOpts)
		}
		return newBadgerStore(path, badgerOpts)
	}
	purge := func(ctx context.Context, opts *options.NodeStoreOptions) error {
		store, err := constructor(ctx, opts)
		if err != nil {
			return err
		}
		err = store.(corekv.Dropable).DropAll()
		if err != nil {
			return err
		}
		return store.Close()
	}

	storeConstructors[options.NodeBadgerStore] = constructor
	storePurgeFuncs[options.NodeBadgerStore] = purge

	storeConstructors[options.NodeDefaultStore] = constructor
	storePurgeFuncs[options.NodeDefaultStore] = purge
}

const (
	// valueLogGCInterval is how often value log GC runs on a persistent store.
	valueLogGCInterval = 5 * time.Minute
	// valueLogGCDiscardRatio is the fraction of a value log file that must be
	// reclaimable before badger rewrites it.
	valueLogGCDiscardRatio = 0.1
)

// badgerStore wraps a persistent badger datastore and periodically runs value log
// GC. Badger does not reclaim the value log space of deleted or overwritten
// entries on its own, so the GC keeps the on-disk store bounded.
type badgerStore struct {
	*badger.Datastore

	db       *badgerds.DB
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

// newBadgerStore opens a persistent badger store at path and starts value log GC.
func newBadgerStore(path string, opts badgerds.Options) (*badgerStore, error) {
	opts.Dir = path
	opts.ValueDir = path
	opts.Logger = nil // badger's default logger is very verbose
	db, err := badgerds.Open(opts)
	if err != nil {
		return nil, err
	}

	store := &badgerStore{
		Datastore: badger.NewDatastoreFrom(db),
		db:        db,
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
	}
	go store.runValueLogGC()
	return store, nil
}

// Close stops value log GC and closes the underlying store.
func (s *badgerStore) Close() error {
	s.stopOnce.Do(func() {
		close(s.stop)
		<-s.done
	})
	return s.Datastore.Close()
}

// runValueLogGC reclaims value log space on a fixed interval until the store closes.
func (s *badgerStore) runValueLogGC() {
	defer close(s.done)

	ticker := time.NewTicker(valueLogGCInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.reclaimValueLog()
		}
	}
}

// reclaimValueLog runs value log GC repeatedly until there is nothing left to
// reclaim, an error occurs, or the store starts closing. Each successful call
// rewrites one file; ErrNoRewrite means no file was eligible and ends the loop.
// Any other error is logged, since it means GC could not make progress.
func (s *badgerStore) reclaimValueLog() {
	start := time.Now()
	reclaimed := 0
	for {
		select {
		case <-s.stop:
			return
		default:
		}
		if err := s.db.RunValueLogGC(valueLogGCDiscardRatio); err != nil {
			if !errors.Is(err, badgerds.ErrNoRewrite) {
				log.ErrorE("Badger value log GC failed", err)
			}
			break
		}
		reclaimed++
	}
	if reclaimed > 0 {
		log.Info("Reclaimed badger value log files",
			corelog.Int("files", reclaimed),
			corelog.Duration("duration", time.Since(start)))
	}
}
