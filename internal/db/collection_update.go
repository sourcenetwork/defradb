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

	"github.com/sourcenetwork/immutable"
	"github.com/valyala/fastjson"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/planner"
)

// UpdateWithFilter updates using a filter to target documents for update.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	res, err := c.updateWithFilter(ctx, filter, updater)
	if err != nil {
		return nil, err
	}
	return res, txn.Commit(ctx)
}

func (c *collection) updateWithFilter(
	ctx context.Context,
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
	selectionPlan, err := c.makeSelectionPlan(ctx, filter)
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
		doc, err := client.NewDocFromMap(docAsMap, c.Definition())
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

		err = c.update(ctx, doc)
		if err != nil {
			return nil, err
		}

		// add successful updated doc to results
		results.DocIDs = append(results.DocIDs, doc.ID().String())
		results.Count++
	}

	return results, nil
}

// makeSelectionPlan constructs a simple read-only plan of the collection using the given filter.
// currently it doesn't support any other operations other than filters.
// (IE: No limit, order, etc)
// Additionally it only requests for the root scalar fields of the object
func (c *collection) makeSelectionPlan(
	ctx context.Context,
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

	txn := mustGetContextTxn(ctx)
	planner := planner.New(
		ctx,
		identity.FromContext(ctx),
		c.db.acp,
		c.db,
		txn,
	)

	return planner.MakeSelectionPlan(slct)
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
