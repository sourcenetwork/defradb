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

	"github.com/ipfs/go-datastore/query"
	"github.com/lens-vm/lens/host-go/config"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

// todo: This file, particularly the `lensLocker`` stuff, contains fairly sensitive code that is both
// cumbersome to fully test with integration/benchmark tests, and can have a significant affect on
// the users if broken (deadlocks, large performance degradation).  It should have proper unit tests.
// https://github.com/sourcenetwork/defradb/issues/1596

// lensRegistry is responsible for managing all migration related state within a local
// database instance.
type lensRegistry struct {
	lensPoolSize int
	// lens lockers by source schema version ID
	lensWarehouse map[string]*lensLocker
	// lens configurations by source schema version ID
	lensConfigs map[string]client.LensConfig
}

var _ client.LensRegistry = (*lensRegistry)(nil)

// DefaultPoolSize is the default size of the lens pool for each schema version.
const DefaultPoolSize int = 5

// NewRegistry instantiates a new registery.
//
// It will be of size 5 (per schema version) if a size is not provided.
func NewRegistry(lensPoolSize immutable.Option[int]) *lensRegistry {
	var size int
	if lensPoolSize.HasValue() {
		size = lensPoolSize.Value()
	} else {
		size = DefaultPoolSize
	}

	return &lensRegistry{
		lensPoolSize:  size,
		lensWarehouse: map[string]*lensLocker{},
		lensConfigs:   map[string]client.LensConfig{},
	}
}

func (r *lensRegistry) SetMigration(ctx context.Context, txn datastore.Txn, cfg client.LensConfig) error {
	key := core.NewSchemaVersionMigrationKey(cfg.SourceSchemaVersionID)

	json, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = txn.Systemstore().Put(ctx, key.ToDS(), json)
	if err != nil {
		return err
	}

	err = r.cacheLens(txn, cfg)
	if err != nil {
		return err
	}

	return nil
}

func (r *lensRegistry) cacheLens(txn datastore.Txn, cfg client.LensConfig) error {
	locker, lockerAlreadyExists := r.lensWarehouse[cfg.SourceSchemaVersionID]
	if !lockerAlreadyExists {
		locker = newLocker(r.lensPoolSize, cfg)
	}

	newLensPipes := make([]*lensPipe, r.lensPoolSize)
	for i := 0; i < r.lensPoolSize; i++ {
		var err error
		newLensPipes[i], err = newLensPipe(cfg)
		if err != nil {
			return err
		}
	}

	// todo - handling txns like this means that the migrations are not available within the current
	// transaction if used for stuff (e.g. GQL requests) before commit.
	// https://github.com/sourcenetwork/defradb/issues/1592
	txn.OnSuccess(func() {
		if !lockerAlreadyExists {
			r.lensWarehouse[cfg.SourceSchemaVersionID] = locker
		}

	drainLoop:
		for {
			select {
			case <-locker.safes:
			default:
				break drainLoop
			}
		}

		for _, lensPipe := range newLensPipes {
			locker.returnLens(lensPipe)
		}

		r.lensConfigs[cfg.SourceSchemaVersionID] = cfg
	})

	return nil
}

func (r *lensRegistry) ReloadLenses(ctx context.Context, txn datastore.Txn) error {
	prefix := core.NewSchemaVersionMigrationKey("")
	q, err := txn.Systemstore().Query(ctx, query.Query{
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
				return err
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

		err = r.cacheLens(txn, cfg)
		if err != nil {
			err = q.Close()
			if err != nil {
				return err
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

func (r *lensRegistry) MigrateUp(
	src enumerable.Enumerable[LensDoc],
	schemaVersionID string,
) (enumerable.Enumerable[LensDoc], error) {
	lensLocker, ok := r.lensWarehouse[schemaVersionID]
	if !ok {
		// If there are no migrations for this schema version, just return the given source.
		return src, nil
	}

	lens, err := lensLocker.borrow()
	if err != nil {
		return nil, err
	}

	lens.SetSource(src)

	return lens, nil
}

func (*lensRegistry) MigrateDown(
	src enumerable.Enumerable[LensDoc],
	schemaVersionID string,
) (enumerable.Enumerable[LensDoc], error) {
	// todo: https://github.com/sourcenetwork/defradb/issues/1591
	return src, nil
}

func (r *lensRegistry) Config() []client.LensConfig {
	result := []client.LensConfig{}
	for _, cfg := range r.lensConfigs {
		result = append(result, cfg)
	}
	return result
}

func (r *lensRegistry) HasMigration(schemaVersionID string) bool {
	_, hasMigration := r.lensWarehouse[schemaVersionID]
	return hasMigration
}

// lensLocker provides a pool-like mechanic for caching a limited number of wasm lens modules in
// a thread safe fashion.
//
// Instanstiating a lens module is pretty expensive as it has to spin up the wasm runtime environment
// so we need to limit how frequently we do this.
type lensLocker struct {
	// The config used to create the lenses within this locker.
	cfg client.LensConfig

	// Using a buffered channel provides an easy way to manage a finite
	// number of lenses.
	//
	// We wish to limit this as creating lenses is expensive, and we do not want
	// to be dynamically resizing this collection and spinning up new lens instances
	// in user time, or holding on to large numbers of them.
	safes chan *lensPipe
}

func newLocker(lensPoolSize int, cfg client.LensConfig) *lensLocker {
	return &lensLocker{
		cfg:   cfg,
		safes: make(chan *lensPipe, lensPoolSize),
	}
}

// borrow attempts to borrow a module from the locker, if one is not available
// it will return a new, temporary instance that will not be returned to the locker
// after use.
func (l *lensLocker) borrow() (enumerable.Socket[LensDoc], error) {
	select {
	case lens := <-l.safes:
		return &borrowedEnumerable{
			source: lens,
			locker: l,
		}, nil
	default:
		// If there are no free cached migrations within the locker, create a new temporary one
		// instead of blocking.
		return newLensPipe(l.cfg)
	}
}

// returnLens returns a borrowed module to the locker, allowing it to be reused by other contexts.
func (l *lensLocker) returnLens(lens *lensPipe) {
	l.safes <- lens
}

// borrowedEnumerable is an enumerable tied to a locker.
//
// it exposes the source enumerable and amends the Reset function so that when called, the source
// pipe is returned to the locker.
type borrowedEnumerable struct {
	source *lensPipe
	locker *lensLocker
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
	s.locker.returnLens(s.source)
	s.source.Reset()
}

// lensPipe provides a mechanic where the underlying wasm module can be hidden from consumers
// and allow input sources to be swapped in and out as different actors borrow it from the locker.
type lensPipe struct {
	input      enumerable.Socket[LensDoc]
	enumerable enumerable.Enumerable[LensDoc]
}

var _ enumerable.Socket[LensDoc] = (*lensPipe)(nil)

func newLensPipe(cfg client.LensConfig) (*lensPipe, error) {
	socket := enumerable.NewSocket[LensDoc]()
	enumerable, err := config.Load[LensDoc, LensDoc](cfg.Lens, socket)
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
