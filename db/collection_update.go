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
	"strings"

	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/immutable"
	"github.com/valyala/fastjson"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/planner"
)

// UpdateWith updates a target document using the given updater type. Target
// can be a Filter statement, a single DocID, a single document,
// an array of DocIDs, or an array of documents.
// If you want more type safety, use the respective typed versions of Update.
// Eg: UpdateWithFilter or UpdateWithDocID
func (c *collection) UpdateWith(
	ctx context.Context,
	identity immutable.Option[string],
	target any,
	updater string,
) (*client.UpdateResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.UpdateWithFilter(ctx, identity, t, updater)
	case client.DocID:
		return c.UpdateWithDocID(ctx, identity, t, updater)
	case []client.DocID:
		return c.UpdateWithDocIDs(ctx, identity, t, updater)
	default:
		return nil, client.ErrInvalidUpdateTarget
	}
}

// UpdateWithFilter updates using a filter to target documents for update.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithFilter(
	ctx context.Context,
	identity immutable.Option[string],
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithFilter(ctx, identity, txn, filter, updater)
	if err != nil {
		return nil, err
	}
	return res, c.commitImplicitTxn(ctx, txn)
}

// UpdateWithDocID updates using a DocID to target a single document for update.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithDocID(
	ctx context.Context,
	identity immutable.Option[string],
	docID client.DocID,
	updater string,
) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithDocID(ctx, identity, txn, docID, updater)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

// UpdateWithDocIDs is the same as UpdateWithDocID but accepts multiple DocIDs as a slice.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithDocIDs(
	ctx context.Context,
	identity immutable.Option[string],
	docIDs []client.DocID,
	updater string,
) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithIDs(ctx, identity, txn, docIDs, updater)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

func (c *collection) updateWithDocID(
	ctx context.Context,
	identity immutable.Option[string],
	txn datastore.Txn,
	docID client.DocID,
	updater string,
) (*client.UpdateResult, error) {
	parsedUpdater, err := fastjson.Parse(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	if parsedUpdater.Type() == fastjson.TypeArray {
		isPatch = true
	} else if parsedUpdater.Type() != fastjson.TypeObject {
		return nil, client.ErrInvalidUpdater
	}

	doc, err := c.Get(ctx, identity, docID, false)
	if err != nil {
		return nil, err
	}

	if isPatch {
		// todo
	} else {
		err = doc.SetWithJSON([]byte(updater))
	}
	if err != nil {
		return nil, err
	}

	err = c.update(ctx, identity, txn, doc)
	if err != nil {
		return nil, err
	}

	results := &client.UpdateResult{
		Count:  1,
		DocIDs: []string{docID.String()},
	}
	return results, nil
}

func (c *collection) updateWithIDs(
	ctx context.Context,
	identity immutable.Option[string],
	txn datastore.Txn,
	docIDs []client.DocID,
	updater string,
) (*client.UpdateResult, error) {
	parsedUpdater, err := fastjson.Parse(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	if parsedUpdater.Type() == fastjson.TypeArray {
		isPatch = true
	} else if parsedUpdater.Type() != fastjson.TypeObject {
		return nil, client.ErrInvalidUpdater
	}

	results := &client.UpdateResult{
		DocIDs: make([]string, len(docIDs)),
	}
	for i, docIDs := range docIDs {
		doc, err := c.Get(ctx, identity, docIDs, false)
		if err != nil {
			return nil, err
		}

		if isPatch {
			// todo
		} else {
			err = doc.SetWithJSON([]byte(updater))
		}
		if err != nil {
			return nil, err
		}

		err = c.update(ctx, identity, txn, doc)
		if err != nil {
			return nil, err
		}

		results.DocIDs[i] = docIDs.String()
		results.Count++
	}
	return results, nil
}

func (c *collection) updateWithFilter(
	ctx context.Context,
	identity immutable.Option[string],
	txn datastore.Txn,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	parsedUpdater, err := fastjson.Parse(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	isMerge := false
	switch parsedUpdater.Type() {
	case fastjson.TypeArray:
		isPatch = true
	case fastjson.TypeObject:
		isMerge = true
	default:
		return nil, client.ErrInvalidUpdater
	}

	// Make a selection plan that will scan through only the documents with matching filter.
	selectionPlan, err := c.makeSelectionPlan(ctx, identity, txn, filter)
	if err != nil {
		return nil, err
	}

	err = selectionPlan.Init()
	if err != nil {
		return nil, err
	}

	if err = selectionPlan.Start(); err != nil {
		return nil, err
	}

	// If the plan isn't properly closed at any exit point log the error.
	defer func() {
		if err := selectionPlan.Close(); err != nil {
			log.ErrorContextE(ctx, "Failed to close the selection plan, after filter update", err)
		}
	}()

	results := &client.UpdateResult{
		DocIDs: make([]string, 0),
	}

	docMap := selectionPlan.DocumentMap()

	// Keep looping until results from the selection plan have been iterated through.
	for {
		next, nextErr := selectionPlan.Next()
		if nextErr != nil {
			return nil, err
		}
		// if theres no more records from the request, jump out of the loop
		if !next {
			break
		}

		// Get the document, and apply the patch
		docAsMap := docMap.ToMap(selectionPlan.Value())
		doc, err := client.NewDocFromMap(docAsMap, c.Schema())
		if err != nil {
			return nil, err
		}

		if isPatch {
			// todo
		} else if isMerge { // else is fine here
			err := doc.SetWithJSON([]byte(updater))
			if err != nil {
				return nil, err
			}
		}

		err = c.update(ctx, identity, txn, doc)
		if err != nil {
			return nil, err
		}

		// add successful updated doc to results
		results.DocIDs = append(results.DocIDs, doc.ID().String())
		results.Count++
	}

	return results, nil
}

// isSecondaryIDField returns true if the given field description represents a secondary relation field ID.
func (c *collection) isSecondaryIDField(fieldDesc client.FieldDefinition) (client.FieldDefinition, bool) {
	if fieldDesc.RelationName == "" || fieldDesc.Kind != client.FieldKind_DocID {
		return client.FieldDefinition{}, false
	}

	relationFieldDescription, valid := c.Definition().GetFieldByName(
		strings.TrimSuffix(fieldDesc.Name, request.RelatedObjectID),
	)
	return relationFieldDescription, valid && !relationFieldDescription.IsPrimaryRelation
}

// patchPrimaryDoc patches the (primary) document linked to from the document of the given DocID via the
// given (secondary) relationship field description (hosted on the collection of the document matching the
// given DocID).
//
// The given field value should be the string representation of the DocID of the primary document to be
// patched.
func (c *collection) patchPrimaryDoc(
	ctx context.Context,
	identity immutable.Option[string],
	txn datastore.Txn,
	secondaryCollectionName string,
	relationFieldDescription client.FieldDefinition,
	docID string,
	fieldValue string,
) error {
	primaryDocID, err := client.NewDocIDFromString(fieldValue)
	if err != nil {
		return err
	}

	primaryCol, err := c.db.getCollectionByName(ctx, txn, relationFieldDescription.Kind.Underlying())
	if err != nil {
		return err
	}
	primaryCol = primaryCol.WithTxn(txn)
	primarySchema := primaryCol.Schema()

	primaryField, ok := primaryCol.Description().GetFieldByRelation(
		relationFieldDescription.RelationName,
		secondaryCollectionName,
		relationFieldDescription.Name,
		&primarySchema,
	)
	if !ok {
		return client.NewErrFieldNotExist(relationFieldDescription.RelationName)
	}

	primaryIDField, ok := primaryCol.Definition().GetFieldByName(primaryField.Name + request.RelatedObjectID)
	if !ok {
		return client.NewErrFieldNotExist(primaryField.Name + request.RelatedObjectID)
	}

	doc, err := primaryCol.Get(
		ctx,
		identity,
		primaryDocID,
		false,
	)

	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}

	// If the document doesn't exist then there is nothing to update.
	if doc == nil {
		return nil
	}

	pc := c.db.newCollection(primaryCol.Description(), primarySchema)
	err = pc.validateOneToOneLinkDoesntAlreadyExist(
		ctx,
		identity,
		txn,
		primaryDocID.String(),
		primaryIDField,
		docID,
	)
	if err != nil {
		return err
	}

	existingVal, err := doc.GetValue(primaryIDField.Name)
	if err != nil && !errors.Is(err, client.ErrFieldNotExist) {
		return err
	}

	if existingVal != nil && existingVal.Value() != "" && existingVal.Value() != docID {
		return NewErrOneOneAlreadyLinked(docID, fieldValue, relationFieldDescription.RelationName)
	}

	err = doc.Set(primaryIDField.Name, docID)
	if err != nil {
		return err
	}

	err = primaryCol.Update(ctx, identity, doc)
	if err != nil {
		return err
	}

	return nil
}

// makeSelectionPlan constructs a simple read-only plan of the collection using the given filter.
// currently it doesn't support any other operations other than filters.
// (IE: No limit, order, etc)
// Additionally it only requests for the root scalar fields of the object
func (c *collection) makeSelectionPlan(
	ctx context.Context,
	identity immutable.Option[string],
	txn datastore.Txn,
	filter any,
) (planner.RequestPlan, error) {
	var f immutable.Option[request.Filter]
	var err error
	switch fval := filter.(type) {
	case string:
		if fval == "" {
			return nil, ErrInvalidFilter
		}

		f, err = c.db.parser.NewFilterFromString(c.Name().Value(), fval)
		if err != nil {
			return nil, err
		}
	case immutable.Option[request.Filter]:
		f = fval
	default:
		return nil, ErrInvalidFilter
	}

	slct, err := c.makeSelectLocal(f)
	if err != nil {
		return nil, err
	}

	planner := planner.New(
		ctx,
		identity,
		c.db.WithTxn(txn),
		txn,
	)

	return planner.MakePlan(&request.Request{
		Queries: []*request.OperationDefinition{
			{
				Selections: []request.Selection{
					slct,
				},
			},
		},
	})
}

func (c *collection) makeSelectLocal(filter immutable.Option[request.Filter]) (*request.Select, error) {
	slct := &request.Select{
		Field: request.Field{
			Name: c.Name().Value(),
		},
		Filterable: request.Filterable{
			Filter: filter,
		},
		ChildSelect: request.ChildSelect{
			Fields: make([]request.Selection, 0),
		},
	}

	for _, fd := range c.Schema().Fields {
		if fd.Kind.IsObject() {
			continue
		}
		slct.Fields = append(slct.Fields, &request.Field{
			Name: fd.Name,
		})
	}

	return slct, nil
}
