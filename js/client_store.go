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
	"fmt"
	"syscall/js"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"

	"github.com/sourcenetwork/goji"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"
)

func (c *Client) addSchema(this js.Value, args []js.Value) (js.Value, error) {
	schema, err := stringArg(args, 0, "schema")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.AddSchema()
	setOptIdentity(opt, args, 1)
	cols, err := c.node.DB.AddSchema(ctx, schema, opt)
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
	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.PatchCollection()
	setOptIdentity(opt, args, 2)
	err = c.node.DB.PatchCollection(ctx, patch, migration, opt)
	return js.Undefined(), err
}

func (c *Client) setActiveCollectionVersion(this js.Value, args []js.Value) (js.Value, error) {
	version, err := stringArg(args, 0, "version")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.SetActiveCollectionVersion()
	setOptIdentity(opt, args, 1)
	err = c.node.DB.SetActiveCollectionVersion(ctx, version, opt)
	return js.Undefined(), err
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
	var transformCID immutable.Option[string]
	if err := structArg(args, 2, "transformCID", &transformCID); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 3, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opts := options.AddView()
	if transformCID.HasValue() {
		opts.SetTransformCID(transformCID.Value())
	}
	cols, err := c.node.DB.AddView(ctx, gqlQuery, sdl, opts)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(cols)
}

// collectionFetchOptions is a local type for JSON serialization from the JS client.
type collectionFetchOptions struct {
	CollectionName  immutable.Option[string]
	VersionID       immutable.Option[string]
	CollectionID    immutable.Option[string]
	IncludeInactive immutable.Option[bool]
}

func (c *Client) refreshViews(this js.Value, args []js.Value) (js.Value, error) {
	var input collectionFetchOptions
	if err := structArg(args, 0, "options", &input); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := collectionFetchOptionsToGetCollectionsOptions(input)
	setOptIdentity(opt, args, 1)
	err = c.node.DB.RefreshViews(ctx, opt)
	return js.Undefined(), err
}

func (c *Client) setMigration(this js.Value, args []js.Value) (js.Value, error) {
	var config client.LensConfig
	if err := structArg(args, 0, "config", &config); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	lensID, err := c.node.DB.SetMigration(ctx, config)
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
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.AddLens()
	setOptIdentity(opt, args, 1)
	lensID, err := c.node.DB.AddLens(ctx, lens, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return js.ValueOf(lensID), err
}

func (c *Client) listLenses(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.ListLenses()
	setOptIdentity(opt, args, 0)
	lenses, err := c.node.DB.ListLenses(ctx, opt)
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
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.GetCollectionByName()
	setOptIdentity(opt, args, 1)
	col, err := c.node.DB.GetCollectionByName(ctx, name, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return newCollection(col, c.txns), nil
}

func (c *Client) getCollections(this js.Value, args []js.Value) (js.Value, error) {
	var input collectionFetchOptions
	if err := structArg(args, 0, "options", &input); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := collectionFetchOptionsToGetCollectionsOptions(input)
	setOptIdentity(opt, args, 1)
	cols, err := c.node.DB.GetCollections(ctx, opt)
	if err != nil {
		return js.Undefined(), err
	}
	wrappers := make([]any, len(cols))
	for i, col := range cols {
		wrappers[i] = newCollection(col, c.txns)
	}
	return js.ValueOf(wrappers), nil
}

// collectionFetchOptionsToGetCollectionsOptions converts collectionFetchOptions to GetCollectionsOptions.
func collectionFetchOptionsToGetCollectionsOptions(input collectionFetchOptions) *options.GetCollectionsOptionsBuilder {
	opt := options.GetCollections()
	if input.VersionID.HasValue() {
		opt.SetVersionID(input.VersionID.Value())
	}
	if input.CollectionID.HasValue() {
		opt.SetCollectionID(input.CollectionID.Value())
	}
	if input.CollectionName.HasValue() {
		opt.SetCollectionName(input.CollectionName.Value())
	}
	if input.IncludeInactive.HasValue() {
		opt.SetIncludeInactive(input.IncludeInactive.Value())
	}
	return opt
}

func (c *Client) getAllIndexes(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.GetAllIndexes()
	setOptIdentity(opt, args, 0)
	indexes, err := c.node.DB.GetAllIndexes(ctx, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(indexes)
}

func (c *Client) listAllEncryptedIndexes(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	indexes, err := c.node.DB.ListAllEncryptedIndexes(ctx)
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
	var opt *options.ExecRequestOptionsBuilder
	if args[1].Type() == js.TypeObject {
		opt = options.ExecRequest()
		operationName := args[1].Get("OperationName")
		if operationName.Type() == js.TypeString {
			opt.SetOperationName(operationName.String())
		}
		variables := args[1].Get("Variables")
		if variables.Type() == js.TypeObject {
			var variablesMap map[string]any
			if err := goji.UnmarshalJS(variables, &variablesMap); err != nil {
				return js.Undefined(), fmt.Errorf("failed to parse variables %w", err)
			}
			opt.SetVariables(variablesMap)
		}
	}
	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	if opt == nil {
		opt = options.ExecRequest()
	}
	setOptIdentity(opt, args, 2)
	res := c.node.DB.ExecRequest(ctx, request, opt)
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
// js values so the async iterator can outpu the correct values
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
