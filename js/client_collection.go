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
		"name":             goji.Async(c.name),
		"versionID":        goji.Async(c.versionID),
		"version":          goji.Async(c.version),
		"schemaRoot":       goji.Async(c.schemaRoot),
		"definition":       goji.Async(c.definition),
		"schema":           goji.Async(c.schema),
		"create":           goji.Async(c.create),
		"createMany":       goji.Async(c.createMany),
		"update":           goji.Async(c.update),
		"delete":           goji.Async(c.delete),
		"exists":           goji.Async(c.exists),
		"updateWithFilter": goji.Async(c.updateWithFilter),
		"deleteWithFilter": goji.Async(c.deleteWithFilter),
		"get":              goji.Async(c.get),
		"getAllDocIDs":     goji.Async(c.getAllDocIDs),
		"createIndex":      goji.Async(c.createIndex),
		"dropIndex":        goji.Async(c.dropIndex),
		"getIndexes":       goji.Async(c.getIndexes),
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

func (c *clientCollection) schemaRoot(this js.Value, args []js.Value) (js.Value, error) {
	return js.ValueOf(c.col.SchemaRoot()), nil
}

func (c *clientCollection) definition(this js.Value, args []js.Value) (js.Value, error) {
	return goji.MarshalJS(c.col.Definition())
}

func (c *clientCollection) schema(this js.Value, args []js.Value) (js.Value, error) {
	return goji.MarshalJS(c.col.Schema())
}

func (c *clientCollection) create(this js.Value, args []js.Value) (js.Value, error) {
	var docMap map[string]any
	if err := structArg(args, 0, "doc", &docMap); err != nil {
		return js.Undefined(), err
	}

	opts, err := getCreateOptionsFromArg(args, 1)
	if err != nil {
		return js.Undefined(), err
	}

	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	doc, err := client.NewDocFromMap(docMap, c.col.Definition())
	if err != nil {
		return js.Undefined(), err
	}
	err = c.col.Create(ctx, doc, opts...)
	return js.Undefined(), err
}

func (c *clientCollection) createMany(this js.Value, args []js.Value) (js.Value, error) {
	var docMaps []map[string]any
	if err := structArg(args, 0, "doc", &docMaps); err != nil {
		return js.Undefined(), err
	}

	opts, err := getCreateOptionsFromArg(args, 1)
	if err != nil {
		return js.Undefined(), err
	}

	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var docs []*client.Document
	for _, d := range docMaps {
		doc, err := client.NewDocFromMap(d, c.col.Definition())
		if err != nil {
			return js.Undefined(), err
		}
		docs = append(docs, doc)
	}
	err = c.col.CreateMany(ctx, docs, opts...)
	return js.Undefined(), err
}

func getCreateOptionsFromArg(args []js.Value, argIndex int) ([]client.DocCreateOption, error) {
	var createOptions client.DocCreateOptions
	if err := structArg(args, argIndex, "options", &createOptions); err != nil {
		return nil, err
	}

	opts := []client.DocCreateOption{}
	if len(createOptions.EncryptedFields) > 0 {
		opts = append(opts, client.CreateDocWithEncryptedFields(createOptions.EncryptedFields))
	}

	if createOptions.EncryptDoc {
		opts = append(opts, client.CreateDocEncrypted(true))
	}
	return opts, nil
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
	doc, err := c.col.Get(ctx, docID, true)
	if err != nil {
		return js.Undefined(), err
	}
	if err := doc.SetWithJSON([]byte(patch)); err != nil {
		return js.Undefined(), err
	}
	err = c.col.Update(ctx, doc)
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
	deleted, err := c.col.Delete(ctx, docID)
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
	exists, err := c.col.Exists(ctx, docID)
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
	result, err := c.col.UpdateWithFilter(ctx, filter, updater)
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
	result, err := c.col.DeleteWithFilter(ctx, filter)
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
	doc, err := c.col.Get(ctx, docID, showDeleted)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(doc)
}

func (c *clientCollection) getAllDocIDs(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res, err := c.col.GetAllDocIDs(ctx)
	if err != nil {
		return js.Undefined(), err
	}
	out := make(chan any)
	go func() {
		defer close(out)
		for id := range res {
			out <- js.ValueOf(id.ID.String())
		}
	}()
	return goji.AsyncIteratorOf(out), err
}

func (c *clientCollection) createIndex(this js.Value, args []js.Value) (js.Value, error) {
	var request client.IndexCreateRequest
	if err := structArg(args, 0, "request", &request); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	desc, err := c.col.CreateIndex(ctx, request)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(desc)
}

func (c *clientCollection) dropIndex(this js.Value, args []js.Value) (js.Value, error) {
	name, err := stringArg(args, 0, "name")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = c.col.DropIndex(ctx, name)
	return js.Undefined(), err
}

func (c *clientCollection) getIndexes(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	desc, err := c.col.GetIndexes(ctx)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(desc)
}
