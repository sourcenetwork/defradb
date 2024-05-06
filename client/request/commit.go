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

import "github.com/sourcenetwork/immutable"

var (
	_ Selection = (*CommitSelect)(nil)
)

// CommitSelect represents the selection of database commits to Defra documents.
type CommitSelect struct {
	Field
	ChildSelect

	CIDFilter

	Limitable
	Offsetable
	Orderable
	Groupable

	// DocID is an optional filter which when provided will limit commits to those
	// belonging to the given document.
	DocID immutable.Option[string]

	// FieldID is an optional filter which when provided will limit commits to those
	// belonging to the given field.
	//
	// `C` may be provided for document-level (composite) commits.
	FieldID immutable.Option[string]

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
		Limitable:   c.Limitable,
		Offsetable:  c.Offsetable,
		Orderable:   c.Orderable,
		Groupable:   c.Groupable,
		ChildSelect: c.ChildSelect,
	}
}
