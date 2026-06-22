// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build js

package js

import (
	"context"
	"syscall/js"

	"github.com/sourcenetwork/goji"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

var _ client.Collection = (*Collection)(nil)

type Collection struct {
	client js.Value
}

func (c *Collection) Name() string {
	promise := c.client.Call("name")
	res, err := goji.Await(goji.PromiseValue(promise))
	if err != nil {
		panic(err)
	}
	return res[0].String()
}

func (c *Collection) VersionID() string {
	promise := c.client.Call("versionID")
	res, err := goji.Await(goji.PromiseValue(promise))
	if err != nil {
		panic(err)
	}
	return res[0].String()
}

func (c *Collection) Version() client.CollectionVersion {
	promise := c.client.Call("version")
	res, err := goji.Await(goji.PromiseValue(promise))
	if err != nil {
		panic(err)
	}
	var out client.CollectionVersion
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		panic(err)
	}
	return out
}

func (c *Collection) CollectionID() string {
	promise := c.client.Call("collectionID")
	res, err := goji.Await(goji.PromiseValue(promise))
	if err != nil {
		panic(err)
	}
	return res[0].String()
}

func (c *Collection) NewIndex(
	ctx context.Context,
	indexDesc client.NewIndexRequest,
	opts ...options.Enumerable[options.NewCollectionIndexOptions],
) (client.IndexDescription, error) {
	res, err := execute(ctx, c.client, "newIndex", goji.MustMarshalJS(indexDesc), jsOpts(opts))
	if err != nil {
		return client.IndexDescription{}, err
	}
	var indexDescOut client.IndexDescription
	if err := goji.UnmarshalJS(res[0], &indexDescOut); err != nil {
		return client.IndexDescription{}, err
	}
	return indexDescOut, nil
}

func (c *Collection) DeleteIndex(
	ctx context.Context,
	indexName string,
	opts ...options.Enumerable[options.DeleteCollectionIndexOptions],
) error {
	_, err := execute(ctx, c.client, "deleteIndex", indexName, jsOpts(opts))
	return err
}

func (c *Collection) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListCollectionIndexesOptions],
) ([]client.ListIndexesResult, error) {
	res, err := execute(ctx, c.client, "listIndexes", jsOpts(opts))
	if err != nil {
		return nil, err
	}
	var out []client.ListIndexesResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Collection) NewEncryptedIndex(
	ctx context.Context,
	req client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.NewEncryptedIndexOptions],
) (client.EncryptedIndexDescription, error) {
	res, err := execute(ctx, c.client, "newEncryptedIndex", goji.MustMarshalJS(req), jsOpts(opts))
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	var indexDescOut client.EncryptedIndexDescription
	if err := goji.UnmarshalJS(res[0], &indexDescOut); err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	return indexDescOut, nil
}

func (c *Collection) ListEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListCollectionEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	res, err := execute(ctx, c.client, "listEncryptedIndexes", jsOpts(opts))
	if err != nil {
		return nil, err
	}
	var out []client.EncryptedIndexDescription
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Collection) DeleteEncryptedIndex(
	ctx context.Context,
	fieldName string,
	opts ...options.Enumerable[options.DeleteEncryptedIndexOptions],
) error {
	_, err := execute(ctx, c.client, "deleteEncryptedIndex", fieldName, jsOpts(opts))
	return err
}

func (c *Collection) Truncate(
	ctx context.Context,
	opts ...options.Enumerable[options.TruncateCollectionOptions],
) error {
	_, err := execute(ctx, c.client, "truncate", jsOpts(opts))
	return err
}

func (c *Collection) PurgeByDocIDs(
	ctx context.Context,
	docIDs []client.DocID,
	pruneHistory bool,
	opts ...options.Enumerable[options.TruncateCollectionOptions],
) error {
	rawIDs := make([]string, len(docIDs))
	for i, id := range docIDs {
		rawIDs[i] = id.String()
	}
	args := map[string]any{
		"docIDs":       rawIDs,
		"pruneHistory": pruneHistory,
	}
	_, err := execute(ctx, c.client, "purgeByDocIDs", args)
	return err
}
