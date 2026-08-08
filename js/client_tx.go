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
	"sync"
	"syscall/js"

	"github.com/sourcenetwork/goji"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
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
		"addCollection":              goji.Async(wrapper.addCollection),
		"patchCollection":            goji.Async(wrapper.patchCollection),
		"deleteCollection":           goji.Async(wrapper.deleteCollection),
		"setActiveCollectionVersion": goji.Async(wrapper.setActiveCollectionVersion),
		"addView":                    goji.Async(wrapper.addView),
		"refreshViews":               goji.Async(wrapper.refreshViews),
		"setMigration":               goji.Async(wrapper.setMigration),
		"addLens":                    goji.Async(wrapper.addLens),
		"listLenses":                 goji.Async(wrapper.listLenses),
		"getCollectionByName":        goji.Async(wrapper.getCollectionByName),
		"getCollections":             goji.Async(wrapper.getCollections),
		"purgeDocuments":             goji.Async(wrapper.purgeDocuments),
		"listIndexes":                goji.Async(wrapper.listIndexes),
		"listAllEncryptedIndexes":    goji.Async(wrapper.listAllEncryptedIndexes),
		"execRequest":                goji.Async(wrapper.execRequest),
		"printDump":                  goji.Async(wrapper.printDump),
		"basicImport":                goji.Async(wrapper.basicImport),
		"basicExport":                goji.Async(wrapper.basicExport),
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
		"peerInfo":                   goji.Async(wrapper.peerInfo),
		"activePeers":                goji.Async(wrapper.activePeers),
		"connect":                    goji.Async(wrapper.connect),
		"disconnect":                 goji.Async(wrapper.disconnect),
		"addReplicator":              goji.Async(wrapper.addReplicator),
		"deleteReplicator":           goji.Async(wrapper.deleteReplicator),
		"listReplicators":            goji.Async(wrapper.listReplicators),
		"addP2PCollections":          goji.Async(wrapper.addP2PCollections),
		"deleteP2PCollections":       goji.Async(wrapper.deleteP2PCollections),
		"listP2PCollections":         goji.Async(wrapper.listP2PCollections),
		"addP2PDocuments":            goji.Async(wrapper.addP2PDocuments),
		"deleteP2PDocuments":         goji.Async(wrapper.deleteP2PDocuments),
		"listP2PDocuments":           goji.Async(wrapper.listP2PDocuments),
		"syncDocuments":              goji.Async(wrapper.syncDocuments),
		"syncCollectionVersions":     goji.Async(wrapper.syncCollectionVersions),
		"syncBranchableCollection":   goji.Async(wrapper.syncBranchableCollection),
	})
}

func (t *transaction) commit(this js.Value, args []js.Value) (js.Value, error) {
	return js.Undefined(), t.txn.Commit()
}

func (t *transaction) discard(this js.Value, args []js.Value) (js.Value, error) {
	t.txn.Discard()
	return js.Undefined(), nil
}

func (t *transaction) addCollection(this js.Value, args []js.Value) (js.Value, error) {
	sdl, err := stringArg(args, 0, "sdl")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.AddCollectionOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	cols, err := t.txn.AddCollection(context.Background(), sdl, asOpts(opt))
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
	optsVal := optionsValue(args, 2)
	var opt options.PatchCollectionOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.PatchCollection(context.Background(), patch, migration, asOpts(opt))
}

func (t *transaction) deleteCollection(this js.Value, args []js.Value) (js.Value, error) {
	names, err := stringSliceArg(args, 0, "names")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.DeleteCollectionOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.DeleteCollection(context.Background(), names, asOpts(opt))
}

func (t *transaction) setActiveCollectionVersion(this js.Value, args []js.Value) (js.Value, error) {
	version, err := stringArg(args, 0, "version")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.SetActiveCollectionVersionOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.SetActiveCollectionVersion(context.Background(), version, asOpts(opt))
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
	optsVal := optionsValue(args, 2)
	var opt options.AddViewOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	cols, err := t.txn.AddView(context.Background(), gqlQuery, sdl, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(cols)
}

func (t *transaction) refreshViews(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.RefreshViewsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.RefreshViews(context.Background(), asOpts(opt))
}

func (t *transaction) setMigration(this js.Value, args []js.Value) (js.Value, error) {
	var config client.LensConfig
	if err := structArg(args, 0, "config", &config); err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.SetMigrationOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	lensID, err := t.txn.SetMigration(context.Background(), config, asOpts(opt))
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
	optsVal := optionsValue(args, 1)
	var opt options.AddLensOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	lensID, err := t.txn.AddLens(context.Background(), lens, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return js.ValueOf(lensID), err
}

func (t *transaction) listLenses(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.ListLensesOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	lenses, err := t.txn.ListLenses(context.Background(), asOpts(opt))
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
	optsVal := optionsValue(args, 1)
	var opt options.GetCollectionByNameOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	col, err := t.txn.GetCollectionByName(context.Background(), name, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return newCollection(col), nil
}

func (t *transaction) getCollections(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.GetCollectionsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	cols, err := t.txn.GetCollections(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	wrappers := make([]any, len(cols))
	for i, col := range cols {
		wrappers[i] = newCollection(col)
	}
	return js.ValueOf(wrappers), nil
}

func (t *transaction) purgeDocuments(this js.Value, args []js.Value) (js.Value, error) {
	return purgeDocuments(t.txn, args)
}

func (t *transaction) listIndexes(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.ListIndexesOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	indexes, err := t.txn.ListIndexes(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(indexes)
}

func (t *transaction) listAllEncryptedIndexes(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.ListAllEncryptedIndexesOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	indexes, err := t.txn.ListAllEncryptedIndexes(context.Background(), asOpts(opt))
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
	optsVal := optionsValue(args, 1)
	var opt options.ExecRequestOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res := t.txn.ExecRequest(context.Background(), request, asOpts(opt))
	return marshalRequestResult(res)
}

func (t *transaction) printDump(this js.Value, args []js.Value) (js.Value, error) {
	return js.Undefined(), t.txn.PrintDump(context.Background())
}

func (t *transaction) basicImport(this js.Value, args []js.Value) (js.Value, error) {
	filepath, err := stringArg(args, 0, "filepath")
	if err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.BasicImport(context.Background(), filepath)
}

func (t *transaction) basicExport(this js.Value, args []js.Value) (js.Value, error) {
	filepath, err := stringArg(args, 0, "filepath")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.BasicExportOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.BasicExport(context.Background(), filepath, asOpts(opt))
}

func (t *transaction) addDACPolicy(this js.Value, args []js.Value) (js.Value, error) {
	policy, err := stringArg(args, 0, "policy")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.AddDACPolicyOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.AddDACPolicy(context.Background(), policy, asOpts(opt))
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
	optsVal := optionsValue(args, 4)
	var opt options.AddDACActorRelationshipOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.AddDACActorRelationship(context.Background(), collectionName, docID, relation, targetActor, asOpts(opt))
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
	optsVal := optionsValue(args, 4)
	var opt options.DeleteDACActorRelationshipOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.DeleteDACActorRelationship(context.Background(), collectionName, docID, relation, targetActor, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) getNACStatus(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.GetNACStatusOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.GetNACStatus(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) reEnableNAC(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.ReEnableNACOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.ReEnableNAC(context.Background(), asOpts(opt))
}

func (t *transaction) disableNAC(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.DisableNACOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.DisableNAC(context.Background(), asOpts(opt))
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
	optsVal := optionsValue(args, 2)
	var opt options.AddNACActorRelationshipOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.AddNACActorRelationship(context.Background(), relation, targetActor, asOpts(opt))
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
	optsVal := optionsValue(args, 2)
	var opt options.DeleteNACActorRelationshipOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.DeleteNACActorRelationship(context.Background(), relation, targetActor, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) getNodeIdentity(this js.Value, args []js.Value) (js.Value, error) {
	res, err := t.txn.GetNodeIdentity(context.Background())
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
	pubKeyType := optionalStringArg(args, 1)
	if pubKeyType == "" {
		pubKeyType = string(crypto.KeyTypeSecp256k1)
	}
	blockCID, err := stringArg(args, 2, "blockCID")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 3)
	pubKey, err := crypto.PublicKeyFromString(crypto.KeyType(pubKeyType), pubKeyHex)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.VerifySignatureOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.VerifySignature(context.Background(), blockCID, pubKey, asOpts(opt))
}

func (t *transaction) peerInfo(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.PeerInfoOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.PeerInfo(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) activePeers(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.ActivePeersOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.ActivePeers(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) connect(this js.Value, args []js.Value) (js.Value, error) {
	addresses, err := stringSliceArg(args, 0, "addresses")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.ConnectOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.Connect(context.Background(), addresses, asOpts(opt))
}

func (t *transaction) disconnect(this js.Value, args []js.Value) (js.Value, error) {
	addresses, err := stringSliceArg(args, 0, "addresses")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.DisconnectOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.Disconnect(context.Background(), addresses, asOpts(opt))
}

func (t *transaction) addReplicator(this js.Value, args []js.Value) (js.Value, error) {
	addresses, err := stringSliceArg(args, 0, "addresses")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.AddReplicatorOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.AddReplicator(context.Background(), addresses, asOpts(opt))
}

func (t *transaction) deleteReplicator(this js.Value, args []js.Value) (js.Value, error) {
	id, err := stringArg(args, 0, "id")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.DeleteReplicatorOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.DeleteReplicator(context.Background(), id, asOpts(opt))
}

func (t *transaction) listReplicators(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.ListReplicatorsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.ListReplicators(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) addP2PCollections(this js.Value, args []js.Value) (js.Value, error) {
	collectionIDs, err := stringSliceArg(args, 0, "collectionIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.AddP2PCollectionsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.AddP2PCollections(context.Background(), collectionIDs, asOpts(opt))
}

func (t *transaction) deleteP2PCollections(this js.Value, args []js.Value) (js.Value, error) {
	collectionIDs, err := stringSliceArg(args, 0, "collectionIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.DeleteP2PCollectionsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.DeleteP2PCollections(context.Background(), collectionIDs, asOpts(opt))
}

func (t *transaction) listP2PCollections(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.ListP2PCollectionsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.ListP2PCollections(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) addP2PDocuments(this js.Value, args []js.Value) (js.Value, error) {
	docIDs, err := stringSliceArg(args, 0, "docIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.AddP2PDocumentsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.AddP2PDocuments(context.Background(), docIDs, asOpts(opt))
}

func (t *transaction) deleteP2PDocuments(this js.Value, args []js.Value) (js.Value, error) {
	docIDs, err := stringSliceArg(args, 0, "docIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.DeleteP2PDocumentsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.DeleteP2PDocuments(context.Background(), docIDs, asOpts(opt))
}

func (t *transaction) listP2PDocuments(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	var opt options.ListP2PDocumentsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := t.txn.ListP2PDocuments(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (t *transaction) syncDocuments(this js.Value, args []js.Value) (js.Value, error) {
	collectionName, err := stringArg(args, 0, "collectionName")
	if err != nil {
		return js.Undefined(), err
	}
	docIDs, err := stringSliceArg(args, 1, "docIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 2)
	var opt options.SyncDocumentsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.SyncDocuments(context.Background(), collectionName, docIDs, asOpts(opt))
}

func (t *transaction) syncCollectionVersions(this js.Value, args []js.Value) (js.Value, error) {
	versionIDs, err := stringSliceArg(args, 0, "versionIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.SyncCollectionVersionsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.SyncCollectionVersions(context.Background(), versionIDs, asOpts(opt))
}

func (t *transaction) syncBranchableCollection(this js.Value, args []js.Value) (js.Value, error) {
	collectionID, err := stringArg(args, 0, "collectionID")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	var opt options.SyncBranchableCollectionOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), t.txn.SyncBranchableCollection(context.Background(), collectionID, asOpts(opt))
}
