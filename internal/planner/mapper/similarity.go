// Copyright 2025 Democratized Data Foundation
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

// Similarity represents an cosine similarity operation definition.
type Similarity struct {

	// The vector to compare the target field to.
	Vector any
	// The mapping of this aggregate's parent/host.
	*core.DocumentMapping

	Field

	// The targetted field for the cosine similarity
	SimilarityTarget Targetable
}
