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

// Limitable is an embeddable struct that hosts a consistent set of properties
// for limiting an aspect of a request.
type Limitable struct {
	// Limit is an optional value that caps the number of results to the number provided.
	Limit immutable.Option[uint64]
}
