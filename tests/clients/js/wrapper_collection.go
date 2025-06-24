// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build js

package js

import (
	"context"
	"syscall/js"

	"github.com/sourcenetwork/goji"

	"github.com/sourcenetwork/defradb/client"
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

func (c *Collection) Schema() client.SchemaDescription {
	promise := c.client.Call("schema")
	res, err := goji.Await(goji.PromiseValue(promise))
	if err != nil {
		panic(err)
	}
	var out client.SchemaDescription
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		panic(err)
	}
	return out
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

func (c *Collection) SchemaRoot() string {
	promise := c.client.Call("schemaRoot")
	res, err := goji.Await(goji.PromiseValue(promise))
	if err != nil {
		panic(err)
	}
	return res[0].String()
}

func (c *Collection) Definition() client.CollectionDefinition {
	promise := c.client.Call("definition")
	res, err := goji.Await(goji.PromiseValue(promise))
	if err != nil {
		panic(err)
	}
	var out client.CollectionDefinition
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		panic(err)
	}
	return out
}

func (c *Collection) Create(
	ctx context.Context,
	doc *client.Document,
	opts ...client.DocCreateOption,
) error {
	docVal, err := goji.MarshalJS(doc)
	if err != nil {
		return err
	}
	_, err = execute(ctx, c.client, "create", docVal, makeDocCreateOptions(opts))
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func makeDocCreateOptions(opts []client.DocCreateOption) js.Value {
	createOpts := client.DocCreateOptions{}
	createOpts.Apply(opts)

	optsVal, err := goji.MarshalJS(createOpts)
	if err != nil {
		return js.Undefined()
	}
	return optsVal
}

func (c *Collection) CreateMany(
	ctx context.Context,
	docs []*client.Document,
	opts ...client.DocCreateOption,
) error {
	docsVal, err := goji.MarshalJS(docs)
	if err != nil {
		return err
	}
	_, err = execute(ctx, c.client, "createMany", docsVal, makeDocCreateOptions(opts))
	if err != nil {
		return err
	}
	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

func (c *Collection) Update(
	ctx context.Context,
	doc *client.Document,
) error {
	patch, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	docID := doc.ID().String()
	_, err = execute(ctx, c.client, "update", docID, string(patch))
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) Save(
	ctx context.Context,
	doc *client.Document,
	opts ...client.DocCreateOption,
) error {
	_, err := c.Get(ctx, doc.ID(), true)
	if err == nil {
		return c.Update(ctx, doc)
	}
	if err.Error() == client.ErrDocumentNotFoundOrNotAuthorized.Error() {
		return c.Create(ctx, doc, opts...)
	}
	return err
}

func (c *Collection) Delete(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	res, err := execute(ctx, c.client, "delete", docID.String())
	if err != nil {
		return false, err
	}
	return res[0].Bool(), nil
}

func (c *Collection) Exists(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	res, err := execute(ctx, c.client, "exists", docID.String())
	if err != nil {
		return false, err
	}
	return res[0].Bool(), nil
}

func (c *Collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	res, err := execute(ctx, c.client, "updateWithFilter", filter, updater)
	if err != nil {
		return nil, err
	}
	var out client.UpdateResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Collection) DeleteWithFilter(
	ctx context.Context,
	filter any,
) (*client.DeleteResult, error) {
	res, err := execute(ctx, c.client, "deleteWithFilter", filter)
	if err != nil {
		return nil, err
	}
	var out client.DeleteResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Collection) Get(
	ctx context.Context,
	docID client.DocID,
	showDeleted bool,
) (*client.Document, error) {
	res, err := execute(ctx, c.client, "get", docID.String(), showDeleted)
	if err != nil {
		return nil, err
	}
	var docMap map[string]any
	if err := goji.UnmarshalJS(res[0], &docMap); err != nil {
		return nil, err
	}
	doc, err := client.NewDocWithID(docID, c.Definition())
	if err != nil {
		return nil, err
	}
	for f, v := range docMap {
		if err := doc.Set(f, v); err != nil {
			return nil, err
		}
	}
	doc.Clean()
	return doc, nil
}

func (c *Collection) GetAllDocIDs(
	ctx context.Context,
) (<-chan client.DocIDResult, error) {
	panic("not implemented")
}

func (c *Collection) CreateIndex(
	ctx context.Context,
	indexDesc client.IndexCreateRequest,
) (client.IndexDescription, error) {
	indexDescVal, err := goji.MarshalJS(indexDesc)
	if err != nil {
		return client.IndexDescription{}, err
	}
	res, err := execute(ctx, c.client, "createIndex", indexDescVal)
	if err != nil {
		return client.IndexDescription{}, err
	}
	var indexDescOut client.IndexDescription
	if err := goji.UnmarshalJS(res[0], &indexDescOut); err != nil {
		return client.IndexDescription{}, err
	}
	return indexDescOut, nil
}

func (c *Collection) DropIndex(ctx context.Context, indexName string) error {
	_, err := execute(ctx, c.client, "dropIndex", indexName)
	return err
}

func (c *Collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	res, err := execute(ctx, c.client, "getIndexes")
	if err != nil {
		return nil, err
	}
	var out []client.IndexDescription
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Collection) CreateEncryptedIndex(
	ctx context.Context,
	req client.EncryptedIndexCreateRequest,
) (client.EncryptedIndexDescription, error) {
	// TODO: implement
	return client.EncryptedIndexDescription{}, nil
}

func (c *Collection) GetEncryptedIndexes(ctx context.Context) ([]client.EncryptedIndexDescription, error) {
	// TODO: implement
	return nil, nil
}
