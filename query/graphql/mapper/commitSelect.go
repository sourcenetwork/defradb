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

import "github.com/sourcenetwork/defradb/client"

// CommitType represents a type of [CommitSelect]
type CommitType int

const (
	NoneCommitType = CommitType(iota)
	LatestCommits
	AllCommits
)

// CommitSelect represents a commit request from a consumer.
//
// E.g. allCommits, or latestCommits.
type CommitSelect struct {
	// The underlying Select, defining the information requested.
	Select

	// The type of commit select request.
	Type CommitType

	// The key of the target document for which to get commits for.
	DocKey string

	// The field for which commits have been requested.
	FieldName client.Option[string]

	// The parent Cid for which commit information has been requested.
	Cid string
}

func (s *CommitSelect) CloneTo(index int) Requestable {
	return s.cloneTo(index)
}

func (s *CommitSelect) cloneTo(index int) *CommitSelect {
	return &CommitSelect{
		Select:    *s.Select.cloneTo(index),
		DocKey:    s.DocKey,
		Type:      s.Type,
		FieldName: s.FieldName,
		Cid:       s.Cid,
	}
}
