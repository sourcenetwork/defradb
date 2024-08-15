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
	"github.com/sourcenetwork/defradb/internal/db/description"
)

// collectionHistoryLink represents an item in a particular collection's schema history, it
// links to the previous and next version items if they exist.
type collectionHistoryLink struct {
	// The collection as this point in history.
	collection *client.CollectionDescription

	// The history link to the next collection versions, if there are some
	// (for the most recent schema version this will be empty).
	next []*collectionHistoryLink

	// The history link to the previous collection versions, if there are
	// some (for the initial collection version this will be empty).
	previous []*collectionHistoryLink
}

// targetedCollectionHistoryLink represents an item in a particular collection's schema history, it
// links to the previous and next version items if they exist and are on the path to
// the target schema version.
type targetedCollectionHistoryLink struct {
	// The collection as this point in history.
	collection *client.CollectionDescription

	// The link to next collection version, if there is one
	// (for the most recent collection version this will be None).
	next immutable.Option[*targetedCollectionHistoryLink]

	// The link to the previous collection version, if there is
	// one (for the initial collection version this will be None).
	previous immutable.Option[*targetedCollectionHistoryLink]
}

// getTargetedCollectionHistory returns the history of the schema of the given id, relative
// to the given target schema version id.
//
// This includes any history items that are only known via registered schema migrations.
func getTargetedCollectionHistory(
	ctx context.Context,
	txn datastore.Txn,
	schemaRoot string,
	targetSchemaVersionID string,
) (map[schemaVersionID]*targetedCollectionHistoryLink, error) {
	history, err := getCollectionHistory(ctx, txn, schemaRoot)
	if err != nil {
		return nil, err
	}

	targetHistoryItem, ok := history[targetSchemaVersionID]
	if !ok {
		// If the target schema version is unknown then there are no possible migrations
		// that we can do.
		return nil, nil
	}

	result := map[schemaVersionID]*targetedCollectionHistoryLink{}

	targetLink := &targetedCollectionHistoryLink{
		collection: targetHistoryItem.collection,
	}
	result[targetLink.collection.SchemaVersionID] = targetLink

	linkForwards(targetLink, targetHistoryItem, result)
	linkBackwards(targetLink, targetHistoryItem, result)

	return result, nil
}

// linkForwards traverses and links the history forwards from the given starting point.
//
// Forward collection versions found will in turn be linked both forwards and backwards, allowing
// branches to be correctly mapped to the target schema version.
func linkForwards(
	currentLink *targetedCollectionHistoryLink,
	currentHistoryItem *collectionHistoryLink,
	result map[schemaVersionID]*targetedCollectionHistoryLink,
) {
	for _, nextHistoryItem := range currentHistoryItem.next {
		if _, ok := result[nextHistoryItem.collection.SchemaVersionID]; ok {
			// As the history forms a DAG, this should only ever happen when
			// iterating through the item we were at immediately before the current.
			continue
		}

		nextLink := &targetedCollectionHistoryLink{
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
// Backward collection versions found will in turn be linked both forwards and backwards, allowing
// branches to be correctly mapped to the target schema version.
func linkBackwards(
	currentLink *targetedCollectionHistoryLink,
	currentHistoryItem *collectionHistoryLink,
	result map[schemaVersionID]*targetedCollectionHistoryLink,
) {
	for _, prevHistoryItem := range currentHistoryItem.previous {
		if _, ok := result[prevHistoryItem.collection.SchemaVersionID]; ok {
			// As the history forms a DAG, this should only ever happen when
			// iterating through the item we were at immediately before the current.
			continue
		}

		prevLink := &targetedCollectionHistoryLink{
			collection: prevHistoryItem.collection,
			next:       immutable.Some(currentLink),
		}
		result[prevLink.collection.SchemaVersionID] = prevLink

		linkForwards(prevLink, prevHistoryItem, result)
		linkBackwards(prevLink, prevHistoryItem, result)
	}
}

// getCollectionHistory returns the history of the collection of the given root id as linked list
// with each item mapped by schema version id.
//
// This includes any history items that are only known via registered schema migrations.
func getCollectionHistory(
	ctx context.Context,
	txn datastore.Txn,
	schemaRoot string,
) (map[schemaVersionID]*collectionHistoryLink, error) {
	cols, err := description.GetCollectionsBySchemaRoot(ctx, txn, schemaRoot)
	if err != nil {
		return nil, err
	}

	history := map[schemaVersionID]*collectionHistoryLink{}
	schemaVersionsByColID := map[uint32]schemaVersionID{}

	for _, col := range cols {
		// Convert the temporary types to the cleaner return type:
		history[col.SchemaVersionID] = &collectionHistoryLink{
			collection: &col,
		}
		schemaVersionsByColID[col.ID] = col.SchemaVersionID
	}

	for _, historyItem := range history {
		for _, source := range historyItem.collection.CollectionSources() {
			srcSchemaVersion := schemaVersionsByColID[source.SourceCollectionID]
			src := history[srcSchemaVersion]
			historyItem.previous = append(
				historyItem.previous,
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
