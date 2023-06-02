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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
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
	err := c.applyDelete(ctx, txn, key)
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
		err := c.applyDelete(ctx, txn, dsKey)
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

	err = selectionPlan.Init()
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
		err = c.applyDelete(ctx, txn, key)
		if err != nil {
			return nil, err
		}

		// Add key of successfully deleted document to our list.
		results.DocKeys = append(results.DocKeys, docKey)
	}

	results.Count = int64(len(results.DocKeys))

	return results, nil
}

func (c *collection) applyDelete(
	ctx context.Context,
	txn datastore.Txn,
	key core.PrimaryDataStoreKey,
) error {
	found, isDeleted, err := c.exists(ctx, txn, key)
	if err != nil {
		return err
	}
	if !found {
		return client.ErrDocumentNotFound
	}
	if isDeleted {
		return NewErrDocumentDeleted(key.DocKey)
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
		client.Deleted,
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
