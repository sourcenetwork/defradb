// Copyright 2026 Democratized Data Foundation
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
	"syscall/js"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/goji"
)

func (c *clientCollection) addDocument(this js.Value, args []js.Value) (js.Value, error) {
	var docMap map[string]any
	if err := structArg(args, 0, "doc", &docMap); err != nil {
		return js.Undefined(), err
	}

	opts, err := getAddOptionsFromArg(args, 1, 2)
	if err != nil {
		return js.Undefined(), err
	}

	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	doc, err := client.NewDocFromMap(ctx, docMap, c.col.Version())
	if err != nil {
		return js.Undefined(), err
	}
	err = c.col.AddDocument(ctx, doc, opts...)
	return js.Undefined(), err
}

func (c *clientCollection) addManyDocuments(this js.Value, args []js.Value) (js.Value, error) {
	var docMaps []map[string]any
	if err := structArg(args, 0, "doc", &docMaps); err != nil {
		return js.Undefined(), err
	}

	opts, err := getAddOptionsFromArg(args, 1, 2)
	if err != nil {
		return js.Undefined(), err
	}

	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var docs []*client.Document
	for _, d := range docMaps {
		doc, err := client.NewDocFromMap(ctx, d, c.col.Version())
		if err != nil {
			return js.Undefined(), err
		}
		docs = append(docs, doc)
	}
	err = c.col.AddManyDocuments(ctx, docs, opts...)
	return js.Undefined(), err
}

// addOptionsInput represents the input structure for add options from JS.
type addOptionsInput struct {
	EncryptDoc      bool     `json:"encryptDoc"`
	EncryptedFields []string `json:"encryptedFields"`
}

func getAddOptionsFromArg(args []js.Value, argIndex int, ctxArgIndex int) ([]options.Enumerable[options.AddDocumentOptions], error) {
	var input addOptionsInput
	if err := structArg(args, argIndex, "options", &input); err != nil {
		return nil, err
	}

	opt := options.AddDocument()
	if input.EncryptDoc {
		opt.SetEncryptDoc(true)
	}
	if len(input.EncryptedFields) > 0 {
		opt.SetEncryptedFields(input.EncryptedFields)
	}
	setOptIdentity(opt, args, ctxArgIndex)
	return []options.Enumerable[options.AddDocumentOptions]{opt}, nil
}

func (c *clientCollection) updateDocument(this js.Value, args []js.Value) (js.Value, error) {
	docIDString, err := stringArg(args, 0, "docID")
	if err != nil {
		return js.Undefined(), err
	}
	patch, err := stringArg(args, 1, "patch")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	docID, err := client.NewDocIDFromString(docIDString)
	if err != nil {
		return js.Undefined(), err
	}
	getOpt := options.GetDocument().SetShowDeleted(true)
	setOptIdentity(getOpt, args, 2)
	doc, err := c.col.GetDocument(ctx, docID, getOpt)
	if err != nil {
		return js.Undefined(), err
	}
	if err := doc.SetWithJSON(ctx, []byte(patch)); err != nil {
		return js.Undefined(), err
	}
	opt := options.UpdateDocument()
	setOptIdentity(opt, args, 2)
	err = c.col.UpdateDocument(ctx, doc, opt)
	return js.Undefined(), err
}

func (c *clientCollection) deleteDocument(this js.Value, args []js.Value) (js.Value, error) {
	docIDString, err := stringArg(args, 0, "docID")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	docID, err := client.NewDocIDFromString(docIDString)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.DeleteDocument()
	setOptIdentity(opt, args, 1)
	deleted, err := c.col.DeleteDocument(ctx, docID, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return js.ValueOf(deleted), nil
}

func (c *clientCollection) existsDocument(this js.Value, args []js.Value) (js.Value, error) {
	docIDString, err := stringArg(args, 0, "docID")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	docID, err := client.NewDocIDFromString(docIDString)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.ExistsDocument()
	setOptIdentity(opt, args, 1)
	exists, err := c.col.ExistsDocument(ctx, docID, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return js.ValueOf(exists), nil
}

func (c *clientCollection) updateDocumentsWithFilter(this js.Value, args []js.Value) (js.Value, error) {
	filter, err := stringArg(args, 0, "filter")
	if err != nil {
		return js.Undefined(), err
	}
	updater, err := stringArg(args, 1, "updater")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.UpdateDocumentsWithFilter()
	setOptIdentity(opt, args, 2)
	result, err := c.col.UpdateDocumentsWithFilter(ctx, filter, updater, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(result)
}

func (c *clientCollection) deleteDocumentsWithFilter(this js.Value, args []js.Value) (js.Value, error) {
	filter, err := stringArg(args, 0, "filter")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.DeleteDocumentsWithFilter()
	setOptIdentity(opt, args, 1)
	result, err := c.col.DeleteDocumentsWithFilter(ctx, filter, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(result)
}

func (c *clientCollection) getDocument(this js.Value, args []js.Value) (js.Value, error) {
	docIDString, err := stringArg(args, 0, "docID")
	if err != nil {
		return js.Undefined(), err
	}
	showDeleted, err := boolArg(args, 1, "showDeleted")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	docID, err := client.NewDocIDFromString(docIDString)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.GetDocument().SetShowDeleted(showDeleted)
	setOptIdentity(opt, args, 2)
	doc, err := c.col.GetDocument(ctx, docID, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(doc)
}
