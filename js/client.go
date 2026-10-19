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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/node"
)

type Client struct {
	node  *node.Node
	txns  *sync.Map
	funcs *funcs
}

// funcs tracks the js.Func values created for a client.
//
// syscall/js keeps every js.Func, and everything its closure references, alive until
// Release is called. Without releasing them a closed client keeps its whole node in memory.
//
// Funcs created for transactions and collections are also only released when the client
// closes, so a long-lived client still accumulates them until then.
type funcs struct {
	mu       sync.Mutex
	list     []js.Func
	released bool
}

func (f *funcs) async(fn func(this js.Value, args []js.Value) (js.Value, error)) js.Func {
	jsFn := goji.Async(fn)
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.released {
		// A call that was still running when the client closed must not
		// register funcs that would never be released.
		jsFn.Release()
		return jsFn
	}
	f.list = append(f.list, jsFn)
	return jsFn
}

func (f *funcs) release() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, fn := range f.list {
		fn.Release()
	}
	f.list = nil
	f.released = true
}

// NewClient returns a new JS client.
//
// This method is public so that it can
// be reused within the test framework.
func NewClient(n *node.Node) *Client {
	return &Client{
		node:  n,
		txns:  &sync.Map{},
		funcs: &funcs{},
	}
}

func (c *Client) JSValue() js.Value {
	return js.ValueOf(map[string]any{
		"addCollection":              c.funcs.async(c.addCollection),
		"patchCollection":            c.funcs.async(c.patchCollection),
		"deleteCollection":           c.funcs.async(c.deleteCollection),
		"setActiveCollectionVersion": c.funcs.async(c.setActiveCollectionVersion),
		"addView":                    c.funcs.async(c.addView),
		"refreshViews":               c.funcs.async(c.refreshViews),
		"setMigration":               c.funcs.async(c.setMigration),
		"addLens":                    c.funcs.async(c.addLens),
		"listLenses":                 c.funcs.async(c.listLenses),
		"listActions":                c.funcs.async(c.listActions),
		"getCollectionByName":        c.funcs.async(c.getCollectionByName),
		"getCollections":             c.funcs.async(c.getCollections),
		"listIndexes":                c.funcs.async(c.listIndexes),
		"listAllEncryptedIndexes":    c.funcs.async(c.listAllEncryptedIndexes),
		"execRequest":                c.funcs.async(c.execRequest),
		"printDump":                  c.funcs.async(c.printDump),
		"basicImport":                c.funcs.async(c.basicImport),
		"basicExport":                c.funcs.async(c.basicExport),
		"addDACPolicy":               c.funcs.async(c.addDACPolicy),
		"verifyDACAccess":            c.funcs.async(c.verifyDACAccess),
		"addDACActorRelationship":    c.funcs.async(c.addDACActorRelationship),
		"deleteDACActorRelationship": c.funcs.async(c.deleteDACActorRelationship),
		"getNACStatus":               c.funcs.async(c.getNACStatus),
		"reEnableNAC":                c.funcs.async(c.reEnableNAC),
		"disableNAC":                 c.funcs.async(c.disableNAC),
		"addNACActorRelationship":    c.funcs.async(c.addNACActorRelationship),
		"deleteNACActorRelationship": c.funcs.async(c.deleteNACActorRelationship),
		"getNodeIdentity":            c.funcs.async(c.getNodeIdentity),
		"newTxn":                     c.funcs.async(c.newTxn),
		"verifySignature":            c.funcs.async(c.verifySignature),
		"peerInfo":                   c.funcs.async(c.peerInfo),
		"activePeers":                c.funcs.async(c.activePeers),
		"connect":                    c.funcs.async(c.connect),
		"disconnect":                 c.funcs.async(c.disconnect),
		"addReplicator":              c.funcs.async(c.addReplicator),
		"deleteReplicator":           c.funcs.async(c.deleteReplicator),
		"listReplicators":            c.funcs.async(c.listReplicators),
		"addP2PCollections":          c.funcs.async(c.addP2PCollections),
		"deleteP2PCollections":       c.funcs.async(c.deleteP2PCollections),
		"listP2PCollections":         c.funcs.async(c.listP2PCollections),
		"addP2PDocuments":            c.funcs.async(c.addP2PDocuments),
		"deleteP2PDocuments":         c.funcs.async(c.deleteP2PDocuments),
		"listP2PDocuments":           c.funcs.async(c.listP2PDocuments),
		"syncDocuments":              c.funcs.async(c.syncDocuments),
		"syncCollectionVersions":     c.funcs.async(c.syncCollectionVersions),
		"syncBranchableCollection":   c.funcs.async(c.syncBranchableCollection),
		"close":                      c.funcs.async(c.close),
	})
}

// Close closes the node and releases the js functions created for this client.
func (c *Client) Close(ctx context.Context) error {
	defer c.funcs.release()
	return c.node.Close(ctx)
}

func (c *Client) Transaction(id uint64) (client.Txn, error) {
	tx, ok := c.txns.Load(id)
	if !ok {
		return nil, ErrInvalidTransactionId
	}
	return tx.(client.Txn), nil //nolint:forcetypeassert
}
