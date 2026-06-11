// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"context"
	"encoding/binary"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// MergeOptions controls storage behavior while applying a CRDT delta.
type MergeOptions struct {
	NewDocCreateMode bool
}

// MergeOption configures CRDT merge behavior.
type MergeOption func(*MergeOptions)

// WithNewDocCreateMode marks a merge as applying the first composite for a new document.
func WithNewDocCreateMode() MergeOption {
	return func(opts *MergeOptions) {
		opts.NewDocCreateMode = true
	}
}

// NewMergeOptions returns the effective merge options.
func NewMergeOptions(options ...MergeOption) MergeOptions {
	var result MergeOptions
	for _, option := range options {
		option(&result)
	}
	return result
}

func setPriority(
	ctx context.Context,
	store datastore.Keyedstore,
	key keys.DataStoreKey,
	priority uint64,
) error {
	prioK := key.WithPriorityFlag()
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, priority)
	if n == 0 {
		return ErrEncodingPriority
	}

	return store.Set(ctx, prioK, buf[0:n])
}

// get the current priority for given key
func getPriority(
	ctx context.Context,
	store datastore.Keyedstore,
	key keys.DataStoreKey,
) (uint64, error) {
	pKey := key.WithPriorityFlag()
	pbuf, err := store.Get(ctx, pKey)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	prio, num := binary.Uvarint(pbuf)
	if num <= 0 {
		return 0, ErrDecodingPriority
	}
	return prio, nil
}
