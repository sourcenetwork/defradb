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
	"github.com/sourcenetwork/defradb/internal/utils"
)

func (c *Collection) AddDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	docVal, err := goji.MarshalJS(doc)
	if err != nil {
		return err
	}
	_, err = execute(ctx, c.client, "addDocument", docVal, makeDocAddOptions(opts))
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

func makeDocAddOptions(opts []options.Enumerable[options.AddDocumentOptions]) js.Value {
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

func (c *Collection) AddManyDocuments(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	docsVal, err := goji.MarshalJS(docs)
	if err != nil {
		return err
	}
	_, err = execute(ctx, c.client, "addManyDocuments", docsVal, makeDocAddOptions(opts))
	if err != nil {
		return err
	}
	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

func (c *Collection) UpdateDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.UpdateDocumentOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	patch, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	docID := doc.ID().String()
	_, err = execute(ctx, c.client, "updateDocument", docID, string(patch))
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) SaveDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.SaveDocumentOptions],
) error {
	saveOpts := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, saveOpts)
	_, err := c.GetDocument(ctx, doc.ID(), options.GetDocument().SetShowDeleted(true))
	if err == nil {
		return c.UpdateDocument(ctx, doc)
	}
	if err.Error() == client.ErrDocumentNotFoundOrNotAuthorized.Error() {
		addOpts := options.AddDocument().
			SetEncryptDoc(saveOpts.EncryptDoc).
			SetEncryptedFields(saveOpts.EncryptedFields)
		return c.AddDocument(ctx, doc, addOpts)
	}
	return err
}

func (c *Collection) DeleteDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.DeleteDocumentOptions],
) (bool, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	res, err := execute(ctx, c.client, "deleteDocument", docID.String())
	if err != nil {
		return false, err
	}
	return res[0].Bool(), nil
}

func (c *Collection) ExistsDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.ExistsDocumentOptions],
) (bool, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	res, err := execute(ctx, c.client, "existsDocument", docID.String())
	if err != nil {
		return false, err
	}
	return res[0].Bool(), nil
}

func (c *Collection) UpdateDocumentsWithFilter(
	ctx context.Context,
	filter any,
	updater string,
	opts ...options.Enumerable[options.UpdateDocumentsWithFilterOptions],
) (*client.UpdateResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	res, err := execute(ctx, c.client, "updateDocumentsWithFilter", filter, updater)
	if err != nil {
		return nil, err
	}
	var out client.UpdateResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Collection) DeleteDocumentsWithFilter(
	ctx context.Context,
	filter any,
	opts ...options.Enumerable[options.DeleteDocumentsWithFilterOptions],
) (*client.DeleteResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	res, err := execute(ctx, c.client, "deleteDocumentsWithFilter", filter)
	if err != nil {
		return nil, err
	}
	var out client.DeleteResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Collection) GetDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.GetDocumentOptions],
) (*client.Document, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	showDeleted := opt.ShowDeleted
	res, err := execute(ctx, c.client, "getDocument", docID.String(), showDeleted)
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
