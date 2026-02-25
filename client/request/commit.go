// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

import (
	"slices"

	"github.com/sourcenetwork/immutable"
)

var (
	_ Selection = (*CommitSelect)(nil)
)

// CommitSelect represents the selection of database commits to Defra documents.
type CommitSelect struct {
	Field
	ChildSelect

	CIDFilter
	DocIDsFilter

	Limitable
	Offsetable
	Orderable
	Groupable
	Filterable

	// Depth limits the returned commits to being X places in the history away from the
	// most current.
	//
	// For example if a document has been updated 5 times, and a depth of 2 is provided
	// only commits for the last two updates will be returned.
	Depth immutable.Option[uint64]
}

func (c CommitSelect) ToSelect() *Select {
	return &Select{
		Field: Field{
			Name:  c.Name,
			Alias: c.Alias,
		},
		DocIDsFilter: c.DocIDsFilter,
		CIDFilter:    c.CIDFilter,
		Limitable:    c.Limitable,
		Offsetable:   c.Offsetable,
		Orderable:    c.Orderable,
		Groupable:    c.Groupable,
		Filterable:   c.Filterable,
		ChildSelect:  c.ChildSelect,
	}
}

// ToSubscriptionSelect implements the subscriptionSelector interface in internal/db/subscriptions.go
// We can safely ignore the first parameter (docID) for now
// since its always copied from the original subscription request
func (c CommitSelect) ToSubscriptionSelect(_, cid string) Selection {
	return &CommitSelect{
		Field: Field{
			Name:  c.Name,
			Alias: c.Alias,
		},
		DocIDsFilter: c.DocIDsFilter,
		CIDFilter: CIDFilter{
			immutable.Some([]string{cid}),
		},
		ChildSelect: c.ChildSelect,
	}
}

// CheckCIDFilter checks if the given cid passes the CID filter.
// Returns true if the cid passes the filter, false otherwise.
// If no CID filter is set, it always passes.
func (c CommitSelect) CheckCIDFilter(cid string) bool {
	return !c.CIDs.HasValue() || slices.Contains(c.CIDs.Value(), cid)
}

// CheckDocIDFilter checks if the given docID passes the DocID filter.
// Returns true if the docID passes the filter, false otherwise.
// If no DocID filter is set, it always passes.
func (c CommitSelect) CheckDocIDFilter(docID string) bool {
	return !c.DocIDs.HasValue() || slices.Contains(c.DocIDs.Value(), docID)
}
