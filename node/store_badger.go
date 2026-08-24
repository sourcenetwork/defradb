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
	"os"
	"strconv"
	"sync"
	"time"

	badgerds "github.com/dgraph-io/badger/v4"
	badgeropts "github.com/dgraph-io/badger/v4/options"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/badger"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
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

	// defaultOrphanBlockGCInterval is how often the orphan block sweep runs.
	defaultOrphanBlockGCInterval = 5 * time.Minute
	// defaultOrphanBlockTTL is how long a fetched-but-unmerged block is kept before the
	// sweep reclaims it. It only has to outlast an in-flight merge; anything older is an
	// abandoned fetch.
	defaultOrphanBlockTTL = 30 * time.Minute
	// orphanBlockGCScanLimit bounds how many markers the sweep examines per run, so a
	// large backlog is worked down over several runs rather than one long pass.
	orphanBlockGCScanLimit = 100_000

	// The sweep deletes blocks, so an operator needs to stop or slow it on a running
	// node without a new image.
	envOrphanGCDisabled = "DEFRA_ORPHAN_GC_DISABLED"
	envOrphanGCInterval = "DEFRA_ORPHAN_GC_INTERVAL"
	envOrphanBlockTTL   = "DEFRA_ORPHAN_BLOCK_TTL"
)

// durationFromEnv returns the duration in name, or fallback when it is unset, malformed
// or not positive. A bad value falls back rather than failing the store open.
func durationFromEnv(name string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(os.Getenv(name))
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

// boolFromEnv returns the boolean in name, or false when it is unset or malformed.
// A malformed value is logged so it is not mistaken for the variable being unset.
func boolFromEnv(name string) bool {
	raw, ok := os.LookupEnv(name)
	if !ok {
		return false
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		log.Info("Ignoring malformed boolean environment variable",
			corelog.String("name", name), corelog.String("value", raw))
		return false
	}
	return v
}

// badgerStore wraps a persistent badger datastore and runs its periodic background
// maintenance: value log GC, which reclaims the space of deleted or overwritten
// entries badger does not reclaim on its own, and an orphan block sweep, which
// deletes blocks fetched during sync whose merge never completed.
type badgerStore struct {
	*badger.Datastore

	db *badgerds.DB
	// stop cancels the context the background maintenance runs under. A sweep runs for
	// minutes, so it has to be interruptible mid-run, not only between ticks.
	stop      context.CancelFunc
	wg        sync.WaitGroup
	closeOnce sync.Once

	// Resolved once at open so the sweep loop does not re-read the environment per tick.
	orphanGCInterval time.Duration
	orphanBlockTTL   time.Duration
	orphanGCDisabled bool
}

// newBadgerStore opens a persistent badger store at path and starts its background
// maintenance. The orphan sweep can be disabled and retuned through the environment.
func newBadgerStore(path string, opts badgerds.Options) (*badgerStore, error) {
	opts.Dir = path
	opts.ValueDir = path
	opts.Logger = nil // badger's default logger is very verbose
	db, err := badgerds.Open(opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	store := &badgerStore{
		Datastore:        badger.NewDatastoreFrom(db),
		db:               db,
		stop:             cancel,
		orphanGCInterval: durationFromEnv(envOrphanGCInterval, defaultOrphanBlockGCInterval),
		orphanBlockTTL:   durationFromEnv(envOrphanBlockTTL, defaultOrphanBlockTTL),
		orphanGCDisabled: boolFromEnv(envOrphanGCDisabled),
	}
	store.wg.Add(1)
	go store.runValueLogGC(ctx)

	if store.orphanGCDisabled {
		log.Info("Orphan block sweep disabled")
	} else {
		log.Info("Orphan block sweep enabled",
			corelog.Duration("interval", store.orphanGCInterval),
			corelog.Duration("ttl", store.orphanBlockTTL))
		store.wg.Add(1)
		go store.runOrphanBlockGC(ctx)
	}
	return store, nil
}

// Close stops the background maintenance and closes the underlying store.
func (s *badgerStore) Close() error {
	s.closeOnce.Do(func() {
		s.stop()
		s.wg.Wait()
	})
	return s.Datastore.Close()
}

// runValueLogGC reclaims value log space on a fixed interval until the store closes.
func (s *badgerStore) runValueLogGC(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(valueLogGCInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.reclaimValueLog(ctx)
		}
	}
}

// reclaimValueLog runs value log GC repeatedly until there is nothing left to
// reclaim, an error occurs, or the store starts closing. Each successful call
// rewrites one file; ErrNoRewrite means no file was eligible and ends the loop.
// Any other error is logged, since it means GC could not make progress.
//
// The on-disk size is reported on every pass, reclaimed or not, so whether the store is
// bounded can be read from the log without shell access to the volume.
func (s *badgerStore) reclaimValueLog(ctx context.Context) {
	start := time.Now()
	reclaimed := 0
	for {
		if ctx.Err() != nil {
			return
		}
		if err := s.db.RunValueLogGC(valueLogGCDiscardRatio); err != nil {
			if !errors.Is(err, badgerds.ErrNoRewrite) {
				log.ErrorE("Badger value log GC failed", err)
			}
			break
		}
		reclaimed++
	}
	lsm, vlog := s.db.Size()
	log.Info("Badger value log GC",
		corelog.Int("files", reclaimed),
		corelog.Int64("lsmBytes", lsm),
		corelog.Int64("vlogBytes", vlog),
		corelog.Duration("duration", time.Since(start)))
}

// runOrphanBlockGC sweeps the blockstore for orphaned blocks on a fixed interval
// until the store closes. It carries a cursor across runs so the whole marker index
// is worked through over time, restarting from the beginning after each full pass.
//
// The first sweep is held back one TTL, so nothing can be reclaimed sooner than that after
// the store opens.
func (s *badgerStore) runOrphanBlockGC(ctx context.Context) {
	defer s.wg.Done()

	select {
	case <-ctx.Done():
		return
	case <-time.After(s.orphanBlockTTL):
	}

	ticker := time.NewTicker(s.orphanGCInterval)
	defer ticker.Stop()

	var cursor []byte
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cursor = s.reclaimOrphanBlocks(ctx, cursor)
		}
	}
}

// reclaimOrphanBlocks runs one sweep step from cursor and returns the cursor to
// resume from. On error it keeps the cursor, so a transient failure part-way through
// the index costs one run rather than every marker examined since the last full pass.
func (s *badgerStore) reclaimOrphanBlocks(ctx context.Context, cursor []byte) []byte {
	start := time.Now()
	result, err := datastore.ReclaimOrphanBlocks(
		ctx, s, start.Add(-s.orphanBlockTTL), cursor, orphanBlockGCScanLimit)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			log.ErrorE("Orphan block sweep failed", err)
		}
		return cursor
	}
	if result.Scanned > 0 {
		log.Info("Swept orphan blocks",
			corelog.Int("reclaimed", result.Reclaimed),
			corelog.Int("repaired", result.Repaired),
			corelog.Int("conflicts", result.Conflicts),
			corelog.Int("unparsed", result.Unparsed),
			corelog.Int("scanned", result.Scanned),
			corelog.Bool("completedPass", result.NextKey == nil),
			corelog.Duration("duration", time.Since(start)))
	}
	return result.NextKey
}
