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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/goji"
)

func (c *Client) addDACPolicy(this js.Value, args []js.Value) (js.Value, error) {
	policy, err := stringArg(args, 0, "policy")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.AddDACPolicy(ctx, policy)
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
	ctx, err := contextArg(args, 4, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
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
	ctx, err := contextArg(args, 4, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
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
	ctx, err := contextArg(args, 5, c.txns)
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
	hasAccess, err := c.node.DB.DocumentACP().Value().CheckDocAccess(ctx, docPermission, actorID, policyID, resourceName, docID)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(map[string]interface{}{
		"hasAccess": hasAccess,
	})
}

func (c *Client) getNodeIdentity(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.GetNodeIdentity(ctx)
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
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	txn, err := c.node.DB.NewTxn(ctx, readOnly)
	if err != nil {
		return js.Undefined(), err
	}
	c.txns.Store(txn.ID(), txn)
	return newTransaction(txn, c.txns), nil
}

func (c *Client) newConcurrentTxn(this js.Value, args []js.Value) (js.Value, error) {
	readOnly, err := boolArg(args, 0, "readOnly")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 1, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	txn, err := c.node.DB.NewConcurrentTxn(ctx, readOnly)
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
	pubKeyType, err := stringArg(args, 1, "publicKeyType")
	if pubKeyType == "" {
		pubKeyType = string(crypto.KeyTypeSecp256k1)
	}
	blockCID, err := stringArg(args, 2, "blockCID")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 3, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	pubKey, err := crypto.PublicKeyFromString(crypto.KeyType(pubKeyType), pubKeyHex)
	if err != nil {
		return js.Undefined(), err
	}
	err = c.node.DB.VerifySignature(ctx, blockCID, pubKey)
	return js.Undefined(), err
}

func (c *Client) close(this js.Value, args []js.Value) (js.Value, error) {
	err := c.node.Close(context.Background())
	return js.Undefined(), err
}
