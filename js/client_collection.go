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
	"sync"
	"syscall/js"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/goji"
)

type clientCollection struct {
	col  client.Collection
	txns *sync.Map
}

func newCollection(col client.Collection, txns *sync.Map) js.Value {
	c := &clientCollection{
		col:  col,
		txns: txns,
	}
	return js.ValueOf(map[string]any{
		"name":                 goji.Async(c.name),
		"versionID":            goji.Async(c.versionID),
		"version":              goji.Async(c.version),
		"collectionID":         goji.Async(c.collectionID),
		"add":                  goji.Async(c.add),
		"addMany":              goji.Async(c.addMany),
		"update":               goji.Async(c.update),
		"delete":               goji.Async(c.delete),
		"exists":               goji.Async(c.exists),
		"updateWithFilter":     goji.Async(c.updateWithFilter),
		"deleteWithFilter":     goji.Async(c.deleteWithFilter),
		"get":                  goji.Async(c.get),
		"addIndex":             goji.Async(c.addIndex),
		"deleteIndex":          goji.Async(c.deleteIndex),
		"listIndexes":          goji.Async(c.listIndexes),
		"addEncryptedIndex":    goji.Async(c.addEncryptedIndex),
		"deleteEncryptedIndex": goji.Async(c.deleteEncryptedIndex),
		"listEncryptedIndexes": goji.Async(c.listEncryptedIndexes),
		"truncate":             goji.Async(c.truncate),
	})
}

func (c *clientCollection) name(this js.Value, args []js.Value) (js.Value, error) {
	return js.ValueOf(c.col.Name()), nil
}

func (c *clientCollection) versionID(this js.Value, args []js.Value) (js.Value, error) {
	return js.ValueOf(c.col.VersionID()), nil
}

func (c *clientCollection) version(this js.Value, args []js.Value) (js.Value, error) {
	return goji.MarshalJS(c.col.Version())
}

func (c *clientCollection) collectionID(this js.Value, args []js.Value) (js.Value, error) {
	return js.ValueOf(c.col.CollectionID()), nil
}

func (c *clientCollection) add(this js.Value, args []js.Value) (js.Value, error) {
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
	err = c.col.Add(ctx, doc, opts...)
	return js.Undefined(), err
}

func (c *clientCollection) addMany(this js.Value, args []js.Value) (js.Value, error) {
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
	err = c.col.AddMany(ctx, docs, opts...)
	return js.Undefined(), err
}

// addOptionsInput represents the input structure for add options from JS.
type addOptionsInput struct {
	EncryptDoc      bool     `json:"encryptDoc"`
	EncryptedFields []string `json:"encryptedFields"`
}

func getAddOptionsFromArg(args []js.Value, argIndex int, ctxArgIndex int) ([]options.Enumerable[options.CollectionAddOptions], error) {
	var input addOptionsInput
	if err := structArg(args, argIndex, "options", &input); err != nil {
		return nil, err
	}

	opt := options.CollectionAdd()
	if input.EncryptDoc {
		opt.SetEncryptDoc(true)
	}
	if len(input.EncryptedFields) > 0 {
		opt.SetEncryptedFields(input.EncryptedFields)
	}
	setOptIdentity(opt, args, ctxArgIndex)
	return []options.Enumerable[options.CollectionAddOptions]{opt}, nil
}

func (c *clientCollection) update(this js.Value, args []js.Value) (js.Value, error) {
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
	getOpt := options.CollectionGet().SetShowDeleted(true)
	setOptIdentity(getOpt, args, 2)
	doc, err := c.col.Get(ctx, docID, getOpt)
	if err != nil {
		return js.Undefined(), err
	}
	if err := doc.SetWithJSON(ctx, []byte(patch)); err != nil {
		return js.Undefined(), err
	}
	opt := options.CollectionUpdate()
	setOptIdentity(opt, args, 2)
	err = c.col.Update(ctx, doc, opt)
	return js.Undefined(), err
}

func (c *clientCollection) delete(this js.Value, args []js.Value) (js.Value, error) {
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
	opt := options.CollectionDelete()
	setOptIdentity(opt, args, 1)
	deleted, err := c.col.Delete(ctx, docID, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return js.ValueOf(deleted), nil
}

func (c *clientCollection) exists(this js.Value, args []js.Value) (js.Value, error) {
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
	opt := options.CollectionExists()
	setOptIdentity(opt, args, 1)
	exists, err := c.col.Exists(ctx, docID, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return js.ValueOf(exists), nil
}

func (c *clientCollection) updateWithFilter(this js.Value, args []js.Value) (js.Value, error) {
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
	opt := options.CollectionUpdateWithFilter()
	setOptIdentity(opt, args, 2)
	result, err := c.col.UpdateWithFilter(ctx, filter, updater, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(result)
}

func (c *clientCollection) deleteWithFilter(this js.Value, args []js.Value) (js.Value, error) {
	filter, err := stringArg(args, 0, "filter")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.CollectionDeleteWithFilter()
	setOptIdentity(opt, args, 1)
	result, err := c.col.DeleteWithFilter(ctx, filter, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(result)
}

func (c *clientCollection) get(this js.Value, args []js.Value) (js.Value, error) {
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
	opt := options.CollectionGet().SetShowDeleted(showDeleted)
	setOptIdentity(opt, args, 2)
	doc, err := c.col.Get(ctx, docID, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(doc)
}

func (c *clientCollection) addIndex(this js.Value, args []js.Value) (js.Value, error) {
	var request client.IndexAddRequest
	if err := structArg(args, 0, "request", &request); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.CollectionAddIndex()
	setOptIdentity(opt, args, 1)
	desc, err := c.col.AddIndex(ctx, request, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(desc)
}

func (c *clientCollection) deleteIndex(this js.Value, args []js.Value) (js.Value, error) {
	name, err := stringArg(args, 0, "name")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.CollectionDeleteIndex()
	setOptIdentity(opt, args, 1)
	err = c.col.DeleteIndex(ctx, name, opt)
	return js.Undefined(), err
}

func (c *clientCollection) listIndexes(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.CollectionListIndexes()
	setOptIdentity(opt, args, 0)
	desc, err := c.col.ListIndexes(ctx, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(desc)
}

func (c *clientCollection) addEncryptedIndex(this js.Value, args []js.Value) (js.Value, error) {
	var request client.EncryptedIndexDescription
	if err := structArg(args, 0, "request", &request); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.AddEncryptedIndex()
	setOptIdentity(opt, args, 1)
	desc, err := c.col.AddEncryptedIndex(ctx, request, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(desc)
}

func (c *clientCollection) deleteEncryptedIndex(this js.Value, args []js.Value) (js.Value, error) {
	fieldName, err := stringArg(args, 0, "fieldName")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.DeleteEncryptedIndex()
	setOptIdentity(opt, args, 1)
	err = c.col.DeleteEncryptedIndex(ctx, fieldName, opt)
	return js.Undefined(), err
}

func (c *clientCollection) listEncryptedIndexes(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.CollectionListEncryptedIndexes()
	setOptIdentity(opt, args, 0)
	desc, err := c.col.ListEncryptedIndexes(ctx, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(desc)
}

func (c *clientCollection) truncate(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.CollectionTruncate()
	setOptIdentity(opt, args, 0)
	err = c.col.Truncate(ctx, opt)
	return js.Undefined(), err
}
