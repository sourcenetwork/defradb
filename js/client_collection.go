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
	"syscall/js"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/goji"
)

type clientCollection struct {
	col client.Collection
}

func newCollection(col client.Collection) js.Value {
	c := &clientCollection{
		col: col,
	}
	return js.ValueOf(map[string]any{
		"name":                      goji.Async(c.name),
		"versionID":                 goji.Async(c.versionID),
		"version":                   goji.Async(c.version),
		"collectionID":              goji.Async(c.collectionID),
		"addDocument":               goji.Async(c.addDocument),
		"addManyDocuments":          goji.Async(c.addManyDocuments),
		"updateDocument":            goji.Async(c.updateDocument),
		"deleteDocument":            goji.Async(c.deleteDocument),
		"existsDocument":            goji.Async(c.existsDocument),
		"updateDocumentsWithFilter": goji.Async(c.updateDocumentsWithFilter),
		"deleteDocumentsWithFilter": goji.Async(c.deleteDocumentsWithFilter),
		"getDocument":               goji.Async(c.getDocument),
		"newIndex":                  goji.Async(c.newIndex),
		"deleteIndex":               goji.Async(c.deleteIndex),
		"listIndexes":               goji.Async(c.listIndexes),
		"newEncryptedIndex":         goji.Async(c.newEncryptedIndex),
		"deleteEncryptedIndex":      goji.Async(c.deleteEncryptedIndex),
		"listEncryptedIndexes":      goji.Async(c.listEncryptedIndexes),
		"truncate":                  goji.Async(c.truncate),
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

func (c *clientCollection) newIndex(this js.Value, args []js.Value) (js.Value, error) {
	var request client.NewIndexRequest
	if err := structArg(args, 0, "request", &request); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.NewCollectionIndex()
	setOptIdentity(opt, args, 1)
	desc, err := c.col.NewIndex(ctx, request, opt)
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
	ctx, err := contextArg(args, 1)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.DeleteCollectionIndex()
	setOptIdentity(opt, args, 1)
	err = c.col.DeleteIndex(ctx, name, opt)
	return js.Undefined(), err
}

func (c *clientCollection) listIndexes(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.ListCollectionIndexes()
	setOptIdentity(opt, args, 0)
	desc, err := c.col.ListIndexes(ctx, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(desc)
}

func (c *clientCollection) newEncryptedIndex(this js.Value, args []js.Value) (js.Value, error) {
	var request client.EncryptedIndexDescription
	if err := structArg(args, 0, "request", &request); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.NewEncryptedIndex()
	setOptIdentity(opt, args, 1)
	desc, err := c.col.NewEncryptedIndex(ctx, request, opt)
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
	ctx, err := contextArg(args, 1)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.DeleteEncryptedIndex()
	setOptIdentity(opt, args, 1)
	err = c.col.DeleteEncryptedIndex(ctx, fieldName, opt)
	return js.Undefined(), err
}

func (c *clientCollection) listEncryptedIndexes(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.ListCollectionEncryptedIndexes()
	setOptIdentity(opt, args, 0)
	desc, err := c.col.ListEncryptedIndexes(ctx, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(desc)
}

func (c *clientCollection) truncate(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.TruncateCollection()
	setOptIdentity(opt, args, 0)
	err = c.col.Truncate(ctx, opt)
	return js.Undefined(), err
}
