// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package description

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

// collectionHistoryLink represents an item in a particular collection's version history, it
// links to the previous and next version items if they exist.
type collectionHistoryLink struct {
	// The collection as this point in history.
	collection *client.CollectionVersion

	// The history link to the next collection versions, if there are some
	// (for the most recent collection version this will be empty).
	next []*collectionHistoryLink

	// The history link to the previous collection versions, if there are
	// some (for the initial collection version this will be empty).
	previous []*collectionHistoryLink
}

// TargetedCollectionHistoryLink represents an item in a particular collection's version history, it
// links to the previous and next version items if they exist and are on the path to
// the target collection version.
type TargetedCollectionHistoryLink struct {
	// The collection as this point in history.
	collection *client.CollectionVersion

	// The link to next collection version, if there is one
	// (for the most recent collection version this will be None).
	next immutable.Option[*TargetedCollectionHistoryLink]

	// The link to the previous collection version, if there is
	// one (for the initial collection version this will be None).
	previous immutable.Option[*TargetedCollectionHistoryLink]
}

// Collection returns the collection version at this point in history.
func (t *TargetedCollectionHistoryLink) Collection() *client.CollectionVersion {
	return t.collection
}

// Next returns the link to the next collection version.
func (t *TargetedCollectionHistoryLink) Next() immutable.Option[*TargetedCollectionHistoryLink] {
	return t.next
}

// Previous returns the link to the previous collection version.
func (t *TargetedCollectionHistoryLink) Previous() immutable.Option[*TargetedCollectionHistoryLink] {
	return t.previous
}

// HasMigrations checks if there are any migrations registered for the given collection version
// by examining the full version history DAG.
//
// This properly handles branching version histories by checking if any version
// reachable from the given version has a migration transform.
func HasMigrations(
	ctx context.Context,
	collectionID string,
	versionID string,
) (bool, error) {
	history, err := GetTargetedCollectionHistory(ctx, collectionID, versionID)
	if err != nil {
		return false, err
	}

	if history == nil {
		return false, nil
	}

	for _, historyLink := range history {
		if historyLink.Collection().PreviousVersion.HasValue() {
			prevVersion := historyLink.Collection().PreviousVersion.Value()
			if prevVersion.Transform.HasValue() {
				return true, nil
			}
		}
	}

	return false, nil
}

// GetTargetedCollectionHistory returns the history of the collection of the given id, relative
// to the given target collection version id.
//
// This includes any history items that are only known via registered collection migrations.
func GetTargetedCollectionHistory(
	ctx context.Context,
	collectionRoot string,
	targetCollectionVersionID string,
) (map[string]*TargetedCollectionHistoryLink, error) {
	history, err := getCollectionHistory(ctx, collectionRoot)
	if err != nil {
		return nil, err
	}

	targetHistoryItem, ok := history[targetCollectionVersionID]
	if !ok {
		// If the target collection version is unknown then there are no possible migrations
		// that we can do.
		return nil, nil
	}

	result := map[string]*TargetedCollectionHistoryLink{}

	targetLink := &TargetedCollectionHistoryLink{
		collection: targetHistoryItem.collection,
	}
	result[targetLink.collection.VersionID] = targetLink

	linkForwards(targetLink, targetHistoryItem, result)
	linkBackwards(targetLink, targetHistoryItem, result)

	return result, nil
}

// linkForwards traverses and links the history forwards from the given starting point.
//
// Forward collection versions found will in turn be linked both forwards and backwards, allowing
// branches to be correctly mapped to the target collection version.
func linkForwards(
	currentLink *TargetedCollectionHistoryLink,
	currentHistoryItem *collectionHistoryLink,
	result map[string]*TargetedCollectionHistoryLink,
) {
	for _, nextHistoryItem := range currentHistoryItem.next {
		if _, ok := result[nextHistoryItem.collection.VersionID]; ok {
			// As the history forms a DAG, this should only ever happen when
			// iterating through the item we were at immediately before the current.
			continue
		}

		nextLink := &TargetedCollectionHistoryLink{
			collection: nextHistoryItem.collection,
			previous:   immutable.Some(currentLink),
		}
		result[nextLink.collection.VersionID] = nextLink

		linkForwards(nextLink, nextHistoryItem, result)
		linkBackwards(nextLink, nextHistoryItem, result)
	}
}

// linkBackwards traverses and links the history backwards from the given starting point.
//
// Backward collection versions found will in turn be linked both forwards and backwards, allowing
// branches to be correctly mapped to the target collection version.
func linkBackwards(
	currentLink *TargetedCollectionHistoryLink,
	currentHistoryItem *collectionHistoryLink,
	result map[string]*TargetedCollectionHistoryLink,
) {
	for _, prevHistoryItem := range currentHistoryItem.previous {
		if _, ok := result[prevHistoryItem.collection.VersionID]; ok {
			// As the history forms a DAG, this should only ever happen when
			// iterating through the item we were at immediately before the current.
			continue
		}

		prevLink := &TargetedCollectionHistoryLink{
			collection: prevHistoryItem.collection,
			next:       immutable.Some(currentLink),
		}
		result[prevLink.collection.VersionID] = prevLink

		linkForwards(prevLink, prevHistoryItem, result)
		linkBackwards(prevLink, prevHistoryItem, result)
	}
}

// getCollectionHistory returns the history of the collection of the given root id as linked list
// with each item mapped by collection version id.
//
// This includes any history items that are only known via registered collection migrations.
func getCollectionHistory(
	ctx context.Context,
	collectionRoot string,
) (map[string]*collectionHistoryLink, error) {
	cols, err := GetCollectionsByCollectionID(ctx, collectionRoot)
	if err != nil {
		return nil, err
	}

	history := map[string]*collectionHistoryLink{}

	for _, col := range cols {
		// Convert the temporary types to the cleaner return type:
		history[col.VersionID] = &collectionHistoryLink{
			collection: &col,
		}
	}

	for _, historyItem := range history {
		if historyItem.collection.PreviousVersion.HasValue() {
			src := history[historyItem.collection.PreviousVersion.Value().SourceCollectionID]
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
