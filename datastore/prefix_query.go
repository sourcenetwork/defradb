// Copyright 2023 Democratized Data Foundation
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
	"context"
	"encoding/json"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/errors"
)

// DeserializePrefix deserializes all elements with the given prefix from the given storage.
// It returns the keys and their corresponding elements.
func DeserializePrefix[T any](
	ctx context.Context,
	prefix []byte,
	store corekv.Reader,
) ([][]byte, []T, error) {
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: prefix})
	if err != nil {
		return nil, nil, err
	}

	keys := make([][]byte, 0)
	elements := make([]T, 0)
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, nil, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		value, err := iter.Value()
		if err != nil {
			return nil, nil, errors.Join(err, iter.Close())
		}

		var element T
		err = json.Unmarshal(value, &element)
		if err != nil {
			return nil, nil, errors.Join(NewErrInvalidStoredValue(err), iter.Close())
		}
		keys = append(keys, iter.Key())
		elements = append(elements, element)
	}
	if err := iter.Close(); err != nil {
		return nil, nil, err
	}
	return keys, elements, nil
}

// FetchKeysForPrefix fetches all keys with the given prefix from the given storage.
func FetchKeysForPrefix(
	ctx context.Context,
	prefix []byte,
	store corekv.Reader,
) ([][]byte, error) {
	iter, err := store.Iterator(ctx, corekv.IterOptions{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	keys := make([][]byte, 0)
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		keys = append(keys, iter.Key())
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	return keys, nil
}
