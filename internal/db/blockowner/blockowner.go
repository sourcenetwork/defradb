// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package blockowner reads the systemstore index recording which documents own a
// block. It takes an explicit store rather than a transaction from the context, so
// that layers below the db package can consult ownership too.
package blockowner

import (
	"bytes"
	"context"
	stderrors "errors"

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/internal/keys"
)

// prefix is the key prefix under which every owner DocID of blockCID is stored.
func prefix(blockCID cid.Cid) []byte {
	return append(keys.NewBlockCIDToDocIDKey(blockCID.String(), "").Bytes(), '/')
}

// DocIDs returns every DocID that owns blockCID.
// Field blocks can be byte-identical across documents, so ownership is a set.
// The index stores DocIDs directly so block ownership survives doc-ref cleanup.
func DocIDs(ctx context.Context, store corekv.Reader, blockCID cid.Cid) ([]string, error) {
	if !blockCID.Defined() {
		return nil, nil
	}

	p := prefix(blockCID)
	iter, err := store.Iterator(ctx, corekv.IterOptions{
		Prefix:   p,
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	var docIDs []string
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, stderrors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}
		docID := bytes.TrimPrefix(iter.Key(), p)
		if len(docID) != 0 {
			docIDs = append(docIDs, string(docID))
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return docIDs, nil
}

// HasOwners reports whether any document still owns blockCID. Prefer this over DocIDs
// when only the presence of an owner matters: it stops at the first match rather than
// materializing the whole owner set.
func HasOwners(ctx context.Context, store corekv.Reader, blockCID cid.Cid) (bool, error) {
	return HasOwnersExcept(ctx, store, blockCID, nil)
}

// HasOwnersExcept reports whether any document owns blockCID, ignoring owner edges whose keys
// are in excluded. It lets a caller read ownership from a read-only snapshot while treating edges
// it deleted in a separate, uncommitted transaction as already gone. Excluded keys are
// BlockCIDToDocIDKey bytes, as yielded by the owner-edge iterator; a nil set makes this a plain
// ownership scan. It stops at the first surviving owner.
func HasOwnersExcept(
	ctx context.Context,
	store corekv.Reader,
	blockCID cid.Cid,
	excluded map[string]struct{},
) (bool, error) {
	if !blockCID.Defined() {
		return false, nil
	}

	p := prefix(blockCID)
	iter, err := store.Iterator(ctx, corekv.IterOptions{
		Prefix:   p,
		KeysOnly: true,
	})
	if err != nil {
		return false, err
	}

	for {
		hasNext, err := iter.Next()
		if err != nil {
			return false, stderrors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}
		if len(bytes.TrimPrefix(iter.Key(), p)) == 0 {
			continue
		}
		if _, isExcluded := excluded[string(iter.Key())]; isExcluded {
			continue
		}
		return true, iter.Close()
	}
	return false, iter.Close()
}
