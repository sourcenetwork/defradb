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
	"encoding/json"
	"sync"

	"github.com/ipfs/go-datastore/query"
	"github.com/lens-vm/lens/host-go/config"
	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/lens-vm/lens/host-go/engine/module"
	"github.com/lens-vm/lens/host-go/runtimes/wasmtime"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
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

	lensPoolsBySchemaVersionID     map[string]*lensPool
	reversedPoolsBySchemaVersionID map[string]*lensPool
	poolLock                       sync.RWMutex

	// lens configurations by source schema version ID
	configs    map[string]client.LensConfig
	configLock sync.RWMutex

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
	txn                            datastore.Txn
	lensPoolsBySchemaVersionID     map[string]*lensPool
	reversedPoolsBySchemaVersionID map[string]*lensPool
	configs                        map[string]client.LensConfig
}

func newTxnCtx(txn datastore.Txn) *txnContext {
	return &txnContext{
		txn:                            txn,
		lensPoolsBySchemaVersionID:     map[string]*lensPool{},
		reversedPoolsBySchemaVersionID: map[string]*lensPool{},
		configs:                        map[string]client.LensConfig{},
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
func NewRegistry(lensPoolSize immutable.Option[int], db TxnSource) client.LensRegistry {
	var size int
	if lensPoolSize.HasValue() {
		size = lensPoolSize.Value()
	} else {
		size = DefaultPoolSize
	}

	return &implicitTxnLensRegistry{
		db: db,
		registry: &lensRegistry{
			poolSize:                       size,
			runtime:                        wasmtime.New(),
			modulesByPath:                  map[string]module.Module{},
			lensPoolsBySchemaVersionID:     map[string]*lensPool{},
			reversedPoolsBySchemaVersionID: map[string]*lensPool{},
			configs:                        map[string]client.LensConfig{},
			txnCtxs:                        map[uint64]*txnContext{},
		},
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
		for schemaVersionID, locker := range txnCtx.lensPoolsBySchemaVersionID {
			r.lensPoolsBySchemaVersionID[schemaVersionID] = locker
		}
		for schemaVersionID, locker := range txnCtx.reversedPoolsBySchemaVersionID {
			r.reversedPoolsBySchemaVersionID[schemaVersionID] = locker
		}
		r.poolLock.Unlock()

		r.configLock.Lock()
		for schemaVersionID, cfg := range txnCtx.configs {
			r.configs[schemaVersionID] = cfg
		}
		r.configLock.Unlock()

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

func (r *lensRegistry) setMigration(ctx context.Context, txnCtx *txnContext, cfg client.LensConfig) error {
	key := core.NewSchemaVersionMigrationKey(cfg.SourceSchemaVersionID)

	json, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = txnCtx.txn.Systemstore().Set(ctx, key.ToDS().Bytes(), json)
	if err != nil {
		return err
	}

	err = r.cacheLens(txnCtx, cfg)
	if err != nil {
		return err
	}

	return nil
}

func (r *lensRegistry) cacheLens(txnCtx *txnContext, cfg client.LensConfig) error {
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

	reversedCfg := client.LensConfig{
		SourceSchemaVersionID:      cfg.SourceSchemaVersionID,
		DestinationSchemaVersionID: cfg.DestinationSchemaVersionID,
		Lens: model.Lens{
			Lenses: inversedModuleCfgs,
		},
	}

	err := r.cachePool(txnCtx.txn, txnCtx.lensPoolsBySchemaVersionID, cfg)
	if err != nil {
		return err
	}
	err = r.cachePool(txnCtx.txn, txnCtx.reversedPoolsBySchemaVersionID, reversedCfg)
	// For now, checking this error is the best way of determining if a migration has an inverse.
	// Inverses are optional.
	//nolint:revive
	if err != nil && !errors.Is(errors.New("Export `inverse` does not exist"), err) {
		return err
	}

	txnCtx.configs[cfg.SourceSchemaVersionID] = cfg

	return nil
}

func (r *lensRegistry) cachePool(txn datastore.Txn, target map[string]*lensPool, cfg client.LensConfig) error {
	pool := r.newPool(r.poolSize, cfg)

	for i := 0; i < r.poolSize; i++ {
		lensPipe, err := r.newLensPipe(cfg)
		if err != nil {
			return err
		}
		pool.returnLens(lensPipe)
	}

	target[cfg.SourceSchemaVersionID] = pool

	return nil
}

func (r *lensRegistry) reloadLenses(ctx context.Context, txnCtx *txnContext) error {
	prefix := core.NewSchemaVersionMigrationKey("")
	q, err := txnCtx.txn.Systemstore().Query(ctx, query.Query{
		Prefix: prefix.ToString(),
	})
	if err != nil {
		return err
	}

	for res := range q.Next() {
		// check for Done on context first
		select {
		case <-ctx.Done():
			// we've been cancelled! ;)
			err = q.Close()
			if err != nil {
				return err
			}

			return nil
		default:
			// noop, just continue on the with the for loop
		}

		if res.Error != nil {
			err = q.Close()
			if err != nil {
				return errors.Wrap(err.Error(), res.Error)
			}
			return res.Error
		}

		var cfg client.LensConfig
		err = json.Unmarshal(res.Value, &cfg)
		if err != nil {
			err = q.Close()
			if err != nil {
				return err
			}
			return err
		}

		err = r.cacheLens(txnCtx, cfg)
		if err != nil {
			err = q.Close()
			if err != nil {
				return errors.Wrap(err.Error(), res.Error)
			}
			return err
		}
	}

	err = q.Close()
	if err != nil {
		return err
	}

	return nil
}

func (r *lensRegistry) migrateUp(
	txnCtx *txnContext,
	src enumerable.Enumerable[LensDoc],
	schemaVersionID string,
) (enumerable.Enumerable[LensDoc], error) {
	return r.migrate(r.lensPoolsBySchemaVersionID, txnCtx.lensPoolsBySchemaVersionID, src, schemaVersionID)
}

func (r *lensRegistry) migrateDown(
	txnCtx *txnContext,
	src enumerable.Enumerable[LensDoc],
	schemaVersionID string,
) (enumerable.Enumerable[LensDoc], error) {
	return r.migrate(r.reversedPoolsBySchemaVersionID, txnCtx.reversedPoolsBySchemaVersionID, src, schemaVersionID)
}

func (r *lensRegistry) migrate(
	pools map[string]*lensPool,
	txnPools map[string]*lensPool,
	src enumerable.Enumerable[LensDoc],
	schemaVersionID string,
) (enumerable.Enumerable[LensDoc], error) {
	lensPool, ok := r.getPool(pools, txnPools, schemaVersionID)
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

func (r *lensRegistry) config(txnCtx *txnContext) []client.LensConfig {
	configs := map[string]client.LensConfig{}
	r.configLock.RLock()
	for schemaVersionID, cfg := range r.configs {
		configs[schemaVersionID] = cfg
	}
	r.configLock.RUnlock()

	// If within a txn actively writing to this registry overwrite
	// values from the (commited) registry.
	// Note: Config cannot be removed, only replaced at the moment.
	for schemaVersionID, cfg := range txnCtx.configs {
		configs[schemaVersionID] = cfg
	}

	result := []client.LensConfig{}
	for _, cfg := range configs {
		result = append(result, cfg)
	}
	return result
}

func (r *lensRegistry) hasMigration(txnCtx *txnContext, schemaVersionID string) bool {
	_, hasMigration := r.getPool(r.lensPoolsBySchemaVersionID, txnCtx.lensPoolsBySchemaVersionID, schemaVersionID)
	return hasMigration
}

func (r *lensRegistry) getPool(
	pools map[string]*lensPool,
	txnPools map[string]*lensPool,
	schemaVersionID string,
) (*lensPool, bool) {
	if pool, ok := txnPools[schemaVersionID]; ok {
		return pool, true
	}

	r.poolLock.RLock()
	pool, ok := pools[schemaVersionID]
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
	cfg client.LensConfig

	registry *lensRegistry

	// Using a buffered channel provides an easy way to manage a finite
	// number of lenses.
	//
	// We wish to limit this as creating lenses is expensive, and we do not want
	// to be dynamically resizing this collection and spinning up new lens instances
	// in user time, or holding on to large numbers of them.
	pipes chan *lensPipe
}

func (r *lensRegistry) newPool(lensPoolSize int, cfg client.LensConfig) *lensPool {
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

func (r *lensRegistry) newLensPipe(cfg client.LensConfig) (*lensPipe, error) {
	socket := enumerable.NewSocket[LensDoc]()

	r.moduleLock.Lock()
	enumerable, err := config.LoadInto[LensDoc, LensDoc](r.runtime, r.modulesByPath, cfg.Lens, socket)
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
