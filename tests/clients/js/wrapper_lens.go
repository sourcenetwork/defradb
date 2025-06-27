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
	sysjs "syscall/js"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/goji"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
)

var _ client.LensRegistry = (*LensRegistry)(nil)

type LensRegistry struct {
	client sysjs.Value
}

func (w *LensRegistry) Init(txnSource client.TxnSource) {}

func (w *LensRegistry) SetMigration(ctx context.Context, collectionID string, config model.Lens) error {
	configVal, err := goji.MarshalJS(config)
	if err != nil {
		return err
	}
	promise := w.client.Call("setMigration", collectionID, configVal)
	_, err = goji.Await(goji.PromiseValue(promise))
	return err
}

func (w *LensRegistry) ReloadLenses(ctx context.Context) error {
	promise := w.client.Call("reloadLenses")
	_, err := goji.Await(goji.PromiseValue(promise))
	return err
}

func (w *LensRegistry) MigrateUp(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	collectionID string,
) (enumerable.Enumerable[map[string]any], error) {
	var srcList []map[string]any
	enumerable.ForEach(src, func(item map[string]any) {
		srcList = append(srcList, item)
	})
	srcVal, err := goji.MarshalJS(srcList)
	if err != nil {
		return nil, err
	}
	promise := w.client.Call("migrateUp", srcVal, collectionID)
	res, err := goji.Await(goji.PromiseValue(promise))
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return enumerable.New(out), nil
}

func (w *LensRegistry) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	collectionID string,
) (enumerable.Enumerable[map[string]any], error) {
	var srcList []map[string]any
	enumerable.ForEach(src, func(item map[string]any) {
		srcList = append(srcList, item)
	})
	srcVal, err := goji.MarshalJS(srcList)
	if err != nil {
		return nil, err
	}
	promise := w.client.Call("migrateDown", srcVal, collectionID)
	res, err := goji.Await(goji.PromiseValue(promise))
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return enumerable.New(out), nil
}
