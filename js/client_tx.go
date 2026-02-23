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
	"sync"
	"syscall/js"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/goji"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"
)

type transaction struct {
	txn client.Txn
	// txns is present only temporarily until we reach consensus on
	// the DX of transactions.
	txns *sync.Map
}

func newTransaction(txn client.Txn, txns *sync.Map) js.Value {
	wrapper := &transaction{txn, txns}
	return js.ValueOf(map[string]any{
		"id":                         txn.ID(),
		"commit":                     goji.Async(wrapper.commit),
		"discard":                    goji.Async(wrapper.discard),
		"addSchema":                  goji.Async(wrapper.addSchema),
		"patchCollection":            goji.Async(wrapper.patchCollection),
		"setActiveCollectionVersion": goji.Async(wrapper.setActiveCollectionVersion),
		"addView":                    goji.Async(wrapper.addView),
		"refreshViews":               goji.Async(wrapper.refreshViews),
		"setMigration":               goji.Async(wrapper.setMigration),
		"addLens":                    goji.Async(wrapper.addLens),
		"listLenses":                 goji.Async(wrapper.listLenses),
		"getCollectionByName":        goji.Async(wrapper.getCollectionByName),
		"getCollections":             goji.Async(wrapper.getCollections),
		"listIndexes":                goji.Async(wrapper.listIndexes),
		"listAllEncryptedIndexes":    goji.Async(wrapper.listAllEncryptedIndexes),
		"execRequest":                goji.Async(wrapper.execRequest),
		"addDACPolicy":               goji.Async(wrapper.addDACPolicy),
		"addDACActorRelationship":    goji.Async(wrapper.addDACActorRelationship),
		"deleteDACActorRelationship": goji.Async(wrapper.deleteDACActorRelationship),
		"getNACStatus":               goji.Async(wrapper.getNACStatus),
		"reEnableNAC":                goji.Async(wrapper.reEnableNAC),
		"disableNAC":                 goji.Async(wrapper.disableNAC),
		"addNACActorRelationship":    goji.Async(wrapper.addNACActorRelationship),
		"deleteNACActorRelationship": goji.Async(wrapper.deleteNACActorRelationship),
		"getNodeIdentity":            goji.Async(wrapper.getNodeIdentity),
		"verifySignature":            goji.Async(wrapper.verifySignature),
	})
}

func (t *transaction) commit(this js.Value, args []js.Value) (js.Value, error) {
	err := t.txn.Commit()
	return js.Undefined(), err
}

func (t *transaction) discard(this js.Value, args []js.Value) (js.Value, error) {
	t.txn.Discard()
	return js.Undefined(), nil
}

func (t *transaction) addSchema(this js.Value, args []js.Value) (js.Value, error) {
	schema, err := stringArg(args, 0, "schema")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.AddSchema()
	setOptIdentity(opt, args, 1)
	cols, err := t.txn.AddSchema(ctx, schema, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(cols)
}

func (t *transaction) patchCollection(this js.Value, args []js.Value) (js.Value, error) {
	patch, err := stringArg(args, 0, "patch")
	if err != nil {
		return js.Undefined(), err
	}
	var migration immutable.Option[model.Lens]
	if err := structArg(args, 1, "lens", &migration); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 2, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.PatchCollection()
	setOptIdentity(opt, args, 2)
	err = t.txn.PatchCollection(ctx, patch, migration, opt)
	return js.Undefined(), err
}

func (t *transaction) setActiveCollectionVersion(this js.Value, args []js.Value) (js.Value, error) {
	version, err := stringArg(args, 0, "version")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.SetActiveCollectionVersion()
	setOptIdentity(opt, args, 1)
	err = t.txn.SetActiveCollectionVersion(ctx, version, opt)
	return js.Undefined(), err
}

func (t *transaction) addView(this js.Value, args []js.Value) (js.Value, error) {
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
	ctx, err := contextArg(args, 3, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opts := options.AddView()
	setOptIdentity(opts, args, 3)
	if transformCID.HasValue() {
		opts.SetTransformCID(transformCID.Value())
	}
	cols, err := t.txn.AddView(ctx, gqlQuery, sdl, opts)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(cols)
}

func (t *transaction) refreshViews(this js.Value, args []js.Value) (js.Value, error) {
	var input collectionFetchOptions
	if err := structArg(args, 0, "options", &input); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := collectionFetchOptionsToGetCollectionsOptions(input)
	setOptIdentity(opt, args, 1)
	err = t.txn.RefreshViews(ctx, opt)
	return js.Undefined(), err
}

func (t *transaction) setMigration(this js.Value, args []js.Value) (js.Value, error) {
	var config client.LensConfig
	if err := structArg(args, 0, "config", &config); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.SetMigration()
	setOptIdentity(opt, args, 1)
	lensID, err := t.txn.SetMigration(ctx, config, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return js.ValueOf(lensID), err
}

func (t *transaction) addLens(this js.Value, args []js.Value) (js.Value, error) {
	var lens model.Lens
	if err := structArg(args, 0, "lens", &lens); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.AddLens()
	setOptIdentity(opt, args, 1)
	lensID, err := t.txn.AddLens(ctx, lens, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return js.ValueOf(lensID), err
}

func (t *transaction) listLenses(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.ListLenses()
	setOptIdentity(opt, args, 0)
	lenses, err := t.txn.ListLenses(ctx, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(lenses)
}

func (t *transaction) getCollectionByName(this js.Value, args []js.Value) (js.Value, error) {
	name, err := stringArg(args, 0, "name")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.GetCollectionByName()
	setOptIdentity(opt, args, 1)
	col, err := t.txn.GetCollectionByName(ctx, name, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return newCollection(col, t.txns), nil
}

func (t *transaction) getCollections(this js.Value, args []js.Value) (js.Value, error) {
	var input collectionFetchOptions
	if err := structArg(args, 0, "options", &input); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := collectionFetchOptionsToGetCollectionsOptions(input)
	setOptIdentity(opt, args, 1)
	cols, err := t.txn.GetCollections(ctx, opt)
	if err != nil {
		return js.Undefined(), err
	}
	wrappers := make([]any, len(cols))
	for i, col := range cols {
		wrappers[i] = newCollection(col, t.txns)
	}
	return js.ValueOf(wrappers), nil
}

func (t *transaction) listIndexes(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.ListIndexes()
	setOptIdentity(opt, args, 0)
	indexes, err := t.txn.ListIndexes(ctx, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(indexes)
}

func (t *transaction) listAllEncryptedIndexes(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.ListAllEncryptedIndexes()
	setOptIdentity(opt, args, 0)
	indexes, err := t.txn.ListAllEncryptedIndexes(ctx, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(indexes)
}

func (t *transaction) execRequest(this js.Value, args []js.Value) (js.Value, error) {
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
	ctx, err := contextArg(args, 2, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	if opt == nil {
		opt = options.ExecRequest()
	}
	setOptIdentity(opt, args, 2)
	res := t.txn.ExecRequest(ctx, request, opt)
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

func (t *transaction) addDACPolicy(this js.Value, args []js.Value) (js.Value, error) {
	policy, err := stringArg(args, 0, "policy")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.AddDACPolicy()
	setOptIdentity(opt, args, 1)
	res, err := t.txn.AddDACPolicy(ctx, policy, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) addDACActorRelationship(this js.Value, args []js.Value) (js.Value, error) {
	collectionName, err := stringArg(args, 0, "collectionName")
	if err != nil {
		return js.Undefined(), err
	}
	docID, err := stringArg(args, 1, "docID")
	if err != nil {
		return js.Undefined(), err
	}
	relation, err := stringArg(args, 2, "relation")
	if err != nil {
		return js.Undefined(), err
	}
	targetActor, err := stringArg(args, 3, "targetActor")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 4, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.AddDACActorRelationship()
	setOptIdentity(opt, args, 4)
	res, err := t.txn.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) deleteDACActorRelationship(this js.Value, args []js.Value) (js.Value, error) {
	collectionName, err := stringArg(args, 0, "collectionName")
	if err != nil {
		return js.Undefined(), err
	}
	docID, err := stringArg(args, 1, "docID")
	if err != nil {
		return js.Undefined(), err
	}
	relation, err := stringArg(args, 2, "relation")
	if err != nil {
		return js.Undefined(), err
	}
	targetActor, err := stringArg(args, 3, "targetActor")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 4, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.DeleteDACActorRelationship()
	setOptIdentity(opt, args, 4)
	res, err := t.txn.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) getNACStatus(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.GetNACStatus()
	setOptIdentity(opt, args, 0)
	res, err := t.txn.GetNACStatus(ctx, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) reEnableNAC(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.ReEnableNAC()
	setOptIdentity(opt, args, 0)
	err = t.txn.ReEnableNAC(ctx, opt)
	return js.Undefined(), err
}

func (t *transaction) disableNAC(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.DisableNAC()
	setOptIdentity(opt, args, 0)
	err = t.txn.DisableNAC(ctx, opt)
	return js.Undefined(), err
}

func (t *transaction) addNACActorRelationship(this js.Value, args []js.Value) (js.Value, error) {
	relation, err := stringArg(args, 0, "relation")
	if err != nil {
		return js.Undefined(), err
	}
	targetActor, err := stringArg(args, 1, "targetActor")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 2, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.AddNACActorRelationship()
	setOptIdentity(opt, args, 2)
	res, err := t.txn.AddNACActorRelationship(ctx, relation, targetActor, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) deleteNACActorRelationship(this js.Value, args []js.Value) (js.Value, error) {
	relation, err := stringArg(args, 0, "relation")
	if err != nil {
		return js.Undefined(), err
	}
	targetActor, err := stringArg(args, 1, "targetActor")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 2, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.DeleteNACActorRelationship()
	setOptIdentity(opt, args, 2)
	res, err := t.txn.DeleteNACActorRelationship(ctx, relation, targetActor, opt)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) getNodeIdentity(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.GetNodeIdentity(ctx)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) verifySignature(this js.Value, args []js.Value) (js.Value, error) {
	pubKeyHex, err := stringArg(args, 0, "publicKey")
	if err != nil {
		return js.Undefined(), err
	}
	pubKeyType, err := stringArg(args, 1, "publicKeyType")
	if pubKeyType == "" {
		pubKeyType = string(crypto.KeyTypeSecp256k1)
	}
	blockCID, err := stringArg(args, 2, "blockCID")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 3, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	pubKey, err := crypto.PublicKeyFromString(crypto.KeyType(pubKeyType), pubKeyHex)
	if err != nil {
		return js.Undefined(), err
	}
	opt := options.VerifySignature()
	setOptIdentity(opt, args, 3)
	err = t.txn.VerifySignature(ctx, blockCID, pubKey, opt)
	return js.Undefined(), err
}
