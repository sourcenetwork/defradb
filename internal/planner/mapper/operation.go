// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mapper

import "github.com/sourcenetwork/defradb/internal/core"

// Operation represents an operation such as query or mutation.
//
// It wraps child Selects belonging to this operation.
type Operation struct {
	// The document mapping for this select, describing how items yielded
	// for this select can be accessed and rendered.
	*core.DocumentMapping

	// Selects is the list of selections in the operation.
	Selects []*Select

	// Mutations is the list of mutations in the operation.
	Mutations []*Mutation

	// CommitSelects is the list of commit selections in the operation.
	CommitSelects []*CommitSelect
}
