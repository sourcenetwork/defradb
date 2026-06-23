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
	"syscall/js"

	"github.com/sourcenetwork/goji"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
)

func (c *Client) addDACPolicy(this js.Value, args []js.Value) (js.Value, error) {
	policy, err := stringArg(args, 0, "policy")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 1)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.AddDACPolicyOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := store.AddDACPolicy(context.Background(), policy, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) addDACActorRelationship(this js.Value, args []js.Value) (js.Value, error) {
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
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.AddDACActorRelationshipOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := store.AddDACActorRelationship(context.Background(), collectionName, docID, relation, targetActor, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) deleteDACActorRelationship(this js.Value, args []js.Value) (js.Value, error) {
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
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.DeleteDACActorRelationshipOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := store.DeleteDACActorRelationship(context.Background(), collectionName, docID, relation, targetActor, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) verifyDACAccess(this js.Value, args []js.Value) (js.Value, error) {
	permission, err := stringArg(args, 0, "permission")
	if err != nil {
		return js.Undefined(), err
	}
	actorID, err := stringArg(args, 1, "actorID")
	if err != nil {
		return js.Undefined(), err
	}
	policyID, err := stringArg(args, 2, "policyID")
	if err != nil {
		return js.Undefined(), err
	}
	resourceName, err := stringArg(args, 3, "resourceName")
	if err != nil {
		return js.Undefined(), err
	}
	docID, err := stringArg(args, 4, "docID")
	if err != nil {
		return js.Undefined(), err
	}
	if !c.node.DB.DocumentACP().HasValue() {
		return js.Undefined(), fmt.Errorf("ACP system not available")
	}
	var docPermission acpTypes.DocumentResourcePermission
	switch permission {
	case "read":
		docPermission = acpTypes.DocumentReadPerm
	case "update":
		docPermission = acpTypes.DocumentUpdatePerm
	case "delete":
		docPermission = acpTypes.DocumentDeletePerm
	default:
		return js.Undefined(), fmt.Errorf("invalid permission: %s", permission)
	}
	hasAccess, err := c.node.DB.DocumentACP().Value().CheckDocAccess(
		context.Background(), docPermission, actorID, policyID, resourceName, docID,
	)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(map[string]any{
		"hasAccess": hasAccess,
	})
}

func (c *Client) getNACStatus(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.GetNACStatusOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := store.GetNACStatus(context.Background(), asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) reEnableNAC(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.ReEnableNACOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), store.ReEnableNAC(context.Background(), asOpts(opt))
}

func (c *Client) disableNAC(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.DisableNACOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), store.DisableNAC(context.Background(), asOpts(opt))
}

func (c *Client) addNACActorRelationship(this js.Value, args []js.Value) (js.Value, error) {
	relation, err := stringArg(args, 0, "relation")
	if err != nil {
		return js.Undefined(), err
	}
	targetActor, err := stringArg(args, 1, "targetActor")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 2)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.AddNACActorRelationshipOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := store.AddNACActorRelationship(context.Background(), relation, targetActor, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) deleteNACActorRelationship(this js.Value, args []js.Value) (js.Value, error) {
	relation, err := stringArg(args, 0, "relation")
	if err != nil {
		return js.Undefined(), err
	}
	targetActor, err := stringArg(args, 1, "targetActor")
	if err != nil {
		return js.Undefined(), err
	}
	optsVal := optionsValue(args, 2)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.DeleteNACActorRelationshipOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	res, err := store.DeleteNACActorRelationship(context.Background(), relation, targetActor, asOpts(opt))
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) getNodeIdentity(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res, err := store.GetNodeIdentity(context.Background())
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) newTxn(this js.Value, args []js.Value) (js.Value, error) {
	readOnly, err := boolArg(args, 0, "readOnly")
	if err != nil {
		return js.Undefined(), err
	}
	txn, err := c.node.DB.NewTxn(readOnly)
	if err != nil {
		return js.Undefined(), err
	}
	c.txns.Store(txn.ID(), txn)
	return newTransaction(txn, c.txns), nil
}

func (c *Client) verifySignature(this js.Value, args []js.Value) (js.Value, error) {
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
	store, err := optionsStore(c.node.DB, optsVal, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	pubKey, err := crypto.PublicKeyFromString(crypto.KeyType(pubKeyType), pubKeyHex)
	if err != nil {
		return js.Undefined(), err
	}
	var opt options.VerifySignatureOptions
	if err := parseOptions(optsVal, &opt); err != nil {
		return js.Undefined(), err
	}
	return js.Undefined(), store.VerifySignature(context.Background(), blockCID, pubKey, asOpts(opt))
}

func (c *Client) close(this js.Value, args []js.Value) (js.Value, error) {
	return js.Undefined(), c.node.Close(context.Background())
}
