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
// links to the previous and next version items if they exist.
//
// It also contains a vector which describes the distance and direction to the
// target schema version (given as an input param on construction).
type targetedSchemaHistoryLink struct {
	// The collection as this point in history.
	collection *client.CollectionDescription

	// The link to next schema version, if there is one
	// (for the most recent schema version this will be None).
	next immutable.Option[*targetedSchemaHistoryLink]

	// The link to the previous schema version, if there is
	// one (for the initial schema version this will be None).
	previous immutable.Option[*targetedSchemaHistoryLink]

	// The distance and direction from this history item to the target.
	//
	// A zero value indicates that this is the target item. A positive value
	// indicates that the target is more recent. A negative value indicates
	// that the target predates this history item.
	targetVector int
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

	result := map[schemaVersionID]*targetedSchemaHistoryLink{}

	for _, item := range history {
		result[item.collection.SchemaVersionID] = &targetedSchemaHistoryLink{
			collection: item.collection,
		}
	}

	for _, item := range result {
		schemaHistoryLink := history[item.collection.ID]
		nextHistoryItems := schemaHistoryLink.next
		if len(nextHistoryItems) == 0 {
			continue
		}

		// WARNING: This line assumes that each collection can only have a single source, and so
		// just takes the first item.  If/when collections can have multiple sources we will need to change
		// this slightly.
		nextItem := result[nextHistoryItems[0].collection.SchemaVersionID]
		item.next = immutable.Some(nextItem)
		nextItem.previous = immutable.Some(item)
	}

	orphanSchemaVersions := map[string]struct{}{}

	for schemaVersion, item := range result {
		if item.collection.SchemaVersionID == targetSchemaVersionID {
			continue
		}
		if item.targetVector != 0 {
			continue
		}

		distanceTravelled := 0
		currentItem := item
		wasFound := false
		for {
			if !currentItem.next.HasValue() {
				break
			}

			currentItem = currentItem.next.Value()
			distanceTravelled++
			if currentItem.targetVector != 0 {
				distanceTravelled += currentItem.targetVector
				wasFound = true
				break
			}
			if currentItem.collection.SchemaVersionID == targetSchemaVersionID {
				wasFound = true
				break
			}
		}

		if !wasFound {
			// The target was not found going up the chain, try looking back.
			// This is important for downgrading schema versions.
			for {
				if !currentItem.previous.HasValue() {
					break
				}

				currentItem = currentItem.previous.Value()
				distanceTravelled--
				if currentItem.targetVector != 0 {
					distanceTravelled += currentItem.targetVector
					wasFound = true
					break
				}
				if currentItem.collection.SchemaVersionID == targetSchemaVersionID {
					wasFound = true
					break
				}
			}
		}

		if !wasFound {
			// This may happen if users define schema migrations to unknown schema versions
			// with no migration path to known schema versions, esentially creating orphan
			// migrations. These may become linked later and should remain persisted in the
			// database, but we can drop them from the history here/now.
			orphanSchemaVersions[schemaVersion] = struct{}{}
			continue
		}

		item.targetVector = distanceTravelled
	}

	for schemaVersion := range orphanSchemaVersions {
		delete(result, schemaVersion)
	}

	return result, nil
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
) (map[collectionID]*schemaHistoryLink, error) {
	cols, err := description.GetCollectionsBySchemaRoot(ctx, txn, schemaRoot)
	if err != nil {
		return nil, err
	}

	history := map[collectionID]*schemaHistoryLink{}

	for _, c := range cols {
		col := c
		// Convert the temporary types to the cleaner return type:
		history[col.ID] = &schemaHistoryLink{
			collection: &col,
		}
	}

	for _, historyItem := range history {
		for _, source := range historyItem.collection.CollectionSources() {
			src := history[source.SourceCollectionID]
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
