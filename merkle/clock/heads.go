// Copyright 2020 Source Inc.
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
	"strings"

	"errors"

	"github.com/sourcenetwork/defradb/core"

	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
)

// heads manages the current Merkle-CRDT heads.
type heads struct {
	store     core.DSReaderWriter
	namespace ds.Key
}

// NewHeadSet
func NewHeadSet(store core.DSReaderWriter, namespace ds.Key) *heads {
	return newHeadset(store, namespace)
}

func newHeadset(store core.DSReaderWriter, namespace ds.Key) *heads {
	return &heads{
		store:     store,
		namespace: namespace,
	}
}

func (hh *heads) key(c cid.Cid) ds.Key {
	// /<namespace>/<cid>
	return hh.namespace.Child(dshelp.MultihashToDsKey(c.Hash()))
}

func (hh *heads) load(ctx context.Context, c cid.Cid) (uint64, error) {
	v, err := hh.store.Get(ctx, hh.key(c))
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
	return store.Put(ctx, hh.key(c), buf[0:n])
}

func (hh *heads) delete(ctx context.Context, store ds.Write, c cid.Cid) error {
	err := store.Delete(ctx, hh.key(c))
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
	log.Infof("replacing DAG head: %s -> %s (new height: %d)", h, c, height)
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
	log.Infof("adding new DAG head: %s (height: %d)", c, height)
	return hh.write(ctx, hh.store, c, height)
}

// List returns the list of current heads plus the max height.
// @todo Document Heads.List function
func (hh *heads) List(ctx context.Context) ([]cid.Cid, uint64, error) {
	q := query.Query{
		Prefix:   hh.namespace.String(),
		KeysOnly: false,
	}

	results, err := hh.store.Query(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err := results.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	heads := make([]cid.Cid, 0)
	var maxHeight uint64
	for r := range results.Next() {
		if r.Error != nil {
			return nil, 0, fmt.Errorf("Failed to get next query result : %w", r.Error)
		}
		// fmt.Println(r.Key, hh.namespace.String())
		headKey := ds.NewKey(strings.TrimPrefix(r.Key, hh.namespace.String()))
		hash, err := dshelp.DsKeyToMultihash(headKey)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to get CID from key : %w", err)
		}
		headCid := cid.NewCidV1(cid.Raw, hash)
		height, n := binary.Uvarint(r.Value)
		if n <= 0 {
			return nil, 0, errors.New("error decocding height")
		}
		heads = append(heads, headCid)
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
