// Copyright 2024 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/internal/keys"
)

type CollectionDelta struct {
	Priority            uint64
	CollectionVersionID string
}

var _ Delta = (*CollectionDelta)(nil)

func (d *CollectionDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type CollectionDelta struct {
		priority  			Int
		collectionVersionID String
	}`)
}

func (d *CollectionDelta) GetPriority() uint64 {
	return d.Priority
}

func (d *CollectionDelta) SetPriority(priority uint64) {
	d.Priority = priority
}

type Collection struct {
	headstorePrefix     keys.HeadstoreKey
	collectionVersionID string
}

var _ ReplicatedData = (*Collection)(nil)

func NewCollection(
	collectionVersionID string,
	key keys.HeadstoreColKey,
) *Collection {
	return &Collection{
		collectionVersionID: collectionVersionID,
		headstorePrefix:     key,
	}
}

func (c *Collection) HeadstorePrefix() keys.HeadstoreKey {
	return c.headstorePrefix
}

func (c *Collection) Delta() *CollectionDelta {
	return &CollectionDelta{
		CollectionVersionID: c.collectionVersionID,
	}
}

func (c *Collection) Merge(ctx context.Context, other Delta) error {
	// Collection merges don't actually need to do anything, as the delta is empty,
	// and doc-level merges are handled by the document commits.
	return nil
}
