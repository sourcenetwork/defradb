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

	ds "github.com/ipfs/go-datastore"

	"github.com/ipfs/go-datastore/query"
)

// DeserializePrefix deserializes all elements with the given prefix from the given storage.
// It returns the keys and their corresponding elements.
func DeserializePrefix[T any](
	ctx context.Context,
	prefix string,
	store DSReaderWriter,
) ([]string, []T, error) {
	q, err := store.Query(ctx, query.Query{Prefix: prefix})
	if err != nil {
		return nil, nil, err
	}

	keys := make([]string, 0)
	elements := make([]T, 0)
	for res := range q.Next() {
		if res.Error != nil {
			_ = q.Close()
			return nil, nil, res.Error
		}

		var element T
		err = json.Unmarshal(res.Value, &element)
		if err != nil {
			_ = q.Close()
			return nil, nil, NewErrInvalidStoredValue(err)
		}
		keys = append(keys, res.Key)
		elements = append(elements, element)
	}
	if err := q.Close(); err != nil {
		return nil, nil, err
	}
	return keys, elements, nil
}

// FetchKeysForPrefix fetches all keys with the given prefix from the given storage.
func FetchKeysForPrefix(
	ctx context.Context,
	prefix string,
	store ds.Read,
) ([]ds.Key, error) {
	q, err := store.Query(ctx, query.Query{Prefix: prefix})
	if err != nil {
		return nil, err
	}

	keys := make([]ds.Key, 0)
	for res := range q.Next() {
		if res.Error != nil {
			_ = q.Close()
			return nil, res.Error
		}
		keys = append(keys, ds.NewKey(res.Key))
	}
	if err = q.Close(); err != nil {
		return nil, err
	}

	return keys, nil
}
