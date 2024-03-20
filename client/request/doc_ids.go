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

// DocIDsFilter is an embeddable struct that hosts a consistent set of properties
// for filtering an aspect of a request by document IDs.
type DocIDsFilter struct {
	// DocIDs is an optional value that ensures any records processed by the request
	// will have one of the given document IDs.
	DocIDs immutable.Option[[]string]
}
