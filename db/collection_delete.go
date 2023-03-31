// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"fmt"

	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	query "github.com/ipfs/go-datastore/query"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-libipfs/blocks"
	dag "github.com/ipfs/go-merkledag"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/merkle/clock"
)

// DeleteWith deletes a target document.
//
// Target can be a Filter statement, a single docKey, a single document,
// an array of docKeys, or an array of documents.
//
// If you want more type safety, use the respective typed versions of Delete.
// Eg: DeleteWithFilter or DeleteWithKey
func (c *collection) DeleteWith(
	ctx context.Context,
	target any,
) (*client.DeleteResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.DeleteWithFilter(ctx, t)
	case client.DocKey:
		return c.DeleteWithKey(ctx, t)
	case []client.DocKey:
		return c.DeleteWithKeys(ctx, t)
	default:
		return nil, client.ErrInvalidDeleteTarget
	}
}

// DeleteWithKey deletes using a DocKey to target a single document for delete.
func (c *collection) DeleteWithKey(
	ctx context.Context,
	key client.DocKey,
) (*client.DeleteResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}

	defer c.discardImplicitTxn(ctx, txn)

	dsKey := c.getPrimaryKeyFromDocKey(key)
	res, err := c.deleteWithKey(ctx, txn, dsKey, client.Deleted)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

// DeleteWithKeys is the same as DeleteWithKey but accepts multiple keys as a slice.
func (c *collection) DeleteWithKeys(
	ctx context.Context,
	keys []client.DocKey,
) (*client.DeleteResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}

	defer c.discardImplicitTxn(ctx, txn)

	res, err := c.deleteWithKeys(ctx, txn, keys, client.Deleted)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

// DeleteWithFilter deletes using a filter to target documents for delete.
func (c *collection) DeleteWithFilter(
	ctx context.Context,
	filter any,
) (*client.DeleteResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}

	defer c.discardImplicitTxn(ctx, txn)

	res, err := c.deleteWithFilter(ctx, txn, filter, client.Deleted)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

func (c *collection) deleteWithKey(
	ctx context.Context,
	txn datastore.Txn,
	key core.PrimaryDataStoreKey,
	status client.DocumentStatus,
) (*client.DeleteResult, error) {
	// Check the docKey we have been given to delete with actually has a corresponding
	//  document (i.e. document actually exists in the collection).
	err := c.applyDelete(ctx, txn, key, status)
	if err != nil {
		return nil, err
	}

	// Upon successfull deletion, record a summary.
	results := &client.DeleteResult{
		Count:   1,
		DocKeys: []string{key.DocKey},
	}

	return results, nil
}

func (c *collection) deleteWithKeys(
	ctx context.Context,
	txn datastore.Txn,
	keys []client.DocKey,
	status client.DocumentStatus,
) (*client.DeleteResult, error) {
	results := &client.DeleteResult{
		DocKeys: make([]string, 0),
	}

	for _, key := range keys {
		dsKey := c.getPrimaryKeyFromDocKey(key)

		// Apply the function that will perform the full deletion of this document.
		err := c.applyDelete(ctx, txn, dsKey, status)
		if err != nil {
			return nil, err
		}

		// Add this deleted key to our list.
		results.DocKeys = append(results.DocKeys, key.String())
	}

	// Upon successfull deletion, record a summary of how many we deleted.
	results.Count = int64(len(results.DocKeys))

	return results, nil
}

func (c *collection) deleteWithFilter(
	ctx context.Context,
	txn datastore.Txn,
	filter any,
	status client.DocumentStatus,
) (*client.DeleteResult, error) {
	// Make a selection plan that will scan through only the documents with matching filter.
	selectionPlan, err := c.makeSelectionPlan(ctx, txn, filter)
	if err != nil {
		return nil, err
	}
	if err := selectionPlan.Start(); err != nil {
		return nil, err
	}

	// If the plan isn't properly closed at any exit point log the error.
	defer func() {
		if err := selectionPlan.Close(); err != nil {
			log.ErrorE(ctx, "Failed to close the request plan, after filter delete", err)
		}
	}()

	results := &client.DeleteResult{
		DocKeys: make([]string, 0),
	}

	// Keep looping until results from the selection plan have been iterated through.
	for {
		next, err := selectionPlan.Next()
		if err != nil {
			return nil, err
		}

		// If no results remaining / or gotten then break out of the loop.
		if !next {
			break
		}

		doc := selectionPlan.Value()
		// Extract the dockey in the string format from the document value.
		docKey := doc.GetKey()

		// Convert from string to client.DocKey.
		key := core.PrimaryDataStoreKey{
			CollectionId: fmt.Sprint(c.colID),
			DocKey:       docKey,
		}

		// Delete the document that is associated with this key we got from the filter.
		err = c.applyFullDelete(ctx, txn, key)
		if err != nil {
			return nil, err
		}

		// Add key of successfully deleted document to our list.
		results.DocKeys = append(results.DocKeys, docKey)
	}

	results.Count = int64(len(results.DocKeys))

	return results, nil
}

type dagDeleter struct {
	bstore datastore.DAGStore
}

func newDagDeleter(bstore datastore.DAGStore) dagDeleter {
	return dagDeleter{
		bstore: bstore,
	}
}

func isDeleteStatus(status client.DocumentStatus) bool {
	switch status {
	case client.Deleted:
		return true
	default:
		return false
	}
}

func (c *collection) applyDelete(
	ctx context.Context,
	txn datastore.Txn,
	key core.PrimaryDataStoreKey,
	status client.DocumentStatus,
) error {
	if !isDeleteStatus(status) {
		return NewErrInvalidDeleteStatus(status)
	}
	found, isDeleted, err := c.exists(ctx, txn, key)
	if err != nil {
		return err
	}
	if !found || isDeleted {
		return client.ErrDocumentNotFound
	}

	_, err = c.delete(ctx, txn, key)
	if err != nil {
		return err
	}

	dsKey := key.ToDataStoreKey()

	headset := clock.NewHeadSet(
		txn.Headstore(),
		dsKey.WithFieldId(core.COMPOSITE_NAMESPACE).ToHeadStoreKey(),
	)
	cids, _, err := headset.List(ctx)
	if err != nil {
		return err
	}

	dagLinks := make([]core.DAGLink, len(cids))
	for i, cid := range cids {
		dagLinks[i] = core.DAGLink{
			Name: core.HEAD,
			Cid:  cid,
		}
	}

	headNode, priority, err := c.saveValueToMerkleCRDT(
		ctx,
		txn,
		dsKey,
		client.COMPOSITE,
		[]byte{},
		dagLinks,
		status,
	)
	if err != nil {
		return err
	}

	if c.db.events.Updates.HasValue() {
		txn.OnSuccess(
			func() {
				c.db.events.Updates.Value().Publish(
					events.Update{
						DocKey:   key.DocKey,
						Cid:      headNode.Cid(),
						SchemaID: c.schemaID,
						Block:    headNode,
						Priority: priority,
					},
				)
			},
		)
	}

	return nil
}

// applyFullDelete deletes from all the stores.
//
// Stores the deletion will act upon:
//  1. Deleting the actual blocks (blockstore).
//  2. Deleting datastore state.
//  3. Deleting headstore state.
//
// Here is what our db stores look like:
//
//	/db
//	-> block /blocks => /db/blocks
//	-> datastore /data => /db/data
//	-> headstore /heads => /db/heads
//	-> systemstore /system => /db/system
func (c *collection) applyFullDelete(
	ctx context.Context,
	txn datastore.Txn, dockey core.PrimaryDataStoreKey) error {
	// Check the docKey we have been given to delete with actually has a corresponding
	//  document (i.e. document actually exists in the collection).
	found, _, err := c.exists(ctx, txn, dockey)
	if err != nil {
		return err
	}
	if !found {
		return client.ErrDocumentNotFound
	}

	// 1. =========================== Delete blockstore state ===========================
	// blocks: /db/blocks/CIQSDFKLJGHFKLGHGLHSKLHKJGS => KGLKJFHJKDLGKHDGLHGLFDHGLFDGKGHL

	// Covert dockey to compositeKey as follows:
	//  * dockey: bae-kljhLKHJG-lkjhgkldjhlzkdf-kdhflkhjsklgh-kjdhlkghjs
	//  => compositeKey: bae-kljhLKHJG-lkjhgkldjhlzkdf-kdhflkhjsklgh-kjdhlkghjs/C
	compositeKey := core.HeadStoreKey{
		DocKey:  dockey.DocKey,
		FieldId: core.COMPOSITE_NAMESPACE,
	}
	headset := clock.NewHeadSet(txn.Headstore(), compositeKey)

	// Get all the heads (cids).
	heads, _, err := headset.List(ctx)
	if err != nil {
		return NewErrFailedToGetHeads(err)
	}

	dagDel := newDagDeleter(txn.DAGstore())
	// Delete DAG of all heads (and the heads themselves)
	for _, head := range heads {
		if err = dagDel.run(ctx, head); err != nil {
			return err
		}
	} // ================================================ Successfully deleted the blocks

	// 2. =========================== Delete datastore state ============================
	_, err = c.delete(ctx, txn, dockey)
	if err != nil {
		return err
	}
	// ======================== Successfully deleted the datastore state of this document

	// 3. =========================== Delete headstore state ===========================
	headQuery := query.Query{
		Prefix:   dockey.ToString(),
		KeysOnly: true,
	}
	headResult, err := txn.Headstore().Query(ctx, headQuery)
	for e := range headResult.Next() {
		if e.Error != nil {
			return err
		}
		err = txn.Headstore().Delete(ctx, ds.NewKey(e.Key))
		if err != nil {
			return err
		}
	} // ====================== Successfully deleted the headstore state of this document

	return nil
}

func (d dagDeleter) run(ctx context.Context, targetCid cid.Cid) error {
	// Validate the cid.
	if targetCid == cid.Undef {
		return nil
	}

	// Get the block using the cid.
	block, err := d.bstore.Get(ctx, targetCid)
	if errors.Is(err, ipld.ErrNotFound{Cid: targetCid}) {
		// If we have multiple heads corresponding to a dockey, one of the heads
		//  could have already deleted the parental dag chain.
		// Example: in the diagram below, HEAD#1 with cid1 deleted (represented by `:x`)
		//          all the parental nodes. Currently HEAD#2 goes to delete
		//          itself (represented by `:d`) and it's parental nodes, but as we see
		//          the parents were already deleted by HEAD#1 so we just stop there.
		//
		//                                     | --> (E:x) HEAD#1->cid1
		// (A:x) --> (B:x) --> (C:x) --> (D:x) |
		//                                     | --> (F:d) HEAD#2->cid2
		return nil
	} else if err != nil {
		return err
	}

	// Attempt deleting the current block and it's links (in a mutally recursive fashion)
	return d.delete(ctx, targetCid, block)
}

// (ipld.Block
//	(ipldProtobufNode{
//	                  Data: (cbor(crdt deltaPayload)),
//	                  Links: (_head => parentCid, fieldName => fieldCid)))

func (d dagDeleter) delete(
	ctx context.Context,
	targetCid cid.Cid,
	targetBlock blocks.Block) error {
	targetNode, err := dag.DecodeProtobuf(targetBlock.RawData())
	if err != nil {
		return err
	}

	// delete current block
	if err := d.bstore.DeleteBlock(ctx, targetCid); err != nil {
		return err
	}

	for _, link := range targetNode.Links() {
		// Call run on all the links (eventually delete is called on them too.)
		if err := d.run(ctx, link.Cid); err != nil {
			return err
		}
	}

	return nil
}
