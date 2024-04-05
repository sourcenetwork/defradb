// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package lens

import (
	"context"
	"sync"

	"github.com/lens-vm/lens/host-go/config"
	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/lens-vm/lens/host-go/engine/module"
	"github.com/lens-vm/lens/host-go/runtimes/wasmtime"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/description"
	"github.com/sourcenetwork/defradb/errors"
)

// todo: This file, particularly the `lensPool` stuff, contains fairly sensitive code that is both
// cumbersome to fully test with integration/benchmark tests, and can have a significant affect on
// the users if broken (deadlocks, large performance degradation).  It should have proper unit tests.
// https://github.com/sourcenetwork/defradb/issues/1596

// lensRegistry is responsible for managing all migration related state within a local
// database instance.
type lensRegistry struct {
	poolSize int

	// The runtime used to execute lens wasm modules.
	runtime module.Runtime

	// The modules by file path used to instantiate lens wasm module instances.
	modulesByPath map[string]module.Module
	moduleLock    sync.Mutex

	lensPoolsByCollectionID     map[uint32]*lensPool
	reversedPoolsByCollectionID map[uint32]*lensPool
	poolLock                    sync.RWMutex

	// Writable transaction contexts by transaction ID.
	//
	// Read-only transaction contexts are not tracked.
	txnCtxs map[uint64]*txnContext
	txnLock sync.RWMutex
}

// txnContext contains uncommitted transaction state tracked by the registry,
// stuff within here should be accessible from within this transaction but not
// from outside.
type txnContext struct {
	txn                         datastore.Txn
	lensPoolsByCollectionID     map[uint32]*lensPool
	reversedPoolsByCollectionID map[uint32]*lensPool
}

func newTxnCtx(txn datastore.Txn) *txnContext {
	return &txnContext{
		txn:                         txn,
		lensPoolsByCollectionID:     map[uint32]*lensPool{},
		reversedPoolsByCollectionID: map[uint32]*lensPool{},
	}
}

// TxnSource represents an object capable of constructing the transactions that
// implicit-transaction registries need internally.
type TxnSource interface {
	NewTxn(context.Context, bool) (datastore.Txn, error)
}

// DefaultPoolSize is the default size of the lens pool for each schema version.
const DefaultPoolSize int = 5

// NewRegistry instantiates a new registery.
//
// It will be of size 5 (per schema version) if a size is not provided.
func NewRegistry(db TxnSource, opts ...Option) client.LensRegistry {
	registry := &lensRegistry{
		poolSize:                    DefaultPoolSize,
		runtime:                     wasmtime.New(),
		modulesByPath:               map[string]module.Module{},
		lensPoolsByCollectionID:     map[uint32]*lensPool{},
		reversedPoolsByCollectionID: map[uint32]*lensPool{},
		txnCtxs:                     map[uint64]*txnContext{},
	}

	for _, opt := range opts {
		opt(registry)
	}

	return &implicitTxnLensRegistry{
		db:       db,
		registry: registry,
	}
}

func (r *lensRegistry) getCtx(txn datastore.Txn, readonly bool) *txnContext {
	r.txnLock.RLock()
	if txnCtx, ok := r.txnCtxs[txn.ID()]; ok {
		r.txnLock.RUnlock()
		return txnCtx
	}
	r.txnLock.RUnlock()

	txnCtx := newTxnCtx(txn)
	if readonly {
		return txnCtx
	}

	r.txnLock.Lock()
	r.txnCtxs[txn.ID()] = txnCtx
	r.txnLock.Unlock()

	txnCtx.txn.OnSuccess(func() {
		r.poolLock.Lock()
		for collectionID, locker := range txnCtx.lensPoolsByCollectionID {
			r.lensPoolsByCollectionID[collectionID] = locker
		}
		for collectionID, locker := range txnCtx.reversedPoolsByCollectionID {
			r.reversedPoolsByCollectionID[collectionID] = locker
		}
		r.poolLock.Unlock()

		r.txnLock.Lock()
		delete(r.txnCtxs, txn.ID())
		r.txnLock.Unlock()
	})

	txn.OnError(func() {
		r.txnLock.Lock()
		delete(r.txnCtxs, txn.ID())
		r.txnLock.Unlock()
	})

	txn.OnDiscard(func() {
		// Delete it to help reduce the build up of memory, the txnCtx will be re-contructed if the
		// txn is reused after discard.
		r.txnLock.Lock()
		delete(r.txnCtxs, txn.ID())
		r.txnLock.Unlock()
	})

	return txnCtx
}

func (r *lensRegistry) setMigration(
	ctx context.Context,
	txnCtx *txnContext,
	collectionID uint32,
	cfg model.Lens,
) error {
	inversedModuleCfgs := make([]model.LensModule, len(cfg.Lenses))
	for i, moduleCfg := range cfg.Lenses {
		// Reverse the order of the lenses for the inverse migration.
		inversedModuleCfgs[len(cfg.Lenses)-i-1] = model.LensModule{
			Path: moduleCfg.Path,
			// Reverse the direction of the lens.
			// This needs to be done on a clone of the original cfg or we may end up mutating
			// the original.
			Inverse:   !moduleCfg.Inverse,
			Arguments: moduleCfg.Arguments,
		}
	}

	reversedCfg := model.Lens{
		Lenses: inversedModuleCfgs,
	}

	err := r.cachePool(txnCtx.txn, txnCtx.lensPoolsByCollectionID, cfg, collectionID)
	if err != nil {
		return err
	}
	err = r.cachePool(txnCtx.txn, txnCtx.reversedPoolsByCollectionID, reversedCfg, collectionID)
	// For now, checking this error is the best way of determining if a migration has an inverse.
	// Inverses are optional.
	//nolint:revive
	if err != nil && !errors.Is(errors.New("Export `inverse` does not exist"), err) {
		return err
	}

	return nil
}

func (r *lensRegistry) cachePool(
	txn datastore.Txn,
	target map[uint32]*lensPool,
	cfg model.Lens,
	collectionID uint32,
) error {
	pool := r.newPool(r.poolSize, cfg)

	for i := 0; i < r.poolSize; i++ {
		lensPipe, err := r.newLensPipe(cfg)
		if err != nil {
			return err
		}
		pool.returnLens(lensPipe)
	}

	target[collectionID] = pool

	return nil
}

func (r *lensRegistry) reloadLenses(ctx context.Context, txnCtx *txnContext) error {
	cols, err := description.GetCollections(ctx, txnCtx.txn)
	if err != nil {
		return err
	}

	for _, col := range cols {
		sources := col.CollectionSources()

		if len(sources) == 0 {
			continue
		}

		// WARNING: Here we are only dealing with the first source in the set, this is fine for now as
		// currently collections can only have one source, however this code will need to change if/when
		// collections support multiple sources.

		if !sources[0].Transform.HasValue() {
			continue
		}

		err = r.setMigration(ctx, txnCtx, col.ID, sources[0].Transform.Value())
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *lensRegistry) migrateUp(
	txnCtx *txnContext,
	src enumerable.Enumerable[LensDoc],
	collectionID uint32,
) (enumerable.Enumerable[LensDoc], error) {
	return r.migrate(r.lensPoolsByCollectionID, txnCtx.lensPoolsByCollectionID, src, collectionID)
}

func (r *lensRegistry) migrateDown(
	txnCtx *txnContext,
	src enumerable.Enumerable[LensDoc],
	collectionID uint32,
) (enumerable.Enumerable[LensDoc], error) {
	return r.migrate(r.reversedPoolsByCollectionID, txnCtx.reversedPoolsByCollectionID, src, collectionID)
}

func (r *lensRegistry) migrate(
	pools map[uint32]*lensPool,
	txnPools map[uint32]*lensPool,
	src enumerable.Enumerable[LensDoc],
	collectionID uint32,
) (enumerable.Enumerable[LensDoc], error) {
	lensPool, ok := r.getPool(pools, txnPools, collectionID)
	if !ok {
		// If there are no migrations for this schema version, just return the given source.
		return src, nil
	}

	lens, err := lensPool.borrow()
	if err != nil {
		return nil, err
	}

	lens.SetSource(src)

	return lens, nil
}

func (r *lensRegistry) getPool(
	pools map[uint32]*lensPool,
	txnPools map[uint32]*lensPool,
	collectionID uint32,
) (*lensPool, bool) {
	if pool, ok := txnPools[collectionID]; ok {
		return pool, true
	}

	r.poolLock.RLock()
	pool, ok := pools[collectionID]
	r.poolLock.RUnlock()
	return pool, ok
}

// lensPool provides a pool-like mechanic for caching a limited number of wasm lens modules in
// a thread safe fashion.
//
// Instanstiating a lens module is pretty expensive as it has to spin up the wasm runtime environment
// so we need to limit how frequently we do this.
type lensPool struct {
	// The config used to create the lenses within this locker.
	cfg model.Lens

	registry *lensRegistry

	// Using a buffered channel provides an easy way to manage a finite
	// number of lenses.
	//
	// We wish to limit this as creating lenses is expensive, and we do not want
	// to be dynamically resizing this collection and spinning up new lens instances
	// in user time, or holding on to large numbers of them.
	pipes chan *lensPipe
}

func (r *lensRegistry) newPool(lensPoolSize int, cfg model.Lens) *lensPool {
	return &lensPool{
		cfg:      cfg,
		registry: r,
		pipes:    make(chan *lensPipe, lensPoolSize),
	}
}

// borrow attempts to borrow a module from the locker, if one is not available
// it will return a new, temporary instance that will not be returned to the locker
// after use.
func (l *lensPool) borrow() (enumerable.Socket[LensDoc], error) {
	select {
	case lens := <-l.pipes:
		return &borrowedEnumerable{
			source: lens,
			pool:   l,
		}, nil
	default:
		// If there are no free cached migrations within the locker, create a new temporary one
		// instead of blocking.
		return l.registry.newLensPipe(l.cfg)
	}
}

// returnLens returns a borrowed module to the locker, allowing it to be reused by other contexts.
func (l *lensPool) returnLens(lens *lensPipe) {
	l.pipes <- lens
}

// borrowedEnumerable is an enumerable tied to a locker.
//
// it exposes the source enumerable and amends the Reset function so that when called, the source
// pipe is returned to the locker.
type borrowedEnumerable struct {
	source *lensPipe
	pool   *lensPool
}

var _ enumerable.Socket[LensDoc] = (*borrowedEnumerable)(nil)

func (s *borrowedEnumerable) SetSource(newSource enumerable.Enumerable[LensDoc]) {
	s.source.SetSource(newSource)
}

func (s *borrowedEnumerable) Next() (bool, error) {
	return s.source.Next()
}

func (s *borrowedEnumerable) Value() (LensDoc, error) {
	return s.source.Value()
}

func (s *borrowedEnumerable) Reset() {
	s.pool.returnLens(s.source)
	s.source.Reset()
}

// lensPipe provides a mechanic where the underlying wasm module can be hidden from consumers
// and allow input sources to be swapped in and out as different actors borrow it from the locker.
type lensPipe struct {
	input      enumerable.Socket[LensDoc]
	enumerable enumerable.Enumerable[LensDoc]
}

var _ enumerable.Socket[LensDoc] = (*lensPipe)(nil)

func (r *lensRegistry) newLensPipe(cfg model.Lens) (*lensPipe, error) {
	socket := enumerable.NewSocket[LensDoc]()

	r.moduleLock.Lock()
	enumerable, err := config.LoadInto[LensDoc, LensDoc](r.runtime, r.modulesByPath, cfg, socket)
	r.moduleLock.Unlock()

	if err != nil {
		return nil, err
	}

	return &lensPipe{
		input:      socket,
		enumerable: enumerable,
	}, nil
}

func (p *lensPipe) SetSource(newSource enumerable.Enumerable[LensDoc]) {
	p.input.SetSource(newSource)
}

func (p *lensPipe) Next() (bool, error) {
	return p.enumerable.Next()
}

func (p *lensPipe) Value() (LensDoc, error) {
	return p.enumerable.Value()
}

func (p *lensPipe) Reset() {
	p.input.Reset()
	// WARNING: Currently the wasm module state is not reset by calling reset on the enumerable
	// this means that state from one context may leak to the next useage.  There is a ticket here
	// to fix this: https://github.com/lens-vm/lens/issues/46
	p.enumerable.Reset()
}
