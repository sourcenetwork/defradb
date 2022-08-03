// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mapper

// Requestable is the interface shared by all items that may be
// requested by consumers.
//
// For example, integer fields, document mutations, or aggregates.
type Requestable interface {
	// GetIndex returns the index at which this item can be found upon
	// its parent.
	GetIndex() int

	// GetName returns the name of this item.  For example 'Age', or
	// '_count'.
	GetName() string

	// AsTargetable tries to return the targetable component of this
	// item. If the item-type does not support targeting, it will
	// return nil and false, otherwise it will return a pointer to
	// the targetable component and true.
	AsTargetable() (*Targetable, bool)

	// AsSelect tries to return the select component of this
	// item. If the item-type does not support selection, it will
	// return nil and false, otherwise it will return a pointer to
	// the select component and true.
	AsSelect() (*Select, bool)

	// CloneTo deep clones this item using the provided index instead
	// of the index of this item.
	CloneTo(index int) Requestable
}

var (
	_ Requestable = (*Aggregate)(nil)
	_ Requestable = (*CommitSelect)(nil)
	_ Requestable = (*Field)(nil)
	_ Requestable = (*Mutation)(nil)
	_ Requestable = (*Select)(nil)
)
