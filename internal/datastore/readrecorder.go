// Copyright 2026 Democratized Data Foundation
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
	"bytes"
	"context"

	"github.com/sourcenetwork/corekv"
)

// sequencePrefix is the systemstore prefix under which all sequences live.
var sequencePrefix = []byte("/seq/")

// readRecorder sits between a transaction's namespaced stores and the raw transaction,
// classifying every key the transaction reads by the store it belongs to.
//
// Writes pass through unclassified: the underlying store's conflict detection compares a
// transaction's reads against other transactions' writes, so only reads decide whether a
// commit conflicts.
//
// The recorder must not be handed to corekv as a context transaction. corekv casts the
// context value to its own concrete transaction type, and a wrapper fails that cast.
type readRecorder struct {
	store corekv.ReaderWriter
	kinds *ReadKinds
}

var _ corekv.ReaderWriter = (*readRecorder)(nil)

// kindForKey classifies a rootstore key by its leading store discriminator byte, which
// namespace.Wrap has already prepended by the time the key reaches here.
func kindForKey(key []byte) ReadKind {
	if len(key) == 0 {
		return 0
	}
	switch key[0] {
	case dataStoreKey:
		// Index entries are also datastore keys and encode identically to document keys.
		// The index write path marks those itself; everything else here is a document.
		return ReadDoc
	case headStoreKey:
		return ReadHead
	case blockStoreKey:
		return ReadBlock
	case systemStoreKey:
		if bytes.HasPrefix(key[1:], sequencePrefix) {
			return ReadSequence
		}
		return ReadSystem
	case peerStoreKey:
		return ReadPeer
	case encStoreKey:
		return ReadEnc
	default:
		return 0
	}
}

func (r *readRecorder) Get(ctx context.Context, key []byte) ([]byte, error) {
	r.kinds.Mark(kindForKey(key))
	return r.store.Get(ctx, key)
}

func (r *readRecorder) Has(ctx context.Context, key []byte) (bool, error) {
	r.kinds.Mark(kindForKey(key))
	return r.store.Has(ctx, key)
}

// Iterator records the shape of the iterator as well as what it scans. A start/end
// iterator is the shape whose bounds check reads one key past the requested range, so
// separating the two shapes is what makes that over-read visible.
func (r *readRecorder) Iterator(ctx context.Context, opts corekv.IterOptions) (corekv.Iterator, error) {
	switch {
	case opts.Prefix != nil:
		r.kinds.Mark(ReadPrefixIterator | kindForKey(opts.Prefix))
	case opts.Start != nil:
		r.kinds.Mark(ReadRangeIterator | kindForKey(opts.Start))
	case opts.End != nil:
		r.kinds.Mark(ReadRangeIterator | kindForKey(opts.End))
	}
	return r.store.Iterator(ctx, opts)
}

func (r *readRecorder) Set(ctx context.Context, key, value []byte) error {
	return r.store.Set(ctx, key, value)
}

func (r *readRecorder) Delete(ctx context.Context, key []byte) error {
	return r.store.Delete(ctx, key)
}
