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
// Target can be a Filter statement, a single DocID, a single document,
// an array of DocIDs, or an array of documents.
//
// If you want more type safety, use the respective typed versions of Delete.
// Eg: DeleteWithFilter or DeleteWithDocID
func (c *collection) DeleteWith(
	ctx context.Context,
	target any,
) (*client.DeleteResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.DeleteWithFilter(ctx, t)
	case client.DocID:
		return c.DeleteWithDocID(ctx, t)
	case []client.DocID:
		return c.DeleteWithDocIDs(ctx, t)
	default:
		return nil, client.ErrInvalidDeleteTarget
	}
}

// DeleteWithDocID deletes using a DocID to target a single document for delete.
func (c *collection) DeleteWithDocID(
	ctx context.Context,
	docID client.DocID,
) (*client.DeleteResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}

	defer c.discardImplicitTxn(ctx, txn)

	dsKey := c.getPrimaryKeyFromDocID(docID)
	res, err := c.deleteWithKey(ctx, txn, dsKey, client.Deleted)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

// DeleteWithDocIDs is the same as DeleteWithDocID but accepts multiple DocIDs as a slice.
func (c *collection) DeleteWithDocIDs(
	ctx context.Context,
	docIDs []client.DocID,
) (*client.DeleteResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}

	defer c.discardImplicitTxn(ctx, txn)

	res, err := c.deleteWithIDs(ctx, txn, docIDs, client.Deleted)
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
	// Check the key we have been given to delete with actually has a corresponding
	//  document (i.e. document actually exists in the collection).
	err := c.applyDelete(ctx, txn, key)
	if err != nil {
		return nil, err
	}

	// Upon successfull deletion, record a summary.
	results := &client.DeleteResult{
		Count:  1,
		DocIDs: []string{key.DocID},
	}

	return results, nil
}

func (c *collection) deleteWithIDs(
	ctx context.Context,
	txn datastore.Txn,
	docIDs []client.DocID,
	status client.DocumentStatus,
) (*client.DeleteResult, error) {
	results := &client.DeleteResult{
		DocIDs: make([]string, 0),
	}

	for _, docID := range docIDs {
		pdsKey := c.getPrimaryKeyFromDocID(docID)

		// Apply the function that will perform the full deletion of this document.
		err := c.applyDelete(ctx, txn, pdsKey)
		if err != nil {
			return nil, err
		}

		// Add this deleted docID to our list.
		results.DocIDs = append(results.DocIDs, docID.String())
	}

	// Upon successfull deletion, record a summary of how many we deleted.
	results.Count = int64(len(results.DocIDs))

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
		DocIDs: make([]string, 0),
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

		// Extract the docID in the string format from the document value.
		docID := doc.GetID()

		pdsKey := core.PrimaryDataStoreKey{
			CollectionId: fmt.Sprint(c.ID()),
			DocID:        docID,
		}

		// Delete the document that is associated with this DS key we got from the filter.
		err = c.applyDelete(ctx, txn, pdsKey)
		if err != nil {
			return nil, err
		}

		// Add docID of successfully deleted document to our list.
		results.DocIDs = append(results.DocIDs, docID)
	}

	results.Count = int64(len(results.DocIDs))

	return results, nil
}

func (c *collection) applyDelete(
	ctx context.Context,
	txn datastore.Txn,
	pdsKey core.PrimaryDataStoreKey,
) error {
	found, isDeleted, err := c.exists(ctx, txn, pdsKey)
	if err != nil {
		return err
	}
	if !found {
		return client.ErrDocumentNotFound
	}
	if isDeleted {
		return NewErrDocumentDeleted(pdsKey.DocID)
	}

	dsKey := pdsKey.ToDataStoreKey()

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

	headNode, priority, err := c.saveCompositeToMerkleCRDT(
		ctx,
		txn,
		dsKey,
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
						DocID:      pdsKey.DocID,
						Cid:        headNode.Cid(),
						SchemaRoot: c.Schema().Root,
						Block:      headNode,
						Priority:   priority,
					},
				)
			},
		)
	}

	return nil
}
