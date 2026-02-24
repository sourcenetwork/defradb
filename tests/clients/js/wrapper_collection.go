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
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/utils"
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

func (c *Collection) Add(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.CollectionAddOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	docVal, err := goji.MarshalJS(doc)
	if err != nil {
		return err
	}
	_, err = execute(ctx, c.client, "add", docVal, makeDocAddOptions(opts))
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

// addOptionsJS is used to marshal options for the JS client.
type addOptionsJS struct {
	EncryptDoc      bool     `json:"encryptDoc"`
	EncryptedFields []string `json:"encryptedFields"`
}

func makeDocAddOptions(opts []options.Enumerable[options.CollectionAddOptions]) js.Value {
	jsOpts := addOptionsJS{}
	addOpts := utils.NewOptions(opts...)
	jsOpts.EncryptDoc = addOpts.EncryptDoc
	jsOpts.EncryptedFields = addOpts.EncryptedFields

	optsVal, err := goji.MarshalJS(jsOpts)
	if err != nil {
		return js.Undefined()
	}
	return optsVal
}

func (c *Collection) AddMany(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Enumerable[options.CollectionAddOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	docsVal, err := goji.MarshalJS(docs)
	if err != nil {
		return err
	}
	_, err = execute(ctx, c.client, "addMany", docsVal, makeDocAddOptions(opts))
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
	opts ...options.Enumerable[options.CollectionUpdateOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
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
	opts ...options.Enumerable[options.CollectionSaveOptions],
) error {
	saveOpts := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, saveOpts)
	_, err := c.Get(ctx, doc.ID(), options.CollectionGet().SetShowDeleted(true))
	if err == nil {
		return c.Update(ctx, doc)
	}
	if err.Error() == client.ErrDocumentNotFoundOrNotAuthorized.Error() {
		addOpts := options.CollectionAdd().
			SetEncryptDoc(saveOpts.EncryptDoc).
			SetEncryptedFields(saveOpts.EncryptedFields)
		return c.Add(ctx, doc, addOpts)
	}
	return err
}

func (c *Collection) Delete(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.CollectionDeleteOptions],
) (bool, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	res, err := execute(ctx, c.client, "delete", docID.String())
	if err != nil {
		return false, err
	}
	return res[0].Bool(), nil
}

func (c *Collection) Exists(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.CollectionExistsOptions],
) (bool, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
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
	opts ...options.Enumerable[options.CollectionUpdateWithFilterOptions],
) (*client.UpdateResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
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
	opts ...options.Enumerable[options.CollectionDeleteWithFilterOptions],
) (*client.DeleteResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
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
	opts ...options.Enumerable[options.CollectionGetOptions],
) (*client.Document, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	showDeleted := opt.ShowDeleted
	res, err := execute(ctx, c.client, "get", docID.String(), showDeleted)
	if err != nil {
		return nil, err
	}
	var docMap map[string]any
	if err := goji.UnmarshalJS(res[0], &docMap); err != nil {
		return nil, err
	}
	doc, err := client.NewDocWithID(ctx, docID, c.Version())
	if err != nil {
		return nil, err
	}
	for f, v := range docMap {
		if err := doc.Set(ctx, f, v); err != nil {
			return nil, err
		}
	}
	doc.Clean()
	return doc, nil
}

func (c *Collection) GetAllDocIDs(
	ctx context.Context,
	opts ...options.Enumerable[options.CollectionGetAllDocIDsOptions],
) (<-chan client.DocIDResult, error) {
	panic("not implemented")
}

func (c *Collection) AddIndex(
	ctx context.Context,
	indexDesc client.IndexAddRequest,
	opts ...options.Enumerable[options.CollectionAddIndexOptions],
) (client.IndexDescription, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	indexDescVal, err := goji.MarshalJS(indexDesc)
	if err != nil {
		return client.IndexDescription{}, err
	}
	res, err := execute(ctx, c.client, "addIndex", indexDescVal)
	if err != nil {
		return client.IndexDescription{}, err
	}
	var indexDescOut client.IndexDescription
	if err := goji.UnmarshalJS(res[0], &indexDescOut); err != nil {
		return client.IndexDescription{}, err
	}
	return indexDescOut, nil
}

func (c *Collection) DeleteIndex(ctx context.Context, indexName string, opts ...options.Enumerable[options.CollectionDeleteIndexOptions]) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	_, err := execute(ctx, c.client, "deleteIndex", indexName)
	return err
}

func (c *Collection) ListIndexes(ctx context.Context, opts ...options.Enumerable[options.CollectionListIndexesOptions]) ([]client.IndexDescription, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	res, err := execute(ctx, c.client, "listIndexes")
	if err != nil {
		return nil, err
	}
	var out []client.IndexDescription
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Collection) AddEncryptedIndex(
	ctx context.Context,
	req client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.AddEncryptedIndexOptions],
) (client.EncryptedIndexDescription, error) {
	opt := utils.NewOptions(opts...)
	if opt != nil {
		ctx = ctxWithOptIdentity(ctx, opt)
	}
	indexDescVal, err := goji.MarshalJS(req)
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	res, err := execute(ctx, c.client, "addEncryptedIndex", indexDescVal)
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	var indexDescOut client.EncryptedIndexDescription
	if err := goji.UnmarshalJS(res[0], &indexDescOut); err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	return indexDescOut, nil
}

func (c *Collection) ListEncryptedIndexes(ctx context.Context, opts ...options.Enumerable[options.CollectionListEncryptedIndexesOptions]) ([]client.EncryptedIndexDescription, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	res, err := execute(ctx, c.client, "listEncryptedIndexes")
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
	opt := utils.NewOptions(opts...)
	if opt != nil {
		ctx = ctxWithOptIdentity(ctx, opt)
	}
	_, err := execute(ctx, c.client, "deleteEncryptedIndex", fieldName)
	return err
}

func (c *Collection) Truncate(ctx context.Context, opts ...options.Enumerable[options.CollectionTruncateOptions]) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	_, err := execute(ctx, c.client, "truncate")
	return err
}
