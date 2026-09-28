// Copyright 2026 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client"
)

// retentionFloor returns the retention rule's height field and floor for col, or false if col is not
// judged. A branchable collection is never judged: one of its heads commits many documents, so one
// cannot be refused alone.
func (db *DB) retentionFloor(col *collection) (field string, floor int64, ok bool) {
	if db.retentionRule == nil || col.Version().IsBranchable {
		return "", 0, false
	}
	return db.retentionRule.RetentionFloor(col.CollectionID())
}

// retentionFloorOf is retentionFloor by CollectionID, for checks made before the merge. A collection
// that cannot be read is not judged.
func (db *DB) retentionFloorOf(ctx context.Context, collectionID string) (string, int64, bool) {
	if db.retentionRule == nil {
		return "", 0, false
	}
	col, err := getCollectionFromCollectionID(ctx, db, collectionID)
	if err != nil {
		return "", 0, false
	}
	return db.retentionFloor(col)
}

// checkRetention returns the refusal for doc as merged, or nil. A deleted document is not judged.
func (db *DB) checkRetention(col *collection, doc *client.Document) error {
	if doc == nil {
		return nil
	}
	field, floor, ok := db.retentionFloor(col)
	if !ok {
		return nil
	}
	height, valid := retentionHeight(doc, field)
	return client.CheckRetention(height, valid, floor)
}

// retentionHeight returns doc's height from field, and false when it is unset, null, not an integer
// or negative.
func retentionHeight(doc *client.Document, field string) (int64, bool) {
	value, err := doc.GetValue(field)
	if err != nil || value == nil {
		return 0, false
	}
	height, ok := value.Value().(int64)
	return height, ok && height >= 0
}
