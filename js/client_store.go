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
	cols, err := c.node.DB.AddSchema(ctx, schema)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(cols)
}

func (c *Client) patchSchema(this js.Value, args []js.Value) (js.Value, error) {
	patch, err := stringArg(args, 0, "patch")
	if err != nil {
		return js.Undefined(), err
	}
	var migration immutable.Option[model.Lens]
	if err := structArg(args, 1, "lens", &migration); err != nil {
		return js.Undefined(), err
	}
	setAsDefaultVersion, err := boolArg(args, 2, "setAsDefaultVersion")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 3, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = c.node.DB.PatchSchema(ctx, patch, migration, setAsDefaultVersion)
	return js.Undefined(), err
}

func (c *Client) patchCollection(this js.Value, args []js.Value) (js.Value, error) {
	patch, err := stringArg(args, 0, "patch")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = c.node.DB.PatchCollection(ctx, patch)
	return js.Undefined(), err
}

func (c *Client) setActiveSchemaVersion(this js.Value, args []js.Value) (js.Value, error) {
	version, err := stringArg(args, 0, "version")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = c.node.DB.SetActiveSchemaVersion(ctx, version)
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
	var transform immutable.Option[model.Lens]
	if err := structArg(args, 2, "transform", &transform); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 3, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	cols, err := c.node.DB.AddView(ctx, gqlQuery, sdl, transform)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(cols)
}

func (c *Client) refreshViews(this js.Value, args []js.Value) (js.Value, error) {
	var options client.CollectionFetchOptions
	if err := structArg(args, 0, "options", &options); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = c.node.DB.RefreshViews(ctx, options)
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
	err = c.node.DB.SetMigration(ctx, config)
	return js.Undefined(), err
}

func (c *Client) lensRegistry(this js.Value, args []js.Value) (js.Value, error) {
	return newLensRegistry(c.node.DB.LensRegistry(), c.txns), nil
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
	col, err := c.node.DB.GetCollectionByName(ctx, name)
	if err != nil {
		return js.Undefined(), err
	}
	return newCollection(col, c.txns), nil
}

func (c *Client) getCollections(this js.Value, args []js.Value) (js.Value, error) {
	var options client.CollectionFetchOptions
	if err := structArg(args, 0, "options", &options); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	cols, err := c.node.DB.GetCollections(ctx, options)
	if err != nil {
		return js.Undefined(), err
	}
	wrappers := make([]any, len(cols))
	for i, col := range cols {
		wrappers[i] = newCollection(col, c.txns)
	}
	return js.ValueOf(wrappers), nil
}

func (c *Client) getSchemaByVersionID(this js.Value, args []js.Value) (js.Value, error) {
	version, err := stringArg(args, 0, "version")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	schema, err := c.node.DB.GetSchemaByVersionID(ctx, version)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(schema)
}

func (c *Client) getSchemas(this js.Value, args []js.Value) (js.Value, error) {
	var options client.SchemaFetchOptions
	if err := structArg(args, 0, "options", &options); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	schemas, err := c.node.DB.GetSchemas(ctx, options)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(schemas)
}

func (c *Client) getAllIndexes(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	indexes, err := c.node.DB.GetAllIndexes(ctx)
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
	var opts []client.RequestOption
	if args[1].Type() == js.TypeObject {
		operationName := args[1].Get("operationName")
		if operationName.Type() == js.TypeString {
			opts = append(opts, client.WithOperationName(operationName.String()))
		}
		variables := args[1].Get("variables")
		if variables.Type() == js.TypeObject {
			var variablesMap map[string]any
			if err := goji.UnmarshalJS(variables, &variablesMap); err != nil {
				return js.Undefined(), fmt.Errorf("failed to parse variables %w", err)
			}
			opts = append(opts, client.WithVariables(variablesMap))
		}
	}
	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res := c.node.DB.ExecRequest(ctx, request, opts...)
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
