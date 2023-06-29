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

	"github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

// historyItem represents an item in a particular schema's history, it
// links to the previous and next version items if they exist.
type historyItem struct {
	// The schema version id of this history item.
	schemaVersionID string

	// The history item for the next schema version, if there is one
	// (for the most recent schema version this will be None).
	next immutable.Option[*historyItem]

	// The history item for the previous schema version, if there is
	// one (for the initial schema version this will be None).
	previous immutable.Option[*historyItem]
}

// targetedHistoryItem represents an item in a particular schema's history, it
// links to the previous and next version items if they exist.
//
// It also contains a vector which describes the distance and direction to the
// target schema version (given as an input param on construction).
type targetedHistoryItem struct {
	// The schema version id of this history item.
	schemaVersionID string

	// The history item for the next schema version, if there is one
	// (for the most recent schema version this will be None).
	next immutable.Option[*targetedHistoryItem]

	// The history item for the previous schema version, if there is
	// one (for the initial schema version this will be None).
	previous immutable.Option[*targetedHistoryItem]

	// The distance and direction from this history item to the target.
	//
	// A zero value indicates that this is the target item. A positive value
	// indicates that the target is more recent. A negative value indicates
	// that the target predates this history item.
	targetVector int
}

// getTargetedHistory returns the history of the schema of the given id, relative
// to the given target schema version id.
//
// This includes any history items that are only known via registered
// schema migrations.
func getTargetedHistory(
	ctx context.Context,
	txn datastore.Txn,
	lensConfigs []client.LensConfig,
	schemaID string,
	targetSchemaVersionID string,
) (map[schemaVersionID]*targetedHistoryItem, error) {
	history, err := getHistory(ctx, txn, lensConfigs, schemaID)
	if err != nil {
		return nil, err
	}

	result := map[schemaVersionID]*targetedHistoryItem{}

	for _, item := range history {
		result[item.schemaVersionID] = &targetedHistoryItem{
			schemaVersionID: item.schemaVersionID,
		}
	}

	for _, item := range result {
		historyItem := history[item.schemaVersionID]
		nextHistoryItem := historyItem.next
		if !nextHistoryItem.HasValue() {
			continue
		}
		nextItem := result[nextHistoryItem.Value().schemaVersionID]
		item.next = immutable.Some(nextItem)
		nextItem.previous = immutable.Some(item)
	}

	orphanSchemaVersions := map[string]struct{}{}

	for schemaVersion, item := range result {
		if item.schemaVersionID == targetSchemaVersionID {
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
			if currentItem.schemaVersionID == targetSchemaVersionID {
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
				if currentItem.schemaVersionID == targetSchemaVersionID {
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

type historyPairing struct {
	schemaVersionID     string
	nextSchemaVersionID string
}

// getHistory returns the history of the schema of the given id.
//
// This includes any history items that are only known via registered
// schema migrations.
func getHistory(
	ctx context.Context,
	txn datastore.Txn,
	lensConfigs []client.LensConfig,
	schemaID string,
) (map[schemaVersionID]*historyItem, error) {
	pairings := map[string]*historyPairing{}

	for _, config := range lensConfigs {
		pairings[config.SourceSchemaVersionID] = &historyPairing{
			schemaVersionID:     config.SourceSchemaVersionID,
			nextSchemaVersionID: config.DestinationSchemaVersionID,
		}

		if _, ok := pairings[config.DestinationSchemaVersionID]; !ok {
			pairings[config.DestinationSchemaVersionID] = &historyPairing{
				schemaVersionID: config.DestinationSchemaVersionID,
			}
		}
	}

	prefix := core.NewSchemaHistoryKey(schemaID, "")
	q, err := txn.Systemstore().Query(ctx, query.Query{
		Prefix: prefix.ToString(),
	})
	if err != nil {
		return nil, err
	}

	for res := range q.Next() {
		// check for Done on context first
		select {
		case <-ctx.Done():
			// we've been cancelled! ;)
			return nil, q.Close()
		default:
			// noop, just continue on the with the for loop
		}

		if res.Error != nil {
			err = q.Close()
			if err != nil {
				return nil, err
			}
			return nil, res.Error
		}

		key, err := core.NewSchemaHistoryKeyFromString(res.Key)
		if err != nil {
			err = q.Close()
			if err != nil {
				return nil, err
			}
			return nil, err
		}

		// The local schema version history takes priority over and migration-defined history
		// and overwrites whatever already exists in the pairings (if any)
		pairings[key.PreviousSchemaVersionID] = &historyPairing{
			schemaVersionID:     key.PreviousSchemaVersionID,
			nextSchemaVersionID: string(res.Value),
		}

		if _, ok := pairings[string(res.Value)]; !ok {
			pairings[string(res.Value)] = &historyPairing{
				schemaVersionID: string(res.Value),
			}
		}
	}

	err = q.Close()
	if err != nil {
		return nil, err
	}

	history := map[schemaVersionID]*historyItem{}

	for _, pairing := range pairings {
		// Convert the temporary types to the cleaner return type:
		history[pairing.schemaVersionID] = &historyItem{
			schemaVersionID: pairing.schemaVersionID,
		}
	}

	for _, pairing := range pairings {
		src := history[pairing.schemaVersionID]

		// Use the internal pairings to set the next/previous links
		if next, hasNext := history[pairing.nextSchemaVersionID]; hasNext {
			src.next = immutable.Some(next)
			next.previous = immutable.Some(src)
		}
	}

	return history, nil
}
