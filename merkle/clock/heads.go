// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package clock

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sort"

	"errors"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/logging"

	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

// heads manages the current Merkle-CRDT heads.
type heads struct {
	store     client.DSReaderWriter
	namespace core.HeadStoreKey
}

func NewHeadSet(store client.DSReaderWriter, namespace core.HeadStoreKey) *heads {
	return newHeadset(store, namespace)
}

func newHeadset(store client.DSReaderWriter, namespace core.HeadStoreKey) *heads {
	return &heads{
		store:     store,
		namespace: namespace,
	}
}

func (hh *heads) key(c cid.Cid) core.HeadStoreKey {
	// /<namespace>/<cid>
	return hh.namespace.WithCid(c)
}

func (hh *heads) load(ctx context.Context, c cid.Cid) (uint64, error) {
	v, err := hh.store.Get(ctx, hh.key(c).ToDS())
	if err != nil {
		return 0, err
	}
	height, n := binary.Uvarint(v)
	if n <= 0 {
		return 0, errors.New("error decoding height")
	}
	return height, nil
}

func (hh *heads) write(ctx context.Context, store ds.Write, c cid.Cid, height uint64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, height)
	if n == 0 {
		return errors.New("error encoding height")
	}
	return store.Put(ctx, hh.key(c).ToDS(), buf[0:n])
}

func (hh *heads) delete(ctx context.Context, store ds.Write, c cid.Cid) error {
	err := store.Delete(ctx, hh.key(c).ToDS())
	if err == ds.ErrNotFound {
		return nil
	}
	return err
}

// IsHead returns if a given cid is among the current heads.
func (hh *heads) IsHead(ctx context.Context, c cid.Cid) (bool, uint64, error) {
	height, err := hh.load(ctx, c)
	if err == ds.ErrNotFound {
		return false, 0, nil
	}
	return err == nil, height, err
}

func (hh *heads) Len(ctx context.Context) (int, error) {
	list, _, err := hh.List(ctx)
	return len(list), err
}

// Replace replaces a head with a new cid.
func (hh *heads) Replace(ctx context.Context, h, c cid.Cid, height uint64) error {
	log.Info(
		ctx,
		"Replacing DAG head",
		logging.NewKV("Old", h),
		logging.NewKV("Cid", c),
		logging.NewKV("Height", height))
	var store ds.Write = hh.store
	var err error

	// batchingDs, batching := store.(ds.Batching)
	// if batching {
	// 	store, err = batchingDs.Batch()
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	err = hh.delete(ctx, store, h)
	if err != nil {
		return err
	}

	err = hh.write(ctx, store, c, height)
	if err != nil {
		return err
	}

	// if batching {
	// 	err := store.(ds.Batch).Commit()
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func (hh *heads) Add(ctx context.Context, c cid.Cid, height uint64) error {
	log.Info(ctx, "Adding new DAG head",
		logging.NewKV("Cid", c),
		logging.NewKV("Height", height))
	return hh.write(ctx, hh.store, c, height)
}

// List returns the list of current heads plus the max height.
// @todo Document Heads.List function
func (hh *heads) List(ctx context.Context) ([]cid.Cid, uint64, error) {
	q := query.Query{
		Prefix:   hh.namespace.ToString(),
		KeysOnly: false,
	}

	results, err := hh.store.Query(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err := results.Close()
		if err != nil {
			log.ErrorE(ctx, "Error closing results", err)
		}
	}()

	heads := make([]cid.Cid, 0)
	var maxHeight uint64
	for r := range results.Next() {
		if r.Error != nil {
			return nil, 0, fmt.Errorf("Failed to get next query result : %w", r.Error)
		}

		headKey, err := core.NewHeadStoreKey(r.Key)
		if err != nil {
			return nil, 0, err
		}

		height, n := binary.Uvarint(r.Value)
		if n <= 0 {
			return nil, 0, errors.New("error decocding height")
		}
		heads = append(heads, headKey.Cid)
		if height > maxHeight {
			maxHeight = height
		}
	}
	sort.Slice(heads, func(i, j int) bool {
		ci := heads[i].Bytes()
		cj := heads[j].Bytes()
		return bytes.Compare(ci, cj) < 0
	})

	return heads, maxHeight, nil
}
