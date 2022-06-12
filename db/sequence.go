// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"encoding/binary"
	"errors"

	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/defradb/core"
)

type sequence struct {
	db  *db
	key core.SequenceKey
	val uint64
}

func (db *db) getSequence(ctx context.Context, key string) (*sequence, error) {
	if key == "" {
		return nil, errors.New("key cannot be empty")
	}
	seqKey := core.NewSequenceKey(key)
	seq := &sequence{
		db:  db,
		key: seqKey,
		val: uint64(0),
	}

	_, err := seq.get(ctx)
	if errors.Is(err, ds.ErrNotFound) {
		err = seq.update(ctx)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return seq, nil
}

func (seq *sequence) get(ctx context.Context) (uint64, error) {
	val, err := seq.db.systemstore().Get(ctx, seq.key.ToDS())
	if err != nil {
		return 0, err
	}
	num := binary.BigEndian.Uint64(val)
	seq.val = num
	return seq.val, nil
}

func (seq *sequence) update(ctx context.Context) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], seq.val)
	if err := seq.db.systemstore().Put(ctx, seq.key.ToDS(), buf[:]); err != nil {
		return err
	}

	return nil
}

func (seq *sequence) next(ctx context.Context) (uint64, error) {
	_, err := seq.get(ctx)
	if err != nil {
		return 0, err
	}

	seq.val++
	return seq.val, seq.update(ctx)
}
