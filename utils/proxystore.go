// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package utils

import (
	"context"

	"github.com/ipfs/go-datastore"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

// Proxy store implements the ds.Datastore interface
// and provides proxy functionality from a 'frontend'
// datastore to one or more 'backend' datastores.
type ProxyStore struct {
	frontend ds.Datastore
	backends []ds.Datastore
}

// NewProxyStore returns a ds.Datastore implemented by a ProxyStore with
// the configured frontend and backends
func NewProxyStore(frontend ds.Datastore, backends ...ds.Datastore) ds.Datastore {
	return &ProxyStore{
		frontend: frontend,
		backends: backends,
	}
}

// Get retrieves the object `value` named by `key`.
// Get will return ErrNotFound if the key is not mapped to a value.
func (p *ProxyStore) Get(ctx context.Context, key datastore.Key) (value []byte, err error) {
	panic("not implemented") // TODO: Implement
}

// Has returns whether the `key` is mapped to a `value`.
// In some contexts, it may be much cheaper only to check for existence of
// a value, rather than retrieving the value itself. (e.g. HTTP HEAD).
// The default implementation is found in `GetBackedHas`.
func (p *ProxyStore) Has(ctx context.Context, key datastore.Key) (exists bool, err error) {
	panic("not implemented") // TODO: Implement
}

// GetSize returns the size of the `value` named by `key`.
// In some contexts, it may be much cheaper to only get the size of the
// value rather than retrieving the value itself.
func (p *ProxyStore) GetSize(ctx context.Context, key datastore.Key) (size int, err error) {
	panic("not implemented") // TODO: Implement
}

// Query searches the datastore and returns a query result. This function
// may return before the query actually runs. To wait for the query:
//
//   result, _ := ds.Query(q)
//
//   // use the channel interface; result may come in at different times
//   for entry := range result.Next() { ... }
//
//   // or wait for the query to be completely done
//   entries, _ := result.Rest()
//   for entry := range entries { ... }
//
func (p *ProxyStore) Query(ctx context.Context, q query.Query) (query.Results, error) {
	panic("not implemented") // TODO: Implement
}

// Put stores the object `value` named by `key`.
//
// The generalized Datastore interface does not impose a value type,
// allowing various datastore middleware implementations (which do not
// handle the values directly) to be composed together.
//
// Ultimately, the lowest-level datastore will need to do some value checking
// or risk getting incorrect values. It may also be useful to expose a more
// type-safe interface to your application, and do the checking up-front.
func (p *ProxyStore) Put(ctx context.Context, key datastore.Key, value []byte) error {
	panic("not implemented") // TODO: Implement
}

// Delete removes the value for given `key`. If the key is not in the
// datastore, this method returns no error.
func (p *ProxyStore) Delete(ctx context.Context, key datastore.Key) error {
	panic("not implemented") // TODO: Implement
}

// Sync guarantees that any Put or Delete calls under prefix that returned
// before Sync(prefix) was called will be observed after Sync(prefix)
// returns, even if the program crashes. If Put/Delete operations already
// satisfy these requirements then Sync may be a no-op.
//
// If the prefix fails to Sync this method returns an error.
func (p *ProxyStore) Sync(ctx context.Context, prefix datastore.Key) error {
	panic("not implemented") // TODO: Implement
}

func (p *ProxyStore) Close() error {
	panic("not implemented") // TODO: Implement
}
