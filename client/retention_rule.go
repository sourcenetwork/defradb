// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

// RetentionRule refuses replicated documents by block height. A document in a covered collection is
// refused when its height is at or below the floor, or when it has no valid height. Branchable
// collections are not judged.
type RetentionRule interface {
	// RetentionFloor returns the height field and floor for the collection with this CollectionID,
	// or false if it is not covered. A floor of zero or less refuses no height. It is called for
	// every replicated document and must not block.
	RetentionFloor(collectionID string) (field string, floor int64, ok bool)
}

// CheckRetention returns the refusal for a document at height, or nil if it is kept. valid is false
// when the document has no valid height.
func CheckRetention(height int64, valid bool, floor int64) error {
	if !valid {
		return ErrNoRetentionHeight
	}
	if floor > 0 && height <= floor {
		return ErrBelowRetentionFloor
	}
	return nil
}
