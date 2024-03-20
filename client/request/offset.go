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

// Offsetable is an embeddable struct that hosts a consistent set of properties
// for offsetting an aspect of a request.
type Offsetable struct {
	// Offset is an optional value that skips the given number of results that would have
	// otherwise been returned.  Commonly used alongside the limit argument,
	// this argument will still work on its own.
	Offset immutable.Option[uint64]
}
