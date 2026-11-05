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
	"context"
	"errors"
	"syscall/js"

	"github.com/sourcenetwork/goji"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

func (c *clientCollection) addDocument(this js.Value, args []js.Value) (js.Value, error) {
	var docMap map[string]any
	if err := structArg(args, 0, "doc", &docMap); err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.AddDocumentOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	ctx := context.Background()
	doc, err := client.NewDocFromMap(ctx, docMap, c.col.Version())
	if err != nil {
		return js.Undefined(), err
	}
	err = c.col.AddDocument(ctx, doc, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(client.DocumentIDs([]*client.Document{doc}))
}

func (c *clientCollection) addManyDocuments(this js.Value, args []js.Value) (js.Value, error) {
	var docMaps []map[string]any
	if err := structArg(args, 0, "doc", &docMaps); err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.AddDocumentOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	ctx := context.Background()
	var docs []*client.Document
	for _, d := range docMaps {
		doc, err := client.NewDocFromMap(ctx, d, c.col.Version())
		if err != nil {
			return js.Undefined(), err
		}
		docs = append(docs, doc)
	}
	err := c.col.AddManyDocuments(ctx, docs, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(client.DocumentIDs(docs))
}

// saveDocument applies the given JSON patch to the document with the given
// ID, creating it if it does not yet exist. The wire format mirrors
// updateDocument so that only the dirty fields in the patch produce commits.
func (c *clientCollection) saveDocument(this js.Value, args []js.Value) (js.Value, error) {
	docIDString, err := stringArg(args, 0, "docID")
	if err != nil {
		return js.Undefined(), err
	}
	patch, err := stringArg(args, 1, "patch")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 2)
	docID, err := client.NewDocIDFromString(docIDString)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.SaveDocumentOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	ctx := context.Background()
	getOpt := options.GetDocumentOptions{
		Identity:    opt.Identity,
		ShowDeleted: true,
	}
	doc, err := c.col.GetDocument(ctx, docID, asOpts(getOpt))
	if err != nil {
		if !errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
			return js.Undefined(), err
		}
		doc, err = client.NewDocWithID(ctx, docID, c.col.Version())
		if err != nil {
			return js.Undefined(), err
		}
	}
	if err := doc.SetWithJSON(ctx, []byte(patch)); err != nil {
		return js.Undefined(), err
	}
	if err := c.col.SaveDocument(ctx, doc, asOpts(opt)); err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS([]string{doc.ID().String()})
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
	optsVal := optionsValue(args, 2)
	docID, err := client.NewDocIDFromString(docIDString)
	if err != nil {
		return js.Undefined(), err
	}
	var updateOpt options.UpdateDocumentOptions
	if err := parseOptions(optsVal, &updateOpt); err != nil {
		return js.Undefined(), err
	}
	ctx := context.Background()
	// Use the same identity to fetch the document.
	getOpt := options.GetDocumentOptions{
		Identity:    updateOpt.Identity,
		ShowDeleted: true,
	}
	doc, err := c.col.GetDocument(ctx, docID, asOpts(getOpt))
	if err != nil {
		return js.Undefined(), err
	}
	if err := doc.SetWithJSON(ctx, []byte(patch)); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.col.UpdateDocument(ctx, doc, asOpts(updateOpt))
}

func (c *clientCollection) deleteDocument(this js.Value, args []js.Value) (js.Value, error) {
	docIDString, err := stringArg(args, 0, "docID")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	docID, err := client.NewDocIDFromString(docIDString)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.DeleteDocumentOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	deleted, err := c.col.DeleteDocument(context.Background(), docID, asOpts(opt))
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
	optsVal := optionsValue(args, 1)
	docID, err := client.NewDocIDFromString(docIDString)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ExistsDocumentOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	exists, err := c.col.ExistsDocument(context.Background(), docID, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return js.ValueOf(exists), nil
}

func (c *clientCollection) updateDocumentsWithFilter(this js.Value, args []js.Value) (js.Value, error) {
	filter, err := filterArg(args, 0, "filter")
	if err != nil {
		return js.Undefined(), err
	}
	updater, err := stringArg(args, 1, "updater")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 2)
	var opt options.UpdateDocumentsWithFilterOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	result, err := c.col.UpdateDocumentsWithFilter(context.Background(), filter, updater, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(result)
}

func (c *clientCollection) deleteDocumentsWithFilter(this js.Value, args []js.Value) (js.Value, error) {
	filter, err := filterArg(args, 0, "filter")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.DeleteDocumentsWithFilterOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	result, err := c.col.DeleteDocumentsWithFilter(context.Background(), filter, asOpts(opt))
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
	optsVal := optionsValue(args, 1)
	docID, err := client.NewDocIDFromString(docIDString)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.GetDocumentOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	doc, err := c.col.GetDocument(context.Background(), docID, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(doc)
}
