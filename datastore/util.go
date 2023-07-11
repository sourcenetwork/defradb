package datastore

import (
	"context"
	"encoding/json"

	ds "github.com/ipfs/go-datastore"

	"github.com/ipfs/go-datastore/query"
)

func DeserializePrefix[T any](
	ctx context.Context,
	prefix string,
	storage ds.Read,
) ([]string, []T, error) {
	q, err := storage.Query(ctx, query.Query{Prefix: prefix})
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
			return nil, nil, err
		}
		keys = append(keys, res.Key)
		elements = append(elements, element)
	}
	if err := q.Close(); err != nil {
		return nil, nil, err
	}
	return keys, elements, nil
}

func FetchKeysForPrefix(
	ctx context.Context,
	prefix string,
	storage ds.Read,
) ([]ds.Key, error) {
	q, err := storage.Query(ctx, query.Query{Prefix: prefix})
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
