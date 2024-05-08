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

import "github.com/sourcenetwork/immutable"

// CommitSelect represents a commit request from a consumer.
//
// E.g. commits, or latestCommits.
type CommitSelect struct {
	// The underlying Select, defining the information requested.
	Select

	// The key of the target document for which to get commits for.
	DocID immutable.Option[string]

	// The field for which commits have been requested.
	FieldID immutable.Option[string]

	// The maximum depth to yield results for.
	Depth immutable.Option[uint64]

	// The parent Cid for which commit information has been requested.
	Cid immutable.Option[string]

	// The SchemaVersionID at the time of commit.
	SchemaVersionID immutable.Option[string]
}

func (s *CommitSelect) CloneTo(index int) Requestable {
	return s.cloneTo(index)
}

func (s *CommitSelect) cloneTo(index int) *CommitSelect {
	return &CommitSelect{
		Select:  *s.Select.cloneTo(index),
		DocID:   s.DocID,
		FieldID: s.FieldID,
		Cid:     s.Cid,
	}
}
