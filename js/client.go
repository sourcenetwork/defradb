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

	"github.com/sourcenetwork/goji"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/node"
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
		"addCollection":              goji.Async(c.addCollection),
		"patchCollection":            goji.Async(c.patchCollection),
		"deleteCollection":           goji.Async(c.deleteCollection),
		"setActiveCollectionVersion": goji.Async(c.setActiveCollectionVersion),
		"addView":                    goji.Async(c.addView),
		"refreshViews":               goji.Async(c.refreshViews),
		"setMigration":               goji.Async(c.setMigration),
		"addLens":                    goji.Async(c.addLens),
		"listLenses":                 goji.Async(c.listLenses),
		"getCollectionByName":        goji.Async(c.getCollectionByName),
		"getCollections":             goji.Async(c.getCollections),
		"listIndexes":                goji.Async(c.listIndexes),
		"listAllEncryptedIndexes":    goji.Async(c.listAllEncryptedIndexes),
		"execRequest":                goji.Async(c.execRequest),
		"printDump":                  goji.Async(c.printDump),
		"basicImport":                goji.Async(c.basicImport),
		"basicExport":                goji.Async(c.basicExport),
		"addDACPolicy":               goji.Async(c.addDACPolicy),
		"verifyDACAccess":            goji.Async(c.verifyDACAccess),
		"addDACActorRelationship":    goji.Async(c.addDACActorRelationship),
		"deleteDACActorRelationship": goji.Async(c.deleteDACActorRelationship),
		"getNACStatus":               goji.Async(c.getNACStatus),
		"reEnableNAC":                goji.Async(c.reEnableNAC),
		"disableNAC":                 goji.Async(c.disableNAC),
		"addNACActorRelationship":    goji.Async(c.addNACActorRelationship),
		"deleteNACActorRelationship": goji.Async(c.deleteNACActorRelationship),
		"getNodeIdentity":            goji.Async(c.getNodeIdentity),
		"newTxn":                     goji.Async(c.newTxn),
		"verifySignature":            goji.Async(c.verifySignature),
		"peerInfo":                   goji.Async(c.peerInfo),
		"activePeers":                goji.Async(c.activePeers),
		"connect":                    goji.Async(c.connect),
		"disconnect":                 goji.Async(c.disconnect),
		"addReplicator":              goji.Async(c.addReplicator),
		"deleteReplicator":           goji.Async(c.deleteReplicator),
		"listReplicators":            goji.Async(c.listReplicators),
		"addP2PCollections":          goji.Async(c.addP2PCollections),
		"deleteP2PCollections":       goji.Async(c.deleteP2PCollections),
		"listP2PCollections":         goji.Async(c.listP2PCollections),
		"addP2PDocuments":            goji.Async(c.addP2PDocuments),
		"deleteP2PDocuments":         goji.Async(c.deleteP2PDocuments),
		"listP2PDocuments":           goji.Async(c.listP2PDocuments),
		"syncDocuments":              goji.Async(c.syncDocuments),
		"syncCollectionVersions":     goji.Async(c.syncCollectionVersions),
		"syncBranchableCollection":   goji.Async(c.syncBranchableCollection),
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
