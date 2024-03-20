// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package lens

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/description"
)

// schemaHistoryLink represents an item in a particular schema's history, it
// links to the previous and next version items if they exist.
type schemaHistoryLink struct {
	// The collection as this point in history.
	collection *client.CollectionDescription

	// The history link to the next schema versions, if there are some
	// (for the most recent schema version this will be empty).
	next []*schemaHistoryLink

	// The history link to the previous schema versions, if there are
	// some (for the initial schema version this will be empty).
	previous []*schemaHistoryLink
}

// targetedSchemaHistoryLink represents an item in a particular schema's history, it
// links to the previous and next version items if they exist and are on the path to
// the target schema version.
type targetedSchemaHistoryLink struct {
	// The collection as this point in history.
	collection *client.CollectionDescription

	// The link to next schema version, if there is one
	// (for the most recent schema version this will be None).
	next immutable.Option[*targetedSchemaHistoryLink]

	// The link to the previous schema version, if there is
	// one (for the initial schema version this will be None).
	previous immutable.Option[*targetedSchemaHistoryLink]
}

// getTargetedSchemaHistory returns the history of the schema of the given id, relative
// to the given target schema version id.
//
// This includes any history items that are only known via registered
// schema migrations.
func getTargetedSchemaHistory(
	ctx context.Context,
	txn datastore.Txn,
	schemaRoot string,
	targetSchemaVersionID string,
) (map[schemaVersionID]*targetedSchemaHistoryLink, error) {
	history, err := getSchemaHistory(ctx, txn, schemaRoot)
	if err != nil {
		return nil, err
	}

	targetHistoryItem, ok := history[targetSchemaVersionID]
	if !ok {
		// If the target schema version is unknown then there are no possible migrations
		// that we can do.
		return nil, nil
	}

	result := map[schemaVersionID]*targetedSchemaHistoryLink{}

	targetLink := &targetedSchemaHistoryLink{
		collection: targetHistoryItem.collection,
	}
	result[targetLink.collection.SchemaVersionID] = targetLink

	linkForwards(targetLink, targetHistoryItem, result)
	linkBackwards(targetLink, targetHistoryItem, result)

	return result, nil
}

// linkForwards traverses and links the history forwards from the given starting point.
//
// Forward schema versions found will in turn be linked both forwards and backwards, allowing
// schema branches to be correctly mapped to the target schema version.
func linkForwards(
	currentLink *targetedSchemaHistoryLink,
	currentHistoryItem *schemaHistoryLink,
	result map[schemaVersionID]*targetedSchemaHistoryLink,
) {
	for _, nextHistoryItem := range currentHistoryItem.next {
		if _, ok := result[nextHistoryItem.collection.SchemaVersionID]; ok {
			// As the history forms a DAG, this should only ever happen when
			// iterating through the item we were at immediately before the current.
			continue
		}

		nextLink := &targetedSchemaHistoryLink{
			collection: nextHistoryItem.collection,
			previous:   immutable.Some(currentLink),
		}
		result[nextLink.collection.SchemaVersionID] = nextLink

		linkForwards(nextLink, nextHistoryItem, result)
		linkBackwards(nextLink, nextHistoryItem, result)
	}
}

// linkBackwards traverses and links the history backwards from the given starting point.
//
// Forward schema versions found will in turn be linked both forwards and backwards, allowing
// schema branches to be correctly mapped to the target schema version.
func linkBackwards(
	currentLink *targetedSchemaHistoryLink,
	currentHistoryItem *schemaHistoryLink,
	result map[schemaVersionID]*targetedSchemaHistoryLink,
) {
	for _, prevHistoryItem := range currentHistoryItem.previous {
		if _, ok := result[prevHistoryItem.collection.SchemaVersionID]; ok {
			// As the history forms a DAG, this should only ever happen when
			// iterating through the item we were at immediately before the current.
			continue
		}

		prevLink := &targetedSchemaHistoryLink{
			collection: prevHistoryItem.collection,
			next:       immutable.Some(currentLink),
		}
		result[prevLink.collection.SchemaVersionID] = prevLink

		linkForwards(prevLink, prevHistoryItem, result)
		linkBackwards(prevLink, prevHistoryItem, result)
	}
}

// getSchemaHistory returns the history of the schema of the given id as linked list
// with each item mapped by schema version id.
//
// This includes any history items that are only known via registered
// schema migrations.
func getSchemaHistory(
	ctx context.Context,
	txn datastore.Txn,
	schemaRoot string,
) (map[schemaVersionID]*schemaHistoryLink, error) {
	cols, err := description.GetCollectionsBySchemaRoot(ctx, txn, schemaRoot)
	if err != nil {
		return nil, err
	}

	history := map[schemaVersionID]*schemaHistoryLink{}
	schemaVersionsByColID := map[uint32]schemaVersionID{}

	for _, c := range cols {
		// Todo - this `col := c` can be removed with Go 1.22:
		// https://github.com/sourcenetwork/defradb/issues/2431
		col := c

		// Convert the temporary types to the cleaner return type:
		history[col.SchemaVersionID] = &schemaHistoryLink{
			collection: &col,
		}
		schemaVersionsByColID[col.ID] = col.SchemaVersionID
	}

	for _, historyItem := range history {
		for _, source := range historyItem.collection.CollectionSources() {
			srcSchemaVersion := schemaVersionsByColID[source.SourceCollectionID]
			src := history[srcSchemaVersion]
			historyItem.previous = append(
				historyItem.next,
				src,
			)

			src.next = append(
				src.next,
				historyItem,
			)
		}
	}

	return history, nil
}
