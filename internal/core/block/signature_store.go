// Copyright 2026 Democratized Data Foundation
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
	"context"
	"sort"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// StoredSignatureLink describes a signature block linked to a canonical block.
type StoredSignatureLink struct {
	Identity string
	Link     cidlink.Link
}

func StoreSignatureLink(
	ctx context.Context,
	store corekv.ReaderWriter,
	blockCID cid.Cid,
	sig *Signature,
	sigLink cidlink.Link,
) error {
	key := keys.NewBlockSignatureKey(blockCID, string(sig.Header.Identity))
	return store.Set(ctx, key.Bytes(), sigLink.Cid.Bytes())
}

func GetSignatureLink(
	ctx context.Context,
	store corekv.Reader,
	blockCID cid.Cid,
	identity string,
) (cidlink.Link, bool, error) {
	key := keys.NewBlockSignatureKey(blockCID, identity)
	sigCIDBytes, err := store.Get(ctx, key.Bytes())
	if errors.Is(err, corekv.ErrNotFound) {
		return cidlink.Link{}, false, nil
	}

	if err != nil {
		return cidlink.Link{}, false, err
	}

	sigCID, err := cid.Cast(sigCIDBytes)
	if err != nil {
		return cidlink.Link{}, false, err
	}

	return cidlink.Link{Cid: sigCID}, true, nil
}

func ListSignatureLinks(
	ctx context.Context,
	store corekv.Reader,
	blockCID cid.Cid,
) ([]StoredSignatureLink, error) {
	prefix := keys.NewBlockSignaturePrefix(blockCID)
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: prefix.Bytes()})
	if err != nil {
		return nil, err
	}

	var result []StoredSignatureLink
	for {
		next, err := iter.Next()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		if !next {
			break
		}

		key, err := keys.NewBlockSignatureKeyFromString(string(iter.Key()))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		value, err := iter.Value()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		sigCID, err := cid.Cast(value)
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		result = append(result, StoredSignatureLink{
			Identity: key.Identity,
			Link:     cidlink.Link{Cid: sigCID},
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Identity < result[j].Identity
	})

	return result, iter.Close()
}

func DeleteSignatureLinks(ctx context.Context, store corekv.ReaderWriter, blockCID cid.Cid) ([]cidlink.Link, error) {
	signatures, err := ListSignatureLinks(ctx, store, blockCID)
	if err != nil {
		return nil, err
	}

	links := make([]cidlink.Link, len(signatures))
	for i, sig := range signatures {
		links[i] = sig.Link
		key := keys.NewBlockSignatureKey(blockCID, sig.Identity)
		if err := store.Delete(ctx, key.Bytes()); err != nil {
			return nil, err
		}
	}

	return links, nil
}
