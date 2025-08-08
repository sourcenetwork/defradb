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
	"fmt"
	"sync"
	"syscall/js"

	"github.com/sourcenetwork/defradb/client"
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
		"patchSchema":                goji.Async(wrapper.patchSchema),
		"patchCollection":            goji.Async(wrapper.patchCollection),
		"setActiveSchemaVersion":     goji.Async(wrapper.setActiveSchemaVersion),
		"addView":                    goji.Async(wrapper.addView),
		"refreshViews":               goji.Async(wrapper.refreshViews),
		"setMigration":               goji.Async(wrapper.setMigration),
		"lensRegistry":               goji.Async(wrapper.lensRegistry),
		"getCollectionByName":        goji.Async(wrapper.getCollectionByName),
		"getCollections":             goji.Async(wrapper.getCollections),
		"getSchemaByVersionID":       goji.Async(wrapper.getSchemaByVersionID),
		"getSchemas":                 goji.Async(wrapper.getSchemas),
		"getAllIndexes":              goji.Async(wrapper.getAllIndexes),
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
	err := t.txn.Commit(context.Background())
	return js.Undefined(), err
}

func (t *transaction) discard(this js.Value, args []js.Value) (js.Value, error) {
	t.txn.Discard(context.Background())
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
	cols, err := t.txn.AddSchema(ctx, schema)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(cols)
}

func (t *transaction) patchSchema(this js.Value, args []js.Value) (js.Value, error) {
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
	ctx, err := contextArg(args, 3, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = t.txn.PatchSchema(ctx, patch, migration, setAsDefaultVersion)
	return js.Undefined(), err
}

func (t *transaction) patchCollection(this js.Value, args []js.Value) (js.Value, error) {
	patch, err := stringArg(args, 0, "patch")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = t.txn.PatchCollection(ctx, patch)
	return js.Undefined(), err
}

func (t *transaction) setActiveSchemaVersion(this js.Value, args []js.Value) (js.Value, error) {
	version, err := stringArg(args, 0, "version")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = t.txn.SetActiveSchemaVersion(ctx, version)
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
	var transform immutable.Option[model.Lens]
	if err := structArg(args, 2, "transform", &transform); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 3, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	cols, err := t.txn.AddView(ctx, gqlQuery, sdl, transform)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(cols)
}

func (t *transaction) refreshViews(this js.Value, args []js.Value) (js.Value, error) {
	var options client.CollectionFetchOptions
	if err := structArg(args, 0, "options", &options); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = t.txn.RefreshViews(ctx, options)
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
	err = t.txn.SetMigration(ctx, config)
	return js.Undefined(), err
}

func (t *transaction) lensRegistry(this js.Value, args []js.Value) (js.Value, error) {
	return newLensRegistry(t.txn.LensRegistry(), t.txns), nil
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
	col, err := t.txn.GetCollectionByName(ctx, name)
	if err != nil {
		return js.Undefined(), err
	}
	return newCollection(col, t.txns), nil
}

func (t *transaction) getCollections(this js.Value, args []js.Value) (js.Value, error) {
	var options client.CollectionFetchOptions
	if err := structArg(args, 0, "options", &options); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	cols, err := t.txn.GetCollections(ctx, options)
	if err != nil {
		return js.Undefined(), err
	}
	wrappers := make([]any, len(cols))
	for i, col := range cols {
		wrappers[i] = newCollection(col, t.txns)
	}
	return js.ValueOf(wrappers), nil
}

func (t *transaction) getSchemaByVersionID(this js.Value, args []js.Value) (js.Value, error) {
	version, err := stringArg(args, 0, "version")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	schema, err := t.txn.GetSchemaByVersionID(ctx, version)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(schema)
}

func (t *transaction) getSchemas(this js.Value, args []js.Value) (js.Value, error) {
	var options client.SchemaFetchOptions
	if err := structArg(args, 0, "options", &options); err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	schemas, err := t.txn.GetSchemas(ctx, options)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(schemas)
}

func (t *transaction) getAllIndexes(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	indexes, err := t.txn.GetAllIndexes(ctx)
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
	ctx, err := contextArg(args, 2, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res := t.txn.ExecRequest(ctx, request, opts...)
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
	res, err := t.txn.AddDACPolicy(ctx, policy)
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
	res, err := t.txn.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
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
	res, err := t.txn.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
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
	res, err := t.txn.GetNACStatus(ctx)
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
	err = t.txn.ReEnableNAC(ctx)
	return js.Undefined(), err
}

func (t *transaction) disableNAC(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, t.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = t.txn.DisableNAC(ctx)
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
	res, err := t.txn.AddNACActorRelationship(ctx, relation, targetActor)
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
	res, err := t.txn.DeleteNACActorRelationship(ctx, relation, targetActor)
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
	err = t.txn.VerifySignature(ctx, blockCID, pubKey)
	return js.Undefined(), err
}
