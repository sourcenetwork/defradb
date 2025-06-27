// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package sequence

import (
	"context"
	"encoding/binary"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/txnctx"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type Sequence struct {
	key keys.Key
	val uint64
}

func Get(ctx context.Context, key keys.Key) (*Sequence, error) {
	seq := &Sequence{
		key: key,
		val: uint64(0),
	}

	_, err := seq.Get(ctx)
	if errors.Is(err, corekv.ErrNotFound) {
		err = seq.Update(ctx)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return seq, nil
}

func (seq *Sequence) Get(ctx context.Context) (uint64, error) {
	txn := txnctx.MustGet(ctx)

	val, err := txn.Systemstore().Get(ctx, seq.key.Bytes())
	if err != nil {
		return 0, err
	}
	num := binary.BigEndian.Uint64(val)
	seq.val = num
	return seq.val, nil
}

func (seq *Sequence) Update(ctx context.Context) error {
	txn := txnctx.MustGet(ctx)

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], seq.val)
	if err := txn.Systemstore().Set(ctx, seq.key.Bytes(), buf[:]); err != nil {
		return err
	}

	return nil
}

func (seq *Sequence) Next(ctx context.Context) (uint64, error) {
	_, err := seq.Get(ctx)
	if err != nil {
		return 0, err
	}

	seq.val++
	return seq.val, seq.Update(ctx)
}
