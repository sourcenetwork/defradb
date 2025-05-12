// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"bytes"
	"context"
	"encoding/binary"
	"sort"

	cid "github.com/ipfs/go-cid"
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// heads manages the current Merkle-CRDT heads.
type heads struct {
	store     datastore.DSReaderWriter
	namespace keys.HeadstoreKey
}

func NewHeadSet(store datastore.DSReaderWriter, namespace keys.HeadstoreKey) *heads {
	return &heads{
		store:     store,
		namespace: namespace,
	}
}

func (hh *heads) key(c cid.Cid) keys.HeadstoreKey {
	return hh.namespace.WithCid(c)
}

func (hh *heads) Write(ctx context.Context, c cid.Cid, height uint64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, height)

	return hh.store.Set(ctx, hh.key(c).Bytes(), buf[0:n])
}

// IsHead returns if a given cid is among the current heads.
func (hh *heads) IsHead(ctx context.Context, c cid.Cid) (bool, error) {
	return hh.store.Has(ctx, hh.key(c).Bytes())
}

// Replace replaces a head with a new CID.
func (hh *heads) Replace(ctx context.Context, old cid.Cid, new cid.Cid, height uint64) error {
	log.InfoContext(
		ctx,
		"Replacing DAG head",
		corelog.Any("Old", old),
		corelog.Any("CID", new),
		corelog.Uint64("Height", height))

	err := hh.store.Delete(ctx, hh.key(old).Bytes())
	if err != nil {
		return err
	}

	err = hh.Write(ctx, new, height)
	if err != nil {
		return err
	}

	return nil
}

// List returns the list of current heads plus the max height.
// @todo Document Heads.List function
func (hh *heads) List(ctx context.Context) ([]cid.Cid, uint64, error) {
	iter, err := hh.store.Iterator(ctx, corekv.IterOptions{
		Prefix: hh.namespace.Bytes(),
	})
	if err != nil {
		return nil, 0, err
	}

	heads := make([]cid.Cid, 0)
	var maxHeight uint64
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, 0, errors.Join(NewErrFailedToGetNextQResult(err), iter.Close())
		}
		if !hasNext {
			break
		}

		headKey, err := keys.NewHeadstoreKey(string(iter.Key()))
		if err != nil {
			return nil, 0, errors.Join(err, iter.Close())
		}

		value, err := iter.Value()
		if err != nil {
			return nil, 0, errors.Join(err, iter.Close())
		}

		height, n := binary.Uvarint(value)
		if n <= 0 {
			return nil, 0, errors.Join(ErrDecodingHeight, iter.Close())
		}
		heads = append(heads, headKey.GetCid())
		if height > maxHeight {
			maxHeight = height
		}
	}
	sort.Slice(heads, func(i, j int) bool {
		ci := heads[i].Bytes()
		cj := heads[j].Bytes()
		return bytes.Compare(ci, cj) < 0
	})

	return heads, maxHeight, iter.Close()
}
