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
	"syscall/js"

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

func (c *Client) getAACStatus(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.GetAACStatus(ctx)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) reEnableAAC(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = c.node.DB.ReEnableAAC(ctx)
	return js.Undefined(), err
}

func (c *Client) disableAAC(this js.Value, args []js.Value) (js.Value, error) {
	ctx, err := contextArg(args, 0, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	err = c.node.DB.DisableAAC(ctx)
	return js.Undefined(), err
}

func (c *Client) addAACActorRelationship(this js.Value, args []js.Value) (js.Value, error) {
	relation, err := stringArg(args, 0, "relation")
	if err != nil {
		return js.Undefined(), err
	}
	targetActor, err := stringArg(args, 1, "targetActor")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.AddAACActorRelationship(ctx, relation, targetActor)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
}

func (c *Client) deleteAACActorRelationship(this js.Value, args []js.Value) (js.Value, error) {
	relation, err := stringArg(args, 0, "relation")
	if err != nil {
		return js.Undefined(), err
	}
	targetActor, err := stringArg(args, 1, "targetActor")
	if err != nil {
		return js.Undefined(), err
	}
	ctx, err := contextArg(args, 2, c.txns)
	if err != nil {
		return js.Undefined(), err
	}
	res, err := c.node.DB.DeleteAACActorRelationship(ctx, relation, targetActor)
	if err != nil {
		return js.Undefined(), err
	}
	return goji.MarshalJS(res)
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
