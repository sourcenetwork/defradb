// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/namespace"
)

type datastore struct {
	underlying corekv.ReaderWriter
}

var _ Keyedstore = (*datastore)(nil)

func newDatastore(rootstore corekv.ReaderWriter) *datastore {
	return &datastore{
		underlying: namespace.Wrap(rootstore, []byte{dataStoreKey}),
	}
}

func (s *datastore) Get(ctx context.Context, key Key) ([]byte, error) {
	keyBytes := key.Bytes()
	return s.underlying.Get(ctx, keyBytes)
}

func (s *datastore) Has(ctx context.Context, key Key) (bool, error) {
	keyBytes := key.Bytes()
	return s.underlying.Has(ctx, keyBytes)
}

func (s *datastore) Iterator(ctx context.Context, opts IterOptions) (corekv.Iterator, error) {
	var prefix []byte
	var start []byte
	var end []byte

	if opts.Prefix != nil {
		prefix = opts.Prefix.Bytes()
	}
	if opts.Start != nil {
		start = opts.Start.Bytes()
	}
	if opts.End != nil {
		end = opts.End.Bytes()
	}

	ckvOpts := corekv.IterOptions{
		Prefix:   prefix,
		Start:    start,
		End:      end,
		KeysOnly: opts.KeysOnly,
		Reverse:  opts.Reverse,
	}
	return s.underlying.Iterator(ctx, ckvOpts)
}

func (s *datastore) Set(ctx context.Context, key Key, value []byte) error {
	keyBytes := key.Bytes()
	return s.underlying.Set(ctx, keyBytes, value)
}

func (s *datastore) Delete(ctx context.Context, key Key) error {
	keyBytes := key.Bytes()
	return s.underlying.Delete(ctx, keyBytes)
}
