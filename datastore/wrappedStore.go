// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"

	ds "github.com/ipfs/go-datastore"
	ktds "github.com/ipfs/go-datastore/keytransform"
	"github.com/ipfs/go-datastore/query"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/defradb/datastore/iterable"
)

type wrappedStore struct {
	transform ktds.KeyTransform
	store     DSReaderWriter
}

var _ DSReaderWriter = (*wrappedStore)(nil)

func prefix(root DSReaderWriter, prefix ds.Key) DSReaderWriter {
	return &wrappedStore{
		transform: ktds.PrefixTransform{Prefix: prefix},
		store:     root,
	}
}

func (w *wrappedStore) Get(ctx context.Context, key ds.Key) (value []byte, err error) {
	return w.store.Get(ctx, w.transform.ConvertKey(key))
}

func (w *wrappedStore) Has(ctx context.Context, key ds.Key) (exists bool, err error) {
	return w.store.Has(ctx, w.transform.ConvertKey(key))
}

func (w *wrappedStore) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	return w.store.GetSize(ctx, w.transform.ConvertKey(key))
}

func (w *wrappedStore) Put(ctx context.Context, key ds.Key, value []byte) error {
	return w.store.Put(ctx, w.transform.ConvertKey(key), value)
}

func (w *wrappedStore) Delete(ctx context.Context, key ds.Key) error {
	return w.store.Delete(ctx, w.transform.ConvertKey(key))
}

func (w *wrappedStore) GetIterator(q query.Query) (iterable.Iterator, error) {
	iterator, err := w.store.GetIterator(
		withPrefix(q, w.transform.ConvertKey(ds.NewKey(q.Prefix)).String()),
	)
	if err != nil {
		return nil, err
	}
	return &wrappedIterator{transform: w.transform, iterator: iterator}, nil
}

func withPrefix(q query.Query, prefix string) query.Query {
	return query.Query{
		Prefix:            prefix,
		Filters:           q.Filters,
		Orders:            q.Orders,
		Limit:             q.Limit,
		Offset:            q.Offset,
		KeysOnly:          q.KeysOnly,
		ReturnExpirations: q.ReturnExpirations,
		ReturnsSizes:      q.ReturnsSizes,
	}
}

// NOTE!!! The following lines are copied from the ktds package, they should be unessecary after
// the key refactoring (as the keys will then not contain the prefixes)

// Query implements Query, inverting keys on the way back out.
func (w *wrappedStore) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	nq, cq := w.prepareQuery(q)

	cqr, err := w.store.Query(ctx, cq)
	if err != nil {
		return nil, err
	}

	qr := dsq.ResultsFromIterator(q, dsq.Iterator{
		Next: func() (dsq.Result, bool) {
			r, ok := cqr.NextSync()
			if !ok {
				return r, false
			}
			if r.Error == nil {
				r.Entry.Key = w.transform.InvertKey(ds.RawKey(r.Entry.Key)).String()
			}
			return r, true
		},
		Close: func() error {
			return cqr.Close()
		},
	})
	return dsq.NaiveQueryApply(nq, qr), nil
}

// Split the query into a child query and a naive query. That way, we can make
// the child datastore do as much work as possible.
func (w *wrappedStore) prepareQuery(q dsq.Query) (naive, child dsq.Query) {

	// First, put everything in the child query. Then, start taking things
	// out.
	child = q

	// Always let the child handle the key prefix.
	child.Prefix = w.transform.ConvertKey(ds.NewKey(child.Prefix)).String()

	// Check if the key transform is order-preserving so we can use the
	// child datastore's built-in ordering.
	orderPreserving := false
	switch w.transform.(type) {
	case ktds.PrefixTransform, *ktds.PrefixTransform:
		orderPreserving = true
	}

	// Try to let the child handle ordering.
orders:
	for i, o := range child.Orders {
		switch o.(type) {
		case dsq.OrderByValue, *dsq.OrderByValue,
			dsq.OrderByValueDescending, *dsq.OrderByValueDescending:
			// Key doesn't matter.
			continue
		case dsq.OrderByKey, *dsq.OrderByKey,
			dsq.OrderByKeyDescending, *dsq.OrderByKeyDescending:
			// if the key transform preserves order, we can delegate
			// to the child datastore.
			if orderPreserving {
				// When sorting, we compare with the first
				// Order, then, if equal, we compare with the
				// second Order, etc. However, keys are _unique_
				// so we'll never apply any additional orders
				// after ordering by key.
				child.Orders = child.Orders[:i+1]
				break orders
			}
		}

		// Can't handle this order under transform, punt it to a naive
		// ordering.
		naive.Orders = q.Orders
		child.Orders = nil
		naive.Offset = q.Offset
		child.Offset = 0
		naive.Limit = q.Limit
		child.Limit = 0
		break
	}

	// Try to let the child handle the filters.

	// don't modify the original filters.
	child.Filters = append([]dsq.Filter(nil), child.Filters...)

	for i, f := range child.Filters {
		switch f := f.(type) {
		case dsq.FilterValueCompare, *dsq.FilterValueCompare:
			continue
		case dsq.FilterKeyCompare:
			child.Filters[i] = dsq.FilterKeyCompare{
				Op:  f.Op,
				Key: w.transform.ConvertKey(ds.NewKey(f.Key)).String(),
			}
			continue
		case *dsq.FilterKeyCompare:
			child.Filters[i] = &dsq.FilterKeyCompare{
				Op:  f.Op,
				Key: w.transform.ConvertKey(ds.NewKey(f.Key)).String(),
			}
			continue
		case dsq.FilterKeyPrefix:
			child.Filters[i] = dsq.FilterKeyPrefix{
				Prefix: w.transform.ConvertKey(ds.NewKey(f.Prefix)).String(),
			}
			continue
		case *dsq.FilterKeyPrefix:
			child.Filters[i] = &dsq.FilterKeyPrefix{
				Prefix: w.transform.ConvertKey(ds.NewKey(f.Prefix)).String(),
			}
			continue
		}

		// Not a known filter, defer to the naive implementation.
		naive.Filters = q.Filters
		child.Filters = nil
		naive.Offset = q.Offset
		child.Offset = 0
		naive.Limit = q.Limit
		child.Limit = 0
		break
	}
	return
}

type wrappedIterator struct {
	transform ktds.KeyTransform
	iterator  iterable.Iterator
}

var _ iterable.Iterator = (*wrappedIterator)(nil)

func (w *wrappedIterator) IteratePrefix(
	ctx context.Context,
	startPrefix ds.Key,
	endPrefix ds.Key,
) (dsq.Results, error) {
	return w.iterator.IteratePrefix(
		ctx,
		w.transform.ConvertKey(startPrefix),
		w.transform.ConvertKey(endPrefix),
	)
}

func (w *wrappedIterator) Close() error {
	return w.iterator.Close()
}
