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

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/txnctx"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// DeleteWithFilter deletes using a filter to target documents for delete.
func (c *collection) DeleteWithFilter(
	ctx context.Context,
	filter any,
) (*client.DeleteResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	res, err := c.deleteWithFilter(ctx, filter, client.Deleted)
	if err != nil {
		return nil, err
	}

	return res, txn.Commit(ctx)
}

func (c *collection) deleteWithFilter(
	ctx context.Context,
	filter any,
	_ client.DocumentStatus,
) (*client.DeleteResult, error) {
	// Make a selection plan that will scan through only the documents with matching filter.
	selectionPlan, err := c.makeSelectionPlan(ctx, filter)
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
			log.ErrorContextE(ctx, "Failed to close the request plan, after filter delete", err)
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

		shortID, err := id.GetShortCollectionID(ctx, c.Version().CollectionID)
		if err != nil {
			return nil, err
		}

		primaryKey := keys.PrimaryDataStoreKey{
			CollectionShortID: shortID,
			DocID:             docID,
		}

		// Delete the document that is associated with this DS key we got from the filter.
		err = c.applyDelete(ctx, primaryKey)
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
	primaryKey keys.PrimaryDataStoreKey,
) error {
	// Must also have read permission to delete, inorder to check if document exists.
	found, isDeleted, err := c.exists(ctx, primaryKey)
	if err != nil {
		return err
	}
	if !found {
		return client.ErrDocumentNotFoundOrNotAuthorized
	}
	if isDeleted {
		return NewErrDocumentDeleted(primaryKey.DocID)
	}

	// Stop deletion of document if the correct permissions aren't there.
	canDelete, err := c.checkAccessOfDocWithACP(
		ctx,
		acpTypes.DocumentDeletePerm,
		primaryKey.DocID,
	)

	if err != nil {
		return err
	}
	if !canDelete {
		return client.ErrDocumentNotFoundOrNotAuthorized
	}

	txn := txnctx.MustGet(ctx)

	ident := identity.FromContext(ctx)
	if (!ident.HasValue() || ident.Value().PrivateKey == nil) && c.db.nodeIdentity.HasValue() {
		ctx = identity.WithContext(ctx, c.db.nodeIdentity)
	}

	if !c.db.signingDisabled {
		ctx = coreblock.ContextWithEnabledSigning(ctx)
	}

	merkleCRDT := crdt.NewDocComposite(
		txn.Datastore(),
		c.Schema().VersionID,
		primaryKey.ToDataStoreKey().WithFieldID(core.COMPOSITE_NAMESPACE),
	)

	link, b, err := coreblock.AddDelta(ctx, txn, merkleCRDT, merkleCRDT.DeleteDelta())
	if err != nil {
		return err
	}

	// publish an update event if the txn succeeds
	updateEvent := event.Update{
		DocID:        primaryKey.DocID,
		Cid:          link.Cid,
		CollectionID: c.Version().CollectionID,
		Block:        b,
	}
	txn.OnSuccess(func() {
		c.db.events.Publish(event.NewMessage(event.UpdateName, updateEvent))
	})

	if c.def.Version.IsBranchable {
		shortID, err := id.GetShortCollectionID(ctx, c.Version().CollectionID)
		if err != nil {
			return err
		}

		collectionCRDT := crdt.NewCollection(
			c.Schema().VersionID,
			keys.NewHeadstoreColKey(shortID),
		)

		link, headNode, err := coreblock.AddDelta(
			ctx,
			txn,
			collectionCRDT,
			collectionCRDT.Delta(),
			[]coreblock.DAGLink{{Link: link}}...,
		)
		if err != nil {
			return err
		}

		updateEvent := event.Update{
			Cid:          link.Cid,
			CollectionID: c.Version().CollectionID,
			Block:        headNode,
		}

		txn.OnSuccess(func() {
			c.db.events.Publish(event.NewMessage(event.UpdateName, updateEvent))
		})
	}

	return nil
}
