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

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/goji"
	"github.com/sourcenetwork/immutable/enumerable"
)

type clientLens struct {
	registry client.LensRegistry
	txns     *sync.Map
}

func newLensRegistry(registry client.LensRegistry, txns *sync.Map) js.Value {
	r := &clientLens{
		registry: registry,
		txns:     txns,
	}
	return js.ValueOf(map[string]any{
		"setMigration": goji.Async(r.setMigration),
		"reloadLenses": goji.Async(r.reloadLenses),
		"migrateUp":    goji.Async(r.migrateUp),
		"migrateDown":  goji.Async(r.migrateDown),
	})
}

func (c *clientLens) setMigration(this js.Value, args []js.Value) (js.Value, error) {
	collectionID, err := stringArg(args, 0, "collectionID")
	if err != nil {
		return js.Undefined(), err
	}
	var lens model.Lens
	if err := structArg(args, 1, "lens", &lens); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = c.registry.SetMigration(ctx, collectionID, lens)
	return js.Undefined(), err
}

func (c *clientLens) reloadLenses(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = c.registry.ReloadLenses(ctx)
	return js.Undefined(), err
}

func (c *clientLens) migrateUp(this js.Value, args []js.Value) (js.Value, error) {
	var data []map[string]any
	if err := structArg(args, 0, "data", &data); err != nil {
		return js.Undefined(), err
	}
	collectionID, err := stringArg(args, 1, "collectionID")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	result, err := c.registry.MigrateUp(ctx, enumerable.New(data), collectionID)
	if err != nil {
		return js.Undefined(), err
	}
	var value []map[string]any
	err = enumerable.ForEach(result, func(item map[string]any) {
		value = append(value, item)
	})
	return goji.MarshalJS(value)
}

func (c *clientLens) migrateDown(this js.Value, args []js.Value) (js.Value, error) {
	var data []map[string]any
	if err := structArg(args, 0, "data", &data); err != nil {
		return js.Undefined(), err
	}
	collectionID, err := stringArg(args, 1, "collectionID")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	result, err := c.registry.MigrateDown(ctx, enumerable.New(data), collectionID)
	if err != nil {
		return js.Undefined(), err
	}
	var value []map[string]any
	err = enumerable.ForEach(result, func(item map[string]any) {
		value = append(value, item)
	})
	return goji.MarshalJS(value)
}
