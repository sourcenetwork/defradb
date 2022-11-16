// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package memory

import (
	"context"
	"sync"

	ds "github.com/ipfs/go-datastore"
)

type op struct {
	delete bool
	value  []byte
}

// basicBatch implements the transaction interface for datastores who do
// not have any sort of underlying transactional support
type basicBatch struct {
	syncLock sync.Mutex
	ops      map[ds.Key]op
	target   *Store
}

var _ ds.Batch = (*basicBatch)(nil)

func NewBasicBatch(d *Store) ds.Batch {
	return &basicBatch{
		ops:    make(map[ds.Key]op),
		target: d,
	}
}

func (b *basicBatch) Put(ctx context.Context, key ds.Key, val []byte) error {
	b.syncLock.Lock()
	defer b.syncLock.Unlock()

	b.ops[key] = op{value: val}
	return nil
}

func (b *basicBatch) Delete(ctx context.Context, key ds.Key) error {
	b.syncLock.Lock()
	defer b.syncLock.Unlock()

	b.ops[key] = op{delete: true}
	return nil
}

func (b *basicBatch) Commit(ctx context.Context) error {
	b.syncLock.Lock()
	defer b.syncLock.Unlock()
	b.target.syncLock.Lock()
	defer b.target.syncLock.Unlock()

	for k, op := range b.ops {
		if op.delete {
			delete(b.target.values, k)
		} else {
			b.target.values[k] = op.value
		}
	}

	return nil
}
