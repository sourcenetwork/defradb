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
	res, err := execute(ctx, c.client, "addDocument", goji.MustMarshalJS(doc), jsOpts(opts))
	if err != nil {
		return err
	}
	if err := setDocumentIDsFromJS([]*client.Document{doc}, res[0]); err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func setDocumentIDsFromJS(docs []*client.Document, value js.Value) error {
	var docIDs []string
	if err := goji.UnmarshalJS(value, &docIDs); err != nil {
		return err
	}
	if len(docIDs) != len(docs) {
		return client.NewErrUnexpectedType[[]string]("docIDs", docIDs)
	}
	for i, docIDString := range docIDs {
		docID, err := client.NewDocIDFromString(docIDString)
		if err != nil {
			return err
		}
		client.ApplySavedDocumentID(docs[i], docID)
	}
	return nil
}

func (c *Collection) AddManyDocuments(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	res, err := execute(ctx, c.client, "addManyDocuments", goji.MustMarshalJS(docs), jsOpts(opts))
	if err != nil {
		return err
	}
	if err := setDocumentIDsFromJS(docs, res[0]); err != nil {
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
	patch, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	if _, err := execute(ctx, c.client, "updateDocument", doc.ID().String(), string(patch), jsOpts(opts)); err != nil {
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
	if !doc.ID().IsValid() {
		return c.AddDocument(ctx, doc, opts...)
	}

	patch, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	res, err := execute(ctx, c.client, "saveDocument", doc.ID().String(), string(patch), jsOpts(opts))
	if err != nil {
		return err
	}
	if err := setDocumentIDsFromJS([]*client.Document{doc}, res[0]); err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) DeleteDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.DeleteDocumentOptions],
) (bool, error) {
	res, err := execute(ctx, c.client, "deleteDocument", docID.String(), jsOpts(opts))
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
	res, err := execute(ctx, c.client, "existsDocument", docID.String(), jsOpts(opts))
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
	res, err := execute(ctx, c.client, "updateDocumentsWithFilter",
		utils.NormalizeFilterForJSON(filter), updater, jsOpts(opts))
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
	res, err := execute(ctx, c.client, "deleteDocumentsWithFilter",
		utils.NormalizeFilterForJSON(filter), jsOpts(opts))
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
	res, err := execute(ctx, c.client, "getDocument", docID.String(), jsOpts(opts))
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
