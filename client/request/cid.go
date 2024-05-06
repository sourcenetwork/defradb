// Copyright 2024 Democratized Data Foundation
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

// CIDFilter is an embeddable struct that hosts a consistent set of properties
// for filtering an aspect of a request by commit CID.
type CIDFilter struct {
	// CID is an optional value that selects a single document at the given commit CID
	// for processing by the request.
	//
	// If a commit matching the given CID is not found an error will be returned. The commit
	// does not need to be the latest, and this property allows viewing of the document at
	// prior revisions.
	CID immutable.Option[string]
}
