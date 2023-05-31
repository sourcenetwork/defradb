// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

// IndexDirection is the direction of an index.
type IndexDirection string

const (
	Ascending  IndexDirection = "ASC"
	Descending IndexDirection = "DESC"
)

// IndexFieldDescription describes how a field is being indexed.
type IndexedFieldDescription struct {
	Name      string
	Direction IndexDirection
}

// IndexDescription describes an index.
type IndexDescription struct {
	// Name contains the name of the index.
	Name string
	// ID is the local identifier of this index.
	ID uint32
	// Fields contains the fields that are being indexed.
	Fields []IndexedFieldDescription
	// Unique indicates whether the index is unique.
	Unique bool
}

// CollectionIndexDescription describes an index on a collection.
// It's useful for retrieving a list of indexes without having to
// retrieve the entire collection description.
type CollectionIndexDescription struct {
	// CollectionName contains the name of the collection.
	CollectionName string
	// Index contains the index description.
	Index IndexDescription
}
