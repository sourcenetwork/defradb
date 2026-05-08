// Copyright 2026 Democratized Data Foundation
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

	"github.com/sourcenetwork/goji"

	"github.com/sourcenetwork/defradb/client/options"
)

func (c *Client) peerInfo(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.PeerInfoOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.PeerInfo(ctx, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) activePeers(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ActivePeersOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.ActivePeers(ctx, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) connect(this js.Value, args []js.Value) (js.Value, error) {
	addresses, err := stringSliceArg(args, 0, "addresses")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ConnectOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.node.DB.Connect(ctx, addresses, asOpts(opt))
}

func (c *Client) addReplicator(this js.Value, args []js.Value) (js.Value, error) {
	addresses, err := stringSliceArg(args, 0, "addresses")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.AddReplicatorOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.node.DB.AddReplicator(ctx, addresses, asOpts(opt))
}

func (c *Client) deleteReplicator(this js.Value, args []js.Value) (js.Value, error) {
	id, err := stringArg(args, 0, "id")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.DeleteReplicatorOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.node.DB.DeleteReplicator(ctx, id, asOpts(opt))
}

func (c *Client) listReplicators(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ListReplicatorsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.ListReplicators(ctx, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) addP2PCollections(this js.Value, args []js.Value) (js.Value, error) {
	collectionIDs, err := stringSliceArg(args, 0, "collectionIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.AddP2PCollectionsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.node.DB.AddP2PCollections(ctx, collectionIDs, asOpts(opt))
}

func (c *Client) deleteP2PCollections(this js.Value, args []js.Value) (js.Value, error) {
	collectionIDs, err := stringSliceArg(args, 0, "collectionIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.DeleteP2PCollectionsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.node.DB.DeleteP2PCollections(ctx, collectionIDs, asOpts(opt))
}

func (c *Client) listP2PCollections(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ListP2PCollectionsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.ListP2PCollections(ctx, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) addP2PDocuments(this js.Value, args []js.Value) (js.Value, error) {
	docIDs, err := stringSliceArg(args, 0, "docIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.AddP2PDocumentsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.node.DB.AddP2PDocuments(ctx, docIDs, asOpts(opt))
}

func (c *Client) deleteP2PDocuments(this js.Value, args []js.Value) (js.Value, error) {
	docIDs, err := stringSliceArg(args, 0, "docIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.DeleteP2PDocumentsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.node.DB.DeleteP2PDocuments(ctx, docIDs, asOpts(opt))
}

func (c *Client) listP2PDocuments(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ListP2PDocumentsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.ListP2PDocuments(ctx, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) syncDocuments(this js.Value, args []js.Value) (js.Value, error) {
	collectionName, err := stringArg(args, 0, "collectionName")
	if err != nil {
		return js.Undefined(), err
	}
	docIDs, err := stringSliceArg(args, 1, "docIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 2)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.SyncDocumentsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.node.DB.SyncDocuments(ctx, collectionName, docIDs, asOpts(opt))
}

func (c *Client) syncCollectionVersions(this js.Value, args []js.Value) (js.Value, error) {
	versionIDs, err := stringSliceArg(args, 0, "versionIDs")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.SyncCollectionVersionsOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.node.DB.SyncCollectionVersions(ctx, versionIDs, asOpts(opt))
}

func (c *Client) syncBranchableCollection(this js.Value, args []js.Value) (js.Value, error) {
	collectionID, err := stringArg(args, 0, "collectionID")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	ctx, err := makeContext(optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.SyncBranchableCollectionOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), c.node.DB.SyncBranchableCollection(ctx, collectionID, asOpts(opt))
}
