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
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

func (c *Client) addCollection(this js.Value, args []js.Value) (js.Value, error) {
	sdl, err := stringArg(args, 0, "sdl")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.AddCollectionOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	cols, err := store.AddCollection(context.Background(), sdl, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(cols)
}

func (c *Client) patchCollection(this js.Value, args []js.Value) (js.Value, error) {
	patch, err := stringArg(args, 0, "patch")
	if err != nil {
		return js.Undefined(), err
	}
	var migration immutable.Option[model.Lens]
	if err := structArg(args, 1, "lens", &migration); err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 2)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.PatchCollectionOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), store.PatchCollection(context.Background(), patch, migration, asOpts(opt))
}

func (c *Client) deleteCollection(this js.Value, args []js.Value) (js.Value, error) {
	names, err := stringSliceArg(args, 0, "names")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.DeleteCollectionOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), store.DeleteCollection(context.Background(), names, asOpts(opt))
}

func (c *Client) setActiveCollectionVersion(this js.Value, args []js.Value) (js.Value, error) {
	version, err := stringArg(args, 0, "version")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.SetActiveCollectionVersionOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), store.SetActiveCollectionVersion(context.Background(), version, asOpts(opt))
}

func (c *Client) addView(this js.Value, args []js.Value) (js.Value, error) {
	gqlQuery, err := stringArg(args, 0, "gqlQuery")
	if err != nil {
		return js.Undefined(), err
	}
	sdl, err := stringArg(args, 1, "sdl")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 2)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.AddViewOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	cols, err := store.AddView(context.Background(), gqlQuery, sdl, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(cols)
}

func (c *Client) refreshViews(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.RefreshViewsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), store.RefreshViews(context.Background(), asOpts(opt))
}

func (c *Client) setMigration(this js.Value, args []js.Value) (js.Value, error) {
	var config client.LensConfig
	if err := structArg(args, 0, "config", &config); err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.SetMigrationOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	lensID, err := store.SetMigration(context.Background(), config, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return js.ValueOf(lensID), err
}

func (c *Client) addLens(this js.Value, args []js.Value) (js.Value, error) {
	var lens model.Lens
	if err := structArg(args, 0, "lens", &lens); err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.AddLensOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	lensID, err := store.AddLens(context.Background(), lens, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return js.ValueOf(lensID), err
}

func (c *Client) listLenses(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ListLensesOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	lenses, err := store.ListLenses(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(lenses)
}

func (c *Client) listActions(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ListActionsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	lenses, err := store.ListActions(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(lenses)
}

func (c *Client) getCollectionByName(this js.Value, args []js.Value) (js.Value, error) {
	name, err := stringArg(args, 0, "name")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.GetCollectionByNameOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	col, err := store.GetCollectionByName(context.Background(), name, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return newCollection(col), nil
}

func (c *Client) getCollections(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.GetCollectionsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	cols, err := store.GetCollections(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	wrappers := make([]any, len(cols))
	for i, col := range cols {
		wrappers[i] = newCollection(col)
	}
	return js.ValueOf(wrappers), nil
}

func (c *Client) purgeDocuments(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 2)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	return purgeDocuments(store, args)
}

func purgeDocuments(store client.Store, args []js.Value) (js.Value, error) {
	collectionName, err := stringArg(args, 0, "collectionName")
	if err != nil {
		return js.Undefined(), err
	}
	var request struct {
		DocIDs       []string `json:"docIDs"`
		PruneHistory bool     `json:"pruneHistory"`
	}
	if err := structArg(args, 1, "request", &request); err != nil {
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

	var opt options.PurgeDocumentsOptions
	if err := parseOptions(optionsValue(args, 2), &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), store.PurgeDocuments(
		context.Background(),
		collectionName,
		docIDs,
		request.PruneHistory,
		asOpts(opt),
	)
}

func (c *Client) listIndexes(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ListIndexesOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	indexes, err := store.ListIndexes(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(indexes)
}

func (c *Client) listAllEncryptedIndexes(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ListAllEncryptedIndexesOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	indexes, err := store.ListAllEncryptedIndexes(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(indexes)
}

func (c *Client) execRequest(this js.Value, args []js.Value) (js.Value, error) {
	request, err := stringArg(args, 0, "request")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ExecRequestOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res := store.ExecRequest(context.Background(), request, asOpts(opt))
	return marshalRequestResult(res)
}

// marshalRequestResult converts a RequestResult into a JS object.
func marshalRequestResult(res *client.RequestResult) (js.Value, error) {
	gql, err := goji.MarshalJS(res.GQL)
	if err != nil {
		return js.Undefined(), err
	}
	out := map[string]any{
		"gql": gql,
	}
	if res.Subscription != nil {
		out["subscription"] = handleSubscription(res.Subscription)
	}
	return js.ValueOf(out), nil
}

// handleSubscription reads gql results and marshals them into
// js values so the async iterator can output the correct values.
func handleSubscription(sub <-chan client.GQLResult) js.Value {
	out := make(chan any)
	go func() {
		defer close(out)
		for res := range sub {
			val, err := goji.MarshalJS(res)
			if err != nil {
				return
			}
			out <- val
		}
	}()
	return goji.AsyncIteratorOf(out)
}

func (c *Client) printDump(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), store.PrintDump(context.Background())
}

func (c *Client) basicImport(this js.Value, args []js.Value) (js.Value, error) {
	filepath, err := stringArg(args, 0, "filepath")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), store.BasicImport(context.Background(), filepath)
}

func (c *Client) basicExport(this js.Value, args []js.Value) (js.Value, error) {
	filepath, err := stringArg(args, 0, "filepath")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.BasicExportOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), store.BasicExport(context.Background(), filepath, asOpts(opt))
}
