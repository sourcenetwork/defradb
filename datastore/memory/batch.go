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
	"sync/atomic"

	ds "github.com/ipfs/go-datastore"
)

type op struct {
	delete bool
	value  []byte
}

// basicBatch implements ds.Batch
type basicBatch struct {
	mu  sync.Mutex
	ops map[ds.Key]op
	ds  *Datastore
}

var _ ds.Batch = (*basicBatch)(nil)

// Put implements ds.Put
func (b *basicBatch) Put(ctx context.Context, key ds.Key, val []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.ops[key] = op{value: val}
	return nil
}

// Delete implements ds.Delete
func (b *basicBatch) Delete(ctx context.Context, key ds.Key) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.ops[key] = op{delete: true}
	return nil
}

// Commit saves the operations to the underlying datastore
func (b *basicBatch) Commit(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	v := atomic.AddUint64(b.ds.version, 1)
	for k, op := range b.ops {
		if op.delete {
			b.ds.values.Set(item{key: k.String(), version: v, isDeleted: true})
		} else {
			b.ds.values.Set(item{key: k.String(), version: v, val: op.value})
		}
	}

	return nil
}
