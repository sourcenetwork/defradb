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
		"saveDocument":              goji.Async(c.saveDocument),
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
		"purgeByDocIDs":             goji.Async(c.purgeByDocIDs),
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
	optsVal := optionsValue(args, 1)
	var opt options.NewCollectionIndexOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	desc, err := c.col.NewIndex(context.Background(), request, asOpts(opt))
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
	optsVal := optionsValue(args, 1)
	var opt options.DeleteCollectionIndexOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.col.DeleteIndex(context.Background(), name, asOpts(opt))
}

func (c *clientCollection) listIndexes(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.ListCollectionIndexesOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	desc, err := c.col.ListIndexes(context.Background(), asOpts(opt))
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
	optsVal := optionsValue(args, 1)
	var opt options.NewEncryptedIndexOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	desc, err := c.col.NewEncryptedIndex(context.Background(), request, asOpts(opt))
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
	optsVal := optionsValue(args, 1)
	var opt options.DeleteEncryptedIndexOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.col.DeleteEncryptedIndex(context.Background(), fieldName, asOpts(opt))
}

func (c *clientCollection) listEncryptedIndexes(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.ListCollectionEncryptedIndexesOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	desc, err := c.col.ListEncryptedIndexes(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(desc)
}

func (c *clientCollection) truncate(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.TruncateCollectionOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.col.Truncate(context.Background(), asOpts(opt))
}

func (c *clientCollection) purgeByDocIDs(this js.Value, args []js.Value) (js.Value, error) {
	var request struct {
		DocIDs       []string `json:"docIDs"`
		PruneHistory bool     `json:"pruneHistory"`
	}
	if err := structArg(args, 0, "request", &request); err != nil {
		return js.Undefined(), err
	}

	docIDs := make([]client.DocID, len(request.DocIDs))
	for i, rawID := range request.DocIDs {
		docID, err := client.NewDocIDFromString(rawID)
		if err != nil {
			return js.Undefined(), err
		}
		docIDs[i] = docID
	}

	optsVal := optionsValue(args, 1)
	var opt options.PurgeByDocIDsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}

	return js.Undefined(), c.col.PurgeByDocIDs(
		context.Background(),
		docIDs,
		request.PruneHistory,
		asOpts(opt),
	)
}
