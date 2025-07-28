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
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/goji"
)

type Client struct {
	node *node.Node
	txns *sync.Map
}

// NewClient returns a new JS client.
//
// This method is public so that it can
// be reused within the test framework.
func NewClient(n *node.Node) *Client {
	return &Client{
		node: n,
		txns: &sync.Map{},
	}
}

func (c *Client) JSValue() js.Value {
	return js.ValueOf(map[string]any{
		"addSchema":                  goji.Async(c.addSchema),
		"patchSchema":                goji.Async(c.patchSchema),
		"patchCollection":            goji.Async(c.patchCollection),
		"setActiveSchemaVersion":     goji.Async(c.setActiveSchemaVersion),
		"addView":                    goji.Async(c.addView),
		"refreshViews":               goji.Async(c.refreshViews),
		"setMigration":               goji.Async(c.setMigration),
		"lensRegistry":               goji.Async(c.lensRegistry),
		"getCollectionByName":        goji.Async(c.getCollectionByName),
		"getCollections":             goji.Async(c.getCollections),
		"getSchemaByVersionID":       goji.Async(c.getSchemaByVersionID),
		"getSchemas":                 goji.Async(c.getSchemas),
		"getAllIndexes":              goji.Async(c.getAllIndexes),
		"execRequest":                goji.Async(c.execRequest),
		"addDACPolicy":               goji.Async(c.addDACPolicy),
		"verifyDACAccess":            goji.Async(c.verifyDACAccess),
		"addDACActorRelationship":    goji.Async(c.addDACActorRelationship),
		"deleteDACActorRelationship": goji.Async(c.deleteDACActorRelationship),
		"getNodeIdentity":            goji.Async(c.getNodeIdentity),
		"newTxn":                     goji.Async(c.newTxn),
		"newConcurrentTxn":           goji.Async(c.newConcurrentTxn),
		"verifySignature":            goji.Async(c.verifySignature),
		"close":                      goji.Async(c.close),
	})
}

func (c *Client) Transaction(id uint64) (client.Txn, error) {
	tx, ok := c.txns.Load(id)
	if !ok {
		return nil, ErrInvalidTransactionId
	}
	return tx.(client.Txn), nil //nolint:forcetypeassert
}
